<template>
  <div class="content-editor">
    <div class="editor-header">
      <div class="current-section">
<span class="section-main">{{ currentSectionTitle }}</span>
        </div>
      <div class="edit-tools">
        <button class="tool-btn save-btn" title="Save" @click="save" :disabled="saving">
          <RiSaveLine size="18" color="#8B7355" />
        </button>
      </div>
    </div>
    <div class="editor-body">
      <div class="content-area">
        <div class="section-title">{{ currentSectionTitle }}</div>
        <EditorContent v-if="editor" :editor="editor" @update="onEditorUpdate" />
        <div v-else class="editor-placeholder">选择章节开始编辑</div>
      </div>
      <div class="ai-generate-area">
        <button class="gen-btn outline-btn" @click="extractOutline">
          <RiFileListLine size="20" color="#8B7355" />
          <span>大纲提取</span>
        </button>
        <button class="gen-btn bid-btn" @click="generateBid">
          <RiFilePaper2Line size="20" color="#8B7355" />
          <span>标书生成</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import { useSettingsStore } from '../stores/settingsStore'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import {
  RiSaveLine,
  RiFileListLine,
  RiFilePaper2Line,
} from '@remixicon/vue'
import { exportDocument } from '../api/client'

const route = useRoute()
const docStore = useDocumentStore()
const settingsStore = useSettingsStore()
const docId = computed(() => route.params.id as string)

const activeSectionId = computed(() => docStore.activeSectionId)
const currentSection = computed(() => {
  if (!activeSectionId.value) return null
  return docStore.sections.get(activeSectionId.value)
})
const currentSectionTitle = computed(() => currentSection.value?.title || '选择章节开始编辑')

const saving = ref(false)

const editor = useEditor({
  extensions: [StarterKit],
  content: '',
  editorProps: {
    attributes: {
      class: 'tiptap-editor',
    },
  },
})

let debounceTimer: ReturnType<typeof setTimeout>
const debouncedSave = (sectionId: string, content: string) => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    docStore.saveSectionContent(docId.value, sectionId, content)
  }, 1000)
}

const onEditorUpdate = ({ editor: ed }: { editor: any }) => {
  if (!activeSectionId.value) return
  const html = ed.getHTML()
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
  if (newId && editor.value) {
    await docStore.loadSection(docId.value, newId)
    const section = docStore.sections.get(newId)
    editor.value.commands.setContent(section?.content || '<p></p>')
  }
}, { immediate: false })

const toggleAI = () => {}

const extractOutline = async () => {
  if (!activeSectionId.value || !editor.value) return
  try {
    const outline = docStore.getFullOutline()
    if (!outline || outline.length === 0) {
      alert('当前文档没有大纲内容')
      return
    }
    const md = buildOutlineMarkdown(outline, 0)
    editor.value.commands.setContent(md)
  } catch (err) {
    console.error('Extract outline failed:', err)
    alert('大纲提取失败，请重试')
  }
}

const generateBid = async () => {
  if (!docId.value) return
  try {
    const outline = docStore.getFullOutline()
    const allSections = collectAllSections(outline)
    if (allSections.length === 0) {
      alert('请先上传或选择模板以生成大纲')
      return
    }
    let allContent = ''
    for (let i = 0; i < allSections.length; i++) {
      const sec = allSections[i]
      const existingContent = sec.content ? `${sec.content}\n\n` : ''
      allContent += `# ${sec.title}\n\n${existingContent}`
      if (i < allSections.length - 1) {
        allContent += '\n---\n\n'
      }
    }
    editor.value?.commands.setContent(allContent)
    await docStore.saveSectionContent(docId.value, 'all', allContent)
    alert('标书内容已生成完成，可点击导出按钮下载')
  } catch (err) {
    console.error('Generate bid failed:', err)
    alert('标书生成失败，请重试')
  }
}

function buildOutlineMarkdown(sections: any[], depth: number): string {
  let md = ''
  for (const s of sections) {
    const prefix = '#'.repeat(depth + 1)
    md += `${prefix} ${s.title}\n`
    if (s.children && s.children.length > 0) {
      md += buildOutlineMarkdown(s.children, depth + 1)
    }
  }
  return md
}

function collectAllSections(sections: any[]): any[] {
  const result: any[] = []
  for (const s of sections) {
    result.push({ title: s.title, content: s.content || '' })
    if (s.children && s.children.length > 0) {
      result.push(...collectAllSections(s.children))
    }
  }
  return result
}

const generateBid = async () => {
  try {
    const fmt = settingsStore.exportFormat
    const res = await exportDocument(docId.value, fmt)
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    const filename = fmt === 'md' ? 'bid-document.md' : 'bid-document.docx'
    link.href = url
    link.setAttribute('download', filename)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  } catch (err) {
    console.error('Export failed:', err)
    alert('导出失败，请重试')
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
  padding: 12px 20px;
  flex-shrink: 0;
}

.current-section {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-main {
  font-size: 14px;
  font-weight: 600;
  color: #3D2B1F;
}

.section-sub {
  font-size: 12px;
  color: #8B7355;
}

.edit-tools {
  display: flex;
  gap: 8px;
}

.tool-btn {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: #F0E8D5;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.tool-btn:hover {
  background: #E8DCC8;
}

.tool-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.editor-body {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
}

.content-area {
  flex: 1;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #3D2B1F;
  margin-bottom: 16px;
}

.editor-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #8B7355;
  font-size: 14px;
}

.tiptap-editor {
  outline: none;
  font-size: 14px;
  line-height: 1.7;
  color: #3D2B1F;
}

.tiptap-editor :deep(p) {
  margin: 0.6em 0;
}

.tiptap-editor :deep(h1),
.tiptap-editor :deep(h2),
.tiptap-editor :deep(h3) {
  margin-top: 1em;
  margin-bottom: 0.5em;
  color: #3D2B1F;
}

.tiptap-editor :deep(ul),
.tiptap-editor :deep(ol) {
  padding-left: 1.5em;
  margin: 0.5em 0;
}

.tiptap-editor :deep(blockquote) {
  border-left: 3px solid #E0D5C0;
  padding-left: 12px;
  margin: 0.5em 0;
  color: #8B7355;
}

.ai-generate-area {
  display: flex;
  gap: 12px;
  margin-top: 24px;
  background: #FFF8F0;
  border-radius: 12px;
  padding: 16px;
  flex-shrink: 0;
  justify-content: center;
}

.gen-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 32px;
  border-radius: 12px;
  border: none;
  cursor: pointer;
  font-size: 15px;
  font-weight: 500;
  transition: background 0.2s;
}

.outline-btn {
  background: #F0E8D5;
  color: #8B7355;
}

.outline-btn:hover {
  background: #E8DCC8;
}

.bid-btn {
  background: #C23B22;
  color: #fff;
}

.bid-btn:hover {
  background: #A83028;
}

.bid-btn span {
  color: #fff;
}
</style>