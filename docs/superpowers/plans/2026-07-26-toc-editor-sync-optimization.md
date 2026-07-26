# TOC提取与编辑器联动优化 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 优化 TOC 提取与 `@eigenpal/docx-editor-vue` 的导航/同步机制，增强中文 Heading 识别

**架构：** 前端实时 DOM 扫描替代启动轮询缓存，MutationObserver 驱动编辑器→大纲同步，后端增强 Heading 模式匹配并支持关键词可配置

**技术栈：** Go/Gin + Vue3/Pinia + @eigenpal/docx-editor-vue

---

## 修改文件清单

| 文件 | 职责 |
|------|------|
| `backend/internal/model/document.go` | Document 新增 `RawBuffer` 字段 |
| `backend/internal/service/store.go` | Store 存储/读取 raw buffer |
| `backend/internal/service/docx_service.go` | 增强 isHeadingUnioffice + 新增 reparse 方法 |
| `backend/internal/service/docx_service_test.go` | 测试新增 heading 模式 |
| `backend/internal/handler/handler.go` | 新增 POST /document/:id/reparse 端点 |
| `frontend/src/stores/documentStore.ts` | 新增 syncHeadingFromEditor + headingPositions 缓存 |
| `frontend/src/api/client.ts` | 新增 reparseDocument API |
| `frontend/src/views/EditorView.vue` | 导航实时查询 + MutationObserver + 显示改进 |

---

### 任务 1：Document 模型新增 RawBuffer

**文件：**
- 修改：`backend/internal/model/document.go:14-20`

- [ ] **步骤 1：添加 RawBuffer 字段**

修改 `model/document.go`，在 Document 结构体中新增 `RawBuffer` 字段（不序列化为 JSON）：

```go
type Document struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Outline   []Section `json:"outline"`
	RawBuffer []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

- [ ] **步骤 2：编译验证**

运行：`cd /home/jingle/opc/bid-maker/backend && go build ./...`
预期：编译通过

- [ ] **步骤 3：Commit**

```bash
git add backend/internal/model/document.go
git commit -m "feat: add RawBuffer field to Document model"
```

---

### 任务 2：Store 保存 raw buffer

**文件：**
- 修改：`backend/internal/service/store.go`

- [ ] **步骤 1：读取 store.go 了解当前 store 实现**

```bash
cat /home/jingle/opc/bid-maker/backend/internal/service/store.go
```

- [ ] **步骤 2：确保 StoreDocument 和 UpdateDocument 保持 RawBuffer**

确认现有 `StoreDocument` 和 `UpdateDocument` 函数接受 `*model.Document`（已是指针），RawBuffer 随整个 struct 一起存储，无需额外代码。若 store 用 map 存 `*model.Document`，则已自动包含 RawBuffer。

- [ ] **步骤 3：Commit**

```bash
git add backend/internal/service/store.go
git commit -m "chore: RawBuffer auto-stored via Document pointer"
```

（如果 store 确实不需要修改，此步骤为空 commit）

---

### 任务 3：后端增强 Heading 检测

**文件：**
- 修改：`backend/internal/service/docx_service.go:35-56`

- [ ] **步骤 1：编写失败的测试（先确认现有测试模式）**

读取现有测试文件确认测试方式：

```bash
cat /home/jingle/opc/bid-maker/backend/internal/service/docx_service_test.go
```

- [ ] **步骤 2：增强 isHeadingUnioffice**

新增函数 `isHeadingUnioffice` 检测更多中文 heading 模式：

```go
var chineseLevelMap = map[string]int{
	"一": 1, "二": 2, "三": 3, "四": 4, "五": 5,
	"六": 6, "七": 7, "八": 8, "九": 9,
}

func isHeadingUnioffice(para document.Paragraph) (bool, int) {
	props := para.X().PPr
	if props != nil && props.PStyle != nil {
		styleVal := props.PStyle.ValAttr
		// Standard patterns (Heading1, 标题1, etc.)
		for i := 1; i <= 9; i++ {
			patterns := []string{
				fmt.Sprintf("Heading%d", i), fmt.Sprintf("Heading %d", i),
				fmt.Sprintf("heading%d", i), fmt.Sprintf("heading %d", i),
				fmt.Sprintf("标题%d", i), fmt.Sprintf("标题 %d", i),
				fmt.Sprintf("一级标题", i), // only level 1-4 are common but keep all
			}
			if i <= 4 {
				patterns = append(patterns, fmt.Sprintf("第%s章", chineseNum(i)))
			}
			if i <= 2 {
				patterns = append(patterns, fmt.Sprintf("第%s节", chineseNum(i)))
			}
			for _, p := range patterns {
				if styleVal == p {
					return true, i
				}
			}
		}
		// Chinese level patterns: 一级标题, 二级标题...
		for level := 1; level <= 9; level++ {
			levelStr := chineseNum(level)
			if styleVal == levelStr+"级标题" {
				return true, level
			}
			if level <= 4 {
				if styleVal == "第"+levelStr+"章" {
					return true, level
				}
			}
			if level <= 2 {
				if styleVal == "第"+levelStr+"节" {
					return true, level
				}
			}
		}
		// Check outlineLvl attribute
		if props.OutlineLvl != nil {
			// outlineLvl="0" -> level 1
			val := props.OutlineLvl.ValAttr
			if val >= 0 && val <= 8 {
				return true, val + 1
			}
		}
	}

	// Fallback: detect numbered headings by text content via regex
	text := paragraphText(para)
	text = strings.TrimSpace(text)
	if text == "" {
		return false, 0
	}
	// Pattern: ^[1-9]、 or ^[1-9]\. or ^第[一二三四五六七八九十]章
	if matched, _ := regexp.MatchString(`^第[一二三四五六七八九十]章`, text); matched {
		return true, 1
	}
	if matched, _ := regexp.MatchString(`^第[一二三四五六七八九十]节`, text); matched {
		return true, 2
	}
	if matched, _ := regexp.MatchString(`^\d+[、\.]`, text); matched {
		// Level based on digit count: 1 -> 1, 1.1 -> 2, 1.1.1 -> 3
		digits := strings.Count(text[:strings.IndexAny(text, "、.")], ".")
		return true, digits + 1
	}
	if matched, _ := regexp.MatchString(`^[一二三四五六七八九十]+[、]`, text); matched {
		return true, 1
	}

	return false, 0
}

func chineseNum(n int) string {
	nums := []string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九"}
	if n >= 1 && n <= 9 {
		return nums[n]
	}
	return ""
}
```

在文件头部添加 `import "regexp"`。

- [ ] **步骤 3：编译验证**

```bash
cd /home/jingle/opc/bid-maker/backend && go build ./...
```

- [ ] **步骤 4：Commit**

```bash
git add backend/internal/service/docx_service.go
git commit -m "feat: enhance Chinese heading detection with outlineLvl and numbered patterns"
```

---

### 任务 4：后端 reparse 方法 + 上传关键词参数

**文件：**
- 修改：`backend/internal/service/docx_service.go`
- 修改：`backend/internal/handler/handler.go`

- [ ] **步骤 1：DocxService 新增 Reparse 方法**

在 `docx_service.go` 新增：

```go
func (s *DocxService) Reparse(doc *model.Document, keyword string) *model.Document {
	if len(doc.RawBuffer) == 0 {
		return doc
	}
	if keyword != "" {
		s.Keyword = keyword
	}
	parsed, err := s.ParseDocument(doc.RawBuffer)
	if err != nil {
		return doc
	}
	doc.Outline = parsed.Outline
	doc.Title = parsed.Title
	doc.UpdatedAt = time.Now().UTC()
	return doc
}
```

同时修改 `UploadDocument` handler 从 query 读取 keyword：

在 handler.go 的 UploadDocument 函数中，从 `c.Query("keyword")` 读取可选关键词，若不为空则设置到 docxService.Keyword。

- [ ] **步骤 2：handler 新增 reparse 端点**

在 `handler.go` 的 `RegisterRoutes` 中 doc group 下新增路由：

```go
doc.POST("/:id/reparse", h.ReparseDocument)
```

新增 handler 方法：

```go
func (h *Handler) ReparseDocument(c *gin.Context) {
	id := c.Param("id")
	doc, ok := service.GetDocument(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	var req struct {
		Keyword string `json:"keyword"`
	}
	c.ShouldBindJSON(&req)
	h.docxService.Reparse(doc, req.Keyword)
	service.UpdateDocument(doc)
	c.JSON(http.StatusOK, gin.H{"outline": doc.Outline})
}
```

- [ ] **步骤 3：编译验证**

```bash
cd /home/jingle/opc/bid-maker/backend && go build ./...
```

- [ ] **步骤 4：Commit**

```bash
git add backend/internal/service/docx_service.go backend/internal/handler/handler.go
git commit -m "feat: add reparse endpoint and upload keyword parameter"
```

---

### 任务 5：后端测试增强的 Heading 检测

**文件：**
- 修改：`backend/internal/service/docx_service_test.go`

- [ ] **步骤 1：编写测试**

```go
package service

import (
	"testing"
)

func TestChineseHeadingLevelMap(t *testing.T) {
	tests := []struct {
		name     string
		num      int
		expected string
	}{
		{"1->一", 1, "一"},
		{"9->九", 9, "九"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := chineseNum(tc.num); got != tc.expected {
				t.Errorf("chineseNum(%d) = %q, want %q", tc.num, got, tc.expected)
			}
		})
	}
}
```

（注意：完整的 heading 测试依赖 DOCX 文件，单元测试无法覆盖完整流程。此测试验证辅助函数。）

- [ ] **步骤 2：运行测试**

```bash
cd /home/jingle/opc/bid-maker/backend && go test ./internal/service/ -run TestChineseHeadingLevelMap -v
```

预期：PASS

- [ ] **步骤 3：Commit**

```bash
git add backend/internal/service/docx_service_test.go
git commit -m "test: add Chinese heading helper tests"
```

---

### 任务 6：前端 API 客户端新增 reparse

**文件：**
- 修改：`frontend/src/api/client.ts`

- [ ] **步骤 1：新增 reparseDocument**

```typescript
export const reparseDocument = (docId: string, keyword?: string) =>
  api.post(`/document/${docId}/reparse`, { keyword })
```

- [ ] **步骤 2：验证 TypeScript 编译**

```bash
cd /home/jingle/opc/bid-maker/frontend && npx tsc --noEmit 2>&1 | head -20
```

预期：无类型错误

- [ ] **步骤 3：Commit**

```bash
git add frontend/src/api/client.ts
git commit -m "feat: add reparseDocument API"
```

---

### 任务 7：前端 documentStore 新增同步方法 + 缓存

**文件：**
- 修改：`frontend/src/stores/documentStore.ts:61-97`

- [ ] **步骤 1：新增 headingPositions 缓存和 syncHeadingFromEditor 方法**

```typescript
// headingPositions cache: sectionId -> { pmPos, timestamp }
const headingPositions = ref<Map<string, { pmPos: number; timestamp: number }>>(new Map())
const CACHE_TTL = 3000 // 3 seconds

const getSectionPmPos = (sectionId: string): number | undefined => {
  // Check cache first
  const cached = headingPositions.value.get(sectionId)
  if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
    return cached.pmPos
  }
  // Fallback to outline tree (legacy after sync)
  const tree = outline.value
  for (let i = 0; i < tree.length; i++) {
    if (tree[i].id === sectionId) return tree[i].pmPos
    if (tree[i].children?.length) {
      for (const child of tree[i].children) {
        if (child.id === sectionId) return child.pmPos
      }
    }
  }
  return undefined
}

const setHeadingPosition = (sectionId: string, pmPos: number) => {
  headingPositions.value.set(sectionId, { pmPos, timestamp: Date.now() })
}

const clearHeadingPositions = () => {
  headingPositions.value.clear()
}

const syncHeadingFromEditor = (headings: Array<{ text: string; pmPos: number }>) => {
  const tree = outline.value
  for (const h of headings) {
    for (let i = 0; i < tree.length; i++) {
      if (tree[i].title === h.text || h.text.includes(tree[i].title)) {
        tree[i].pmPos = h.pmPos
        headingPositions.value.set(tree[i].id, { pmPos: h.pmPos, timestamp: Date.now() })
        break
      }
      if (tree[i].children?.length) {
        for (const child of tree[i].children) {
          if (child.title === h.text || h.text.includes(child.title)) {
            child.pmPos = h.pmPos
            headingPositions.value.set(child.id, { pmPos: h.pmPos, timestamp: Date.now() })
            break
          }
        }
      }
    }
  }
}

const syncTitleFromEditor = (pmPos: number, newTitle: string) => {
  const tree = outline.value
  for (let i = 0; i < tree.length; i++) {
    if (tree[i].pmPos === pmPos) {
      tree[i].title = newTitle
      return
    }
    if (tree[i].children?.length) {
      for (const child of tree[i].children) {
        if (child.pmPos === pmPos) {
          child.title = newTitle
          return
        }
      }
    }
  }
}
```

Return 中添加：`headingPositions, getSectionPmPos, setHeadingPosition, clearHeadingPositions, syncHeadingFromEditor, syncTitleFromEditor`

- [ ] **步骤 2：删除旧的 getSectionPmPos（由新版替代）**

确认旧 `getSectionPmPos` 被新版本替换。

- [ ] **步骤 3：验证 TypeScript**

```bash
cd /home/jingle/opc/bid-maker/frontend && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **步骤 4：Commit**

```bash
git add frontend/src/stores/documentStore.ts
git commit -m "feat: add heading position cache and editor sync methods"
```

---

### 任务 8：前端 EditorView 重构导航

**文件：**
- 修改：`frontend/src/views/EditorView.vue:112-194`

- [ ] **步骤 1：替换 handleEditorReady + handleSelectSection**

```typescript
// Reactive heading map for quick lookup
let editorViewport: HTMLElement | null = null
let headingScanAttempts = 0

function handleEditorReady() {
  editorViewport = document.querySelector('.docx-editor-vue__pages-viewport') as HTMLElement | null
  if (!editorViewport) {
    // Retry a few times if viewport not ready
    if (headingScanAttempts < 20) {
      headingScanAttempts++
      setTimeout(handleEditorReady, 200)
    }
    return
  }
  // Initial sync of all heading positions
  scanAndSyncHeadings()
  // Setup MutationObserver for ongoing sync
  setupHeadingObserver()
}

function scanHeadings(): Array<{ text: string; pmPos: number }> {
  if (!editorViewport) return []
  const elements = editorViewport.querySelectorAll('[data-pm-start][data-pm-end]')
  const results: Array<{ text: string; pmPos: number }> = []
  elements.forEach((el) => {
    const text = el.textContent?.trim()
    if (!text) return
    const pmPos = Number(el.getAttribute('data-pm-start')) || 0
    if (pmPos) results.push({ text, pmPos })
  })
  return results
}

function scanAndSyncHeadings() {
  const headings = scanHeadings()
  if (headings.length > 0) {
    docStore.syncHeadingFromEditor(headings)
  }
}

function findPmPosForSection(sectionId: string): number | undefined {
  // Check cached positions
  const cached = docStore.headingPositions.get(sectionId)
  if (cached && Date.now() - cached.timestamp < 3000) {
    return cached.pmPos
  }
  // Fallback: real-time scan by title
  const section = findSectionInTree(docStore.outline, sectionId)
  if (!section || !section.title) return undefined
  const headings = scanHeadings()
  for (const h of headings) {
    if (h.text === section.title || h.text.includes(section.title)) {
      docStore.setHeadingPosition(sectionId, h.pmPos)
      return h.pmPos
    }
  }
  return undefined
}

function findSectionInTree(sections: Section[], id: string): Section | null {
  for (const s of sections) {
    if (s.id === id) return s
    if (s.children) {
      const found = findSectionInTree(s.children, id)
      if (found) return found
    }
  }
  return null
}

const handleSelectSection = (sectionId: string) => {
  const pmPos = findPmPosForSection(sectionId)
  if (pmPos !== undefined && editorRef.value?.scrollToPosition) {
    editorRef.value.scrollToPosition(pmPos)
  }
  docStore.loadSection(props.id, sectionId)
}

let headingObserver: MutationObserver | null = null
let syncDebounceTimer: ReturnType<typeof setTimeout> | null = null

function setupHeadingObserver() {
  if (!editorViewport || headingObserver) return
  headingObserver = new MutationObserver(() => {
    // Clear heading position cache
    docStore.clearHeadingPositions()
    // Debounced sync
    if (syncDebounceTimer) clearTimeout(syncDebounceTimer)
    syncDebounceTimer = setTimeout(() => {
      const headings = scanHeadings()
      if (headings.length > 0) {
        // Sync titles: find changed headings and update outline titles
        headings.forEach(h => {
          docStore.syncTitleFromEditor(h.pmPos, h.text)
        })
      }
    }, 300)
  })
  headingObserver.observe(editorViewport, {
    childList: true,
    subtree: true,
    characterData: true,
  })
}

onUnmounted(() => {
  if (headingObserver) {
    headingObserver.disconnect()
    headingObserver = null
  }
  if (syncDebounceTimer) {
    clearTimeout(syncDebounceTimer)
  }
})
```

同时更新 `<script>` 的 import 添加 `onUnmounted`（已在现有代码中）。

- [ ] **步骤 2：验证 TypeScript 编译**

```bash
cd /home/jingle/opc/bid-maker/frontend && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **步骤 3：Commit**

```bash
git add frontend/src/views/EditorView.vue
git commit -m "feat: real-time heading scanning and MutationObserver sync"
```

---

### 任务 9：前端内容显示改进

**文件：**
- 修改：`frontend/src/views/EditorView.vue`

- [ ] **步骤 1：隐藏编辑器自带大纲 + 骨架屏**

```diff
- :show-outline="true"
+ :show-outline="false"
```

在 `<DocxEditor>` 同一 section 内添加骨架屏：

```html
<section class="center-panel">
  <div v-if="!editorReady" class="editor-skeleton">
    <div class="skeleton-toolbar" />
    <div class="skeleton-content">
      <div class="skeleton-line" style="width: 60%" />
      <div class="skeleton-line" style="width: 80%" />
      <div class="skeleton-line" style="width: 40%" />
      <div class="skeleton-line" style="width: 70%" />
    </div>
  </div>
  <DocxEditor
    v-if="docStore.docxBuffer"
    ref="editorRef"
    :document-buffer="docStore.docxBuffer"
    :show-menu-bar="true"
    :show-toolbar="true"
    :show-outline="false"
    :read-only="false"
    @ready="handleEditorReady"
  />
</section>
```

在 script setup 中添加 `const editorReady = ref(false)`，在 `handleEditorReady` 末尾设置 `editorReady.value = true`。

添加骨架屏样式：

```css
.editor-skeleton {
  padding: 40px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.skeleton-toolbar {
  height: 40px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 8px;
}

.skeleton-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.skeleton-line {
  height: 16px;
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 4px;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
```

- [ ] **步骤 2：验证 TypeScript 编译**

```bash
cd /home/jingle/opc/bid-maker/frontend && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **步骤 3：Commit**

```bash
git add frontend/src/views/EditorView.vue
git commit -m "feat: hide editor outline panel and add loading skeleton"
```

---

## 验证

全部任务完成后：

```bash
cd /home/jingle/opc/bid-maker/backend && go build ./... && go test ./...
cd /home/jingle/opc/bid-maker/frontend && npx tsc --noEmit && npx vite build 2>&1 | tail -5
```
