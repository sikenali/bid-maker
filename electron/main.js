'use strict';

/** 文制星 桌面客户端主进程.
 *
 * 进程编排(生产模式):
 *   parse-service(Node)  ELECTRON_RUN_AS_NODE=1  默认 :18000
 *   Go 后端(bin/)                 默认 :18080,并托管 frontend/dist
 *   BrowserWindow  -> http://127.0.0.1:<GO_PORT>
 *
 * 开发模式(NODE_ENV=development):不启动任何子进程,直接加载 vite dev server。
 * 端口被占用时自动改用空闲端口,避免与本机开发实例(8000/8080/3000)冲突。
 */

const { app, BrowserWindow, ipcMain, dialog } = require('electron');
const path = require('path');
const net = require('net');
const http = require('http');
const { spawn } = require('child_process');

const isDev = process.env.NODE_ENV === 'development';
const APP_NAME = '文制星';
const PREF_PARSE_PORT = 18000;
const PREF_BACKEND_PORT = 18080;

// 开发:仓库根目录;生产:app.asar 根目录
const APP_ROOT = isDev ? path.join(__dirname, '..') : app.getAppPath();
// extraResources 落点。注意:app.getPath('resources') 必须在 app ready 之后调用,
// 故用惰性函数而非顶层常量(顶层调用会抛 "Failed to get 'resources' path")。
function resourcesRoot() {
  return isDev ? APP_ROOT : app.getPath('resources');
}

let backendProc = null;
let parseProc = null;
let ports = { parse: PREF_PARSE_PORT, backend: PREF_BACKEND_PORT };

// ── 单实例锁:防止二次启动导致端口占用与进程泄漏 ─────────────────────────
if (!app.requestSingleInstanceLock()) {
  app.quit();
  process.exit(0);
}
app.on('second-instance', () => {
  const [win] = BrowserWindow.getAllWindows();
  if (win) {
    if (win.isMinimized()) win.restore();
    win.focus();
  }
});

// ── 端口分配 ───────────────────────────────────────────────────────────
function portFree(port) {
  return new Promise((resolve) => {
    const srv = net.createServer();
    srv.once('error', () => resolve(false));
    srv.listen(port, '127.0.0.1', () => {
      srv.close(() => resolve(true));
    });
  });
}

/** 首选端口空闲则用之,否则向系统申请随机空闲端口。 */
async function pickPort(preferred) {
  if (await portFree(preferred)) return preferred;
  return new Promise((resolve) => {
    const srv = net.createServer();
    srv.listen(0, '127.0.0.1', () => {
      const p = srv.address().port;
      srv.close(() => resolve(p));
    });
  });
}

// ── HTTP 就绪探测 ──────────────────────────────────────────────────────
function waitHttp(url, timeout = 60000) {
  const start = Date.now();
  return new Promise((resolve, reject) => {
    const probe = () => {
      const req = http.get(url, (res) => {
        res.resume();
        if (res.statusCode === 200) return resolve(res);
        retry();
      });
      req.on('error', retry);
      req.setTimeout(2000, () => req.destroy(new Error('probe timeout')));
    };
    const retry = () => {
      if (Date.now() - start > timeout) return reject(new Error(`服务未在 ${Math.round(timeout / 1000)}s 内就绪: ${url}`));
      setTimeout(probe, 250);
    };
    probe();
  });
}

// ── 子进程:解析服务(Node) ────────────────────────────────────────────
function startParseService(port) {
  // 生产:extraResources/parse-service(随包分发完整 node_modules);开发:仓库内
  const entry = isDev
    ? path.join(APP_ROOT, 'parse-service', 'server.js')
    : path.join(resourcesRoot(), 'parse-service', 'server.js');
  parseProc = spawn(process.execPath, [entry], {
    env: {
      ...process.env,
      ELECTRON_RUN_AS_NODE: '1',
      PORT: String(port),
      // 打包后禁用 ELECTRON_RUN_AS_NODE 环境残留对子进程的影响
      NODE_ENV: 'production',
    },
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  wireProcLogs(parseProc, 'parse-service');
  return parseProc;
}

// ── 子进程:Go 后端 ────────────────────────────────────────────────────
/** 后端二进制候选名。
 *  Windows/Linux 单架构打包:extraResources 直接落到 bin/<exe>
 *  macOS 双架构打包:v25 不支持 mac.arm64 子配置,资源按 -x64/-arm64 后缀区分,
 *  主进程按 process.arch 命中对应文件。
 */
function backendBinNames() {
  const exe = process.platform === 'win32' ? 'bid-maker-server.exe' : 'bid-maker-server';
  const names = [exe];
  if (process.platform === 'darwin') names.push(`${exe}-${process.arch}`);
  return names;
}

function backendBinaryCandidates() {
  const dirs = [path.join(resourcesRoot(), 'bin'), path.join(APP_ROOT, 'bin')];
  const out = [];
  for (const dir of dirs) {
    for (const name of backendBinNames()) out.push(path.join(dir, name));
  }
  // 开发/手动编译:<os>-<arch> 子目录
  const sub = path.join(APP_ROOT, 'bin', `${process.platform}-${process.arch === 'x64' ? 'amd64' : process.arch}`);
  for (const name of backendBinNames()) out.push(path.join(sub, name));
  return out;
}

function backendBinary() {
  for (const c of backendBinaryCandidates()) {
    try {
      if (require('fs').accessSync(c, require('fs').constants.X_OK)) return c;
    } catch { /* next */ }
  }
  return null;
}

function startBackend(port) {
  const bin = backendBinary();
  if (!bin) {
    throw new Error(
      `未找到 Go 后端二进制,请先运行 npm run build:backend\n已尝试:\n  ${backendBinaryCandidates().join('\n  ')}`
    );
  }
  const userData = app.getPath('userData');
  backendProc = spawn(bin, [], {
    env: {
      ...process.env,
      PORT: String(port),
      // Go 端变量名沿用 PYTHON_SERVICE 指向解析服务(保持服务端兼容)
      PARSE_SERVICE: `http://127.0.0.1:${ports.parse}`,
      PYTHON_SERVICE: `http://127.0.0.1:${ports.parse}`,
      // 前端构建产物:开发在仓库内,生产在 extraResources(真目录;
      // Go 子进程无法读取 asar 归档,必须指向 resources 下的实体文件)
      FRONTEND_DIST: isDev
        ? path.join(APP_ROOT, 'frontend', 'dist')
        : path.join(resourcesRoot(), 'frontend', 'dist'),
      BID_MAKER_DATA_DIR: userData,
      GIN_MODE: 'release',
    },
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  wireProcLogs(backendProc, 'backend');
  return backendProc;
}

function wireProcLogs(proc, name) {
  if (!proc.stdout || !proc.stderr) return;
  const relay = (stream, tag) => {
    let buf = '';
    stream.setEncoding('utf8');
    stream.on('data', (chunk) => {
      buf += chunk;
      const lines = buf.split('\n');
      buf = lines.pop() || '';
      for (const line of lines) if (line) console.log(`[${tag}] ${line}`);
    });
  };
  relay(proc.stdout, name);
  relay(proc.stderr, name);
  proc.on('exit', (code, signal) => {
    console.log(`[${name}] 进程退出 code=${code} signal=${signal}`);
    if (!app.isQuitting) {
      // 子进程意外退出:提示后退出应用,避免留下半死状态
      dialog.showErrorBox('启动失败', `${name} 进程意外退出(code=${code})`);
      app.quit();
    }
  });
}

// ── 启动编排 ───────────────────────────────────────────────────────────
async function bootServices() {
  ports.parse = await pickPort(PREF_PARSE_PORT);
  ports.backend = await pickPort(PREF_BACKEND_PORT);
  console.log(`[electron] 分配端口: parse=${ports.parse} backend=${ports.backend}`);

  startParseService(ports.parse);
  await waitHttp(`http://127.0.0.1:${ports.parse}/health`);
  console.log('[electron] parse-service 已就绪');

  startBackend(ports.backend);
  await waitHttp(`http://127.0.0.1:${ports.backend}/health`);
  console.log('[electron] Go 后端已就绪');
}

function createWindow() {
  const win = new BrowserWindow({
    width: 1280,
    height: 860,
    minWidth: 1100,
    minHeight: 700,
    show: false,
    backgroundColor: '#FAF7F2',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: false,
    },
  });

  win.setTitle(`${APP_NAME} v${app.getVersion()} (${process.platform}/${process.arch})`);
  win.once('ready-to-show', () => win.show());

  if (isDev) {
    win.loadURL(process.env.FRONTEND_URL || 'http://localhost:3000');
  } else {
    bootServices()
      .then(() => win.loadURL(`http://127.0.0.1:${ports.backend}`))
      .catch((err) => {
        console.error('[electron]', err);
        dialog.showErrorBox('启动失败', err.message);
        app.quit();
      });
  }

  // 阻止窗口内外部链接跳走浏览器
  win.webContents.setWindowOpenHandler(({ url }) => {
    if (/^https?:/.test(url)) require('electron').shell.openExternal(url);
    return { action: 'deny' };
  });
}

function stopChildren() {
  app.isQuitting = true;
  for (const proc of [backendProc, parseProc]) {
    if (!proc || proc.killed) continue;
    try {
      proc.kill();
    } catch { /* ignore */ }
  }
  backendProc = null;
  parseProc = null;
}

app.whenReady().then(() => {
  createWindow();
  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});
app.on('before-quit', () => stopChildren());
app.on('will-quit', () => stopChildren());

// ── IPC ────────────────────────────────────────────────────────────────
ipcMain.handle('get-app-info', () => ({
  version: app.getVersion(),
  platform: process.platform,
  arch: process.arch,
  isDesktop: true,
  userData: app.getPath('userData'),
}));

ipcMain.handle('get-runtime-info', () => ({
  backendPort: ports.backend,
  parseServicePort: ports.parse,
  backendBinary: backendBinary(),
}));

ipcMain.handle('show-about-dialog', () => {
  dialog.showMessageBox({
    type: 'info',
    title: `关于 ${APP_NAME}`,
    message: `${APP_NAME}\n版本 ${app.getVersion()}\n${process.platform}/${process.arch}`,
    detail: `用户数据目录: ${app.getPath('userData')}`,
    buttons: ['确定'],
  });
});
