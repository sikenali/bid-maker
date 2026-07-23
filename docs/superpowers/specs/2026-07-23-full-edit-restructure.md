# 全文编辑重构：unioffice outline + @eigenpal/docx-editor-vue

**日期:** 2026-07-23
**状态:** 待审核

## 背景

当前项目有三大功能模块（上传解析、大纲树展示、富文本编辑、AI Chat），各模块独立实现且存在冗余。重构目标是统一编辑引擎，简化后端依赖。

## 架构总览

```
上传 DOCX
  ├──→ 后端 unioffice 解析 heading → 返回树形 Section[] → 前端 OutlineTree 展示
  └──→ 前端 @eigenpal/docx-editor-vue 加载完整 DOCX → 全文编辑 + 导出
```

## 后端改动

### 1. docx_service.go — 保留 unioffice 解析，移除 GenerateDocument

**保留的接口（不变）：**
- `ParseDocument(data []byte) (*model.Document, error)` — unioffice 解析 + XML fallback
- `filterKeywordOutline(sections []Section, keyword string) []Section` — 关键词匹配
- `extractSectionsWithKeyword(paras []Paragraph, keyword string) []Section` — 从关键词位置提取

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
- 或保留 DOCX 文件存储服务（未来扩展）

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

```typescript
interface Section {
  id: string
  title: string
  level: number
  parent_id: string
  content: string
  children: Section[]
  // 新增
  paraId?: string  // Word 内部 paraId，用于 scrollToParaId
}
```

**变更的方法：**

| 方法 | 旧行为 | 新行为 |
|------|--------|--------|
| `loadOutline(docId)` | GET `/api/document/:id/outline` | 不变 |
| `loadSection(docId, sectionId)` | GET 获取单章内容 | 获取 `paraId` 后调用 `editorRef.scrollToParaId(paraId)` |
| `saveSectionContent(docId, sectionId, content)` | PUT 保存单章内容 | 调用 `editorRef.setContentControlContent()` 或修改选中区域 |
| `updateOutlineTree(docId, newOutline)` | PUT 更新大纲树 | 保留，同步后端 outline 变更 |

### 3. OutlineTree.vue — 调整点击行为

**旧行为：** 点击章节 → emit select(id) → loadSection 请求后端 API → 渲染 HTML
**新行为：** 点击章节 → emit select(id) → 查找对应 paraId → 调用 `editorRef.scrollToParaId(paraId)`

**OutlineTree 数据结构调整：**
后端 ParseDocument 返回的 Section 需要包含 `paraId` 字段（从 unioffice 的 paragraph ID 映射）。

**但当前模型不支持 paraId** → 方案选择：
A. 改造 model.Section 增加 `paraId` 字段（需改前后端）
B. OutlineTree 仅作为导航展示，不提供跳转（简单）
C. 前端维护 sectionId → paraId 的映射缓存

我推荐 C：前端通过 `@eigenpal` 的 `useOutlineSidebar` composable 获取 `HeadingInfo[]` 并缓存 paraId 映射。

### 4. AIChat.vue — 改造

**新功能：**
- 全局指令模式：发送 "帮我改写第一章" → 后端 LLM → 替换文档对应章节
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
    → documentStore.outline = res.data.outline
    → docxEditorRef.loadDocumentBuffer(rawFileBytes)
  
点击左侧 OutlineTree 章节
  → 查找 section.paraId（从 @eigenpal HeadingInfo 缓存）
  → editorRef.scrollToParaId(paraId)
  
AI Chat 编辑
  → 用户输入 "改写第一章"
  → POST /api/chat { document_id, section_id, message }
    → 后端 LLM 处理 → 返回改写后的段落文本
  → 前端 editorRef.setContentControlContent() 或 DOM 替换
```

## 不做的事情

- 不改前端路由（仍是 /editor/:id）
- 不改 OutlineTree 组件 UI 样式
- 不改 AIChat 组件 UI 样式
- 不改 Document 存储方式（仍用内存 store）
- 不添加新的持久化层
