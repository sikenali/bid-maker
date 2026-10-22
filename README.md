# 文制星 — AI 智能标书编写系统

<p align="center">
  <strong>招标文件解析 · AI 大纲生成 · 逐章流式补写 · Word 专业导出</strong>
</p>

<p align="center">
  <img src="frontend/public/logo-icon.svg" width="120" height="120" alt="文制星 Logo">
</p>

<p align="center">
  <img alt="Vue" src="https://img.shields.io/badge/Vue_3-3.5-4FC08D?logo=vue.js&logoColor=white">
  <img alt="TypeScript" src="https://img.shields.io/badge/TypeScript-5.4-3178C6?logo=typescript&logoColor=white">
  <img alt="Vite" src="https://img.shields.io/badge/Vite-5-646CFF?logo=vite&logoColor=white">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white">
  <img alt="License" src="https://img.shields.io/badge/License-MIT-yellow">
</p>

## 软件介绍

文制星是一套基于大语言模型的 AI 标书编写系统，覆盖从招标文件解析、结构化大纲生成、逐章智能补写到 Word 专业导出的全流程。支持 docx / pdf / xlsx / pptx / md / txt / html 多格式上传，Markdown（Tiptap）与 Word（DocxEditor）双模式编辑器实时同步，具备技能系统与多模型路由能力，是标书编写场景的一站式智能工具。

**核心功能**

| 功能 | 说明 |
|------|------|
| 文件解析 | 支持 docx/pdf/xlsx/pptx/md/txt/html，提取技术指标、商务要求、评分点、废标条款 |
| 大纲管理 | auto（纯 AI）/ custom（纯用户输入）/ hybrid（结合解析结果）三种生成模式 |
| 逐章生成 | SSE 流式输出，实时预览，进度追踪 |
| 全文生成 | 自动按大纲顺序逐章生成，支持断点续跑 |
| 双模式编辑 | Markdown（Tiptap）/ Word（DocxEditor）实时切换，双向同步 |
| AI 助手 | 支持多模型路由，带上下文感知的智能问答 |
| 技能系统 | 本地技能扫描 + 自定义技能导入，支持 `/` 触发命令面板 |
| 模板管理 | 内置标准投标/技术方案/商务标模板，支持自定义上传 |
| 导出 | Markdown / Word 双格式，专业排版样式 |

## 技术实现

### 解析 → 大纲 → 逐章生成 → 导出

系统由三个服务构成：Node.js 解析服务（parse-service）通过 officeparser 提取技术指标、商务要求、评分点、废标条款；Go 后端（Gin）负责会话、大纲、生成、导出等 API 编排，并调用配置的 LLM Provider；Vue 3 前端提供上传、大纲、双模式编辑、AI 助手与设置界面。逐章生成通过 SSE 流式推送，支持断点续跑与实时预览。

### 前端架构

Vue 3 + TypeScript + Vite + Pinia 单页应用。`EditorView` 为主编辑器，集成 `OutlineTree`（大纲树）、Tiptap（Markdown）与 DocxEditor（Word）双模式，配合 `GenerateFlowDialog` 与 `chatStore` 完成 AI 生成与对话。8 个设置模块通过 `SettingsModal` 统一管理，全部持久化至 localStorage。

### 后端模式

Go + Gin 后端（8080）承担会话管理、大纲锁定、章节生成、Word 导出等能力；Node.js 解析服务（8000）独立运行，专注文档解析；两者通过 HTTP 通信。支持通过环境变量 `LLM_PROVIDERS`（JSON）配置多厂商模型路由，含连通性测试接口。

### 桌面客户端

基于 Electron 构建 Windows / macOS / Linux 三平台原生应用，打包时会把 Go 后端二进制与 Node.js 解析服务一同打入 extraResources。

```bash
# 安装依赖
npm i
cd parse-service && npm i && cd ..
cd frontend && npm i && cd ..
cd backend && go mod download && cd ..

# 开发模式（同时启动 parse-service + Go 后端 + Vue 前端 + Electron）
npm run electron:dev

# 打包 Windows exe
npm run package:desktop:win

# 打包 macOS dmg
npm run package:desktop:mac

# 打包 Linux AppImage
npm run package:desktop:linux
```

---

## 声明

1. **使用目的**：本工具旨在辅助标书内容生成，用户应自行核对最终文档的合规性与准确性。
2. **数据使用**：调用外部 LLM 时，用户输入内容将发送至所配置的 API 端点，请谨慎选择可信模型服务。
3. **免责声明**：本工具按"现有状态"提供，不作任何形式的明示或默示保证。

## License

[MIT](./LICENSE)
