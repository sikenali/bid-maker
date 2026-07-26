# TOC提取与编辑器联动优化设计

## 背景

当前项目中 TOC（目录）提取与 `@eigenpal/docx-editor-vue` 编辑器的联动存在三个核心问题：

1. **大纲导航脆弱** — 点击左栏大纲节点后，通过启动时轮询 DOM 缓存的 `pmPos` 定位编辑器位置，编辑器内容变更后缓存失效
2. **大纲同步缺失** — 用户在编辑器中修改标题，左栏 OutlineTree 无反应
3. **TOC 提取不完善** — Heading 检测只匹配有限样式名，关键词硬编码

## 设计

### 1. 导航改进（Outline → Editor）

**现状**：`EditorView.vue` 中的 `handleEditorReady()` 使用 `setInterval` 每 100ms 扫描编辑器 DOM，最多 60 次，提取 `data-pm-start` 属性后按标题文本匹配并缓存 pmPos。之后点击大纲节点时使用缓存值。

**问题**：用户编辑后缓存失效，导航指向错误位置或不可用。

**方案**：实时查询 + 短时缓存

- 删除 `handleEditorReady()` 中的 `setInterval` 轮询缓存逻辑
- 每次点击大纲节点时，实时扫描编辑器 DOM 查找当前标题文本对应的 `data-pm-start`
  - 只扫描标记 `[data-pm-start]` 的元素，非全 DOM 扫描
  - 按标题文本精确匹配优先，降级到 `includes` + level 联合匹配
- 引入 3 秒短时缓存 `Map<string, {pmPos, timestamp}>`
  - 缓存 key 为 section.id，避免文本匹配歧义
  - 编辑器 DOM 发生任何 `input` 或 `characterData` 变更时清空缓存

**涉及文件**：
- `frontend/src/views/EditorView.vue` — 重构 handleEditorReady + handleSelectSection
- `frontend/src/stores/documentStore.ts` — 新增 headingPositions 缓存管理

### 2. 同步改进（Editor → Outline）

**现状**：无任何同步机制。

**方案**：MutationObserver 监听编辑器 DOM 文本变化

- EditorView 挂载后，在 `.docx-editor-vue__pages-viewport` 上挂载 `MutationObserver`
  - 配置 `{ childList: true, subtree: true, characterData: true }`
  - 过滤：仅处理有 `[data-pm-start]` 属性元素的文本变化
- 检测到字符变更后，300ms debounce，然后：
  - 重新扫描所有 `[data-pm-start]` 元素，提取文本 + pmPos
  - 与 `docStore.outline` 中对应 pmPos 的 section 对比
  - 若文本变化则更新 section.title（仅更新 title 字段，不触发后端同步）
- 防抖 300ms，避免高频编辑导致频繁 DOM 扫描

**涉及文件**：
- `frontend/src/views/EditorView.vue` — 新增 MutationObserver 逻辑
- `frontend/src/stores/documentStore.ts` — 新增 syncHeadingFromEditor 方法

### 3. TOC 提取改进

**现状**：`docx_service.go` 的 `isHeadingUnioffice` 只匹配 `Heading1`/`标题1` 等样式名。关键词 `"投标文件"` 在 `NewDocxService()` 中硬编码。

**方案**：

**3.1 Heading 模式增强**

新增识别模式（按优先级）：

| 优先级 | 模式 | 示例 | Level 映射 |
|--------|------|------|-----------|
| 1 | `Heading{N}` / `heading{N}` / `Heading {N}` | `Heading1` | N |
| 2 | `标题{N}` / `标题 {N}` | `标题1` | N |
| 3 | `{X}级标题` | `一级标题` | 一→1, 二→2 … 九→9 |
| 4 | `第{X}章` | `第一章` | 1 |
| 5 | `第{X}节` | `第一节` | 2 |
| 6 | `OutlineLevel` 样式属性 | outlineLvl="0" | 0→1, 1→2 … |
| 7 | 数字编号正则 | `^[1-9]\.` / `^\d+\.\d+` | 1 / 2 |

**3.2 关键词可配置**

- Upload API 新增可选 query 参数: `POST /upload?keyword=招标文件`
- 新增后端 API: `POST /document/:id/reparse`
  - Body: `{ "keyword": "招标文件" }`
  - 使用后端存储的原始 DOCX buffer 重新提取
- Document 模型新增 `RawBuffer []byte` 字段（仅后端存储，不返回前端）
- Store 存储时保存原始 buffer

**涉及文件**：
- `backend/internal/service/docx_service.go` — 增强 isHeading + 新增 reparse 方法
- `backend/internal/handler/handler.go` — 新增 reparse 端点
- `backend/internal/model/document.go` — 新增 RawBuffer 字段
- `backend/internal/service/store.go` — 确保 buffer 随 doc 存储
- `frontend/src/api/client.ts` — 新增 reparseDocument API

### 4. 内容显示改进

**现状**：编辑器自带大纲 (`show-outline="true"`) 与左栏 OutlineTree 重复，无加载状态。

**方案**：
- `<DocxEditor :show-outline="false" />` 隐藏编辑器自带大纲
- 编辑器加载完成前显示骨架屏（灰色占位区域）
- 编辑器加载失败显示错误信息和重试按钮

**涉及文件**：
- `frontend/src/views/EditorView.vue` — 修改 show-outline + 骨架屏

## 不变内容

- 后端 OutlineTree 的增删改操作（`promoteLevel`/`demoteLevel`/`addChild`/`removeSection`）保持不变，继续通过 `PUT /document/:id/outline` 同步
- AI 聊天推送到编辑器的逻辑（`AIChat.vue` 的 `tryApplyAIContent`）不变
- 导出功能不变
- 路由和页面布局不变

## 测试策略

- 前端：手动测试导航点击跳转准确性、编辑器改标题后大纲更新
- 后端：`docx_service_test.go` 新增中文 heading 模式的单元测试
- 后端：新增 `POST /document/:id/reparse` 的集成测试
