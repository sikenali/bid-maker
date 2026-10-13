<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, reactive, nextTick, defineComponent, h } from 'vue'
import { useRouter } from 'vue-router'
import {
  getSession,
  updateSession,
  generateFull,
  runOutlineSkill,
  getGenerateStream,
  getMarkdown,
  exportDocx,
} from '../api/client'
import { saveAs } from 'file-saver'
import { setSectionMarkdownContent } from '../utils/markdown'
import { useChatStore } from '../stores/chatStore'
import { useSettingsStore } from '../stores/settingsStore'
import {
  RiFileListLine,
  RiExpandLeftLine,
  RiEditLine,
  RiDeleteBinLine,
  RiDownload2Line,
  RiListUnordered,
  RiFileWordLine,
  RiMarkdownLine,
  RiSparklingLine,
  RiArrowDownSLine,
  RiSendPlaneFill,
  RiLoader4Line,
  RiPlayCircleLine,
  RiArrowDropRightLine,
  RiArrowDropDownLine,
} from '@remixicon/vue'

const props = defineProps<{ id: string }>()
const router = useRouter()
const chatStore = useChatStore()
const settingsStore = useSettingsStore()

// ── State ──────────────────────────────────────────────────────────
const session = ref<any>(null)
const outline = ref<any[]>([])
const markdown = ref('')
const activeSectionId = ref('')
const isGenerating = ref(false)
const generationProgress = ref(0)
const isGeneratingOutline = ref(false)
const isExporting = ref(false)
const editorMode = ref<'word' | 'markdown'>('word')
const isOutlineCollapsed = ref(false)
const outlineExpanded = reactive<Record<string, boolean>>({})
const chatInput = ref('')
const modelDropdownOpen = ref(false)
const lastSync = ref<string>('刚刚')
const toast = ref<{ show: boolean; text: string; type: 'info' | 'error' }>({ show: false, text: '', type: 'info' })

const sectionBuffers = new Map<string, string>()
let mdSaveTimer: ReturnType<typeof setTimeout> | null = null
let genTimeout: ReturnType<typeof setTimeout> | null = null
let sses: EventSource[] = []

// ── Computed ───────────────────────────────────────────────────────
const visibleOutline = computed(() => {
  if (!isOutlineCollapsed.value) return outline.value
  return outline.value.map(s => ({ ...s, children: [] }))
})

const availableModels = computed(() => {
  const keys = settingsStore.apiKeys.filter(k => k.enabled)
  if (keys.length === 0) return settingsStore.allModels
  return keys.map((k: any) => ({ id: k.id, name: k.modelName || k.model, provider: k.provider, model: k.model }))
})

const currentModelName = computed(() => {
  const cur = availableModels.value.find(m => m.id === settingsStore.selectedModelId)
  return cur?.name || '选择模型'
})

// ── Lifecycle ──────────────────────────────────────────────────────
onMounted(async () => {
  try {
    const res = await getSession(props.id)
    session.value = res.data
    if (res.data.outline?.sections) {
      outline.value = normalizeOutline(res.data.outline.sections)
      for (const s of outline.value) outlineExpanded[s.section_id] = true
    }
    if (res.data.markdown) {
      markdown.value = res.data.markdown
    } else {
      try {
        const md = await getMarkdown(props.id)
        markdown.value = md.data.markdown || ''
      } catch {}
    }
    settingsStore.fetchModels()
  } catch (err) {
    console.error('Failed to load session:', err)
    showToast('加载失败，请重试', 'error')
  }

  window.addEventListener('gen-chunk', handleGenChunk as EventListener)
  window.addEventListener('gen-done', handleGenDone as EventListener)
})

onUnmounted(() => {
  if (mdSaveTimer) clearTimeout(mdSaveTimer)
  if (genTimeout) clearTimeout(genTimeout)
  sses.forEach(es => es.close())
  window.removeEventListener('gen-chunk', handleGenChunk as EventListener)
  window.removeEventListener('gen-done', handleGenDone as EventListener)
})

watch(() => markdown.value, (val) => {
  if (!val) return
  if (mdSaveTimer) clearTimeout(mdSaveTimer)
  lastSync.value = '正在同步...'
  mdSaveTimer = setTimeout(() => saveMarkdown(val), 1500)
})

watch(() => chatStore.messages.length, () => {
  nextTick(() => {
    const chatEl = document.querySelector('.chat-area')
    if (chatEl) chatEl.scrollTop = chatEl.scrollHeight
  })
})

// ── Persistence ────────────────────────────────────────────────────
async function saveMarkdown(md: string) {
  try {
    await updateSession(props.id, { markdown: md })
    lastSync.value = '刚刚'
  } catch {
    lastSync.value = '同步失败'
    setTimeout(() => { lastSync.value = '刚刚' }, 3000)
  }
}

// ── Toast ──────────────────────────────────────────────────────────
function showToast(text: string, type: 'info' | 'error' = 'info') {
  toast.value = { show: true, text, type }
  setTimeout(() => { toast.value.show = false }, 2600)
}

// ── Outline helpers ────────────────────────────────────────────────
// Backend returns sections flat with children: string[] of IDs.
// Normalize to nested tree: sections where children is an array of node objects.
function normalizeOutline(sections: any[]): any[] {
  if (!sections || sections.length === 0) return []
  // If already nested (children is array of objects), return as-is
  const sample = sections[0]
  if (sample?.children?.[0]?.section_id) return sections
  // Build map by ID
  const map = new Map<string, any>()
  for (const s of sections) map.set(s.section_id, { ...s, children: [] })
  const roots: any[] = []
  map.forEach((node) => {
    // Resolve child IDs to objects
    if (Array.isArray(node.children)) {
      const childIds: string[] = node.children
      node.children = childIds.map(id => map.get(id)).filter(Boolean)
    }
    if (node.parent_id && map.has(node.parent_id)) {
      const parent = map.get(node.parent_id)
      if (parent && parent.children && !parent.children.includes(node)) {
        // only add if not already resolved via parent's children list
        if (!parent.children.some((c: any) => c.section_id === node.section_id)) {
          parent.children.push(node)
        }
      }
    } else if (!node.parent_id) {
      roots.push(node)
    }
  })
  return roots
}

function toggleExpand(id: string) {
  outlineExpanded[id] = !outlineExpanded[id]
}

function selectSection(sectionId: string) {
  activeSectionId.value = sectionId
  nextTick(() => {
    const el = document.querySelector(`[data-section-id="${sectionId}"]`)
    if (el && 'scrollIntoView' in el) {
      ;(el as HTMLElement).scrollIntoView({ behavior: 'smooth', block: 'start' })
    }
  })
}

function deleteSection(sectionId: string) {
  outline.value = removeSection(outline.value, sectionId)
}

function removeSection(sections: any[], id: string): any[] {
  return sections.filter(s => s.section_id !== id).map(s => ({
    ...s,
    children: s.children ? removeSection(s.children, id) : s.children,
  }))
}

async function handleGenerateOutline() {
  if (isGeneratingOutline.value) return
  isGeneratingOutline.value = true
  try {
    const res = await runOutlineSkill(props.id, { mode: 'hybrid' })
    if (res.data.sections) {
      outline.value = normalizeOutline(res.data.sections)
      for (const s of outline.value) outlineExpanded[s.section_id] = true
    }
    showToast('大纲已生成')
  } catch (err: any) {
    showToast(err?.response?.data?.error || '生成大纲失败', 'error')
  } finally {
    isGeneratingOutline.value = false
  }
}

// ── Generation ─────────────────────────────────────────────────────
async function handleGenerateSection(sectionId: string) {
  if (isGenerating.value) return
  isGenerating.value = true
  generationProgress.value = 0

  try {
    const es = getGenerateStream(props.id, sectionId)
    sses.push(es)
    es.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.chunk) {
          const section = findSectionById(outline.value, sectionId)
          if (section) {
            const buffer = (sectionBuffers.get(sectionId) || '') + data.chunk
            sectionBuffers.set(sectionId, buffer)
            markdown.value = setSectionMarkdownContent(markdown.value || '', section.title, section.level, buffer)
            generationProgress.value = Math.min(95, generationProgress.value + 2)
          }
        } else if (data.done) {
          es.close()
          sses = sses.filter(x => x !== es)
          sectionBuffers.delete(sectionId)
          isGenerating.value = false
          generationProgress.value = 0
          window.dispatchEvent(new CustomEvent('gen-done', { detail: { sectionId } }))
        }
      } catch {}
    }
    es.onerror = () => {
      es.close()
      sses = sses.filter(x => x !== es)
      isGenerating.value = false
      generationProgress.value = 0
    }
    if (genTimeout) clearTimeout(genTimeout)
    genTimeout = setTimeout(() => {
      es.close()
      sses = sses.filter(x => x !== es)
      isGenerating.value = false
      generationProgress.value = 0
      sectionBuffers.delete(sectionId)
    }, 90000)
  } catch (err) {
    console.error('Generate failed:', err)
    isGenerating.value = false
    generationProgress.value = 0
  }
}

async function handleGenerateFull() {
  if (isGenerating.value) return
  isGenerating.value = true
  generationProgress.value = 0
  try {
    await generateFull(props.id, {})
    showToast('标书生成已启动')
  } catch (err: any) {
    showToast(err?.response?.data?.error || '生成失败', 'error')
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
  generationProgress.value = Math.min(95, generationProgress.value + 1)
}

function handleGenDone(e: CustomEvent) {
  const { sectionId } = e.detail as { sectionId: string }
  const section = findSectionById(outline.value, sectionId)
  const buffer = sectionBuffers.get(sectionId)
  if (section && buffer) {
    markdown.value = setSectionMarkdownContent(markdown.value || '', section.title, section.level, buffer)
  }
  sectionBuffers.delete(sectionId)
  isGenerating.value = false
  generationProgress.value = 0
}

function findSectionById(sections: any[], id: string): any {
  for (const s of sections) {
    if (s.section_id === id) return s
    if (s.children) {
      const found = findSectionById(s.children, id)
      if (found) return found
    }
  }
  return null
}

// ── Export ─────────────────────────────────────────────────────────
async function handleExport() {
  if (isExporting.value) return
  isExporting.value = true
  if (mdSaveTimer) { clearTimeout(mdSaveTimer); mdSaveTimer = null }
  try {
    const blob = await exportDocx(props.id)
    const filename = `${session.value?.title || '标书'}_${new Date().toISOString().slice(0, 10)}.docx`
    saveAs(blob, filename)
    showToast('导出成功')
  } catch (err) {
    showToast('导出失败，请重试', 'error')
  } finally {
    isExporting.value = false
  }
}

// ── Chat ───────────────────────────────────────────────────────────
async function handleSendChat() {
  const text = chatInput.value.trim()
  if (!text || chatStore.isSending) return
  chatInput.value = ''
  await chatStore.sendMessage(text, activeSectionId.value || undefined, props.id, settingsStore.selectedModelId)
}

function toggleModelDropdown() {
  modelDropdownOpen.value = !modelDropdownOpen.value
}

function selectModel(id: string) {
  settingsStore.setModel(id)
  modelDropdownOpen.value = false
}

// ── Markdown Renderer (Word-like rich text) ────────────────────────
interface Block { type: string; html: string }

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function inlineMd(s: string): string {
  let out = escapeHtml(s)
  out = out.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  out = out.replace(/(^|[^*])\*([^*]+)\*/g, '$1<em>$2</em>')
  out = out.replace(/`([^`]+)`/g, '<code>$1</code>')
  return out
}

const renderedBlocks = computed<Block[]>(() => {
  const lines = (markdown.value || '').split('\n')
  const blocks: Block[] = []
  let listBuf: string[] = []
  let quoteBuf: string[] = []
  let paraBuf: string[] = []
  let inCode = false
  let codeBuf: string[] = []

  const flushList = () => {
    if (listBuf.length) {
      const isOrdered = /^ \d+\./.test(listBuf[0]) || /^\d+\./.test(listBuf[0])
      const tag = isOrdered ? 'ol' : 'ul'
      blocks.push({
        type: 'list',
        html: `<${tag}>` + listBuf.map(l => `<li>${inlineMd(l.replace(/^[-*]\s+|^\d+\.\s+/, ''))}</li>`).join('') + `</${tag}>`,
      })
      listBuf = []
    }
  }
  const flushQuote = () => {
    if (quoteBuf.length) {
      blocks.push({
        type: 'quote',
        html: '<blockquote>' + quoteBuf.map(inlineMd).join('<br/>') + '</blockquote>',
      })
      quoteBuf = []
    }
  }
  const flushPara = () => {
    if (paraBuf.length) {
      blocks.push({ type: 'p', html: '<p>' + paraBuf.map(inlineMd).join('<br/>') + '</p>' })
      paraBuf = []
    }
  }

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    const t = line.trim()

    if (inCode) {
      if (t === '```') {
        blocks.push({ type: 'code', html: '<pre><code>' + escapeHtml(codeBuf.join('\n')) + '</code></pre>' })
        codeBuf = []
        inCode = false
      } else {
        codeBuf.push(line)
      }
      continue
    }

    if (t.startsWith('```')) { inCode = true; continue }

    if (t === '---' || t === '***' || t === '___') {
      flushList(); flushQuote(); flushPara()
      blocks.push({ type: 'hr', html: '<hr/>' })
      continue
    }

    const h1 = t.match(/^#\s+(.*)$/)
    const h2 = t.match(/^##\s+(.*)$/)
    const h3 = t.match(/^###\s+(.*)$/)
    const h4 = t.match(/^####\s+(.*)$/)

    if (h1) { flushList(); flushQuote(); flushPara(); blocks.push({ type: 'title', html: `<h1>${inlineMd(h1[1])}</h1>` }); continue }
    if (h2) { flushList(); flushQuote(); flushPara(); blocks.push({ type: 'h2', html: `<h2>${inlineMd(h2[1])}</h2>` }); continue }
    if (h3) { flushList(); flushQuote(); flushPara(); blocks.push({ type: 'h3', html: `<h3>${inlineMd(h3[1])}</h3>` }); continue }
    if (h4) { flushList(); flushQuote(); flushPara(); blocks.push({ type: 'h4', html: `<h4>${inlineMd(h4[1])}</h4>` }); continue }

    if (/^>\s?/.test(t)) {
      flushList(); flushPara()
      quoteBuf.push(t.replace(/^>\s?/, ''))
      continue
    }

    if (/^[-*]\s+/.test(t) || /^\d+\.\s+/.test(t)) {
      flushQuote(); flushPara()
      listBuf.push(t)
      continue
    }

    if (t === '') { flushList(); flushQuote(); flushPara(); continue }

    flushList(); flushQuote()
    paraBuf.push(t)
  }

  flushList(); flushQuote(); flushPara()
  return blocks
})

// ── Recursive Outline Node (inline via defineComponent) ───────────
const OutlineNodeRecursive = defineComponent({
  name: 'OutlineNodeRecursive',
  props: {
    sections: { type: Array, required: true },
    level: { type: Number, required: true },
    activeId: { type: String, default: '' },
    expanded: { type: Object, default: () => ({}) },
  },
  emits: ['select', 'toggle', 'delete', 'generate'],
  setup(props, { emit }) {
    return () =>
      props.sections.map((s: any) => {
        const isExpanded = !!props.expanded[s.section_id]
        const hasChildren = s.children && s.children.length > 0
        return h(
          'div',
          { class: 'outline-node-wrap', key: s.section_id },
          [
            h(
              'div',
              {
                class: ['outline-node', { active: props.activeId === s.section_id }],
                style: { paddingLeft: `${20 + props.level * 18}px` },
                onClick: () => emit('select', s.section_id),
              },
              [
                hasChildren
                  ? h(
                      'button',
                      {
                        class: 'expand-btn',
                        onClick: (e: MouseEvent) => { e.stopPropagation(); emit('toggle', s.section_id) },
                      },
                      [h(isExpanded ? RiArrowDropDownLine : RiArrowDropRightLine, { size: 14 })],
                    )
                  : h('span', { class: 'expand-spacer' }),
                h('span', { class: 'node-title' }, s.title),
                h(
                  'span',
                  { class: 'node-actions' },
                  [
                    h(
                      'button',
                      { class: 'action-btn', title: '生成', onClick: (e: MouseEvent) => { e.stopPropagation(); emit('generate', s.section_id) } },
                      [h(RiPlayCircleLine, { size: 14 })],
                    ),
                    h(
                      'button',
                      { class: 'action-btn', title: '编辑', onClick: (e: MouseEvent) => { e.stopPropagation(); emit('select', s.section_id) } },
                      [h(RiEditLine, { size: 14 })],
                    ),
                    h(
                      'button',
                      { class: 'action-btn action-btn-danger', title: '删除', onClick: (e: MouseEvent) => { e.stopPropagation(); emit('delete', s.section_id) } },
                      [h(RiDeleteBinLine, { size: 14 })],
                    ),
                  ],
                ),
              ],
            ),
            hasChildren && isExpanded
              ? h(OutlineNodeRecursive, {
                  sections: s.children,
                  level: props.level + 1,
                  activeId: props.activeId,
                  expanded: props.expanded,
                  onSelect: (id: string) => emit('select', id),
                  onToggle: (id: string) => emit('toggle', id),
                  onDelete: (id: string) => emit('delete', id),
                  onGenerate: (id: string) => emit('generate', id),
                })
              : null,
          ],
        )
      })
  },
})
</script>

<template>
  <div class="editor-page">
    <!-- Left: Outline Panel -->
    <aside class="outline-panel">
      <div class="outline-header">
        <div class="outline-title-wrap">
          <div class="outline-title-icon">
            <RiFileListLine size="20" color="#fff" />
          </div>
          <span class="outline-title">大纲</span>
        </div>
        <button class="icon-btn" @click="isOutlineCollapsed = !isOutlineCollapsed" :title="isOutlineCollapsed ? '展开全部' : '收起全部'">
          <RiExpandLeftLine size="18" />
        </button>
      </div>

      <div class="outline-tree">
        <div v-if="visibleOutline.length === 0" class="outline-empty">
          <RiListUnordered size="32" color="var(--color-text-faint)" />
          <span>暂无大纲节点</span>
          <p>点击下方"生成大纲"按钮开始</p>
        </div>
        <OutlineNodeRecursive
          v-else
          :sections="visibleOutline"
          :level="0"
          :active-id="activeSectionId"
          :expanded="outlineExpanded"
          @select="selectSection"
          @toggle="toggleExpand"
          @delete="deleteSection"
          @generate="handleGenerateSection"
        />
      </div>

      <div class="outline-footer">
        <button class="btn-export" @click="handleExport" :disabled="isExporting">
          <RiDownload2Line size="18" />
          <span>{{ isExporting ? '导出中...' : '导出文件' }}</span>
        </button>
      </div>
    </aside>

    <!-- Center: Editor -->
    <section class="editor-center">
      <div class="editor-toolbar">
        <div class="mode-switch">
          <button class="mode-pill" :class="{ active: editorMode === 'word' }" @click="editorMode = 'word'">
            <RiFileWordLine size="15" />
            <span>Word 模式</span>
          </button>
          <button class="mode-pill" :class="{ active: editorMode === 'markdown' }" @click="editorMode = 'markdown'">
            <RiMarkdownLine size="15" />
            <span>Markdown 模式</span>
          </button>
        </div>
        <div class="toolbar-right">
          <span v-if="isGenerating" class="gen-badge">
            <RiLoader4Line size="14" class="spin-icon" />
            生成中 {{ generationProgress }}%
          </span>
          <span v-else class="sync-status">
            <span class="dot"></span>
            {{ lastSync }}
          </span>
        </div>
      </div>

      <div class="editor-scroll">
        <div class="doc-card">
          <div v-if="editorMode === 'word'" class="doc-render">
            <div v-if="renderedBlocks.length === 0" class="doc-placeholder">
              <RiFileWordLine size="48" color="var(--color-text-faint)" />
              <p>暂无文档内容</p>
              <p class="hint">点击右下角"生成标书"按钮开始创作</p>
            </div>
            <template v-else>
              <div
                v-for="(b, i) in renderedBlocks"
                :key="i"
                class="block"
                v-html="b.html"
              />
            </template>
          </div>

          <textarea
            v-else
            v-model="markdown"
            class="doc-md-editor"
            placeholder="# 标书标题&#10;&#10;在此输入 Markdown 内容..."
            spellcheck="false"
          />
        </div>
      </div>

      <div class="editor-footer">
        <button class="btn-outline-md" @click="handleGenerateOutline" :disabled="isGeneratingOutline">
          <RiLoader4Line v-if="isGeneratingOutline" size="18" class="spin-icon" />
          <RiListUnordered v-else size="18" />
          <span>{{ isGeneratingOutline ? '生成中...' : '生成大纲' }}</span>
        </button>
        <button class="btn-primary-md" @click="handleGenerateFull" :disabled="isGenerating">
          <RiLoader4Line v-if="isGenerating" size="18" class="spin-icon" />
          <RiFileWordLine v-else size="18" />
          <span>{{ isGenerating ? '生成中...' : '生成标书' }}</span>
        </button>
      </div>
    </section>

    <!-- Right: Assistant Panel -->
    <aside class="assistant-panel">
      <div class="assistant-header">
        <div class="assistant-title">
          <RiSparklingLine size="22" color="var(--color-primary)" />
          <span>智能助手</span>
        </div>
        <div class="model-select" @click.stop="toggleModelDropdown">
          <span class="model-name">{{ currentModelName }}</span>
          <RiArrowDownSLine size="14" :style="{ transform: modelDropdownOpen ? 'rotate(180deg)' : '' }" />
        </div>
        <div v-if="modelDropdownOpen" class="model-dropdown" @click.stop>
          <div
            v-for="m in availableModels"
            :key="m.id"
            class="model-option"
            :class="{ active: m.id === settingsStore.selectedModelId }"
            @click="selectModel(m.id)"
          >
            <span>{{ m.name }}</span>
            <span class="model-provider">{{ m.provider }}</span>
          </div>
        </div>
      </div>

      <div class="chat-area">
        <div v-if="chatStore.messages.length === 0" class="chat-empty">
          <RiSparklingLine size="40" color="var(--color-text-faint)" />
          <p>有什么可以帮您？</p>
          <p class="hint">提问关于标书撰写、内容优化、模板推荐等</p>
        </div>
        <template v-else>
          <div
            v-for="(msg, i) in chatStore.messages"
            :key="i"
            class="msg-row"
            :class="msg.role === 'user' ? 'msg-user' : 'msg-ai'"
          >
            <div v-if="msg.role === 'ai'" class="avatar avatar-ai">
              <RiSparklingLine size="14" color="var(--color-primary)" />
            </div>
            <div class="bubble">{{ msg.content }}</div>
          </div>
        </template>
      </div>

      <div class="chat-input-wrap">
        <div class="chat-input-row">
          <input
            v-model="chatInput"
            class="chat-input"
            placeholder="请输入你的问题..."
            @keydown.enter.prevent="handleSendChat"
            :disabled="chatStore.isSending"
          />
          <button
            class="send-btn"
            @click="handleSendChat"
            :disabled="!chatInput.trim() || chatStore.isSending"
            title="发送"
          >
            <RiLoader4Line v-if="chatStore.isSending" size="16" class="spin-icon" />
            <RiSendPlaneFill v-else size="16" />
          </button>
        </div>
      </div>
    </aside>

    <transition name="fade">
      <div v-if="toast.show" class="toast" :class="toast.type">
        {{ toast.text }}
      </div>
    </transition>
  </div>
</template>

<style scoped>
.editor-page {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: var(--color-bg-secondary);
  font-family: "SourceHanSans-Regular", "SourceHanSans-Medium", var(--font-family-cn);
  color: var(--color-text-primary);
}

/* ── Outline Panel ─────────────────────────────────────────── */
.outline-panel {
  width: 350px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-white);
  border-right: 1px solid var(--color-border-soft);
}

.outline-header {
  height: 74px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--color-border-soft);
  flex-shrink: 0;
}

.outline-title-wrap {
  display: flex;
  align-items: center;
  gap: 12px;
}

.outline-title-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
}

.outline-title {
  font-size: 17px;
  font-weight: 600;
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
  color: var(--color-text-primary);
}

.icon-btn {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.icon-btn:hover {
  background: var(--color-bg-secondary);
  color: var(--color-primary);
}

.outline-tree {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
}

.outline-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: var(--color-text-secondary);
  gap: 12px;
  text-align: center;
}

.outline-empty p {
  font-size: 13px;
  color: var(--color-text-muted);
  margin: 0;
}

.outline-node-wrap {
  margin-bottom: 2px;
}

.outline-node {
  height: 46px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding-right: 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
  position: relative;
}

.outline-node:hover {
  background: var(--color-bg-secondary);
}

.outline-node.active {
  background: var(--color-primary-soft);
}

.outline-node.active .node-title {
  color: var(--color-primary);
  font-weight: 500;
}

.expand-btn {
  width: 22px;
  height: 22px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border-radius: 4px;
  transition: all 0.15s;
}

.expand-btn:hover {
  background: var(--color-border-soft);
  color: var(--color-primary);
}

.expand-spacer {
  width: 22px;
  flex-shrink: 0;
}

.node-title {
  flex: 1;
  font-size: 14px;
  color: var(--color-text-body);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.4;
}

.node-actions {
  display: none;
  gap: 4px;
  align-items: center;
}

.outline-node:hover .node-actions {
  display: flex;
}

.outline-node:hover .node-title {
  flex: 0 1 auto;
  max-width: 50%;
}

.action-btn {
  width: 26px;
  height: 26px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.action-btn:hover {
  background: var(--color-bg-white);
  color: var(--color-primary);
}

.action-btn-danger:hover {
  background: var(--color-error-soft);
  color: var(--color-error);
}

.outline-footer {
  height: 92px;
  padding: 20px 24px;
  border-top: 1px solid var(--color-border-soft);
  flex-shrink: 0;
}

.btn-export {
  width: 100%;
  height: 52px;
  border-radius: 10px;
  border: none;
  background: var(--color-primary);
  color: #fff;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(192, 57, 43, 0.2);
}

.btn-export:hover:not(:disabled) {
  background: #A93226;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(192, 57, 43, 0.3);
}

.btn-export:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ── Editor Center ─────────────────────────────────────────── */
.editor-center {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-secondary);
  overflow: hidden;
  min-width: 0;
}

.editor-toolbar {
  height: 46px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  background: var(--color-bg-secondary);
}

.mode-switch {
  display: flex;
  gap: 0;
  background: var(--color-bg-white);
  border: 1px solid var(--color-border);
  border-radius: 23px;
  padding: 3px;
}

.mode-pill {
  height: 38px;
  padding: 0 18px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  border-radius: 20px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;
}

.mode-pill.active {
  background: var(--color-primary);
  color: #fff;
  box-shadow: 0 1px 4px rgba(192, 57, 43, 0.25);
}

.mode-pill:not(.active):hover {
  background: var(--color-bg-secondary);
  color: var(--color-text-primary);
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.gen-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 14px;
  background: var(--color-primary-soft);
  color: var(--color-primary);
  font-size: 12px;
  font-weight: 500;
}

.sync-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--color-text-muted);
  font-size: 12px;
}

.sync-status .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--color-success);
}

.editor-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.doc-card {
  background: var(--color-bg-white);
  border-radius: 8px;
  padding: 56px 60px;
  max-width: 990px;
  margin: 0 auto;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  min-height: 720px;
  border: 1px solid var(--color-border-soft);
}

.doc-render {
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  color: var(--color-text-body);
  line-height: 1.85;
  font-size: 16px;
}

.doc-render h1 {
  font-family: "SourceHanSans-Bold", "SourceHanSans-Medium", var(--font-family-cn);
  font-size: 32px;
  font-weight: 700;
  text-align: center;
  color: var(--color-text-primary);
  margin: 0 0 40px;
  line-height: 1.4;
  letter-spacing: 2px;
}

.doc-render h2 {
  font-size: 22px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 32px 0 16px;
  line-height: 1.5;
  padding-left: 12px;
  border-left: 4px solid var(--color-primary);
}

.doc-render h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 24px 0 12px;
  line-height: 1.5;
}

.doc-render h4 {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 20px 0 10px;
  line-height: 1.5;
}

.doc-render p {
  margin: 12px 0;
  text-indent: 2em;
  color: var(--color-text-body);
  font-size: 15.5px;
  line-height: 1.85;
}

.doc-render ul,
.doc-render ol {
  margin: 12px 0;
  padding-left: 40px;
}

.doc-render li {
  margin: 8px 0;
  padding-left: 4px;
  color: var(--color-text-body);
  font-size: 15.5px;
  line-height: 1.8;
}

.doc-render blockquote {
  margin: 16px 0;
  padding: 12px 20px;
  border-left: 3px solid var(--color-primary);
  background: var(--color-bg-secondary);
  color: var(--color-text-secondary);
  font-style: italic;
  border-radius: 0 6px 6px 0;
  font-size: 15px;
  line-height: 1.7;
}

.doc-render hr {
  border: none;
  border-top: 1px dashed var(--color-border);
  margin: 24px 0;
}

.doc-render code {
  font-family: 'Menlo', 'Consolas', monospace;
  background: var(--color-bg-secondary);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 14px;
  color: var(--color-primary);
}

.doc-render pre {
  margin: 16px 0;
  padding: 16px 20px;
  background: var(--color-bg-secondary);
  border-radius: 6px;
  overflow-x: auto;
}

.doc-render pre code {
  background: none;
  padding: 0;
  color: var(--color-text-body);
  font-size: 14px;
}

.doc-render strong {
  font-weight: 600;
  color: var(--color-text-primary);
}

.doc-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 120px 40px;
  color: var(--color-text-muted);
  gap: 12px;
}

.doc-placeholder p {
  margin: 0;
  text-indent: 0;
  color: var(--color-text-secondary);
  font-size: 15px;
}

.doc-placeholder .hint {
  font-size: 13px;
  color: var(--color-text-muted);
}

.doc-md-editor {
  width: 100%;
  min-height: 720px;
  border: none;
  outline: none;
  resize: none;
  font-family: 'Menlo', 'Consolas', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.8;
  color: var(--color-text-body);
  background: transparent;
}

.editor-footer {
  height: 82px;
  padding: 12px 24px;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 20px;
  flex-shrink: 0;
  background: var(--color-bg-secondary);
  border-top: 1px solid var(--color-border-soft);
}

.btn-outline-md,
.btn-primary-md {
  height: 58px;
  width: 280px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.btn-outline-md {
  background: var(--color-bg-white);
  color: var(--color-text-primary);
  border: 1px solid var(--color-border);
}

.btn-outline-md:hover:not(:disabled) {
  background: var(--color-primary-soft);
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.btn-primary-md {
  background: var(--color-primary);
  color: #fff;
  border: 1px solid var(--color-primary);
  box-shadow: 0 2px 8px rgba(192, 57, 43, 0.2);
}

.btn-primary-md:hover:not(:disabled) {
  background: #A93226;
  border-color: #A93226;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(192, 57, 43, 0.3);
}

.btn-outline-md:disabled,
.btn-primary-md:disabled {
  opacity: 0.55;
  cursor: not-allowed;
  transform: none;
}

/* ── Assistant Panel ───────────────────────────────────────── */
.assistant-panel {
  width: 460px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-white);
  border-left: 1px solid var(--color-border-soft);
  position: relative;
}

.assistant-header {
  height: 78px;
  padding: 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--color-border-soft);
  flex-shrink: 0;
  gap: 12px;
}

.assistant-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 17px;
  font-weight: 600;
  color: var(--color-primary);
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
}

.model-select {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 38px;
  padding: 0 14px;
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  color: var(--color-text-primary);
  transition: all 0.2s;
  font-family: inherit;
  min-width: 0;
}

.model-select:hover {
  border-color: var(--color-primary);
}

.model-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 130px;
}

.model-select > [data-icon] {
  transition: transform 0.2s;
}

.model-dropdown {
  position: absolute;
  top: 74px;
  right: 12px;
  width: 260px;
  max-height: 320px;
  overflow-y: auto;
  background: var(--color-bg-white);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
  z-index: 20;
  padding: 6px;
}

.model-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--color-text-body);
  transition: background 0.15s;
  gap: 8px;
}

.model-option:hover {
  background: var(--color-bg-secondary);
}

.model-option.active {
  background: var(--color-primary-soft);
  color: var(--color-primary);
  font-weight: 500;
}

.model-provider {
  color: var(--color-text-muted);
  font-size: 11px;
  flex-shrink: 0;
}

.chat-area {
  flex: 1;
  overflow-y: auto;
  padding: 20px 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.chat-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: var(--color-text-muted);
  gap: 8px;
  text-align: center;
  flex: 1;
}

.chat-empty p {
  font-size: 14px;
  color: var(--color-text-secondary);
  margin: 0;
}

.chat-empty .hint {
  font-size: 12px;
  color: var(--color-text-muted);
  max-width: 260px;
}

.msg-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  max-width: 100%;
}

.msg-row.msg-ai {
  align-self: flex-start;
}

.msg-row.msg-user {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: var(--color-primary-soft);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.bubble {
  max-width: 340px;
  padding: 12px 16px;
  border-radius: 12px;
  font-size: 14px;
  line-height: 1.65;
  word-break: break-word;
}

.msg-ai .bubble {
  background: #F7F3EA;
  color: var(--color-text-body);
  border-top-left-radius: 4px;
}

.msg-user .bubble {
  background: var(--color-primary);
  color: #fff;
  border-top-right-radius: 4px;
}

.chat-input-wrap {
  height: 88px;
  padding: 16px;
  border-top: 1px solid var(--color-border-soft);
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

.chat-input-row {
  position: relative;
  width: 100%;
  display: flex;
  align-items: center;
}

.chat-input {
  width: 100%;
  height: 56px;
  padding: 0 50px 0 18px;
  border: 1px solid var(--color-border);
  border-radius: 28px;
  background: var(--color-bg-secondary);
  outline: none;
  font-size: 14px;
  color: var(--color-text-primary);
  font-family: inherit;
  transition: all 0.2s;
}

.chat-input:focus {
  border-color: var(--color-primary);
  background: var(--color-bg-white);
}

.chat-input::placeholder {
  color: var(--color-text-muted);
}

.send-btn {
  position: absolute;
  right: 8px;
  width: 38px;
  height: 38px;
  border-radius: 50%;
  border: none;
  background: var(--color-primary);
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.send-btn:hover:not(:disabled) {
  background: #A93226;
  transform: scale(1.05);
}

.send-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  background: var(--color-text-muted);
}

/* ── Toast ─────────────────────────────────────────────────── */
.toast {
  position: fixed;
  top: 30px;
  left: 50%;
  transform: translateX(-50%);
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 14px;
  color: #fff;
  z-index: 100;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.toast.info { background: var(--color-success-strong); }
.toast.error { background: var(--color-error); }

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translate(-50%, -8px);
}

.spin-icon {
  animation: spin 1.2s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
