<template>
  <div class="page">
    <header class="navbar">
      <div class="logo-area" @click="goHome">
        <div class="logo-icon">
          <RiRadarFill size="22" color="#fff" />
        </div>
        <div class="brand-texts">
          <span class="brand-zh">文制星</span>
          <span class="brand-en">Boomerang</span>
        </div>
      </div>
      <div class="nav-actions">
        <button class="nav-btn" title="导出" @click="handleExportDocx">
          <span class="nav-btn-content">
            <RiDownloadLine size="20" />
            <span class="nav-btn-label">导出</span>
          </span>
        </button>
        <button class="nav-btn" title="上传" @click="goHome">
          <span class="nav-btn-content">
            <RiQuestionLine size="20" />
            <span class="nav-btn-label">上传</span>
          </span>
        </button>
        <button class="nav-btn" title="设置" @click="goSettings">
          <span class="nav-btn-content">
            <RiSettingsLine size="20" />
            <span class="nav-btn-label">设置</span>
          </span>
        </button>
      </div>
    </header>

    <main class="editor-body">
      <aside class="left-panel">
        <OutlineTree @select="handleSelectSection" />
        <ProgressTracker v-if="genStore.phase === 'generating' || genStore.phase === 'done'" />
      </aside>
      <section class="center-panel">
        <MarkdownEditor
          v-if="docStore.markdown !== undefined"
          v-model="docStore.markdown"
          class="center-md-editor"
        />
      </section>
      <aside class="right-panel">
        <AIChat :doc-id="props.id" :editor-ref="editorRef" :docx-buffer="docStore.docxBuffer" @export-docx="handleExportDocx" />
      </aside>
    </main>

    <GenerateFlowDialog
      v-if="genStore.phase === 'preview'"
      @confirm="onConfirmGeneration"
      @cancel="genStore.reset()"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import OutlineTree from '../components/OutlineTree.vue'
import AIChat from '../components/AIChat.vue'
import GenerateFlowDialog from '../components/GenerateFlowDialog.vue'
import ProgressTracker from '../components/ProgressTracker.vue'
import MarkdownEditor from '../components/MarkdownEditor.vue'
import { useGenerateStore } from '../stores/generateStore'
import { exportDocx } from '../api/client'
import {
  RiRadarFill,
  RiQuestionLine,
  RiSettingsLine,
  RiDownloadLine,
} from '@remixicon/vue'

const props = defineProps<{ id: string }>()
const router = useRouter()
const docStore = useDocumentStore()
const genStore = useGenerateStore()

const editorRef = ref<{
  setContent: (content: string) => void
  insertContent: (content: string) => void
}>({
  setContent: (content: string) => {
    docStore.markdown = content
  },
  insertContent: (content: string) => {
    docStore.markdown = (docStore.markdown || '') + content
  },
})

const goSettings = () => router.push('/settings')
const goHome = () => router.push('/')

let mdSaveTimer: ReturnType<typeof setTimeout> | null = null

onMounted(async () => {
  try {
    await docStore.loadOutline(props.id)
    await docStore.loadMarkdown(props.id)
  } catch (err) {
    console.error('Failed to load outline:', err)
  }
  window.addEventListener('gen-chunk', handleGenChunk as EventListener)
  window.addEventListener('gen-done', handleGenDone as EventListener)
})

onUnmounted(() => {
  if (mdSaveTimer) {
    clearTimeout(mdSaveTimer)
    mdSaveTimer = null
  }
  window.removeEventListener('gen-chunk', handleGenChunk as EventListener)
  window.removeEventListener('gen-done', handleGenDone as EventListener)
})

watch(() => docStore.markdown, (val) => {
  if (mdSaveTimer) clearTimeout(mdSaveTimer)
  mdSaveTimer = setTimeout(() => docStore.saveDocumentMarkdown(props.id, val), 1000)
})

async function handleExportDocx() {
  try {
    const res = await exportDocx(props.id)
    const blob = new Blob([res.data], { type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `标书_${props.id}_${new Date().toISOString().slice(0, 10)}.docx`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (err) {
    console.error('Export failed:', err)
    alert('导出失败，请重试')
  }
}

const handleSelectSection = (sectionId: string) => {
  docStore.loadSection(props.id, sectionId)
}

function handleGenChunk(e: CustomEvent) {
  const { chunk } = e.detail
  editorRef.value.insertContent(chunk)
}

function handleGenDone(_e: CustomEvent) {
}

function onConfirmGeneration() {
  genStore.confirmGeneration()
  docStore.updateOutlineTree(props.id, genStore.outline)
}
</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #FDF6E3;
}

.navbar {
  height: 64px;
  padding: 0 32px;
  background: #FDF6E3;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.logo-icon {
  width: 44px;
  height: 44px;
  background: #C23B22;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.brand-texts {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.brand-zh {
  font-size: 22px;
  font-weight: 700;
  color: #3D2B1F;
  line-height: 1.2;
}

.brand-en {
  font-size: 12px;
  color: #8B7355;
  line-height: 1.3;
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.nav-btn {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: #F5EFE0;
  border: 0.7px solid #E0D5C0;
  cursor: pointer;
  transition: all 0.2s;
  overflow: hidden;
  white-space: nowrap;
  color: #5C4033;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-btn:hover {
  width: 90px;
  background: #C23B22;
  border-color: transparent;
  color: #fff;
}

.nav-btn-content {
  display: flex;
  align-items: center;
  gap: 6px;
}

.nav-btn-label {
  font-size: 13px;
  font-weight: 500;
  color: inherit;
  display: none;
}

.nav-btn:hover .nav-btn-label {
  display: inline;
}

.editor-body {
  flex: 1;
  display: flex;
  gap: 16px;
  padding: 0 32px 24px;
  overflow: hidden;
}

.left-panel {
  width: 380px;
  flex-shrink: 0;
  background: #FBF7EF;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.center-panel {
  flex: 1;
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.right-panel {
  width: 380px;
  flex-shrink: 0;
  background: #FBF7EF;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.center-md-editor {
  height: 100%;
}
</style>
