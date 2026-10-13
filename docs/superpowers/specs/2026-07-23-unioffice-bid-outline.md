# unioffice 提取投标文件大纲

**日期:** 2026-07-23
**状态:** 待审核

## 背景

当前后端 `docx_service.go` 使用手动 XML 解析 DOCX 文件（通过 `archive/zip` + `encoding/xml`），仅支持基础 heading 样式识别，解析能力有限。需要引入 `unioffice` 库替换手动解析，并实现按关键词（如"投标文件"）提取子章节大纲的功能。

## 需求

1. 上传 DOCX 后，后端用 unioffice 解析
2. 找到标题包含"投标文件"的 Heading 章节
3. 提取该章节下的所有子章节作为大纲（树形结构，保留层级）
4. 忽略其他章节内容
5. 关键词"投标文件"可配置，后续可扩展
6. 前端 OutlineTree 支持增删节点、点击查看内容（已有功能，无需改动）

## 改动范围

### 后端

#### 1. 依赖变更

`go.mod` 新增：
```
require github.com/unidoc/unioffice/v2 v2.12.0
```
替换为 fork：
```
replace github.com/unidoc/unioffice/v2 => github.com/sikenali/unioffice/v2 v2.0.0-20260701084101-423612299d83
```

#### 2. docx_service.go 重写

**保留的接口签名：**
- `ParseDocument(data []byte) (*model.Document, error)` — 保持不变
- `GenerateDocument(doc *model.Document) ([]byte, error)` — 保持不变
- `GenerateMarkdown(doc *model.Document) []byte` — 保持不变

**内部实现变更：**

`ParseDocument` 流程：
1. `document.Open()` 从 `bytes.Reader` 打开 DOCX
2. 遍历 `doc.Paragraphs()`，用 `paragraphText()` 提取文本
3. 用 `isHeading()` 检测 heading 级别（通过 `para.X().PPr.PStyle.ValAttr`）
4. 构建完整 Section 树（保留现有树形逻辑）
5. 在完整 Section 树中搜索匹配关键词"投标文件"的节点
6. 提取该节点的子树作为 outline
7. 若未找到匹配节点，返回空 outline

**新增辅助函数：**
- `paragraphText(para document.Paragraph) string` — 拼接段落文本
- `isHeading(para document.Paragraph) (bool, int)` — 检测 heading 级别
- `extractSections(paras []document.Paragraph) []model.Section` — 从 unioffice 段落构建 Section 树
- `findKeywordSubtree(sections []model.Section, keyword string) []model.Section` — 搜索匹配关键词的子树
- `filterKeywordOutline(sections []model.Section, keyword string) []model.Section` — 顶层入口，提取匹配关键词的子树

**数据模型：** 不变（`Section` 树形结构）

#### 3. handler.go 变更

`UploadDocument` 处理流程不变，仍调用 `ParseDocument`。

`ParseDocument` 新增可选参数 `keyword string`（默认"投标文件"），通过 `model.Document` 中新增字段或通过初始化参数传递。

**方案选择：** `DocxService` 增加 `Keyword` 字段，通过 `NewDocxService(keyword)` 或 setter 方法设置。

### 前端

无需改动。现有 `OutlineTree.vue` 已支持：
- 递归渲染树形大纲
- 新增节点（`addTopSection` / `addChild`）
- 删除节点（`removeSection`）
- 点击选中 → 显示内容

## 数据流

```
上传 DOCX
  → handler.UploadDocument()
    → docxService.ParseDocument(buf.Bytes())
      → document.Open(bytes.NewReader(data))
      → 遍历 Paragraphs()，构建完整 Section 树
      → findKeywordSubtree(sections, "投标文件")
        → 遍历树，找到标题包含"投标文件"的节点
        → 返回该节点的 Children 作为 outline
      → 返回 *model.Document{Outline: subtree}
  → 前端接收 JSON
    → documentStore.outline 更新
      → OutlineTree 渲染
```

## 关键实现细节

### isHeading 检测

```go
func isHeading(para document.Paragraph) (bool, int) {
    props := para.X().PPr
    if props != nil && props.PStyle != nil {
        styleVal := props.PStyle.ValAttr
        // 匹配 "Heading1", "Heading 1", "heading1", "标题1" 等
        for i := 1; i <= 9; i++ {
            patterns := []string{
                fmt.Sprintf("Heading%d", i),
                fmt.Sprintf("Heading %d", i),
                fmt.Sprintf("heading%d", i),
                fmt.Sprintf("标题%d", i),
                fmt.Sprintf("标题 %d", i),
            }
            for _, p := range patterns {
                if styleVal == p {
                    return true, i
                }
            }
        }
    }
    return false, 0
}
```

### 关键词子树匹配

```go
func findKeywordSubtree(sections []model.Section, keyword string) []model.Section {
    for _, sec := range sections {
        if strings.Contains(sec.Title, keyword) {
            return sec.Children
        }
        if found := findKeywordSubtree(sec.Children, keyword); found != nil {
            return found
        }
    }
    return nil
}
```

### 段落提取

```go
func paragraphText(para document.Paragraph) string {
    var sb strings.Builder
    for _, r := range para.Runs() {
        sb.WriteString(r.Text())
    }
    return sb.String()
}
```

## 不做的事情

- 不改前端代码
- 不改数据模型（`Section` / `Document`）
- 不改现有 API 路由
- 不处理 PDF 或其他格式
- 不实现通配符匹配（延后）