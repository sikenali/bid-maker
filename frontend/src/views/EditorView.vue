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
      </aside>
      <section class="center-panel">
        <DocxEditor
          v-if="docStore.docxBuffer"
          ref="editorRef"
          :document-buffer="docStore.docxBuffer"
          :show-menu-bar="true"
          :show-toolbar="true"
          :show-outline="true"
          :read-only="false"
          @ready="handleEditorReady"
        />
      </section>
      <aside class="right-panel">
        <AIChat :doc-id="props.id" :editor-ref="editorRef" :docx-buffer="docStore.docxBuffer" @export-docx="handleExportDocx" />
      </aside>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import OutlineTree from '../components/OutlineTree.vue'
import AIChat from '../components/AIChat.vue'
import {
  RiRadarFill,
  RiQuestionLine,
  RiSettingsLine,
  RiDownloadLine,
} from '@remixicon/vue'
import { DocxEditor } from '@eigenpal/docx-editor-vue'

const props = defineProps<{ id: string }>()
const router = useRouter()
const docStore = useDocumentStore()
const editorRef = ref<InstanceType<typeof DocxEditor> | null>(null)

const goSettings = () => router.push('/settings')
const goHome = () => router.push('/')

onMounted(() => {
  try {
    docStore.loadOutline(props.id)
  } catch (err) {
    console.error('Failed to load outline:', err)
  }
})

onUnmounted(() => {})

async function handleExportDocx() {
  if (!editorRef.value) {
    alert('编辑器尚未就绪')
    return
  }
  try {
    const buffer = await editorRef.value.save()
    const blob = new Blob([buffer], { type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' })
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

// Extract heading pmPos values by scanning the rendered document DOM.
// @eigenpal renders spans with data-pm-start for each paragraph element,
// allowing us to map section titles to ProseMirror positions accurately.
let headingSyncStarted = false

function handleEditorReady() {
  if (headingSyncStarted) return
  headingSyncStarted = true

  let attempts = 0
  const maxAttempts = 60
  const interval = 100
  const timer = setInterval(() => {
    attempts++
    try {
      const viewport = document.querySelector('.docx-editor-vue__pages-viewport') as HTMLElement | null
      if (!viewport) {
        if (attempts >= maxAttempts) clearInterval(timer)
        return
      }

      // Scan all paragraph-like elements for data-pm-start/pm-end spans
      const headingTexts = docStore.getFullOutline().map((s: any) => s.title).filter(Boolean)
      if (!headingTexts.length) {
        if (attempts >= maxAttempts) clearInterval(timer)
        return
      }

      // Find paragraph containers that have data-pm attributes
      const allParas = viewport.querySelectorAll('[data-pm-start][data-pm-end]')
      const pmHeadings: Array<{text: string; pmPos: number}> = []

      allParas.forEach((el: Element) => {
        const fullText = el.textContent?.trim()
        if (!fullText) return
        // Check if this element's text matches any heading title
        const matchIdx = headingTexts.findIndex(
          (title: string) => title && fullText.includes(title.trim())
        )
        if (matchIdx !== -1) {
          pmHeadings.push({
            text: fullText,
            pmPos: Number(el.getAttribute('data-pm-start')) || 0,
          })
        }
      })

      if (pmHeadings.length > 0) {
        // Match by title — use substring matching since Word formatting may add extra text
        const tree = docStore.outline
        for (const h of pmHeadings) {
          for (let i = 0; i < tree.length; i++) {
            const s = tree[i]
            if (s.title && h.text.includes(s.title)) {
              if (s.pmPos === undefined) s.pmPos = h.pmPos
              continue
            }
            if (s.children?.length) {
              for (const child of s.children) {
                if (child.title && h.text.includes(child.title)) {
                  if (child.pmPos === undefined) child.pmPos = h.pmPos
                }
              }
            }
          }
        }
        clearInterval(timer)
      } else if (attempts >= maxAttempts) {
        clearInterval(timer)
      }
    } catch {
      /* ignore */
    }
  }, interval)
}

const handleSelectSection = (sectionId: string) => {
  const pmPos = docStore.getSectionPmPos(sectionId)
  if (pmPos !== undefined && editorRef.value?.scrollToPosition) {
    editorRef.value.scrollToPosition(pmPos)
  }
  docStore.loadSection(props.id, sectionId)
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
</style>
