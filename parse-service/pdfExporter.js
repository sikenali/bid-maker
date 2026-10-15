'use strict';

/** PDF 导出引擎:HTML -> headless Chrome(CDP printToPDF) -> PDF 二进制.
 *
 * 依赖: puppeteer-core + 系统 Chrome(无需下载 Chromium).
 * Chrome 路径: CHROME_PATH 环境变量优先,否则按常见路径自动探测.
 * 容器内须带 --no-sandbox(已默认).
 *
 * 设置项映射(全部真实生效):
 * - vectorHD: 打印背景图形开/关(文字恒为矢量)
 * - bookmarks: CDP generateDocumentOutline 开/关
 * - watermark: 每页固定水印层开/关
 * 页码统一走 footerTemplate(第 X 页/共 Y 页),与开关无关.
 */

// puppeteer-core v25 起为纯 ESM,而本服务是 CJS:
// 不用顶层 require,改为首次导出 PDF 时动态 import(失败不影响其他功能)。
let puppeteer = null;
let puppeteerLoader = null;
function loadPuppeteer() {
  if (!puppeteerLoader) {
    puppeteerLoader = import('puppeteer-core').then((mod) => {
      puppeteer = mod.default || mod;
      return puppeteer;
    });
  }
  return puppeteerLoader;
}
const fs = require('fs');

/** 浏览器候选路径,按可用性排序:Chrome > Edge > 其他 Chromium 内核。
 * 桌面客户端不随包分发 Chrome,依赖用户机器已装任一内核。
 */
function chromeCandidates() {
  const winEnv = (name) => process.env[name];
  const pf = winEnv('ProgramFiles');
  const pf86 = winEnv('ProgramFiles(x86)');
  const local = winEnv('LOCALAPPDATA');

  const explicit = [
    '/usr/bin/google-chrome',
    '/usr/bin/google-chrome-stable',
    '/usr/bin/chromium',
    '/usr/bin/chromium-browser',
    '/usr/bin/microsoft-edge',
    '/usr/bin/microsoft-edge-stable',
    '/opt/google/chrome/chrome',
    '/opt/google/chrome/google-chrome',
    '/opt/microsoft/msedge/msedge',
    '/opt/chromium/chromium',
    '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    '/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary',
    '/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge',
    '/Applications/Brave Browser.app/Contents/MacOS/Brave Browser',
    '/Applications/Chromium.app/Contents/MacOS/Chromium',
  ];

  const win = [
    'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe',
    'C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe',
    'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe',
    'C:\\Program Files\\Microsoft\\Edge\\Application\\msedge.exe',
    'C:\\Program Files\\BraveSoftware\\Brave-Browser\\Application\\brave.exe',
    ...(pf ? [`${pf}\\Google\\Chrome\\Application\\chrome.exe`, `${pf}\\Microsoft\\Edge\\Application\\msedge.exe`] : []),
    ...(pf86 ? [`${pf86}\\Google\\Chrome\\Application\\chrome.exe`, `${pf86}\\Microsoft\\Edge\\Application\\msedge.exe`] : []),
    ...(local ? [`${local}\\Google\\Chrome\\Application\\chrome.exe`, `${local}\\Microsoft\\Edge\\Application\\msedge.exe`] : []),
  ];

  // puppeteer 的 chrome-headless-shell 缓存(若用户手动 @puppeteer/browsers install 过)
  const cacheDir = process.env.PUPPETEER_CACHE_DIR ||
    (process.platform === 'win32' && process.env.LOCALAPPDATA
      ? `${process.env.LOCALAPPDATA}\\puppeteer`
      : `${(process.env.HOME || process.env.USERPROFILE || '')}/.cache/puppeteer`);
  const headlessShell = chromeHeadlessShellCache(cacheDir);

  return [process.env.CHROME_PATH, ...headlessShell, ...(process.platform === 'win32' ? win : explicit)].filter(Boolean);
}

// 在 puppeteer 缓存目录里找 chrome-headless-shell / chrome-* 可执行文件
function chromeHeadlessShellCache(dir) {
  const out = [];
  let entries = [];
  try { entries = fs.readdirSync(dir); } catch { return out; }
  for (const name of entries) {
    const base = `${dir}/${name}`;
    // chrome-headless-shell-xxx/chrome-headless-shell-linux64/chrome-headless-shell
    // chrome-xxx/chrome-linux/chrome
    try {
      for (const rel of [
        `chrome-headless-shell-linux64/chrome-headless-shell`,
        `chrome-headless-shell-linux-arm64/chrome-headless-shell`,
        `chrome-headless-shell-win64/chrome-headless-shell.exe`,
        `chrome-headless-shell-win-arm64/chrome-headless-shell.exe`,
        `chrome-headless-shell-mac-x64/chrome-headless-shell`,
        `chrome-headless-shell-mac-arm64/chrome-headless-shell`,
        `chrome-linux/chrome`,
        `chrome-linux64/chrome`,
        `chrome-win64/chrome.exe`,
        `chrome-win-arm64/chrome.exe`,
        `chrome-mac-x64/Google Chrome.app/Contents/MacOS/Google Chrome`,
        `chrome-mac-arm64/Google Chrome.app/Contents/MacOS/Google Chrome`,
      ]) {
        out.push(`${base}/${rel}`);
      }
    } catch { /* ignore */ }
  }
  return out;
}

let cachedChrome = null;
let chromeResolved = false;

function resolveChrome() {
  if (chromeResolved) return cachedChrome;
  for (const p of chromeCandidates()) {
    try { fs.accessSync(p, fs.constants.X_OK); cachedChrome = p; chromeResolved = true; return p; } catch { /* next */ }
  }
  cachedChrome = null;
  chromeResolved = true;
  return null;
}

/** 浏览器可用性快照,供 /health 上报(缓存后 O(1))。 */
function getChromeStatus() {
  const path = resolveChrome();
  return path ? { available: true, path } : { available: false, hint: '未找到 Chrome/Edge,PDF 导出不可用;可设置 CHROME_PATH 指定可执行文件' };
}

async function getBrowser() {
  if (!browserPromise) {
    browserPromise = (async () => {
      const exe = resolveChrome();
      if (!exe) {
        const err = new Error('未找到可用的 Chrome/Edge,PDF 导出不可用;请安装浏览器或设置 CHROME_PATH 环境变量');
        err.code = 'CHROME_NOT_FOUND';
        err.tried = chromeCandidates();
        throw err;
      }
      const pe = await loadPuppeteer();
      return pe.launch({
        executablePath: exe,
        args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-gpu', '--disable-dev-shm-usage'],
      });
    })();
    browserPromise.catch(() => { browserPromise = null; });
  }
  return browserPromise;
}

// ── HTML 转义 ────────────────────────────────────────────────────────
function esc(s) {
  return String(s == null ? '' : s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function inlineMd(s) {
  return esc(s).replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
}

// ── 正文 md-lite 块渲染 ──────────────────────────────────────────────
// 支持: ```代码块 / > 引用 / - 无序 / 1. 有序 / | 表格 / # 残留标题 / 段落
function renderBody(content) {
  const lines = String(content || '').split('\n');
  let html = '';
  let inCode = false;
  let codeBuf = [];
  let listTag = null; // 'ul' | 'ol' | null

  const closeList = () => {
    if (listTag) { html += `</${listTag}>`; listTag = null; }
  };
  const pushPara = (t) => {
    if (t.trim()) html += `<p>${inlineMd(t.trim())}</p>`;
  };

  let pendingTable = [];
  const flushTable = () => {
    if (!pendingTable.length) return;
    // 分隔行(|---|---|)与全空行丢弃:逐格判定,避免数组转字符串引入逗号误判
    const rows = pendingTable.filter((r) => !r.cells.every((c) => /^[\s:~-]*$/.test(c)));
    if (rows.length) {
      html += '<table><thead><tr>' + rows[0].cells.map((c) => `<th>${inlineMd(c)}</th>`).join('') + '</tr></thead>';
      if (rows.length > 1) {
        html += '<tbody>' + rows.slice(1).map((r) => '<tr>' + r.cells.map((c) => `<td>${inlineMd(c)}</td>`).join('') + '</tr>').join('') + '</tbody>';
      }
      html += '</table>';
    }
    pendingTable = [];
  };
  const isTableRow = (t) => /^\|.*\|$/.test(t.trim());
  const splitRow = (t) => t.trim().replace(/^\||\|$/g, '').split('|').map((c) => c.trim());

  for (const raw of lines) {
    const line = raw.replace(/\r$/, '');
    const t = line.trim();
    if (t.startsWith('```')) {
      if (inCode) {
        html += `<pre>${esc(codeBuf.join('\n'))}</pre>`;
        codeBuf = [];
        inCode = false;
      } else {
        closeList(); flushTable();
        inCode = true;
      }
      continue;
    }
    if (inCode) { codeBuf.push(line); continue; }
    if (isTableRow(t)) {
      closeList();
      pendingTable.push({ cells: splitRow(t) });
      continue;
    }
    flushTable();
    if (!t) { closeList(); continue; }
    const hm = /^(#{1,6})\s+(.+)/.exec(t);
    if (hm) {
      closeList();
      const lv = Math.min(hm[1].length + 3, 6);
      html += `<h${lv}>${inlineMd(hm[2])}</h${lv}>`;
      continue;
    }
    if (/^>\s?/.test(t)) {
      closeList();
      html += `<blockquote>${inlineMd(t.replace(/^>\s?/, ''))}</blockquote>`;
      continue;
    }
    const ulm = /^[-*+]\s+(.+)/.exec(t);
    if (ulm) {
      if (listTag !== 'ul') { closeList(); html += '<ul>'; listTag = 'ul'; }
      html += `<li>${inlineMd(ulm[1])}</li>`;
      continue;
    }
    const olm = /^\d+[.)]\s+(.+)/.exec(t);
    if (olm) {
      if (listTag !== 'ol') { closeList(); html += '<ol>'; listTag = 'ol'; }
      html += `<li>${inlineMd(olm[1])}</li>`;
      continue;
    }
    closeList();
    pushPara(t);
  }
  if (inCode && codeBuf.length) html += `<pre>${esc(codeBuf.join('\n'))}</pre>`;
  flushTable();
  closeList();
  return html;
}

// ── 整文档 HTML ──────────────────────────────────────────────────────
function buildHtml(data, options) {
  const opts = { vectorHD: true, bookmarks: true, watermark: false, watermarkText: '内部稿', ...(options || {}) };
  const title = String(data.title || '投标文件');
  const bidNo = String(data.bid_no || '');
  const exportTime = data.export_time || new Date().toISOString().slice(0, 10);
  const sections = Array.isArray(data.sections) ? data.sections : [];

  let body = `<div class="cover"><h1 class="cover-title">${esc(title)}</h1>`;
  if (bidNo) body += `<div class="cover-meta">招标编号：${esc(bidNo)}</div>`;
  body += `<div class="cover-meta">导出时间：${esc(exportTime)}</div></div>`;

  for (const s of sections) {
    const lv = Math.min(Math.max(s.level || 1, 1), 3);
    const st = String(s.title || '');
    if (st) body += `<h${lv}>${esc(st)}</h${lv}>`;
    body += renderBody(s.content);
  }

  const wm = opts.watermark
    ? `<div class="watermark">${esc(opts.watermarkText || '内部稿')}</div>`
    : '';

  return `<!DOCTYPE html><html lang="zh-CN"><head><meta charset="utf-8"><style>
@page { size: A4; margin: 20mm 15mm 22mm 15mm; }
* { box-sizing: border-box; }
body { font-family: "Noto Serif CJK SC", "Songti SC", "SimSun", serif; font-size: 12pt; line-height: 1.8; color: #1a1a1a; margin: 0; }
.cover { text-align: center; padding-top: 220px; page-break-after: always; }
.cover-title { font-size: 26pt; margin: 0 0 32px; }
.cover-meta { font-size: 13pt; color: #444; margin-top: 12px; }
h1 { font-size: 20pt; border-bottom: 2px solid #333; padding-bottom: 8px; margin: 28px 0 14px; page-break-after: avoid; }
h2 { font-size: 16pt; margin: 24px 0 12px; page-break-after: avoid; }
h3 { font-size: 13pt; margin: 20px 0 10px; page-break-after: avoid; }
h4, h5, h6 { font-size: 12pt; margin: 16px 0 8px; }
p { margin: 8px 0; text-align: justify; }
ul, ol { margin: 8px 0; padding-left: 28px; }
li { margin: 4px 0; }
blockquote { margin: 12px 0; padding: 8px 16px; border-left: 3px solid #999; color: #555; background: #f7f7f7; }
pre { background: #f4f4f4; padding: 12px 16px; font-size: 10pt; overflow-x: auto; border-radius: 4px; }
table { width: 100%; border-collapse: collapse; margin: 12px 0; font-size: 10.5pt; }
th, td { border: 1px solid #666; padding: 6px 10px; text-align: left; }
th { background: #eee; font-weight: bold; }
.watermark { position: fixed; top: 42%; left: 50%; transform: translate(-50%, -50%) rotate(-30deg); font-size: 72pt; font-weight: bold; color: rgba(180, 40, 40, 0.08); white-space: nowrap; pointer-events: none; z-index: 9999; }
</style></head><body>${wm}${body}</body></html>`;
}

/**
 * @param data {title, bid_no, export_time?, sections: [{title, content, level?}], options?}
 * @param options {vectorHD?, bookmarks?, watermark?, watermarkText?}
 * @returns Buffer PDF 二进制
 */
async function renderPdf(data, options) {
  const opts = { vectorHD: true, bookmarks: true, watermark: false, ...(options || {}), ...((data && data.options) || {}) };
  const html = buildHtml(data, opts);
  const browser = await getBrowser();
  const page = await browser.newPage();
  try {
    await page.setContent(html, { waitUntil: 'load', timeout: 30000 });
    const cdp = await page.createCDPSession();
    const { data: b64 } = await cdp.send('Page.printToPDF', {
      landscape: false,
      displayHeaderFooter: true,
      headerTemplate: '<span></span>',
      footerTemplate: '<div style="width:100%;text-align:center;font-size:9px;color:#888;font-family:sans-serif;"><span>第 <span class="pageNumber"></span> 页 / 共 <span class="totalPages"></span> 页</span></div>',
      printBackground: opts.vectorHD !== false,
      preferCSSPageSize: true,
      generateTaggedPDF: true,
      generateDocumentOutline: opts.bookmarks !== false,
      margin: { top: '20mm', bottom: '22mm', left: '15mm', right: '15mm' },
    });
    return Buffer.from(b64, 'base64');
  } finally {
    await page.close().catch(() => {});
  }
}

module.exports = { renderPdf, buildHtml, renderBody, getChromeStatus };
