'use strict';

/** 交叉编译 Go 后端,产物落到 bin/<os>-<arch>/ 供 electron-builder 按平台打包。
 *
 * 用法:
 *   node scripts/build-backend.js          # 仅当前平台(本地开发)
 *   node scripts/build-backend.js --all     # win-x64 / linux-x64 / mac-x64 / mac-arm64
 *   node scripts/build-backend.js --arch x64
 *
 * 说明:项目无 CGO 依赖(gin/sonic 走纯 Go 路径),CGO_ENABLED=0 可保证
 * 静态二进制,从 Linux 交叉出 win/mac 无需 mingw/llvm。
 */

const { execFileSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const REPO_ROOT = path.resolve(__dirname, '..');
const BACKEND_DIR = path.join(REPO_ROOT, 'backend');
const BIN_ROOT = path.join(REPO_ROOT, 'bin');

const args = new Set(process.argv.slice(2));
const all = args.has('--all');

const TARGETS = all
  ? [
      { os: 'windows', arch: 'amd64', ext: '.exe' },
      { os: 'linux', arch: 'amd64', ext: '' },
      { os: 'darwin', arch: 'amd64', ext: '' },
      { os: 'darwin', arch: 'arm64', ext: '' },
    ]
  : [{ os: process.platform, arch: process.arch === 'x64' ? 'amd64' : process.arch, ext: process.platform === 'windows' ? '.exe' : '' }];

function goVersion() {
  try {
    return execFileSync('go', ['version'], { encoding: 'utf8' }).trim();
  } catch {
    throw new Error('未找到 go 命令,请先安装 Go');
  }
}

console.log(`[build-backend] ${goVersion()}`);

for (const t of TARGETS) {
  const outDir = path.join(BIN_ROOT, `${t.os}-${t.arch}`);
  fs.mkdirSync(outDir, { recursive: true });
  const out = path.join(outDir, `bid-maker-server${t.ext}`);

  console.log(`[build-backend] ${t.os}/${t.arch} -> ${path.relative(REPO_ROOT, out)}`);
  execFileSync('go', ['build', '-trimpath', '-ldflags', '-s -w', '-o', out, './cmd/server/'], {
    cwd: BACKEND_DIR,
    env: { ...process.env, CGO_ENABLED: '0', GOOS: t.os, GOARCH: t.arch },
    stdio: 'inherit',
  });
}

console.log(`[build-backend] 完成,产物目录: ${path.relative(REPO_ROOT, BIN_ROOT)}`);
