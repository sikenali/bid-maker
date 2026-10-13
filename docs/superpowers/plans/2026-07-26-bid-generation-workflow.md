# Bid Generation Workflow 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 实现标书生成工作流——用户在 AIChat 中选"生成标书"skill → AI 生成 Markdown 大纲 → 弹框确认 → 左侧大纲树显示结构 → 用户逐章点击生成内容（SSE 流式写入编辑器）。

**架构：** 后端新增 2 个 API（`/api/generate-outline` + `/api/generate-section SSE`），前端新增 generateStore + 弹框/进度组件，AIChat 检测 generate 型 skill 送入生成流程而非普通聊天。

**技术栈：** Go/Gin + go-openai + SSE, Vue3/Pinia/TypeScript, @eigenpal/docx-editor-vue

---

## 文件结构

| 文件 | 操作 | 职责 |
|------|------|------|
| `backend/internal/service/generate_service.go` | 创建 | AI 大纲生成 + MD 解析 + 章节内容流式生成 |
| `backend/internal/handler/generate_handler.go` | 创建 | GenerateOutline / GenerateSection HTTP 处理 |
| `backend/internal/handler/handler.go` | 修改 | 注册新路由 |
| `frontend/src/api/client.ts` | 修改 | 新增 `generateOutline` + `generateSectionStream` |
| `frontend/src/stores/generateStore.ts` | 创建 | Pinia store：phase/outline/sectionStates |
| `frontend/src/stores/settingsStore.ts` | 修改 | Skill 接口增加 `type` 字段，添加"生成标书"skill |
| `frontend/src/components/GenerateFlowDialog.vue` | 创建 | 大纲预览 + 确认弹框 |
| `frontend/src/components/ProgressTracker.vue` | 创建 | 生成进度面板 |
| `frontend/src/components/AIChat.vue` | 修改 | detect generate skill → 触发生成流程 |
| `frontend/src/components/OutlineTree.vue` | 修改 | 显示章节状态 pending/generating/done/error |
| `frontend/src/views/EditorView.vue` | 修改 | 集成 GenerateFlowDialog + ProgressTracker |

---

### 任务 1：Backend — Generate Service

**文件：** 创建 `backend/internal/service/generate_service.go`

- [ ] **步骤 1：编写 MD 解析 Section 测试**

```go
// backend/internal/service/generate_service_test.go
package service

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParseOutlineFromMD(t *testing.T) {
	md := `# 第一章 项目概况
## 1.1 项目背景
## 1.2 建设目标
### 1.2.1 总体目标
# 第二章 技术方案`
	sections := ParseOutlineFromMD(md)
	assert.Len(t, sections, 2)
	assert.Equal(t, "第一章 项目概况", sections[0].Title)
	assert.Equal(t, 1, sections[0].Level)
	assert.Len(t, sections[0].Children, 2)
	assert.Equal(t, "1.1 项目背景", sections[0].Children[0].Title)
	assert.Equal(t, "1.2 建设目标", sections[0].Children[1].Title)
	assert.Len(t, sections[0].Children[1].Children, 1)
	assert.Equal(t, "总体目标", sections[1].Title) // "第二章 技术方案" has no children
}

func TestParseOutlineFromMD_Empty(t *testing.T) {
	sections := ParseOutlineFromMD("")
	assert.Len(t, sections, 0)
}

func TestParseOutlineFromMD_NoHeadings(t *testing.T) {
	sections := ParseOutlineFromMD("纯文本内容\n没有标题\n")
	assert.Len(t, sections, 0)
}
```

运行：`cd backend && go test ./internal/service/ -run TestParseOutlineFromMD -v`
预期：FAIL（函数未定义）

- [ ] **步骤 2：实现 MD 解析函数**

```go
// backend/internal/service/generate_service.go
package service

import (
	"regexp"
	"strings"

	"github.com/example/bid-maker-backend/internal/model"
	"github.com/google/uuid"
)

var reHeading = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

func ParseOutlineFromMD(md string) []model.Section {
	lines := strings.Split(md, "\n")
	var stack []model.Section
	var root []model.Section

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		matches := reHeading.FindStringSubmatch(trimmed)
		if matches == nil {
			continue
		}
		level := len(matches[1])
		title := strings.TrimSpace(matches[2])

		sec := model.Section{
			ID:    uuid.NewString(),
			Title: title,
			Level: level,
		}

		// Pop stack until parent found
		for len(stack) > 0 && stack[len(stack)-1].Level >= level {
			stack = stack[:len(stack)-1]
		}
		if len(stack) > 0 {
			parent := &stack[len(stack)-1]
			sec.ParentID = parent.ID
			parent.Children = append(parent.Children, sec)
		} else {
			root = append(root, sec)
		}
		stack = append(stack, sec)
	}
	return root
}
```

- [ ] **步骤 3：运行测试验证通过**

运行：`cd backend && go test ./internal/service/ -run TestParseOutlineFromMD -v`
预期：PASS

- [ ] **步骤 4：实现 GenerateOutline 方法**

```go
// backend/internal/service/generate_service.go — add to file

type GenerateService struct {
	LLM *LLMRegistry
}

func NewGenerateService(llm *LLMRegistry) *GenerateService {
	return &GenerateService{LLM: llm}
}

type GenerateOutlineRequest struct {
	Message     string `json:"message"`
	SkillPrompt string `json:"skill_prompt"`
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	Endpoint    string `json:"endpoint"`
	Format      string `json:"format"`
	APIKey      string `json:"apiKey"`
}

func (s *GenerateService) GenerateOutline(req GenerateOutlineRequest) ([]model.Section, error) {
	// Build system prompt: skill prompt + instruction to return markdown outline
	var sb strings.Builder
	sb.WriteString(req.SkillPrompt)
	sb.WriteString("\n\n根据以上要求生成标书大纲。")
	sb.WriteString("以 Markdown 格式输出，使用 # 表示章，## 表示节，### 表示小节。")
	sb.WriteString("只输出大纲，不需要其他说明文字。")

	systemPrompt := sb.String()

	// Call AI (non-streaming)
	// Check if custom endpoint
	var reply string
	var err error

	// Reuse chat service logic
	if req.Endpoint != "" {
		// Custom endpoint
		messages := []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": req.Message},
		}
		reply, err = chatWithCustomEndpoint(req.Endpoint, req.Format, req.APIKey, req.Model, messages, 120)
	} else {
		// Use registered provider
		client, ok := s.LLM.GetProvider(req.Provider)
		if !ok {
			return nil, fmt.Errorf("provider %s not found", req.Provider)
		}
		messages := []openai.ChatCompletionMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: req.Message},
		}
		resp, err := client.Chat(ctx, messages, req.Model)
		if err != nil {
			return nil, err
		}
		reply = resp
	}
	if err != nil {
		return nil, err
	}

	return ParseOutlineFromMD(reply), nil
}
```

Note: add `"fmt"` and `"github.com/google/uuid"` to imports.
Note: `chatWithCustomEndpoint` is in `chat_service.go` — needs to be exported or moved to a shared util.

- [ ] **步骤 5：实现 GenerateSection 流式方法**

```go
// backend/internal/service/generate_service.go — add

type GenerateSectionRequest struct {
	DocumentID     string `json:"document_id"`
	SectionID      string `json:"section_id"`
	SectionTitle   string `json:"section_title"`
	SectionPath    []string `json:"section_path"`
	OutlineContext string `json:"outline_context"`
	Provider       string `json:"provider"`
	Model          string `json:"model"`
	Endpoint       string `json:"endpoint"`
	Format         string `json:"format"`
	APIKey         string `json:"apiKey"`
}

type SectionChunk struct {
	SectionID string `json:"section_id"`
	Chunk     string `json:"chunk,omitempty"`
	Done      bool   `json:"done,omitempty"`
	Error     string `json:"error,omitempty"`
}

func (s *GenerateService) GenerateSectionStream(req GenerateSectionRequest, w http.ResponseWriter) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	pathStr := strings.Join(req.SectionPath, " > ")
	systemPrompt := fmt.Sprintf(
		`你是标书撰写专家。当前正在编写章节: %s（ID: %s）

完整大纲上下文:
%s

要求:
- 根据大纲上下文，为当前章节生成完整标书内容
- 使用正式商务文档风格
- 包含完整段落、必要的数据和论证
- 直接输出内容，不需要章节标题`,
		pathStr, req.SectionID, req.OutlineContext,
	)

	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("请编写章节「%s」的内容", req.SectionTitle)},
	}

	client, ok := s.LLM.GetProvider(req.Provider)
	var stream *openai.ChatCompletionStream
	if !ok && req.Endpoint != "" {
		// Use custom endpoint streaming (simplified — use go-openai compatible)
		err := fmt.Errorf("custom endpoint streaming not yet implemented")
		writeSSEError(w, flusher, req.SectionID, err.Error())
		return
	} else if !ok {
		writeSSEError(w, flusher, req.SectionID, fmt.Sprintf("provider %s not found", req.Provider))
		return
	}

	// Streaming via go-openai
	var err error
	stream, err = client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		writeSSEError(w, flusher, req.SectionID, err.Error())
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeSSEError(w, flusher, req.SectionID, err.Error())
			return
		}
		chunk := response.Choices[0].Delta.Content
		if chunk == "" {
			continue
		}
		writeSSEChunk(w, flusher, req.SectionID, chunk)
	}
	writeSSEDone(w, flusher, req.SectionID)
}

func writeSSEChunk(w http.ResponseWriter, flusher http.Flusher, sectionID, chunk string) {
	data, _ := json.Marshal(SectionChunk{SectionID: sectionID, Chunk: chunk})
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

func writeSSEError(w http.ResponseWriter, flusher http.Flusher, sectionID, errMsg string) {
	data, _ := json.Marshal(SectionChunk{SectionID: sectionID, Error: errMsg})
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

func writeSSEDone(w http.ResponseWriter, flusher http.Flusher, sectionID string) {
	data, _ := json.Marshal(SectionChunk{SectionID: sectionID, Done: true})
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}
```

Note: Need to add streaming support to `LLMClient` interface or `OpenAIProvider`. Requires:
```go
// In llm_service.go — add to LLMClient interface
CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error)

// Implement in OpenAIProvider
func (p *OpenAIProvider) CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error) {
	return p.client.CreateChatCompletionStream(ctx, req)
}
```

- [ ] **步骤 6：运行后端编译检查**

运行：`cd backend && go build ./...`
预期：PASS

- [ ] **步骤 7：Commit**

```bash
git add backend/internal/service/generate_service.go backend/internal/service/generate_service_test.go
git commit -m "feat: add generate service with MD parsing and SSE streaming"
```

---

### 任务 2：Backend — Generate Handler + Routes

**文件：** 创建 `backend/internal/handler/generate_handler.go`，修改 `backend/internal/handler/handler.go`

- [ ] **步骤 1：实现 GenerateOutline 和 GenerateSection HTTP handlers**

```go
// backend/internal/handler/generate_handler.go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/example/bid-maker-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type GenerateHandler struct {
	svc *service.GenerateService
}

func NewGenerateHandler(svc *service.GenerateService) *GenerateHandler {
	return &GenerateHandler{svc: svc}
}

func (h *GenerateHandler) GenerateOutline(c *gin.Context) {
	var req service.GenerateOutlineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	outline, err := h.svc.GenerateOutline(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"outline": outline})
}

func (h *GenerateHandler) GenerateSection(c *gin.Context) {
	var req service.GenerateSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.svc.GenerateSectionStream(req, c.Writer)
}
```

- [ ] **步骤 2：注册路由**

在 `backend/internal/handler/handler.go` 的 `RegisterRoutes` 中：

```go
// Add to the imports at top
// generateHandler := NewGenerateHandler(s.generateService)
// And in the RouteGroup:
// Register in RegisterRoutes method
api.POST("/generate-outline", generateHandler.GenerateOutline)
api.POST("/generate-section", generateHandler.GenerateSection)
```

在 `Handler` struct 中添加 `generateService`：

```go
type Handler struct {
	docxService   *service.DocxService
	chatService   *service.ChatService
	llmRegistry   *service.LLMRegistry
	skillService  *service.SkillService
	generateService *service.GenerateService
}
```

在 NewHandler/构造函数中初始化 `generateService`。

- [ ] **步骤 3：运行后端编译检查**

运行：`cd backend && go build ./...`
预期：PASS

- [ ] **步骤 4：Commit**

```bash
git add backend/internal/handler/generate_handler.go backend/internal/handler/handler.go
git commit -m "feat: add generate outline and section handlers"
```

---

### 任务 3：Frontend — API Client + GenerateStore

**文件：** 修改 `frontend/src/api/client.ts`，创建 `frontend/src/stores/generateStore.ts`，修改 `frontend/src/stores/settingsStore.ts`

- [ ] **步骤 1：API client 新增方法**

```typescript
// frontend/src/api/client.ts — add

export async function generateOutline(data: {
  message: string
  skill_prompt: string
  provider: string
  model: string
  endpoint: string
  format: string
  apiKey: string
}): Promise<{ outline: Section[] }> {
  const res = await api.post('/generate-outline', data)
  return res.data
}

export async function generateSectionStream(
  data: {
    document_id: string
    section_id: string
    section_title: string
    section_path: string[]
    outline_context: string
    provider: string
    model: string
    endpoint: string
    format: string
    apiKey: string
  },
  onChunk: (sectionId: string, chunk: string) => void,
  onDone: (sectionId: string) => void,
  onError: (sectionId: string, error: string) => void
): Promise<void> {
  const response = await fetch('/api/generate-section', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!response.ok) {
    throw new Error(`generate-section failed: ${response.statusText}`)
  }
  const reader = response.body!.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''
    for (const line of lines) {
      const trimmed = line.trim()
      if (!trimmed.startsWith('data: ')) continue
      try {
        const parsed = JSON.parse(trimmed.slice(6))
        if (parsed.error) {
          onError(parsed.section_id, parsed.error)
        } else if (parsed.done) {
          onDone(parsed.section_id)
        } else if (parsed.chunk) {
          onChunk(parsed.section_id, parsed.chunk)
        }
      } catch { /* skip malformed */ }
    }
  }
}
```

- [ ] **步骤 2：创建 generateStore**

```typescript
// frontend/src/stores/generateStore.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { generateOutline, generateSectionStream } from '../api/client'
import type { Section } from '../api/client'

export type GenerationPhase = 'idle' | 'preview' | 'generating' | 'done' | 'error'
export type SectionState = 'pending' | 'generating' | 'done' | 'error'

interface ModelConfig {
  provider: string
  model: string
  endpoint: string
  format: string
  apiKey: string
}

export const useGenerateStore = defineStore('generate', () => {
  const phase = ref<GenerationPhase>('idle')
  const outline = ref<Section[]>([])
  const sectionStates = ref<Map<string, SectionState>>(new Map())
  const currentSectionId = ref<string | null>(null)
  const error = ref<string | null>(null)
  const modelConfig = ref<ModelConfig | null>(null)
  const docId = ref<string>('')
  const userMessage = ref<string>('')

  // Computed
  const totalSections = computed(() => {
    let count = 0
    const countRecursive = (secs: Section[]) => {
      for (const s of secs) {
        count++
        countRecursive(s.children)
      }
    }
    countRecursive(outline.value)
    return count
  })

  const completedSections = computed(() => {
    let count = 0
    sectionStates.value.forEach(s => { if (s === 'done') count++ })
    return count
  })

  const progressPercent = computed(() => {
    if (totalSections.value === 0) return 0
    return Math.round((completedSections.value / totalSections.value) * 100)
  })

  // Actions
  async function generateOutlineAction(
    id: string,
    message: string,
    skillPrompt: string,
    config: ModelConfig
  ) {
    docId.value = id
    userMessage.value = message
    modelConfig.value = config
    error.value = null

    try {
      const result = await generateOutline({
        message,
        skill_prompt: skillPrompt,
        ...config,
      })
      outline.value = result.outline
      phase.value = 'preview'
    } catch (err: any) {
      error.value = err.message || '生成大纲失败'
      phase.value = 'error'
    }
  }

  function confirmGeneration() {
    phase.value = 'generating'
    // Initialize all sections as pending
    const states = new Map<string, SectionState>()
    const initRecursive = (secs: Section[]) => {
      for (const s of secs) {
        states.set(s.id, 'pending')
        initRecursive(s.children)
      }
    }
    initRecursive(outline.value)
    sectionStates.value = states
  }

  async function generateSection(sectionId: string, sectionTitle: string, sectionPath: string[], outlineContext: string) {
    if (!modelConfig.value) return
    sectionStates.value.set(sectionId, 'generating')
    currentSectionId.value = sectionId

    try {
      await generateSectionStream(
        {
          document_id: docId.value,
          section_id: sectionId,
          section_title: sectionTitle,
          section_path: sectionPath,
          outline_context: outlineContext,
          ...modelConfig.value,
        },
        (sid, chunk) => {
          // Emit chunk event — EditorView will listen
          window.dispatchEvent(new CustomEvent('gen-chunk', { detail: { sectionId: sid, chunk } }))
        },
        (sid) => {
          sectionStates.value.set(sid, 'done')
          currentSectionId.value = null
          window.dispatchEvent(new CustomEvent('gen-done', { detail: { sectionId: sid } }))
          // Check if all done
          let allDone = true
          sectionStates.value.forEach(s => { if (s !== 'done') allDone = false })
          if (allDone) phase.value = 'done'
        },
        (sid, err) => {
          sectionStates.value.set(sid, 'error')
          currentSectionId.value = null
          error.value = err
        }
      )
    } catch (err: any) {
      sectionStates.value.set(sid, 'error')
      error.value = err.message || '生成失败'
    }
  }

  function retrySection(sectionId: string) {
    sectionStates.value.set(sectionId, 'pending')
    // Re-trigger generation
  }

  function reset() {
    phase.value = 'idle'
    outline.value = []
    sectionStates.value = new Map()
    currentSectionId.value = null
    error.value = null
    modelConfig.value = null
    docId.value = ''
    userMessage.value = ''
  }

  function getSectionState(sectionId: string): SectionState {
    return sectionStates.value.get(sectionId) || 'pending'
  }

  return {
    phase, outline, sectionStates, currentSectionId, error,
    modelConfig, docId, userMessage,
    totalSections, completedSections, progressPercent,
    generateOutline: generateOutlineAction,
    confirmGeneration, generateSection, retrySection, reset, getSectionState,
  }
})
```

- [ ] **步骤 3：修改 Skill 接口 + settingsStore 添加"生成标书"skill**

```typescript
// frontend/src/stores/settingsStore.ts — modify Skill interface
export interface Skill {
  id: string
  name: string
  description: string
  prompt: string
  type?: 'chat' | 'generate'  // ← add this
}

// In built-in skills, add "生成标书" entry
// Find where builtInSkills are defined (search for "outline" / "expand" / "summarize" / "format")
// and add:
{
  id: 'generate-bid',
  name: '生成标书',
  description: '根据需求自动生成标书大纲和内容',
  type: 'generate',
  prompt: `你是资深的投标文件撰写专家，精通各类政府采购、工程招标、信息化服务标书的编写规范。

根据用户提供的需求描述，生成一份完整的标书大纲。

输出格式要求：
- 使用 Markdown 标题语法（# ## ###）表示章节层级
- # 表示章（如"第一章 项目概况"）
- ## 表示节（如"1.1 项目背景"）
- ### 表示小节（如"1.1.1 项目背景分析"）

只输出大纲，不要额外说明。`,
}
```

- [ ] **步骤 4：Commit**

```bash
git add frontend/src/api/client.ts frontend/src/stores/generateStore.ts frontend/src/stores/settingsStore.ts
git commit -m "feat: add generate API client, store, and bid skill type"
```

---

### 任务 4：Frontend — GenerateFlowDialog + ProgressTracker

**文件：** 创建 `frontend/src/components/GenerateFlowDialog.vue`，创建 `frontend/src/components/ProgressTracker.vue`

- [ ] **步骤 1：实现 GenerateFlowDialog.vue**

```vue
<template>
  <div class="dialog-overlay" @click.self="$emit('cancel')">
    <div class="dialog">
      <div class="dialog-header">
        <RiFileList3Line size="20" />
        <span>标书生成预览</span>
      </div>
      <div class="dialog-body">
        <p class="dialog-query">"{{ genStore.userMessage }}"</p>
        <div class="outline-preview">
          <div v-for="sec in genStore.outline" :key="sec.id" class="preview-node">
            <PreviewNode :section="sec" :depth="0" />
          </div>
        </div>
      </div>
      <div class="dialog-footer">
        <button class="btn-cancel" @click="$emit('cancel')">取消</button>
        <button class="btn-confirm" @click="$emit('confirm')">确认生成</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useGenerateStore } from '../stores/generateStore'
import { RiFileList3Line } from '@remixicon/vue'
import PreviewNode from './PreviewNode.vue'

defineEmits<{ confirm: []; cancel: [] }>()
const genStore = useGenerateStore()
</script>

<style scoped>
.dialog-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.4);
  display: flex; align-items: center; justify-content: center; z-index: 1000;
}
.dialog {
  background: #fff; border-radius: 16px; width: 640px; max-height: 80vh;
  display: flex; flex-direction: column; box-shadow: 0 8px 32px rgba(0,0,0,0.15);
}
.dialog-header {
  display: flex; align-items: center; gap: 10px; padding: 20px 24px 0;
  font-size: 18px; font-weight: 600; color: #3D2B1F;
}
.dialog-body { flex: 1; overflow-y: auto; padding: 16px 24px; }
.dialog-query {
  font-size: 14px; color: #8B7355; padding: 12px; background: #FBF7EF;
  border-radius: 8px; margin-bottom: 16px;
}
.outline-preview { font-size: 14px; color: #3D2B1F; }
.dialog-footer {
  display: flex; justify-content: flex-end; gap: 12px; padding: 16px 24px;
  border-top: 1px solid #E0D5C0;
}
.btn-cancel {
  padding: 8px 24px; border-radius: 8px; border: 1px solid #E0D5C0;
  background: #fff; color: #5C4033; cursor: pointer; font-size: 14px;
}
.btn-confirm {
  padding: 8px 24px; border-radius: 8px; border: none;
  background: #C23B22; color: #fff; cursor: pointer; font-size: 14px;
}
.btn-confirm:hover { background: #A8321D; }
</style>
```

Create `PreviewNode.vue` as a simple recursive tree:
```vue
<template>
  <div class="preview-node" :style="{ paddingLeft: depth * 20 + 'px' }">
    <div class="preview-title">
      <RiFileLine size="16" />
      <span>{{ section.title }}</span>
    </div>
    <PreviewNode v-for="child in section.children" :key="child.id"
      :section="child" :depth="depth + 1" />
  </div>
</template>
<script setup lang="ts">
import { RiFileLine } from '@remixicon/vue'
defineProps<{ section: any; depth: number }>()
</script>
<style scoped>
.preview-node { margin: 4px 0; }
.preview-title { display: flex; align-items: center; gap: 8px; padding: 6px 8px; border-radius: 6px; }
.preview-title:hover { background: #F5EFE0; }
</style>
```

- [ ] **步骤 2：实现 ProgressTracker.vue**

```vue
<template>
  <div class="progress-tracker">
    <div class="progress-header">
      <span class="progress-title">生成进度</span>
      <span class="progress-count">{{ genStore.completedSections }}/{{ genStore.totalSections }}</span>
    </div>
    <div class="progress-bar">
      <div class="progress-fill" :style="{ width: genStore.progressPercent + '%' }" />
    </div>
    <div class="section-states">
      <div v-for="sec in flatSections" :key="sec.id" class="state-row"
        :class="getSectionState(sec.id)">
        <span class="state-icon">{{ iconFor(getSectionState(sec.id)) }}</span>
        <span class="state-title" :style="{ paddingLeft: (sec._level || 0) * 16 + 'px' }">
          {{ sec.title }}
        </span>
        <button v-if="getSectionState(sec.id) === 'error'"
          class="retry-btn" @click="genStore.retrySection(sec.id)">重试</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useGenerateStore } from '../stores/generateStore'

const genStore = useGenerateStore()

const flatSections = computed(() => {
  const flat: any[] = []
  const flatten = (secs: any[], level: number) => {
    for (const s of secs) {
      flat.push({ ...s, _level: level })
      flatten(s.children, level + 1)
    }
  }
  flatten(genStore.outline, 0)
  return flat
})

function getSectionState(id: string) {
  return genStore.getSectionState(id)
}

function iconFor(state: string) {
  switch (state) {
    case 'done': return '✅'
    case 'generating': return '⏳'
    case 'error': return '❌'
    default: return '⬜'
  }
}
</script>

<style scoped>
.progress-tracker {
  padding: 8px 16px; background: #FBF7EF; border-top: 1px solid #E0D5C0;
}
.progress-header { display: flex; justify-content: space-between; margin-bottom: 8px; }
.progress-title { font-size: 13px; font-weight: 600; color: #3D2B1F; }
.progress-count { font-size: 13px; color: #8B7355; }
.progress-bar { height: 6px; background: #E0D5C0; border-radius: 3px; margin-bottom: 12px; overflow: hidden; }
.progress-fill { height: 100%; background: #C23B22; border-radius: 3px; transition: width 0.3s; }
.section-states { max-height: 200px; overflow-y: auto; }
.state-row { display: flex; align-items: center; gap: 8px; padding: 6px 4px; font-size: 13px; }
.state-icon { width: 18px; text-align: center; }
.state-title { flex: 1; color: #3D2B1F; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.retry-btn { padding: 2px 8px; font-size: 11px; border: 1px solid #C23B22; border-radius: 4px; background: #fff; color: #C23B22; cursor: pointer; }
.state-row.generating .state-title { color: #C23B22; }
.state-row.done .state-title { color: #8B7355; }
</style>
```

- [ ] **步骤 3：Commit**

```bash
git add frontend/src/components/GenerateFlowDialog.vue frontend/src/components/ProgressTracker.vue frontend/src/components/PreviewNode.vue
git commit -m "feat: add outline preview dialog and progress tracker"
```

---

### 任务 5：Frontend — AIChat + EditorView + OutlineTree 集成

**文件：** 修改 `frontend/src/components/AIChat.vue`，`frontend/src/views/EditorView.vue`，`frontend/src/components/OutlineTree.vue`

- [ ] **步骤 1：修改 AIChat handleSend 检测 generate skill**

在 `frontend/src/components/AIChat.vue` 的 `<script>` 中：

```typescript
// Add to imports
import { useGenerateStore } from '../stores/generateStore'
const genStore = useGenerateStore()
```

修改 `handleSend` 函数（找到 `chatStore.sendMessage(...)` 调用处，根据 activeSkillObj 类型分流）：

```typescript
async function handleSend() {
  const text = inputText.value.trim()
  if (!text || chatStore.isSending) return

  const skill = activeSkillObj.value
  const finalText = skill ? text : text

  if (skill?.type === 'generate') {
    // Generation flow
    const modelEntry = settingsStore.apiKeys.find(k => k.id === selectedModelId.value)
    if (!modelEntry) {
      chatStore.messages.push({ role: 'ai', content: '请先在设置中配置 API 密钥' })
      return
    }
    await genStore.generateOutline(
      props.docId || '',
      text,
      skill.prompt,
      {
        provider: modelEntry.provider,
        model: modelEntry.model,
        endpoint: modelEntry.endpoint || '',
        format: modelEntry.format || 'openai',
        apiKey: modelEntry.key,
      }
    )
    inputText.value = ''
    showSkillPopup.value = false
    activeSkillObj.value = null
    return
  }

  // Existing chat logic
  // ...
  // chatStore.sendMessage(...)
}
```

- [ ] **步骤 2：修改 EditorView 集成 Dialog + ProgressTracker + 监听 gen-chunk/gen-done**

```typescript
// frontend/src/views/EditorView.vue — <script> additions

// Import generate store
import { useGenerateStore } from '../stores/generateStore'
const genStore = useGenerateStore()

// Listen for gen-chunk and gen-done events (in onMounted)
onMounted(() => {
  // ... existing code
  window.addEventListener('gen-chunk', handleGenChunk)
  window.addEventListener('gen-done', handleGenDone)
})
onUnmounted(() => {
  // ... existing cleanup
  window.removeEventListener('gen-chunk', handleGenChunk)
  window.removeEventListener('gen-done', handleGenDone)
})

// Handle streaming chunks
function handleGenChunk(e: CustomEvent) {
  const { sectionId, chunk } = e.detail
  // Append chunk to editor content
  if (editorRef.value?.insertContent) {
    editorRef.value.insertContent(chunk)
  } else if (editorRef.value?.setContent) {
    // Fallback: append to existing content
    editorRef.value.insertContent?.(chunk)
  }
}

// Handle section done
function handleGenDone(e: CustomEvent) {
  // Section generation complete, update outline
}
```

Template additions (inside `<template>`):

```vue
<!-- After .center-panel -->
<GenerateFlowDialog
  v-if="genStore.phase === 'preview'"
  @confirm="onConfirmGeneration"
  @cancel="genStore.reset()"
/>

<!-- Inside .left-panel, after OutlineTree -->
<ProgressTracker v-if="genStore.phase === 'generating' || genStore.phase === 'done'" />
```

And the confirm handler:

```typescript
function onConfirmGeneration() {
  genStore.confirmGeneration()
  // Fill outline into docStore
  docStore.updateOutlineTree(props.id, genStore.outline)
}
```

- [ ] **步骤 3：修改 OutlineTree 显示章节状态**

在 `frontend/src/components/OutlineTree.vue` 中，对每个 `OutlineTreeNode` 传入状态：

```vue
<OutlineTreeNode
  v-for="section in docStore.outline"
  :key="section.id"
  :section="section"
  :depth="0"
  :active-section-id="docStore.activeSectionId"
  :open-menu-id="openMenuId"
  :section-state="genStore.getSectionState(section.id)"
  @select="selectSection"
  ...
/>
```

在 `OutlineTreeNode.vue` 中，根据状态添加样式：

```vue
<div class="outline-node" :class="{
  active: section.id === activeSectionId,
  'state-generating': sectionState === 'generating',
  'state-done': sectionState === 'done',
  'state-error': sectionState === 'error',
}" @click="handleClick">
  <span class="state-icon">{{ iconFor(sectionState) }}</span>
  <span class="node-title">{{ section.title }}</span>
  ...
</div>

<script setup>
// Add prop
defineProps<{ ..., sectionState?: string }>()

function iconFor(state: string) {
  switch (state) {
    case 'done': return '✅'
    case 'generating': return '⏳'
    case 'error': return '❌'
    default: return ''
  }
}
</script>
```

同时，当用户点击一个 `pending` 状态的章节时，触发生成：

在 `OutlineTree.vue` 的 `selectSection` 中：

```typescript
function selectSection(id: string) {
  // ... existing selection logic
  if (genStore.phase === 'generating' && genStore.getSectionState(id) === 'pending') {
    // Find section in outline
    const section = findSectionById(genStore.outline, id)
    if (section) {
      const path = getSectionPath(genStore.outline, id)
      genStore.generateSection(id, section.title, path, JSON.stringify(genStore.outline))
    }
  }
}
```

- [ ] **步骤 4：Commit**

```bash
git add frontend/src/components/AIChat.vue frontend/src/views/EditorView.vue frontend/src/components/OutlineTree.vue frontend/src/components/OutlineTreeNode.vue
git commit -m "feat: integrate generation workflow into AIChat, EditorView and OutlineTree"
```

---

## 验证

### 端到端测试流程

1. 启动后端：`cd backend && go run ./cmd/server/main.go`
2. 启动前端：`cd frontend && npm run dev`
3. 上传一份空 DOCX 模板
4. 在 AIChat 中选择"生成标书"skill
5. 输入需求（如"生成一份服务器采购标书"）
6. 点击发送 → 等待 AI 返回大纲 → 弹框预览
7. 确认 → 大纲填入左侧树（全部 pending）
8. 点击章节 → SSE 流式生成 → 内容填入编辑器
9. 章节状态从 ⏳ → ✅
10. 全部完成后，进度条显示 100%

### 验证检查点

| 检查项 | 预期 |
|--------|------|
| generate-outline 返回正确大纲 JSON | AI 回复的 md 被正确解析为 Section 树 |
| generate-section SSE 流式响应 | 每个 chunk 写入 `data: {chunk: "..."}\n\n` |
| 前端解析 SSE | reader.read() 逐块解析，无遗漏 |
| 内容写入编辑器 | `editorRef.value.insertContent(chunk)` 逐块追加 |
| 章节状态同步 | Map 状态正确反映 pending/generating/done/error |
| 错误重试 | error 章节显示重试按钮，点击后清状态重新生成 |

## 执行

计划已完成并保存到 `docs/superpowers/plans/2026-07-26-bid-generation-workflow.md`。

两种执行方式：

**1. 子代理驱动（推荐）** - 每个任务调度一个新的子代理，快速迭代

**2. 内联执行** - 在当前会话中执行任务，批量执行并设有检查点

选哪种方式？
