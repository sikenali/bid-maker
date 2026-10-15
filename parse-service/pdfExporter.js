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

const puppeteer = require('puppeteer-core');
const fs = require('fs');

const EXE_CANDIDATES = [
  process.env.CHROME_PATH,
  '/usr/bin/google-chrome',
  '/usr/bin/chromium',
  '/usr/bin/chromium-browser',
  '/opt/google/chrome/chrome',
  '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
  'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe',
].filter(Boolean);

let browserPromise = null;

function resolveChrome() {
  for (const p of EXE_CANDIDATES) {
    try { fs.accessSync(p, fs.constants.X_OK); return p; } catch { /* next */ }
  }
  return null;
}

async function getBrowser() {
  if (!browserPromise) {
    browserPromise = (async () => {
      const exe = resolveChrome();
      if (!exe) {
        throw new Error('未找到可用的 Chrome,请安装 Chrome 或设置 CHROME_PATH 环境变量');
      }
      return puppeteer.launch({
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

module.exports = { renderPdf, buildHtml, renderBody };
