# Markdown 管线重构（AnyDoc 上传解析 + md 单一源 + 回填导出）

## Overview

将投标文档生成流程统一为"Markdown 单一源"管线：上传 `.docx` 用 AnyDoc 转 Markdown → 从 md 提取章节树 + 保留全文 → 原生 Markdown 编辑器编辑 / AI 生成直接产 md → 导出时前端把最新 md 回填到现有 docx 编辑器 buffer，复用 `DocxEditor.save()` 产出 docx。

## 决策记录

| 决策点 | 结论 |
|--------|------|
| 引入 AnyDoc 动机 | 统一为 Markdown 管线 |
| 转换运行位置 | 后端运行 |
| Go 后端集成方式 | subprocess 调 anydoc CLI（`os/exec`） |
| 上传文件范围 | 仍仅 `.docx`，仅解析引擎换成 AnyDoc |
| 编辑器 | 原生 Markdown 编辑 |
| md→docx 导出 | Markdown 为源，导出回填 docx buffer 复用现有 `DocxEditor.save()` |
| unioffice | 彻底移除（解析、GenerateMarkdown、Reparse 及相关测试全部删除，无兜底） |

## 架构与数据流

```
上传 .docx
  → POST /api/upload（前端传文件）
  → Go 后端 os/exec 调 anydoc CLI → 产出 GFM .md 文本
  → ParseOutlineFromMD（扩展）提取章节树 + 各章节正文 Content + 全文 Markdown
  → Document 存储：{Title, Outline[], Markdown, RawBuffer}
  → 返回 document_id + 章节树

编辑（用户 / AI）：
  → AI 生成 = generate-service（已产 Markdown）→ 更新章节 Content + 全文 Markdown
  → 前端原生 Markdown 编辑器 ←→ 后端存 .md

导出：
  → 前端把当前 md → 回填构建 docx buffer → 喂给现有 DocxEditor → save() → 下载 docx
```

## 后端

### 数据模型（`model/document.go`）

`Section` 不变（`ID/Title/Level/ParentID/Content/Children`）。`Document` 新增 `Markdown string` 作为全局单一源：

```go
type Document struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Outline   []Section `json:"outline"`
    Markdown  string    `json:"markdown"`   // 新增：全文 GFM 源
    RawBuffer []byte    `json:"-"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### AnyDocService（新增 `backend/internal/service/anydoc_service.go`）

```go
type AnyDocService struct {
    Binary string // anydoc CLI 路径，默认 "anydoc"
}

// ConvertDocxToMarkdown: 用 os/exec 调 anydoc <in.docx> -o -（或临时文件），
// 返回 GFM markdown 文本。binary 缺失 / 执行失败 / 超时 → 返回 error。
```

关键点：
- `os/exec.Command` + 超时（默认 ~30s），捕获 `Stderr`
- 输入的 docx bytes 先落临时文件，调用后删除
- 产出 md 交给章节解析提取 Outline 与全文 Markdown
- `NewAnyDocService()` 读取 env `ANYDOC_BIN`，未配置时服务仍可启动，仅转换时返回错误

### 章节提取（md→树，扩展 `ParseOutlineFromMD`）

复用 `generate_service.go` 的 `ParseOutlineFromMD` 思路构建章节树，并扩展为同时保留各章节**正文 Content**：

- 按 Markdown 标题行（`#{1,6} `）分段
- 标题行归属章节树（复用现有层级/父子逻辑）
- 标题行到下一标题之间的正文并入该标题对应叶子章节的 `Content`
- 保留全文用于 `Document.Markdown`

API 建议新增独立函数 `ParseSectionsWithContent(md string) ([]model.Section, string)`，返回章节树与可能被清洗后的全文；`ParseOutlineFromMD` 保留给 AI 大纲预览用（无正文）。

### 上传/模板 Handler 改动（`handler.go`）

- 移除对 `docxService` 的依赖，改用 `anyDocService`
- 上传与模板导入路径：调用 `ConvertDocxToMarkdown` → `ParseSectionsWithContent` → 构建 `model.Document`（含 `Markdown`）
- 删除 `ReparseDocument`（依赖 unioffice）与相关路由；如保留"重新解析"能力，改为在前端对已存 md 重新走章节提取

### 移除 unioffice

- 删除 `docx_service.go` 全部内容（或整文件移除），及其测试 `docx_service_test.go`、`debug_test.go` 内用例
- 从 `go.mod` 移除 `github.com/unidoc/unioffice/v2`
- 确认无其他引用后清理 `chineseNum`、`findLeaf` 等不再使用的辅助（若被复用则保留到新模块）

## 前端

### 原生 Markdown 编辑器接入（`EditorView.vue`）

- Markdown 为显示与编辑的源；`DocxEditor` 不再作为主要编辑界面
- 新增/复用 Markdown 编辑面板（如 `CodeMirror` 或极简 textarea，或引入 markdown 编辑库），显示当前 doc 的 `Markdown`
- 编辑内容变更 → 防抖同步到后端（`saveSectionContent` / 新增保存全文 md 的接口）
- AI 生成 chunk：generate-service 已产 md，直接更新 Markdown 源并刷新编辑器

### 导出回填 docx buffer

- 前端维护 `currentMarkdown`
- 导出时：把 `currentMarkdown` → 构建/回填一个新 docx docBuffer（前端 md→docx 序列化，或交互后芯片）→ 喂给 `DocxEditor` → `save()` → 下载
- 因 unioffice 已移除，后端不再负责 md→docx 字节构建；回填必须在前端完成

## 错误处理与降级

- anydoc 二进制缺失 / 执行超时 / 转换失败 → 上传接口返回明确错误，前端提示"解析失败，请检查文件或服务器配置"
- 因 unioffice 已彻底移除，无自动降级（已确认接受该风险）
- 前端在上传前按扩展名拦截非 `.docx`；大文件 / 异常文件由 anydoc 报错

## 测试

- 后端：`ParseSectionsWithContent` 单元测试（标题层级、父子关系、正文归属）；`AnyDocService` 用样例 docx 断言产出含标题与正确章节树
- 删除旧的 unioffice 依赖测试（`docx_service_test.go`、`debug_test.go` 相关用例）

## 范围外（v1）

- 编辑器富文本工具栏增强（表格/复杂样式）——v1 用基础 Markdown 编辑
- 多格式上传（pdf/pptx/xlsx）——仍仅 docx
- anydoc 二进制自动下载 / 内嵌——假设部署时已安装可执行文件
- 数据库持久化——沿用当前内存 store