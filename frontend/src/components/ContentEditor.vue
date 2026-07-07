<template>
  <div class="content-editor">
    <div class="editor-header">
      <span class="current-section">{{ currentSectionTitle }}</span>
      <div class="toolbar">
        <button class="ai-btn" title="AI Assist" @click="toggleAI" :disabled="saving">
          ✨ AI Assist
        </button>
        <button class="save-btn" title="Save" @click="save" :disabled="saving">
          {{ saving ? 'Saving...' : '💾 Save' }}
        </button>
      </div>
    </div>
    <div class="editor-body">
      <Editor
        v-if="editor"
        :editor="editor"
        @update="onEditorUpdate"
      />
      <div v-else class="editor-placeholder">Select a section to edit</div>
    </div>
    <div class="editor-footer">
      <button class="outline-btn" @click="extractOutline">Extract Outline</button>
      <button class="export-btn" @click="generateBid">Generate Bid</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { exportDocument } from '../api/client'
import { useRoute } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import { Editor } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'

const route = useRoute()
const docStore = useDocumentStore()
const docId = computed(() => route.params.id as string)

const activeSectionId = computed(() => docStore.activeSectionId)
const currentSection = computed(() => {
  if (!activeSectionId.value) return null
  return docStore.sections.get(activeSectionId.value)
})
const currentSectionTitle = computed(() => currentSection.value?.title || 'No section selected')
const saving = ref(false)

const editor = ref<Editor | null>(null)

let debounceTimer: ReturnType<typeof setTimeout>
const debouncedSave = (sectionId: string, content: string) => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    docStore.saveSectionContent(docId.value, sectionId, content)
  }, 1000)
}

const onEditorUpdate = ({ editor }: { editor: Editor }) => {
  if (!activeSectionId.value) return
  const html = editor.getHTML()
  debouncedSave(activeSectionId.value, html)
}

const save = async () => {
  if (!activeSectionId.value || !editor.value) return
  saving.value = true
  try {
    const html = editor.value.getHTML()
    await docStore.saveSectionContent(docId.value, activeSectionId.value, html)
  } finally {
    saving.value = false
  }
}

watch(activeSectionId, async (newId) => {
  if (newId) {
    await docStore.loadSection(docId.value, newId)
    const section = docStore.sections.get(newId)
    if (editor.value) {
      editor.value.commands.setContent(section?.content || '<p></p>')
    }
  }
}, { immediate: false })

onMounted(async () => {
  editor.value = new Editor({
    extensions: [StarterKit],
    content: '',
    editorProps: {
      attributes: {
        class: 'tiptap-editor',
      },
    },
  })

  if (editor.value && activeSectionId.value) {
    await docStore.loadSection(docId.value, activeSectionId.value)
    const section = docStore.sections.get(activeSectionId.value)
    editor.value.commands.setContent(section?.content || '<p></p>')
  }
})

onBeforeUnmount(() => {
  if (editor.value) {
    editor.value.destroy()
    editor.value = null
  }
})

const toggleAI = () => {
  console.log('AI Assist triggered')
}

const extractOutline = () => {
  console.log('Outline already extracted on upload. Re-extract placeholder.')
}

const generateBid = async () => {
  try {
    const res = await exportDocument(docId.value)
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', 'bid-document.docx')
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  } catch (err) {
    console.error('Export failed:', err)
    alert('Export failed. Please try again.')
  }
}
</script>

<style scoped>
.content-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
  background: #fafafa;
}

.current-section {
  font-weight: 600;
  font-size: 15px;
  color: #333;
}

.toolbar {
  display: flex;
  gap: 8px;
}

.ai-btn, .save-btn {
  padding: 6px 14px;
  border-radius: 6px;
  border: 1px solid #ddd;
  background: white;
  cursor: pointer;
  font-size: 13px;
  transition: background 0.2s;
}

.ai-btn:hover, .save-btn:hover {
  background: #f0f0f0;
}

.ai-btn:disabled, .save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.editor-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.editor-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #999;
  font-size: 14px;
}

.tiptap-editor {
  max-width: 800px;
  margin: 0 auto;
  min-height: 100%;
  outline: none;
  font-size: 15px;
  line-height: 1.7;
  color: #333;
}

.tiptap-editor :deep(p) {
  margin: 0.6em 0;
}

.tiptap-editor :deep(h1),
.tiptap-editor :deep(h2),
.tiptap-editor :deep(h3) {
  margin-top: 1em;
  margin-bottom: 0.5em;
}

.tiptap-editor :deep(ul),
.tiptap-editor :deep(ol) {
  padding-left: 1.5em;
  margin: 0.5em 0;
}

.tiptap-editor :deep(blockquote) {
  border-left: 3px solid #d0d0d0;
  padding-left: 12px;
  margin: 0.5em 0;
  color: #555;
}

.editor-footer {
  display: flex;
  justify-content: center;
  gap: 16px;
  padding: 16px;
  border-top: 1px solid #eee;
  background: #fafafa;
}

.outline-btn, .export-btn {
  padding: 10px 24px;
  border-radius: 6px;
  border: 1px solid #ddd;
  background: white;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.2s;
}

.outline-btn:hover {
  background: #f5f5f5;
}

.export-btn {
  background: #1677ff;
  color: white;
  border: none;
}

.export-btn:hover {
  background: #4096ff;
}
</style>
