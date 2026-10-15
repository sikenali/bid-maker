'use strict';

/** 桌面客户端打包流水线:
 *
 *   1. parse-service 依赖按生产安装(剔除可选原生模块,缩减体积)
 *   2. 交叉编译 Go 后端到 bin/<os>-<arch>/
 *   3. 构建前端 frontend/dist
 *   4. electron-builder 按平台出包
 *
 * 用法:
 *   node scripts/package-desktop.js              # 当前平台
 *   node scripts/package-desktop.js --win|--mac|--linux
 *
 * 注意:第 1 步会改写 parse-service/node_modules。
 *      若之后需要本地开发含可选依赖的完整依赖,在 parse-service 下重新 npm ci。
 */

const { execFileSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const REPO_ROOT = path.resolve(__dirname, '..');
// 平台参数整体透传给 electron-builder,例如 "--win --x64" / "--mac --x64 --arm64";
// 接受带或不带双横线(win / --win),不带时自动补齐
const platformArgs = process.argv.slice(2).map((a) => (/^--/.test(a) ? a : `--${a}`));
const platformFlag = platformArgs.join(' ');

// Windows 上 npm 是 .cmd shim,execFileSync 需显式带扩展名
const NPM = process.platform === 'win32' ? 'npm.cmd' : 'npm';

/** 目录体积(MB),用于核对依赖裁剪效果 */
function dirSizeMb(dir) {
  let total = 0;
  let entries = [];
  try { entries = fs.readdirSync(dir, { withFileTypes: true }); } catch { return 0; }
  const stack = [dir];
  while (stack.length) {
    const base = stack.pop();
    let items = [];
    try { items = fs.readdirSync(base, { withFileTypes: true }); } catch { continue; }
    for (const it of items) {
      const p = path.join(base, it.name);
      try {
        if (it.isDirectory()) stack.push(p);
        else total += fs.statSync(p).size;
      } catch { /* ignore */ }
    }
  }
  void entries;
  return Math.round((total / 1048576) * 10) / 10;
}

const steps = [
  {
    label: '生产化安装 parse-service 依赖(omit optional)',
    run: () => {
      execFileSync(NPM, ['ci', '--omit=optional'], {
        cwd: path.join(REPO_ROOT, 'parse-service'),
        stdio: 'inherit',
      });
      const nm = path.join(REPO_ROOT, 'parse-service', 'node_modules');
      console.log(`[package] parse-service/node_modules: ${dirSizeMb(nm)} MB`);
    },
  },
  {
    label: '交叉编译 Go 后端(全部目标平台)',
    run: () => execFileSync(process.execPath, [path.join(__dirname, 'build-backend.js'), '--all'], {
      cwd: REPO_ROOT,
      stdio: 'inherit',
    }),
  },
  {
    label: '构建前端',
    run: () => execFileSync(NPM, ['run', 'build'], { cwd: REPO_ROOT, stdio: 'inherit' }),
  },
];

console.log(`[package] 目标平台: ${platformFlag || '当前平台'}`);
for (const step of steps) {
  console.log(`\n[package] ▶ ${step.label}`);
  step.run();
}

console.log(`\n[package] ▶ electron-builder ${platformFlag || ''}`.trim());
// 优先用本地 devDependency(electron-builder ^25,与本仓库 build 配置匹配);缺失时回退 npx
const localBuilder = path.join(REPO_ROOT, 'node_modules', 'electron-builder', 'cli.js');
const useLocal = fs.existsSync(localBuilder);
const cmd = useLocal ? process.execPath : NPM;
const cmdArgs = useLocal
  ? [localBuilder, ...platformArgs]
  : ['exec', 'electron-builder', ...platformArgs];
execFileSync(cmd, cmdArgs, { cwd: REPO_ROOT, stdio: 'inherit' });

console.log(`\n[package] 完成,产物目录: ${path.join(REPO_ROOT, 'release')}`);
