# Bid Document Generation Workflow

## Overview

A multi-step AI-powered workflow for generating bid documents from user requirements. Users describe their needs in AIChat, an AI skill generates a structured outline (Markdown → parsed tree), the user confirms, then sections are generated one-by-one on demand with streaming progress.

## User Flow

```
AIChat → 选"生成标书"skill + 输入需求 → 
  POST /api/generate-outline → AI 返回 Markdown 大纲 → 
  弹框预览大纲树 → 用户确认 →
  大纲填入左侧 OutlineTree（状态: pending）→
  用户点击某章节 → POST /api/generate-section（SSE）→
  流式内容填入编辑器 → 章节标记 done →
  用户继续点下一章...
```

## Backend API

### POST /api/generate-outline

Generate a structured outline from user requirements + skill prompt.

**Request:**
```json
{
  "message": "生成一份服务器采购标书，预算500万",
  "skill_prompt": "你是一个标书生成专家，根据需求生成标书大纲...",
  "provider": "qwen",
  "model": "qwen-max",
  "endpoint": "",
  "format": "openai",
  "apiKey": "sk-xxx"
}
```

**Response:**
```json
{
  "outline": [
    {
      "id": "sec_xxx",
      "title": "第一章 项目概况",
      "level": 1,
      "parent_id": "",
      "content": "",
      "children": [
        {
          "id": "sec_yyy",
          "title": "1.1 项目背景",
          "level": 2,
          "parent_id": "sec_xxx",
          "content": "",
          "children": []
        }
      ]
    }
  ]
}
```

**Backend logic:**
1. Construct system prompt from skill_prompt + instructions to return Markdown
2. Call AI (non-streaming, via existing go-openai or custom endpoint)
3. Parse AI response Markdown → Section tree (lines starting with `#` / `##` / `###`)
4. Assign unique IDs, build parent-child relationships
5. Return structured outline

### POST /api/generate-section (SSE)

Generate content for a single section, streamed via SSE.

**Request:**
```json
{
  "document_id": "abc123",
  "section_id": "sec_yyy",
  "section_title": "1.1 项目背景",
  "section_path": ["第一章 项目概况", "1.1 项目背景"],
  "outline_context": "full outline JSON string",
  "provider": "qwen",
  "model": "qwen-max",
  "endpoint": "",
  "format": "openai",
  "apiKey": "sk-xxx"
}
```

**Response (SSE stream):**
```
data: {"section_id": "sec_yyy", "chunk": "本项目旨在..."}
data: {"section_id": "sec_yyy", "chunk": "采购高性能服务器..."}
data: {"section_id": "sec_yyy", "done": true}
```

**Backend logic:**
1. Construct system prompt:
   - Current section title + full path
   - Full outline tree as context
   - Instructions: write formal bid document content
2. Call AI streaming API (`go-openai` `CreateChatCompletionStream`)
3. For each chunk, write SSE event to ResponseWriter
4. On completion, send `done: true`

## Frontend Components

### GenerateFlowDialog.vue

Modal dialog showing the parsed outline tree for user confirmation.

- Props: `outline: Section[]`
- Emits: `confirm`, `cancel`
- Display: tree view of sections with level indentation (read-only)
- Actions: "确认生成" / "取消" buttons

### ProgressTracker.vue

Inline progress panel showing generation status.

- Shown in the left panel below the outline
- Section-level status indicators: `pending` / `generating` / `done` / `error`
- Overall progress bar (e.g., "3/8 章节")
- Retry button for failed sections

### stores/generateStore.ts

```typescript
interface GenerateState {
  outline: Section[]
  phase: 'idle' | 'preview' | 'generating' | 'done' | 'error'
  sectionStates: Map<string, 'pending' | 'generating' | 'done' | 'error'>
  currentSectionId: string | null
  totalSections: number
  completedSections: number
  error: string | null
}
```

Key actions:
- `generateOutline(docId, message, skillPrompt, modelConfig)` → sets phase=preview
- `confirmGeneration()` → fills docStore.outline → starts generating first section
- `generateNextSection()` → calls SSE endpoint → streams content into DocxEditor
- `retrySection(sectionId)` → resets section state → re-generates
- `reset()` → clears all state

## EditorView Integration

- `AIChat` detects "生成标书" skill → triggers `genStore.generateOutline()`
- On `phase === 'preview'`: show `GenerateFlowDialog`
- On confirm: `OutlineTree` displays outline with status icons
- User clicks section → `genStore.generateNextSection(sectionId)`
- SSE chunks → `editorRef.value.setContent(content)` or `insertContent()`
- Section state updates reflected in `OutlineTree` node styling

## Section States in OutlineTree

| State | Icon | Node Style |
|-------|------|------------|
| pending | ⬜ | default |
| generating | ⏳ | pulse animation |
| done | ✅ | dimmed text |
| error | ❌ | red text + retry button |

## Error Handling

- Outline generation fails → error toast, allow retry
- Section generation fails → mark section as `error`, show retry button
- SSE connection drops → auto-retry with exponential backoff (max 3)
- API key invalid → propagate error from existing chat error handling

## Out of Scope (v1)

- Pause/resume generation mid-stream
- Batch generation (generate all sections at once)
- Editing outline before confirmation
- Database persistence (in-memory store only)
