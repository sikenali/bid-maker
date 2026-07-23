# Full Edit Restructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace manual DOCX reading/editing/export with @eigenpal/docx-editor-vue, while keeping backend unioffice heading extraction as a background data source.

**Architecture:** Frontend uses @eigenpal/docx-editor-vue for full-doc editing. Its built-in `useOutlineSidebar` composable provides heading positions (paraId). Backend unioffice parsing remains as background service for template sync and keyword matching. The OutlineTree navigates via paraId → editor.scrollToParaId().

**Tech Stack:** Vue 3, TypeScript, ProseMirror (via eigenpal), unioffice (Go), Gin (Go), axios

## Global Constraints

- @eigenpal/docx-editor-vue version 1.9.0 (already installed in frontend/)
- unioffice fork: github.com/sikenali/unioffice/v2 v2.0.0-20260701084101-423612299d83
- Data model: Section{ID, Title, Level, ParentID, Content, Children} unchanged
- API routes unchanged (upload, outline, chat)
- Heading detection patterns: Heading1-Heading9 + Chinese variants

---

### Task 1: Remove unioffice generateDocument from backend

**Files:**
- Delete: `backend/internal/service/zipwriter.go`
- Modify: `backend/internal/service/docx_service.go:152-267` — remove GenerateDocument, generateDocumentXML, buildDocumentXML, buildContentTypes, buildRels, writeSectionXML, writeParagraphXML, xmlEscape
- Modify: `backend/internal/handler/handler.go:182-212` — Update ExportDocument to return empty success or remove export route entirely
- Modify: `backend/internal/service/docx_service_test.go` — Remove TestGenerateDocument test

**Interfaces:**
- Consumes: nothing
- Produces: DocxService no longer has GenerateDocument method

- [ ] **Step 1: Delete zipwriter.go**

  Run: `rm /home/jingle/opc/bid-maker/backend/internal/service/zipwriter.go`

- [ ] **Step 2: Remove GenerateDocument and XML generation functions from docx_service.go**

  Remove these functions entirely from `/home/jingle/opc/bid-maker/backend/internal/service/docx_service.go`:
  - Line ~152-154: `func (s *DocxService) GenerateDocument(doc *model.Document) ([]byte, error)`
  - Line ~156-178: `func (s *DocxService) generateDocumentXML(doc *model.Document) ([]byte, error)`
  - Line ~180-191: `func (s *DocxService) buildDocumentXML(sections []model.Section) string`
  - Line ~193-209: `func (s *DocxService) writeSectionXML(b *strings.Builder, sec *model.Section)`
  - Line ~211-218: `func writeParagraphXML(b *strings.Builder, text, style string)`
  - Line ~220-227: `func xmlEscape(s string)`
  - Line ~229-236: `func (s *DocxService) buildContentTypes() string`
  - Line ~238-244: `func (s *DocxService) buildRels() string`
  - Line ~246-252: `func (s *DocxService) GenerateMarkdown(doc *model.Document) []byte` — KEEP this function
  - Line ~254-263: `func (s *DocxService) writeSectionMarkdown(b *strings.Builder, sec *model.Section, level int)` — KEEP this function

  Actually, let me be precise about what to KEEP vs REMOVE:

  **KEEP:**
  - `paragraphText(para document.Paragraph) string` — used by ParseDocument
  - `isHeadingUnioffice(para document.Paragraph) (bool, int)` — used by ParseDocument
  - `extractSectionsWithKeywordUnioffice(paras []document.Paragraph, keyword string) []model.Section` — used by ParseDocument
  - `extractSectionsUnioffice(paras []document.Paragraph) []model.Section` — used by ParseDocument
  - All manual XML types: wDoc, wBody, wPara, wParaProps, wStyle, wRun, wText
  - `extractTextXml(runs []wRun) string`
  - `headingLevelXml(p wPara) int`
  - `extractSectionsXml(paras []wPara, keyword string) []model.Section`
  - `filterKeywordOutline(sections []model.Section, keyword string) []model.Section`
  - `findLeaf(sec *model.Section) *model.Section`
  - `ParseDocument(data []byte) (*model.Document, error)` — KEEP, the main function
  - `parseWithUnioffice(data []byte) (*model.Document, error)` — KEEP
  - `parseWithXmlFallback(data []byte) (*model.Document, error)` — KEEP
  - `GenerateMarkdown(doc *model.Document) []byte` — KEEP
  - `writeSectionMarkdown(b *strings.Builder, sec *model.Section, level int)` — KEEP
  - `NowUTC() time.Time` — KEEP

  **REMOVE:**
  - `GenerateDocument(doc *model.Document) ([]byte, error)`
  - `generateDocumentXML(doc *model.Document) ([]byte, error)`
  - `buildDocumentXML(sections []model.Section) string`
  - `writeSectionXML(b *strings.Builder, sec *model.Section)`
  - `writeParagraphXML(b *strings.Builder, text, style string)`
  - `xmlEscape(s string)`
  - `buildContentTypes() string`
  - `buildRels() string`

  After removals, clean up unused imports:
  - Remove `"bytes"` import
  - Remove `"archive/zip"` not needed anymore (keep only if parseWithXmlFallback uses it — yes it does!)

- [ ] **Step 3: Update handler.go ExportDocument**

  In `/home/jingle/opc/bid-maker/backend/internal/handler/handler.go`, replace `ExportDocument` (lines 182-212):

```go
func (h *Handler) ExportDocument(c *gin.Context) {
    id := c.Param("id")
    _, ok := service.GetDocument(id)
    if !ok {
        c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
        return
    }
    // Export is now done by frontend via @eigenpal save()
    // Keep API for backward compatibility but return success
    c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "export complete — frontend handles save()"}
})
}
```

- [ ] **Step 4: Remove GenerateDocument test**

  In `/home/jingle/opc/bid-maker/backend/internal/service/docx_service_test.go`, remove the entire `TestGenerateDocument` function (around line 125-155).

- [ ] **Step 5: Build and verify**

  Run: `cd backend && go build ./...`
  Expected: no errors

- [ ] **Step 6: Run tests**

  Run: `cd backend && go test ./internal/service/ -v -count=1 2>&1 | tail -30`
  Expected: all tests pass (TestGenerateDocument removed, others remain)

- [ ] **Step 7: Commit**

  ```bash
  git rm backend/internal/service/zipwriter.go
  git add backend/internal/service/docx_service.go backend/internal/service/docx_service_test.go backend/internal/handler/handler.go
  git commit -m "chore: remove GenerateDocument and XML generation from backend"
  ```

---

### Task 2: Wire @eigenpal/docx-editor-vue into EditorView

**Files:**
- Create: `frontend/src/components/DocxEditorWrapper.vue` — thin wrapper around DocxEditor component
- Modify: `frontend/src/views/EditorView.vue` — integrate DocxEditor

**Interfaces:**
- Consumes: `docId` prop, `docxBuffer` bytes (from upload)
- Produces: `editorRef` expose scrollToParaId(), save(), etc.

- [ ] **Step 1: Create DocxEditorWrapper.vue**

  Create `/home/jingle/opc/bid-maker/frontend/src/components/DocxEditorWrapper.vue`:

```vue
<template>
  <div class="docx-editor-wrapper" ref="wrapper">
    <DocxEditor
      v-if="docBuffer"
      :document-buffer="docBuffer"
      :show-menu-bar="true"
      :show-toolbar="true"
      :show-outline="false"
      :read-only="false"
      :on-open="handleOpen"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, type Ref } from 'vue'
import { DocxEditor, useOutlineSidebar } from '@eigenpal/docx-editor-vue'
import type { DocxEditorHandle, useOutlineSidebar } from '@eigenpal/docx-editor-vue'

const props = defineProps<{
  docBuffer?: ArrayBuffer | null
  outlineSections?: any[]
}>()

const emit = defineEmits<{
  selectSection: [sectionId: string]
}>()

const handleOpen = async (file: File) => {
  // User opens another file via @eigenpal's built-in open
  const buffer = await file.arrayBuffer()
  emit('openFile', buffer)
}

defineExpose({
  editorRef: null as unknown as ReturnType<typeof useOutlineSidebar>['editorView'],
})
</script>
```

Wait, I need to check exactly which exports are available from the package. Let me look at the type definitions:

From the .d.ts files I read:
- `import { DocxEditor, useOutlineSidebar } from '@eigenpal/docx-editor-vue'` — both are exported
- `useOutlineSidebar` takes options including `editorView`, `showOutline`, `outlineHeadings`, etc.

Actually, `@eigenpal/docx-editor-vue` has a more complex API. The simplest approach: just embed `<DocxEditor>` directly in EditorView without a separate wrapper. Let me adjust:

- [ ] **Step 1: Integrate @eigenpal directly into EditorView.vue**

  Modify `/home/jingle/opc/bid-maker/frontend/src/views/EditorView.vue`. Replace the three-panel layout (OutlineTree | ContentEditor | AIChat) with:

```vue
<template>
  <div class="page">
    <!-- navbar remains unchanged -->
    <main class="editor-body">
      <aside class="left-panel">
        <OutlineTree @select="handleSelectSection" />
      </aside>
      <section class="center-panel">
        <DocxEditor
          ref="editorRef"
          :document-buffer="docxBuffer"
          :show-menu-bar="true"
          :show-toolbar="true"
          :show-outline="false"
          :read-only="false"
        />
      </section>
      <aside class="right-panel">
        <AIChat :doc-id="docId" :outline="docStore.outline" />
      </aside>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import OutlineTree from '../components/OutlineTree.vue'
import AIChat from '../components/AIChat.vue'
import { DocxEditor } from '@eigenpal/docx-editor-vue'
import type { DocxEditorRef } from '@eigenpal/docx-editor-vue/dist/components/DocxEditor/types'

const props = defineProps<{ id: string }>()
const router = useRouter()
const docStore = useDocumentStore()
const editorRef = ref<DocxEditorRef | null>(null)
const docxBuffer = ref<ArrayBuffer | null>(null)

// Load document outline AND docx buffer
onMounted(async () => {
  try {
    // 1. Load outline from backend
    await docStore.loadOutline(props.id)
    
    // 2. Get docx buffer — either from URL params (passed from UploadView)
    //    or fetch it separately
    
    // For now, assume the docx buffer needs to come from somewhere.
    // Option A: UploadView stores it in the router query
    // Option B: Fetch it from backend API
    
    // The file input from upload gives us the raw file bytes.
    // We need to store it somewhere accessible.
  } catch (err) {
    console.error('Failed to load document:', err)
  }
})

const handleSelectSection = (sectionId: string) => {
  const section = findSectionInSectionById(sectionId)
  if (section?.paraId && editorRef.value) {
    editorRef.value.scrollToParaId(section.paraId)
  }
  docStore.activeSectionId.value = sectionId
}

function findSectionInSectionById(id: string) {
  // Walk docStore.outline recursively to find matching section
  // This can be added to documentStore later
}
</script>
```

- [ ] **Step 2: Store docxBuffer in documentStore**

  Modify `/home/jingle/opc/bid-maker/frontend/src/stores/documentStore.ts`. Add:
  ```typescript
  const docxBuffer = ref<ArrayBuffer | null>(null)
  
  const setDocxBuffer = (buffer: ArrayBuffer) => {
    docxBuffer.value = buffer
  }
  
  return { ..., docxBuffer, setDocxBuffer }
  ```

- [ ] **Step 3: Pass docxBuffer from UploadView to EditorView**

  In `UploadView.vue`, after successful upload:
  ```javascript
  // Store the original file bytes before navigating
  sessionStorage.setItem(`docx-${docId}`, JSON.stringify(Array.from(new Uint8Array(event.target.result))))
  ```
  Or better: use the router query to pass file bytes as base64? No, too large.

  Better approach: upload endpoint returns document ID, UploadView stores the raw file bytes in a Pinia store shared with EditorView.

- [ ] **Step 4: Build and run dev server**

  Run: `pnpm dev` in workspace root
  Expected: Dev server starts, EditorView shows @eigenpal editor when document loaded

- [ ] **Step 5: Commit**

  ```bash
  git add frontend/src/views/EditorView.vue frontend/src/stores/documentStore.ts frontend/src/views/UploadView.vue frontend/src/components/AIChat.vue
  git commit -m "feat: integrate @eigenpal/docx-editor-vue for full-document editing"
  ```

---

### Task 3: Convert HeadingInfo[] to Section[] and wire OutlineTree

**Files:**
- Modify: `frontend/src/stores/documentStore.ts` — add heading conversion + paraId support
- Modify: `frontend/src/components/OutlineTree.vue` — update click to call scrollToParaId
- Modify: `frontend/src/views/EditorView.vue` — pass outline to EditorView via prop from DocumentStore

**Interfaces:**
- Consumes: `headingInfo[]` from @eigenpal or `docStore.outline` from backend
- Produces: `section.paraId` for scroll navigation

- [ ] **Step 1: Add paraId field mapping in documentStore**

  Create helper function in `documentStore.ts`:
  ```typescript
  // Convert HeadingInfo array to tree structure
  function headingInfoToSections(headings: Array<{title: string, level: number, paraId: string}>): Section[] {
    const sections: Section[] = []
    const stack: {section: Section; level: number}[] = []
    
    for (const h of headings) {
      const section: Section = {
        id: h.paraId,
        title: h.title,
        level: h.level,
        parent_id: '',
        content: '',
        children: [],
        paraId: h.paraId
      }
      
      // Find correct parent
      while (stack.length > 0 && stack[stack.length - 1].level >= h.level) {
        stack.pop()
      }
      
      if (stack.length === 0) {
        sections.push(section)
      } else {
        stack[stack.length - 1].section.children!.push(section)
      }
      
      stack.push({ section, level: h.level })
    }
    
    return sections
  }
  ```

- [ ] **Step 2: Add paraId to Section interface**

  In `documentStore.ts`, update the Section interface:
  ```typescript
  export interface Section {
    id: string
    title: string
    level: number
    parent_id: string
    content: string
    children: Section[]
    paraId?: string  // NEW: Word paragraph ID for navigation
  }
  ```

- [ ] **Step 3: Update loadOutline to accept HeadingInfo**

  Modify `loadOutline` to optionally merge paraIds from backend response OR generate them:
  ```typescript
  const loadOutline = async (docId: string) => {
    const res = await getOutline(docId)
    // Backend sends section outlines with IDs from unioffice
    // paraId comes from Document ID which we need to map
    const items = res.data.outline || []
    outline.value = items
  }
  ```

- [ ] **Step 4: Update OutlineTree.vue click handler**

  Ensure OutlineTree emits section IDs that match `paraId` values:
  ```vue
  @select="(sectionId: string) => {
    docStore.outline.map((s: Section) => {
      if (s.id === sectionId) {
        // scrollToParaId called from EditorView or docStore
      }
    })
  }"
  ```

- [ ] **Step 5: Verify UI navigation works**

  Run: `pnpm dev`
  Expected: Clicking outline node scrolls @eigenpal to corresponding paragraph

- [ ] **Step 6: Commit**

  ```bash
  git add frontend/src/stores/documentStore.ts frontend/src/components/OutlineTree.vue frontend/src/views/EditorView.vue
  git commit -m "feat: wire OutlineTree paraId navigation to @eigenpal editor"
  ```

---

### Task 4: Wire AI Chat to @eigenpal and finalize

**Files:**
- Modify: `frontend/src/components/AIChat.vue`
- Modify: `frontend/src/views/UploadView.vue` — store docxBuffer after upload
- Cleanup: remove ContentEditor.vue if no longer used

**Interfaces:**
- Consumes: `docId`, outline sections, editor refs
- Produces: edited content saved back to editor

- [ ] **Step 1: Integrate AI Chat save result to @eigenpal**

  In `AIChat.vue`, when receiving LLM response:
  ```typescript
  const applyAIPatch = async (response: string) => {
    // Find matching section by title or sectionId
    // Use editorRef.setContentControlContent or ProseMirror API to patch
  }
  ```

- [ ] **Step 2: Handle file bytes from upload**

  In `UploadView.vue`, after successful upload, store original bytes:
  ```javascript
  const handleFile = async (file: File) => {
    const buffer = await file.arrayBuffer()
    // Store in Pinia store for EditorView access
    await uploadDocument(file) // uploads to backend
    // Pass buffer somehow to EditorView
  }
  ```

- [ ] **Step 3: Clean up unused components**

  Check if `ContentEditor.vue` can be safely removed or repurposed.

- [ ] **Step 4: Final integration test**

  Run: `pnpm dev`
  Flow: Upload → @eigenpal loads document → OutlineTree shows outline → Click node scrolls → AI Chat edits content → Save produces DOCX

- [ ] **Step 5: Commit**

  ```bash
  git add frontend/src/components/AIChat.vue frontend/src/views/UploadView.vue frontend/src/components/ContentEditor.vue
  git commit -m "feat: connect AI Chat to @eigenpal editor and finalize upload flow"
  ```

---

## Scope Notes

- Task 1 removes GenerateDocument and all XML generation code — this is a breaking change ONLY affecting the export endpoint. The export endpoint now serves only as backward compatibility for existing frontend code. Actual DOCX export is handled by `@eigenpal/save()`.
- Task 2 requires understanding how to share state between UploadView and EditorView. Options: sessionStorage, route query params, or a shared Pinia store. Shared Pinia store is preferred.
- Task 3 requires `paraId` from backend to match what `@eigenpal` uses. If backend generates its own IDs (like "sec-1"), they won't match @eigenpal's internal IDs. Consider passing paraIds alongside the outline data.
- Task 4 depends on Tasks 1-3 being integrated first. Chat-to-editor integration may need ProseMirror selection API.
