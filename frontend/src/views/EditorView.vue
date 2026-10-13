<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { getSession, updateOutline, lockOutline, generateSection, generateFull, getGenerateStream, getMarkdown, exportDocx } from '../api/client'
import { saveAs } from 'file-saver'
import OutlineTree from '../components/OutlineTree.vue'
import AIChat from '../components/AIChat.vue'
import GenerateFlowDialog from '../components/GenerateFlowDialog.vue'
import ProgressTracker from '../components/ProgressTracker.vue'
import MarkdownEditor from '../components/MarkdownEditor.vue'
import {
  RiRadarFill,
  RiQuestionLine,
  RiSettingsLine,
  RiDownloadLine,
  RiPlayCircleLine,
  RiLockLine,
  RiLockUnlockLine,
} from '@remixicon/vue'

const props = defineProps<{ id: string }>()
const router = useRouter()

// State
const session = ref<any>(null)
const outline = ref<any[]>([])
const markdown = ref('')
const activeSectionId = ref('')
const isGenerating = ref(false)
const generationProgress = ref(0)
const isLocked = ref(false)

const editorRef = ref<{
  setContent: (content: string) => void
  insertContent: (content: string) => void
  setContentControlContent: (sectionId: string, content: string) => void
}>({
  setContent: (content: string) => {
    markdown.value = content
  },
  insertContent: (content: string) => {
    markdown.value = (markdown.value || '') + content
  },
  setContentControlContent: (sectionId: string, content: string) => {
    const section = findSectionById(outline.value, sectionId)
    if (!section) {
      markdown.value = content
      return
    }
    markdown.value = setSectionMarkdownContent(markdown.value || '', section.title, section.level, content)
  },
})

const sectionBuffers = new Map<string, string>()

const goSettings = () => router.push('/settings')
const goHome = () => router.push('/')

let mdSaveTimer: ReturnType<typeof setTimeout> | null = null

onMounted(async () => {
  try {
    // 加载会话信息
    const res = await getSession(props.id)
    session.value = res.data
    isLocked.value = res.data.outline?.locked || false
    
    // 加载大纲
    if (res.data.outline?.sections) {
      outline.value = res.data.outline.sections
    }
    
    // 加载 Markdown
    if (res.data.markdown) {
      markdown.value = res.data.markdown
    }
  } catch (err) {
    console.error('Failed to load session:', err)
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

watch(() => markdown.value, (val) => {
  if (mdSaveTimer) clearTimeout(mdSaveTimer)
  mdSaveTimer = setTimeout(() => saveMarkdown(val), 1000)
})

async function saveMarkdown(md: string) {
  markdown.value = md
  // 保存到会话（TODO: 实现 saveSession API）
}

async function handleExportDocx() {
  if (mdSaveTimer) {
    clearTimeout(mdSaveTimer)
    mdSaveTimer = null
  }
  
  try {
    // 使用 file-saver 导出
    const blob = await exportDocx(props.id)
    const filename = `${session.value?.title || '标书'}_${new Date().toISOString().slice(0, 10)}.docx`
    saveAs(blob, filename)
  } catch (err) {
    console.error('Export failed:', err)
    alert('导出失败，请重试')
  }
}

const handleSelectSection = (sectionId: string) => {
  activeSectionId.value = sectionId
}

async function handleGenerateSection(sectionId: string) {
  if (isGenerating.value || isLocked.value) return
  
  isGenerating.value = true
  generationProgress.value = 0
  
  try {
    // 使用 SSE 流式生成
    const eventSource = getGenerateStream(props.id, sectionId)
    
    eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data)
      if (data.chunk) {
        // 更新章节内容
        const section = findSectionById(outline.value, sectionId)
        if (section) {
          const buffer = (sectionBuffers.get(sectionId) || '') + data.chunk
          sectionBuffers.set(sectionId, buffer)
          markdown.value = setSectionMarkdownContent(markdown.value || '', section.title, section.level, buffer)
          generationProgress.value = Math.min(90, generationProgress.value + 5)
        }
      }
    }
    
    eventSource.onerror = () => {
      eventSource.close()
      isGenerating.value = false
      generationProgress.value = 0
    }
    
    // 定时关闭（生成完成）
    setTimeout(() => {
      eventSource.close()
      isGenerating.value = false
      generationProgress.value = 0
      sectionBuffers.delete(sectionId)
      
      // 触发完成事件
      window.dispatchEvent(new CustomEvent('gen-done', { detail: { sectionId } }))
    }, 60000) // 60秒超时
  } catch (err) {
    console.error('Generate failed:', err)
    isGenerating.value = false
    generationProgress.value = 0
  }
}

async function handleGenerateAll() {
  if (isGenerating.value || isLocked.value) return
  
  isGenerating.value = true
  generationProgress.value = 0
  
  try {
    const res = await generateFull(props.id, {})
    const taskId = res.data.task_id
    
    // TODO: 轮询任务状态
    console.log('Full generation started:', taskId)
    
    setTimeout(() => {
      isGenerating.value = false
      generationProgress.value = 0
    }, 5000)
  } catch (err) {
    console.error('Full generate failed:', err)
    isGenerating.value = false
  }
}

function handleGenChunk(e: CustomEvent) {
  const { sectionId, chunk } = e.detail as { sectionId: string; chunk: string }
  const section = findSectionById(outline.value, sectionId)
  if (!section) return
  const buffer = (sectionBuffers.get(sectionId) || '') + (chunk || '')
  sectionBuffers.set(sectionId, buffer)
  markdown.value = setSectionMarkdownContent(markdown.value || '', section.title, section.level, buffer)
}

function handleGenDone(e: CustomEvent) {
  const { sectionId } = e.detail as { sectionId: string }
  const section = findSectionById(outline.value, sectionId)
  const buffer = sectionBuffers.get(sectionId)
  if (section) {
    markdown.value = setSectionMarkdownContent(markdown.value || '', section.title, section.level, buffer || '')
  }
  sectionBuffers.delete(sectionId)
  isGenerating.value = false
  generationProgress.value = 0
}

async function handleLockOutline() {
  try {
    await lockOutline(props.id)
    isLocked.value = true
  } catch (err) {
    alert(err?.response?.data?.error || '锁定失败')
  }
}

async function handleUnlockOutline() {
  isLocked.value = false
  // TODO: 实现解锁 API
}

// Helpers
function findSectionById(sections: any[], id: string): any {
  for (const s of sections) {
    if (s.section_id === id) return s
    if (s.children) {
      const found = findSectionById(s.children, id)
      if (found) return found
    }
  }
  return nil
}

function setSectionMarkdownContent(md: string, title: string, level: number, content: string): string {
  const heading = ''.padStart(level, '#') + ' ' + title
  // 简单实现：追加到末尾
  if (!md.trim()) {
    return heading + '\n\n' + content
  }
  return md + '\n\n' + heading + '\n\n' + content
}
</script>

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
        <button class="nav-btn" :title="isLocked ? '大纲已锁定' : '导出'" @click="handleExportDocx" :disabled="isGenerating">
          <span class="nav-btn-content">
            <RiDownloadLine size="20" />
            <span class="nav-btn-label">导出</span>
          </span>
        </button>
        <button class="nav-btn" title="生成全部" @click="handleGenerateAll" :disabled="isGenerating || isLocked">
          <span class="nav-btn-content">
            <RiPlayCircleLine size="20" />
            <span class="nav-btn-label">生成</span>
          </span>
        </button>
        <button class="nav-btn" :title="isLocked ? '解锁大纲' : '锁定大纲'" @click="isLocked ? handleUnlockOutline() : handleLockOutline()">
          <span class="nav-btn-content">
            <RiLockLine v-if="isLocked" size="20" />
            <RiLockUnlockLine v-else size="20" />
            <span class="nav-btn-label">{{ isLocked ? '已锁定' : '锁定' }}</span>
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
        <OutlineTree 
          :outline="outline" 
          :active-id="activeSectionId"
          @select="handleSelectSection"
          @generate="handleGenerateSection"
        />
        <ProgressTracker 
          v-if="isGenerating" 
          :progress="generationProgress"
          :is-locked="isLocked"
        />
      </aside>
      <section class="center-panel">
        <MarkdownEditor
          v-if="markdown !== undefined"
          v-model="markdown"
          class="center-md-editor"
        />
      </section>
      <aside class="right-panel">
        <AIChat 
          :session-id="props.id" 
          :editor-ref="editorRef"
          :outline="outline"
          @export-docx="handleExportDocx"
        />
      </aside>
    </main>

    <GenerateFlowDialog
      v-if="isGenerating"
      :progress="generationProgress"
      :is-locked="isLocked"
      @cancel="isGenerating = false"
    />
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: #FDF6E3;
}

.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 64px;
  padding: 0 32px;
  background: #FDF6E3;
  display: flex;
  align-items: center;
  justify-content: space-between;
  z-index: 50;
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

.nav-btn:hover:not(:disabled) {
  width: 90px;
  background: #C23B22;
  border-color: transparent;
  color: #fff;
}

.nav-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
  display: flex;
  height: calc(100vh - 64px);
  padding-top: 64px;
}

.left-panel {
  width: 280px;
  background: #fff;
  border-right: 1px solid #E8DCC8;
  overflow-y: auto;
  padding: 16px;
}

.center-panel {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.center-md-editor {
  max-width: 800px;
  margin: 0 auto;
}

.right-panel {
  width: 320px;
  background: #fff;
  border-left: 1px solid #E8DCC8;
  overflow-y: auto;
  padding: 16px;
}
</style>
