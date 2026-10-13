# 全文编辑重构：unioffice outline + @eigenpal/docx-editor-vue

**日期:** 2026-07-23
**状态:** 待审核

## 背景

当前项目有三大功能模块（上传解析、大纲树展示、富文本编辑、AI Chat），各模块独立实现且存在冗余。重构目标是统一编辑引擎，简化后端依赖。

## 架构总览

```
上传 DOCX
  ├──→ 后端 unioffice 解析 heading → 返回树形 Section[] → 后台数据源（AI Chat 上下文等）
  └──→ 前端 @eigenpal/docx-editor-vue 加载完整 DOCX → 全文编辑 + 导出
          │
          └→ useOutlineSidebar → HeadingInfo[] → 转换为 Section[] → OutlineTree 导航展示
```

## OutlineTree 数据源设计

**双数据源模式：**

1. **后端 unioffice 解析**（后台） — 上传时仍调用 `ParseDocument`，提取 heading → 树形 `Section[]` → 供 AI Chat 按章节上下文使用、模板同步
2. **前端 @eigenpal HeadingInfo**（前台，UI 主数据源） — `useOutlineSidebar` composable 返回 `HeadingInfo[]`（含 title, level, paraId, pmPos），转换为 `Section[]` 渲染到 OutlineTree
3. **paraId 映射缓存** — 前端维护 `headingInfo[] → Section[]` 的转换和缓存，点击 OutlineTree 节点时用 paraId 调用 `editorRef.scrollToParaId()`
4. **双向同步** — 用户在 OutlineTree 手动增删节点时，同时更新前端缓存 + 后端 outline（通过 `updateOutlineTree` API）

**数据结构转换：**

```typescript
// @eigenpal HeadingInfo
interface HeadingInfo {
  title: string
  level: number
  paraId: string
  pmPos: number
}

// 转换为 Section（树形）
type Section = {
  id: string        // paraId
  title: string     // title
  level: number     // level
  parent_id: string // 同级前一个同 level 的 paraId 或 ''
  content: ''       // 空（大纲导航模式）
  children: Section[]
  paraId?: string   // 原值保留
}
```

## 后端改动

### 1. docx_service.go — 移除 GenerateDocument 相关代码

**保留的接口（不变）：**
- `ParseDocument(data []byte) (*model.Document, error)` — unioffice 解析 + XML fallback
- `filterKeywordOutline(sections []Section, keyword string) []Section` — 关键词匹配

**移除的方法：**
- `GenerateDocument(*model.Document) ([]byte, error)` — 不再需要手动 XML 生成 DOCX
- `generateDocumentXML()`, `buildDocumentXML()`, `buildContentTypes()`, `buildRels()` — 相关辅助函数
- `NewZipWriter` / `zipwriter.go` — ZIP 写入工具
- `writeSectionXML()`, `writeParagraphXML()`, `xmlEscape()` — XML 生成辅助

**新增/改造的接口：**
- `GenerateMarkdown(doc *model.Document) []byte` — **保留**（用于 Markdown 导出兼容）

**数据模型：** `model.Section` / `model.Document` — **不变**

### 2. handler.go — export API 改造

`ExportDocument` 新流程：
- 不再调用 `docxService.GenerateDocument()`
- 改为返回空响应 `{status: "ok"}`，实际导出由前端 @eigenpal save() 完成

### 3. 其余保持不变

- `chat_service.go` — LLM chat API 不变
- `store.go` — 内存文档存储不变
- `template_service.go` — 模板服务不变

## 前端改动

### 1. ContentEditor.vue → 替换为 @eigenpal/docx-editor-vue

**组件 API 使用：**

```vue
<DocxEditor
  :documentBuffer="docxBuffer"
  :showMenuBar="true"
  :showToolbar="true"
  :readOnly="false"
  :showOutline="true"
  :onDocumentNameChange="onNameChange"
  :onChange="onDocChange"
/>
```

**关键 ref 方法：**
- `editorRef.save()` → Promise<ArrayBuffer> — 导出 DOCX
- `editorRef.scrollToParaId(paraId)` — 跳转到指定段落
- `editorRef.loadDocument(document)` — 加载 Document 对象
- `editorRef.loadDocumentBuffer(buffer)` — 加载 DOCX bytes

### 2. documentStore.ts — 改造

**新增字段：** `section.paraId`（用于 scrollToParaId 跳转）

**变更的方法：**

| 方法 | 旧行为 | 新行为 |
|------|--------|--------|
| `loadOutline(docId)` | GET `/api/document/:id/outline` | 不变，但 outline 数据来自 @eigenpal HeadingInfo |
| `loadSection(docId, sectionId)` | GET 获取单章内容 | 获取 `paraId` 后调用 `editorRef.scrollToParaId(paraId)` |
| `saveSectionContent(docId, sectionId, content)` | PUT 保存单章内容 | 不再需要（内容在 @eigenpal 中直接管理） |
| `updateOutlineTree(docId, newOutline)` | PUT 更新大纲树 | 保留，同步手动增删节点到后端 |

### 3. EditorView.vue — 整合 @eigenpal

在 EditorView 中初始化 `@eigenpal`：
1. 上传完成后，原始文件 bytes 传递给 `docxEditorRef.loadDocumentBuffer()`
2. 使用 `provide/useOutlineSidebar` composable 获取 `headingInfo[]`
3. 将 `headingInfo[]` 转换为 `Section[]` 注入 `documentStore.outline`
4. OutlineTree 点击节点头部 → 通过 `editorRef.scrollToParaId(paraId)` 跳转

### 4. AIChat.vue — 改造

**新功能：**
- 全局指令模式：发送 "帮我改写第一章" → 后端 LLM → 前端替换文档对应章节
- 选区模式：用户在前端 @eigenpal 中选中一段文字 → Chat 面板显示 "编辑选区" 按钮
- Chat 结果通过 `editorRef.setContentControlContent()` 或直接使用 ProseMirror API 写入编辑器

**API 不变：**
- `POST /api/chat` — 支持 `section_id` 参数（可选）
- `section_id` 为空时进行全文对话
- `section_id` 有值时绑定到特定章节上下文

## 数据流

```
用户上传 DOCX
  → POST /api/upload (multipart)
    → 后端 unioffice parse → storeDocument(doc) → 返回 {id, outline: Section[]}
  → 前端接收 {id, outline}
    → docxEditorRef.loadDocumentBuffer(rawFileBytes)
    → @eigenpal useOutlineSidebar → headingInfo[]
    → convert headingInfo[] → Section[] → documentStore.outline

点击左侧 OutlineTree 章节
  → 查找 section.paraId（从 @eigenpal HeadingInfo 缓存）
  → editorRef.scrollToParaId(paraId)

AI Chat 编辑
  → 用户输入 "改写第一章"
  → POST /api/chat { document_id, section_id, message }
    → 后端 LLM 处理 → 返回改写后的段落文本
  → 前端 editorRef.setContentControlContent() 或 DOM 替换

导出 DOCX
  → 点击 "标书生成" / 导出按钮
  → editorRef.save() → ArrayBuffer
  → POST /api/export { format: 'docx', data: arraybuffer }
    → 后端保存文件（未来扩展）
  → 前端直接触发 Blob 下载
```

## 不做的事情

- 不改前端路由（仍是 /editor/:id）
- 不改 OutlineTree 组件 UI 样式
- 不改 AIChat 组件 UI 样式
- 不改 Document 存储方式（仍用内存 store）
- 不添加新的持久化层
