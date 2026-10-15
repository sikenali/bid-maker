'use strict';

/** 桌面客户端打包流水线:
 *
 *   1. parse-service 依赖按生产安装(剔除可选原生模块 @napi-rs/canvas,~220MB -> ~40MB)
 *   2. 交叉编译 Go 后端到 bin/<os>-<arch>/
 *   3. 构建前端 frontend/dist
 *   4. electron-builder 按平台出包
 *
 * 用法:
 *   node scripts/package-desktop.js            # 当前平台
 *   node scripts/package-desktop.js --win|--mac|--linux
 *
 * 注意:第 1 步会改写 parse-service/node_modules(移除 canvas)。
 *      若之后需要本地开发含 canvas 的完整依赖,在 parse-service 下重新 npm ci。
 */

const { execFileSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const REPO_ROOT = path.resolve(__dirname, '..');
const arg = process.argv[2];
// 接受 "--win"/"win" 两种写法;透传给 electron-builder 的应是单横线形式
const platformFlag = /^--/.test(arg) ? arg : arg ? `--${arg}` : '';

const steps = [
  {
    label: '生产化安装 parse-service 依赖(omit optional)',
    run: () => {
      execFileSync('npm', ['ci', '--omit=optional'], {
        cwd: path.join(REPO_ROOT, 'parse-service'),
        stdio: 'inherit',
      });
      const canvasDir = path.join(REPO_ROOT, 'parse-service', 'node_modules', '@napi-rs');
      if (fs.existsSync(canvasDir)) {
        console.warn('[package] 警告:@napi-rs 仍存在于 node_modules,体积未缩减');
      }
    },
  },
  {
    label: '交叉编译 Go 后端(全部目标平台)',
    run: () => execFileSync('node', [path.join(__dirname, 'build-backend.js'), '--all'], {
      cwd: REPO_ROOT,
      stdio: 'inherit',
    }),
  },
  {
    label: '构建前端',
    run: () => execFileSync('npm', ['run', 'build'], { cwd: REPO_ROOT, stdio: 'inherit' }),
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
const cmd = useLocal ? process.execPath : 'npx';
const cmdArgs = useLocal
  ? [localBuilder, platformFlag].filter(Boolean)
  : ['electron-builder', platformFlag].filter(Boolean);
execFileSync(cmd, cmdArgs, { cwd: REPO_ROOT, stdio: 'inherit' });

console.log(`\n[package] 完成,产物目录: ${path.join(REPO_ROOT, 'release')}`);
