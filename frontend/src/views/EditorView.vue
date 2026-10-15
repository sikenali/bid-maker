<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, reactive, nextTick, defineComponent, h, type PropType } from 'vue'
import { useRouter } from 'vue-router'
import {
  getSession,
  updateSession,
  updateOutline,
  generateFull,
  getGenerateFullStream,
  runOutlineSkill,
  getGenerateStream,
  getMarkdown,
  exportDocx,
} from '../api/client'
import { getTemplateList, getExportSettings, type TemplateData, type TemplateSection } from '../api/settings'
import { saveAs } from 'file-saver'
import TopNav from '@/components/layout/TopNav.vue'
import { useChatStore } from '../stores/chatStore'
import { useSettingsStore } from '../stores/settingsStore'
import { DocxEditor, createEmptyDocument } from '@eigenpal/docx-editor-vue'
import { serializeDocumentBody } from '@eigenpal/docx-editor-core'
import { markdownToDocx } from '@/utils/markdownDocx'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
// eslint-disable-next-line @typescript-eslint/no-unused-vars,unused-imports/no-unused-imports
import { marked } from 'marked'
import TurndownService from 'turndown'
import {
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
  RiMoreLine,
  RiAddLine,
  RiArrowDownLine,
} from '@remixicon/vue'

const props = defineProps<{ id: string }>()
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const _router = useRouter()
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
const openMenuId = ref<string | null>(null)
const chatInput = ref('')
const modelDropdownOpen = ref(false)
const lastSync = ref<string>('')
const lastSaveTime = ref<string>('')
const toast = ref<{ show: boolean; text: string; type: 'info' | 'error' }>({ show: false, text: '', type: 'info' })

// Template state
const templates = ref<TemplateData[]>([])
const selectedTemplateId = ref<number | null>(null)
const templateDropdownOpen = ref(false)

// Skill command state
const skillPopupOpen = ref(false)
const skillPopupFilter = ref('')
const skillPopupIndex = ref(0)
const activeSkill = ref<any>(null)

const availableSkills = computed(() => {
  return settingsStore.allSkills.filter((s: any) => s.enabled !== false)
})

const filteredSkills = computed(() => {
  if (!skillPopupFilter.value) return availableSkills.value
  const filter = skillPopupFilter.value.toLowerCase()
  return availableSkills.value.filter((s: any) => 
    s.name.toLowerCase().includes(filter) || 
    s.description?.toLowerCase().includes(filter)
  )
})

// Editor refs
const docxEditorRef = ref<InstanceType<typeof DocxEditor> | null>(null)
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const mdEditorRef = ref<any>(null)
const wordDocLoading = ref(false)
let wordSaveTimer: ReturnType<typeof setTimeout> | null = null

// turndown service for converting HTML to markdown
const turndownService = new TurndownService({
  headingStyle: 'atx',
  codeBlockStyle: 'fenced',
  bulletListMarker: '-',
})

// tiptap editor for markdown mode
const mdEditor = useEditor({
  extensions: [StarterKit],
  content: '',
  editorProps: {
    attributes: {
      class: 'tiptap-editor',
    },
  },
  onUpdate: ({ editor }) => {
    const html = editor.getHTML()
    const md = turndownService.turndown(html)
    markdown.value = md
    scheduleSave()
  },
})

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
      // Load into tiptap editor
      nextTick(() => {
        if (mdEditor.value) {
          mdEditor.value.commands.setContent(res.data.markdown)
        }
        if (editorMode.value === 'word' && docxEditorRef.value) {
          syncWordMode()
        }
      })
      } else {
        try {
          const md = await getMarkdown(props.id)
          markdown.value = md.data.markdown || ''
          nextTick(() => {
            if (mdEditor.value) {
              mdEditor.value.commands.setContent(markdown.value)
            }
            if (editorMode.value === 'word' && docxEditorRef.value) {
              syncWordMode()
            }
          })
        } catch {}
      }
    settingsStore.fetchModels()
    // Load templates
    const tplRes = getTemplateList()
    templates.value = tplRes.templates || []
  } catch (err) {
    console.error('Failed to load session:', err)
    showToast('加载失败，请重试', 'error')
  }

  window.addEventListener('gen-chunk', handleGenChunk as EventListener)
  window.addEventListener('gen-done', handleGenDone as EventListener)
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  if (mdSaveTimer) clearTimeout(mdSaveTimer)
  if (outlineSaveTimer) clearTimeout(outlineSaveTimer)
  if (wordSaveTimer) clearTimeout(wordSaveTimer)
  if (genTimeout) clearTimeout(genTimeout)
  sses.forEach(es => es.close())
  window.removeEventListener('gen-chunk', handleGenChunk as EventListener)
  window.removeEventListener('gen-done', handleGenDone as EventListener)
  document.removeEventListener('click', handleClickOutside)
})

watch(() => editorMode.value, (mode) => {
  if (mode === 'markdown' && mdEditor.value && markdown.value) {
    nextTick(() => {
      mdEditor.value?.commands.setContent(markdown.value)
    })
  }
  if (mode === 'word' && markdown.value) {
    nextTick(() => syncWordMode())
  }
})

watch(() => chatStore.messages.length, () => {
  nextTick(() => {
    const chatEl = document.querySelector('.chat-area')
    if (chatEl) chatEl.scrollTop = chatEl.scrollHeight
  })
})

// ── Click outside handler ──────────────────────────────────────────
function handleClickOutside(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.template-select') && templateDropdownOpen.value) {
    templateDropdownOpen.value = false
  }
  if (!target.closest('.model-select') && modelDropdownOpen.value) {
    modelDropdownOpen.value = false
  }
  if (!target.closest('.skill-popup') && !target.closest('.chat-input')) {
    skillPopupOpen.value = false
    activeSkill.value = null
  }
  if (!target.closest('.node-menu-wrap') && openMenuId.value) {
    openMenuId.value = null
  }
}

// ── Template handlers ──────────────────────────────────────────────
function toggleTemplateDropdown() {
  templateDropdownOpen.value = !templateDropdownOpen.value
}

function selectTemplate(id: number) {
  selectedTemplateId.value = id
  templateDropdownOpen.value = false
}

function clearTemplate() {
  selectedTemplateId.value = null
}

const selectedTemplate = computed(() => {
  return templates.value.find(t => t.id === selectedTemplateId.value) || null
})

// ── Skill command handlers ─────────────────────────────────────────
function handleChatInput(e: Event) {
  const target = e.target as HTMLInputElement
  const value = target.value
  const match = value.match(/\/(\S*)$/)
  
  if (match) {
    skillPopupOpen.value = true
    skillPopupFilter.value = match[1]
    skillPopupIndex.value = 0
  } else {
    skillPopupOpen.value = false
    activeSkill.value = null
  }
}

function handleSkillKeyDown(e: KeyboardEvent) {
  if (!skillPopupOpen.value) return
  
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    skillPopupIndex.value = (skillPopupIndex.value + 1) % filteredSkills.value.length
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    skillPopupIndex.value = (skillPopupIndex.value - 1 + filteredSkills.value.length) % filteredSkills.value.length
  } else if (e.key === 'Enter' && skillPopupIndex.value < filteredSkills.value.length) {
    e.preventDefault()
    selectSkill(filteredSkills.value[skillPopupIndex.value])
  } else if (e.key === 'Escape') {
    skillPopupOpen.value = false
    activeSkill.value = null
  }
}

function selectSkill(skill: any) {
  activeSkill.value = skill
  chatInput.value = ''
  skillPopupOpen.value = false
  // Focus back to input
  nextTick(() => {
    (document.querySelector('.chat-input') as HTMLElement | null)?.focus()
  })
}

function removeActiveSkill() {
  activeSkill.value = null
}

// ── Persistence ────────────────────────────────────────────────────
function scheduleSave() {
  if (mdSaveTimer) clearTimeout(mdSaveTimer)
  lastSync.value = '正在同步...'
  mdSaveTimer = setTimeout(() => saveMarkdown(markdown.value), 1500)
}

async function saveMarkdown(md: string) {
  try {
    await updateSession(props.id, { markdown: md })
    lastSync.value = '已自动保存'
    const now = new Date()
    lastSaveTime.value = now.getHours().toString().padStart(2, '0') + ':' + now.getMinutes().toString().padStart(2, '0')
  } catch {
    lastSync.value = '同步失败'
    setTimeout(() => { lastSync.value = '已自动保存' }, 3000)
  }
}

// ── Toast ──────────────────────────────────────────────────────────
function showToast(text: string, type: 'info' | 'error' = 'info') {
  toast.value = { show: true, text, type }
  setTimeout(() => { toast.value.show = false }, 2600)
}

// ── Outline helpers ────────────────────────────────────────────────
function normalizeOutline(sections: any[]): any[] {
  if (!sections || sections.length === 0) return []
  const sample = sections[0]
  // Already nested (children are objects, not string IDs)
  if (sample?.children?.[0]?.section_id) return sections
  const map = new Map<string, any>()
  for (const s of sections) {
    // Preserve original children (string IDs) before replacing with empty array
    const originalChildren = Array.isArray(s.children) ? [...s.children] : []
    map.set(s.section_id, { ...s, children: [], _originalChildren: originalChildren })
  }
  const roots: any[] = []
  map.forEach((node) => {
    // Resolve string child IDs to actual objects
    if (node._originalChildren && node._originalChildren.length > 0) {
      node.children = node._originalChildren.map((id: string) => map.get(id)).filter(Boolean)
    }
    delete node._originalChildren
    if (node.parent_id && map.has(node.parent_id)) {
      const parent = map.get(node.parent_id)
      if (parent && parent.children && !parent.children.includes(node)) {
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

function toggleMenu(id: string) {
  openMenuId.value = openMenuId.value === id ? null : id
}

// ── Outline persistence ────────────────────────────────────────────
// 后端要求扁平结构(children 为 id 数组),嵌套对象需拍平后 PUT
let outlineSaveTimer: ReturnType<typeof setTimeout> | null = null
function flattenOutline(sections: any[]): any[] {
  const out: any[] = []
  const walk = (nodes: any[]) => {
    for (const n of nodes) {
      const kids: any[] = Array.isArray(n.children) ? n.children : []
      out.push({ ...n, children: kids.map((c: any) => c.section_id) })
      if (kids.length) walk(kids)
    }
  }
  walk(sections)
  return out
}

function scheduleOutlineSave() {
  if (outlineSaveTimer) clearTimeout(outlineSaveTimer)
  outlineSaveTimer = setTimeout(saveOutline, 800)
}

async function saveOutline() {
  try {
    await updateOutline(props.id, {
      outline_id: session.value?.outline?.outline_id || `outline-${Date.now()}`,
      source_mode: session.value?.outline?.source_mode || 'custom',
      locked: !!session.value?.outline?.locked,
      sections: flattenOutline(outline.value),
    } as any)
    lastSync.value = '大纲已同步'
  } catch (e: any) {
    showToast(e?.response?.data?.error || '大纲同步失败', 'error')
  }
}

function isOutlineLocked(): boolean {
  if (session.value?.outline?.locked) {
    showToast('大纲已锁定，无法修改结构', 'error')
    return true
  }
  return false
}

// 递归查找节点所在容器数组及其下标
function findNodeContainer(sections: any[], id: string): { arr: any[]; idx: number } | null {
  for (let i = 0; i < sections.length; i++) {
    if (sections[i].section_id === id) return { arr: sections, idx: i }
    if (sections[i].children?.length) {
      const r = findNodeContainer(sections[i].children, id)
      if (r) return r
    }
  }
  return null
}

function maxOrder(sections: any[]): number {
  let m = 0
  const walk = (nodes: any[]) => {
    for (const n of nodes) {
      if (typeof n.order === 'number' && n.order > m) m = n.order
      if (n.children?.length) walk(n.children)
    }
  }
  walk(sections)
  return m
}

function addLevelDelta(node: any, delta: number) {
  node.level = Math.min(Math.max(node.level + delta, 1), 3)
  for (const c of node.children || []) addLevelDelta(c, delta)
}

// 降级:挂到上一兄弟节点下;无上一兄弟则原地加深一级(上限 L3)
function demoteSection(id: string) {
  openMenuId.value = null
  if (isOutlineLocked()) return
  const found = findNodeContainer(outline.value, id)
  if (!found) return
  const { arr, idx } = found
  const node = arr[idx]
  if (idx > 0) {
    const prev = arr[idx - 1]
    const oldLevel = node.level
    arr.splice(idx, 1)
    prev.children = prev.children || []
    prev.children.push(node)
    node.parent_id = prev.section_id
    node.level = Math.min(prev.level + 1, 3)
    for (const c of node.children || []) addLevelDelta(c, node.level - oldLevel)
    outlineExpanded[prev.section_id] = true
  } else {
    if (node.level >= 3) {
      showToast('已是最深层级', 'error')
      return
    }
    addLevelDelta(node, 1)
  }
  activeSectionId.value = id
  scheduleOutlineSave()
}

// 新增:在当前节点后插入同级章节
function addSiblingSection(id: string) {
  openMenuId.value = null
  if (isOutlineLocked()) return
  const found = findNodeContainer(outline.value, id)
  if (!found) return
  const { arr, idx } = found
  const ref = arr[idx]
  const node = {
    section_id: `sec-custom-${Date.now()}`,
    title: '新章节',
    level: ref.level,
    order: maxOrder(outline.value) + 1,
    required: false,
    parent_id: ref.parent_id || null,
    source: 'manual',
    requirements: [],
    response_strategy: 'satisfy',
    response_basis: '',
    children: [],
    content: '',
    status: 'pending',
  }
  arr.splice(idx + 1, 0, node)
  outlineExpanded[node.section_id] = true
  activeSectionId.value = node.section_id
  scheduleOutlineSave()
  showToast('已新增章节，可点击标题编辑正文')
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
  openMenuId.value = null
  if (isOutlineLocked()) return
  outline.value = removeSection(outline.value, sectionId)
  if (activeSectionId.value === sectionId) activeSectionId.value = ''
  scheduleOutlineSave()
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
    // Use selected template if available
    if (selectedTemplate.value && selectedTemplate.value.content && selectedTemplate.value.content.length > 0) {
      const templateSections = selectedTemplate.value.content.map((s: TemplateSection, idx: number) => ({
        section_id: `sec-${Date.now()}-${idx}`,
        title: s.title,
        level: s.level,
        order: idx,
        required: true,
        parent_id: null,
        source: 'template',
        requirements: [],
        response_strategy: 'satisfy',
        response_basis: '',
        children: s.children || [],
        content: '',
        status: 'pending',
      }))
      outline.value = normalizeOutline(templateSections)
      for (const s of outline.value) outlineExpanded[s.section_id] = true
      showToast(`已使用模板「${selectedTemplate.value.name}」生成大纲`)
    } else {
      const res = await runOutlineSkill(props.id, { mode: 'auto' })
      if (res.data.sections) {
        outline.value = normalizeOutline(res.data.sections)
        for (const s of outline.value) outlineExpanded[s.section_id] = true
      }
      showToast('大纲已生成')
    }
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
            if (editorMode.value === 'markdown' && mdEditor.value) {
              mdEditor.value.commands.setContent(markdown.value)
            }
            if (editorMode.value === 'word' && docxEditorRef.value) {
              syncWordMode()
            }
            generationProgress.value = Math.min(95, generationProgress.value + 2)
          }
        } else if (data.done) {
          es.close()
          sses = sses.filter(x => x !== es)
          sectionBuffers.delete(sectionId)
          isGenerating.value = false
          generationProgress.value = 0
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
    const es = getGenerateFullStream(props.id)
    sses.push(es)
    es.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.chunk) {
          const section = findSectionById(outline.value, data.section_id)
          if (section) {
            const buffer = (sectionBuffers.get(data.section_id) || '') + (data.chunk || '')
            sectionBuffers.set(data.section_id, buffer)
            markdown.value = setSectionMarkdownContent(markdown.value || '', section.title, section.level, buffer)
            if (editorMode.value === 'markdown' && mdEditor.value) {
              mdEditor.value.commands.setContent(markdown.value)
            }
            generationProgress.value = data.progress || Math.min(95, generationProgress.value + 2)
          }
        } else if (data.progress !== undefined) {
          generationProgress.value = data.progress
        } else if (data.done) {
          es.close()
          sses = sses.filter(x => x !== es)
          sectionBuffers.clear()
          isGenerating.value = false
          generationProgress.value = 100
          // Reload session to pick up backend-persisted content
          setTimeout(() => reloadSession(), 300)
        } else if (data.error) {
          es.close()
          sses = sses.filter(x => x !== es)
          isGenerating.value = false
          showToast(`生成失败：${data.error}`, 'error')
        }
      } catch {}
    }
    es.onerror = () => {
      es.close()
      sses = sses.filter(x => x !== es)
      isGenerating.value = false
      generationProgress.value = 0
      showToast('生成连接中断，请重试', 'error')
    }
    if (genTimeout) clearTimeout(genTimeout)
    genTimeout = setTimeout(() => {
      es.close()
      sses = sses.filter(x => x !== es)
      isGenerating.value = false
      generationProgress.value = 0
      sectionBuffers.clear()
      showToast('生成超时，请重试', 'error')
    }, 300000)
    showToast('全文生成已启动，请稍候…')
  } catch (err: any) {
    showToast(err?.response?.data?.error || '生成失败', 'error')
    isGenerating.value = false
  }
}

// Reload session content after full generation completes
async function reloadSession() {
  try {
    const res = await getSession(props.id)
    if (res.data.markdown) {
      markdown.value = res.data.markdown
      if (mdEditor.value) mdEditor.value.commands.setContent(markdown.value)
      if (editorMode.value === 'word' && docxEditorRef.value) syncWordMode()
    }
  } catch {}
}

function handleGenChunk(e: CustomEvent) {
  const { sectionId, chunk } = e.detail as { sectionId: string; chunk: string }
  const section = findSectionById(outline.value, sectionId)
  if (!section) return
  const buffer = (sectionBuffers.get(sectionId) || '') + (chunk || '')
  sectionBuffers.set(sectionId, buffer)
  markdown.value = setSectionMarkdownContent(markdown.value || '', section.title, section.level, buffer)
  if (editorMode.value === 'markdown' && mdEditor.value) {
    mdEditor.value.commands.setContent(markdown.value)
  }
  // Word mode only synced on gen-done, not per-chunk (performance)
  generationProgress.value = Math.min(95, generationProgress.value + 1)
}

function handleGenDone(e: CustomEvent) {
  const { sectionId } = e.detail as { sectionId: string }
  const section = findSectionById(outline.value, sectionId)
  const buffer = sectionBuffers.get(sectionId)
  if (section && buffer) {
    markdown.value = setSectionMarkdownContent(markdown.value || '', section.title, section.level, buffer)
    if (editorMode.value === 'markdown' && mdEditor.value) {
      mdEditor.value.commands.setContent(markdown.value)
    }
    if (editorMode.value === 'word' && docxEditorRef.value) {
      syncWordMode()
    }
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
  const title = session.value?.title || '标书'
  const dateStr = new Date().toISOString().slice(0, 10)
  try {
    const exportSettings = await getExportSettings()
    const format = exportSettings.format || 'docx'
    if (format === 'markdown' || format === 'md') {
      const mdRes = await getMarkdown(props.id)
      const mdContent = mdRes.data.markdown || markdown.value || ''
      if (!mdContent.trim()) {
        showToast('没有可导出的内容', 'error')
        return
      }
      const blob = new Blob([mdContent], { type: 'text/markdown;charset=utf-8' })
      saveAs(blob, `${title}_${dateStr}.md`)
      showToast('Markdown 导出成功')
    } else {
      const blob = await exportDocx(props.id)
      saveAs(blob, `${title}_${dateStr}.docx`)
      showToast('Word 导出成功')
    }
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
  
  let message = text
  if (activeSkill.value) {
    // Prepend skill prompt to message
    const skillPrompt = activeSkill.value.prompt || activeSkill.value.description || ''
    message = skillPrompt ? `${skillPrompt}\n\n${text}` : text
    activeSkill.value = null
  }
  
  chatInput.value = ''
  await chatStore.sendMessage(message, activeSectionId.value || undefined, props.id, settingsStore.selectedModelId)
}

function toggleModelDropdown() {
  modelDropdownOpen.value = !modelDropdownOpen.value
}

function selectModel(id: string) {
  settingsStore.setModel(id)
  modelDropdownOpen.value = false
}

// ── Markdown helpers ───────────────────────────────────────────────
function escapeRegExp(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function setSectionMarkdownContent(md: string, title: string, level: number, content: string): string {
  const prefix = '#'.repeat(level) + ' ' + title
  const safePrefix = escapeRegExp(prefix)
  const sectionPattern = new RegExp(`^${safePrefix}\\.?\\s+\\n[\\s\\S]*?(?=^#{1,${level}} |\\Z)`, 'm')
  const newSection = `${prefix}\n\n${content}\n\n`
  if (sectionPattern.test(md)) {
    return md.replace(sectionPattern, newSection)
  }
  return md + '\n\n' + newSection
}

// ── DocxEditor handlers ────────────────────────────────────────────
async function syncWordMode() {
  if (!markdown.value || !docxEditorRef.value) return
  wordDocLoading.value = true
  try {
    const buf = await markdownToDocx(markdown.value)
    await docxEditorRef.value.loadDocumentBuffer(buf)
  } catch (e) {
    console.error('Word mode sync failed:', e)
  } finally {
    wordDocLoading.value = false
  }
}

// Word 模式编辑内容回写 markdown(保留标题层级,防抖保存);
// 程序化加载(syncWordMode)期间跳过,避免回环覆盖
function handleWordChange(_doc: any) {
  if (wordDocLoading.value) return
  try {
    const xml: string = serializeDocumentBody(_doc.package.document)
    const unesc = (s: string) => s.replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').replace(/&amp;/g, '&')
    const paras = xml.match(/<w:p[\s>][\s\S]*?<\/w:p>/g) || []
    const lines: string[] = []
    for (const p of paras) {
      const style = /<w:pStyle[^>]*w:val="([^"]+)"/.exec(p)?.[1] || 'Normal'
      const text = [...p.matchAll(/<w:t[^>]*>([^<]*)<\/w:t>/g)].map((m) => unesc(m[1])).join('').trim()
      if (!text) continue
      const prefix = style === 'Heading1' ? '# ' : style === 'Heading2' ? '## ' : style === 'Heading3' ? '### ' : ''
      lines.push(prefix + text)
    }
    const mdContent = lines.join('\n\n')
    if (mdContent && mdContent !== markdown.value) {
      markdown.value = mdContent
      scheduleSave()
    }
  } catch {}
}

function handleModeSwitch(mode: 'word' | 'markdown') {
  editorMode.value = mode
}

// ── Recursive Outline Node ─────────────────────────────────────────
// @ts-ignore
const OutlineNodeRecursive = defineComponent({
  name: 'OutlineNodeRecursive',
  props: {
    sections: { type: Array, required: true },
    level: { type: Number, required: true },
    activeId: { type: String, default: '' },
    expanded: { type: Object, default: () => ({}) },
    openMenuId: { type: String as PropType<string | null>, default: null },
  },
  emits: ['select', 'toggle', 'delete', 'generate', 'menu', 'demote', 'add'],
  setup(props: any, { emit }: any) {
    // @ts-ignore
    return (): any =>
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
                class: ['outline-node', { active: props.activeId === s.section_id, child: props.level > 0 }],
                style: { paddingLeft: `${16 + props.level * 20}px` },
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
                      [h(isExpanded ? RiArrowDropDownLine : RiArrowDropRightLine, { size: '14' })],
                    )
                  : h('span', { class: 'expand-spacer' }),
                h('span', { class: 'node-title' }, s.title),
                h(
                  'span',
                  { class: 'node-menu-wrap', onClick: (e: MouseEvent) => e.stopPropagation() },
                  [
                    h(
                      'button',
                      {
                        class: ['node-menu-btn', { open: props.openMenuId === s.section_id }],
                        title: '更多操作',
                        onClick: (e: MouseEvent) => { e.stopPropagation(); emit('menu', s.section_id) },
                      },
                      [h(RiMoreLine, { size: '16' })],
                    ),
                    props.openMenuId === s.section_id
                      ? h('div', { class: 'node-menu' }, [
                          h(
                            'button',
                            {
                              class: 'node-menu-item',
                              onClick: (e: MouseEvent) => { e.stopPropagation(); emit('demote', s.section_id) },
                            },
                            [h(RiArrowDownLine, { size: '15' }), h('span', '降级')],
                          ),
                          h(
                            'button',
                            {
                              class: 'node-menu-item',
                              onClick: (e: MouseEvent) => { e.stopPropagation(); emit('add', s.section_id) },
                            },
                            [h(RiAddLine, { size: '15' }), h('span', '新增')],
                          ),
                          h('div', { class: 'node-menu-divider' }),
                          h(
                            'button',
                            {
                              class: 'node-menu-item danger',
                              onClick: (e: MouseEvent) => { e.stopPropagation(); emit('delete', s.section_id) },
                            },
                            [h(RiDeleteBinLine, { size: '15' }), h('span', '删除')],
                          ),
                        ])
                      : null,
                  ],
                ),
                h(
                  'span',
                  { class: 'node-actions' },
                  [
                    h(
                      'button',
                      { class: 'action-btn', title: '生成', onClick: (e: MouseEvent) => { e.stopPropagation(); emit('generate', s.section_id) } },
                      [h(RiPlayCircleLine, { size: '14' })],
                    ),
                    h(
                      'button',
                      { class: 'action-btn', title: '编辑', onClick: (e: MouseEvent) => { e.stopPropagation(); emit('select', s.section_id) } },
                      [h(RiEditLine, { size: '14' })],
                    ),
                    h(
                      'button',
                      { class: 'action-btn action-btn-danger', title: '删除', onClick: (e: MouseEvent) => { e.stopPropagation(); emit('delete', s.section_id) } },
                      [h(RiDeleteBinLine, { size: '14' })],
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
                  openMenuId: props.openMenuId,
                  onSelect: (id: string) => emit('select', id),
                  onToggle: (id: string) => emit('toggle', id),
                  onDelete: (id: string) => emit('delete', id),
                  onGenerate: (id: string) => emit('generate', id),
                  onMenu: (id: string) => emit('menu', id),
                  onDemote: (id: string) => emit('demote', id),
                  onAdd: (id: string) => emit('add', id),
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
    <TopNav title="编辑标书" :show-back="true" />
    <div class="editor-layout">
      <!-- Left: Outline Panel -->
    <aside class="outline-panel">
      <div class="outline-header">
        <span class="outline-title">标书大纲</span>
        <span class="outline-count" @click="isOutlineCollapsed = !isOutlineCollapsed">{{ outline.length }} 节</span>
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
          :open-menu-id="openMenuId"
          @select="selectSection"
          @toggle="toggleExpand"
          @delete="deleteSection"
          @generate="handleGenerateSection"
          @menu="toggleMenu"
          @demote="demoteSection"
          @add="addSiblingSection"
        />
      </div>

      <div class="outline-footer">
        <div class="template-select" @click.stop="toggleTemplateDropdown">
          <span class="template-label">模板</span>
          <span class="template-name" :class="{ 'placeholder': !selectedTemplate }">
            {{ selectedTemplate ? selectedTemplate.name : '选择模板' }}
          </span>
          <RiArrowDownSLine size="14" :style="{ transform: templateDropdownOpen ? 'rotate(180deg)' : '' }" />
        </div>
        <div v-if="templateDropdownOpen" class="template-dropdown" @click.stop>
          <div
            v-for="tpl in templates"
            :key="tpl.id"
            class="template-option"
            :class="{ active: tpl.id === selectedTemplateId }"
            @click="selectTemplate(tpl.id!)"
          >
            <span class="template-option-name">{{ tpl.name }}</span>
            <span class="template-option-category">{{ tpl.category }}</span>
          </div>
          <div v-if="templates.length === 0" class="template-empty">暂无模板</div>
          <div v-if="selectedTemplateId" class="template-clear" @click="clearTemplate">清除选择</div>
        </div>
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
          <button class="mode-pill" :class="{ active: editorMode === 'word' }" @click="handleModeSwitch('word')">
            <RiFileWordLine size="15" />
            <span>Word 模式</span>
          </button>
          <button class="mode-pill" :class="{ active: editorMode === 'markdown' }" @click="handleModeSwitch('markdown')">
            <RiMarkdownLine size="15" />
            <span>Markdown 模式</span>
          </button>
        </div>
        <div class="toolbar-right">
          <span v-if="isGenerating" class="gen-badge">
            <RiLoader4Line size="14" class="spin-icon" />
            生成中 {{ generationProgress }}%
          </span>
          <span v-else-if="lastSaveTime" class="sync-status">已自动保存 · {{ lastSaveTime }}</span>
        </div>
      </div>

      <div class="editor-scroll">
        <div class="doc-card">
          <!-- Word Mode: DocxEditor -->
          <div v-if="editorMode === 'word'" class="docx-editor-wrap">
            <DocxEditor
              ref="docxEditorRef"
              :document="createEmptyDocument()"
              :show-toolbar="true"
              :show-menu-bar="true"
              :show-ruler="true"
              document-name="标书文档"
              mode="editing"
              @change="handleWordChange"
            />
            <div v-if="wordDocLoading" class="word-sync-badge">
              <RiLoader4Line size="14" class="spin-icon" />
              正在加载 Word 内容...
            </div>
          </div>

          <!-- Markdown Mode: Tiptap Editor -->
          <div v-else class="md-editor-wrap">
            <EditorContent
              v-if="mdEditor"
              :editor="mdEditor"
              class="tiptap-editor-container"
            />
          </div>
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
          <RiSparklingLine size="22" color="var(--color-info)" />
          <span>ChatGPT 助手</span>
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
            <div class="bubble">{{ msg.content }}</div>
          </div>
        </template>
      </div>

      <div class="chat-input-wrap">
        <div v-if="activeSkill" class="skill-chip">
          <RiCommandLine size="14" />
          <span>{{ activeSkill.name }}</span>
          <button class="skill-chip-close" @click="removeActiveSkill">
            <RiDeleteBinLine size="12" />
          </button>
        </div>
        <div class="chat-input-row">
          <input
            v-model="chatInput"
            class="chat-input"
            placeholder="描述你的需求，输入 / 调用技能…"
            @input="handleChatInput"
            @keydown="handleSkillKeyDown"
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
        <div v-if="skillPopupOpen && filteredSkills.length > 0" class="skill-popup">
          <div
            v-for="(skill, i) in filteredSkills"
            :key="skill.id"
            class="skill-option"
            :class="{ active: i === skillPopupIndex }"
            @mouseenter="skillPopupIndex = i"
            @click="selectSkill(skill)"
          >
            <div class="skill-option-name">{{ skill.name }}</div>
            <div class="skill-option-desc">{{ skill.description }}</div>
          </div>
        </div>
      </div>
    </aside>

    </div>

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
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  background: var(--color-bg-secondary);
  font-family: "SourceHanSans-Regular", "SourceHanSans-Medium", var(--font-family-cn);
  color: var(--color-text-primary);
}

.editor-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
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
  padding: 24px 24px 20px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--color-border-soft);
  flex-shrink: 0;
}

.outline-title {
  font-size: 20px;
  font-weight: 700;
  font-family: "SourceHanSans-Bold", var(--font-family-cn);
  color: var(--color-text-primary);
}

.outline-count {
  font-size: 15px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  color: var(--color-text-muted);
  cursor: pointer;
  user-select: none;
  transition: color 0.2s;
}

.outline-count:hover {
  color: var(--color-primary);
}

.outline-tree {
  flex: 1;
  overflow-y: auto;
  padding: 0 16px;
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
  margin-bottom: 0;
}

.outline-node {
  height: 48px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.2s;
  position: relative;
}

.outline-node:hover {
  background: var(--color-bg-secondary);
}

.outline-node.active {
  background: rgba(247, 239, 223, 1);
}

.outline-node.child .node-title {
  font-size: 15px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  color: var(--color-text-secondary);
}

.outline-node.active.child {
  background: rgba(247, 239, 223, 1);
}

.outline-node.active.child .node-title {
  color: var(--color-text-primary);
  font-weight: 400;
}

.outline-node.child {
  height: 44px;
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
  font-size: 16px;
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.2;
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

/* ── 节点更多操作菜单(圆圈三点按钮 + 下拉子菜单) ─────────────────── */
.node-menu-wrap {
  position: relative;
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
}

.node-menu-btn {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  border: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}

.node-menu-btn:hover,
.node-menu-btn.open {
  border-color: var(--color-primary);
  color: var(--color-primary);
  background: var(--color-bg-secondary);
}

.node-menu {
  position: absolute;
  right: 0;
  top: calc(100% + 6px);
  min-width: 132px;
  background: var(--color-bg-white);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
  padding: 6px;
  z-index: 60;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.node-menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 9px 12px;
  border: none;
  border-radius: 6px;
  background: transparent;
  font-size: 14px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  color: var(--color-text-body);
  cursor: pointer;
  text-align: left;
  transition: background 0.15s;
  white-space: nowrap;
}

.node-menu-item:hover {
  background: var(--color-bg-secondary);
}

.node-menu-item.danger {
  color: var(--color-error);
}

.node-menu-item.danger:hover {
  background: var(--color-error-soft);
}

.node-menu-divider {
  height: 1px;
  background: var(--color-border-soft);
  margin: 4px 6px;
}

.outline-footer {
  padding: 16px 24px 24px 24px;
  border-top: 1px solid var(--color-border-soft);
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.template-select {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 44px;
  padding: 0 16px;
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.template-select:hover {
  border-color: var(--color-primary);
}

.template-label {
  font-size: 13px;
  color: var(--color-text-muted);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}

.template-name {
  flex: 1;
  font-size: 14px;
  color: var(--color-text-primary);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.template-name.placeholder {
  color: var(--color-text-faint);
}

.template-select > [data-icon] {
  color: var(--color-text-secondary);
  transition: transform 0.2s;
  flex-shrink: 0;
}

.template-dropdown {
  position: absolute;
  bottom: 56px;
  left: 0;
  right: 0;
  background: var(--color-bg-white);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
  z-index: 20;
  max-height: 280px;
  overflow-y: auto;
  padding: 6px;
}

.template-option {
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

.template-option:hover {
  background: var(--color-bg-secondary);
}

.template-option.active {
  background: var(--color-primary-soft);
  color: var(--color-primary);
  font-weight: 500;
}

.template-option-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.template-option-category {
  color: var(--color-text-muted);
  font-size: 11px;
  flex-shrink: 0;
}

.template-empty {
  padding: 20px;
  text-align: center;
  font-size: 13px;
  color: var(--color-text-muted);
}

.template-clear {
  padding: 8px 12px;
  font-size: 12px;
  color: var(--color-error);
  cursor: pointer;
  border-top: 1px solid var(--color-border-soft);
  text-align: center;
  margin-top: 4px;
}

.template-clear:hover {
  background: var(--color-error-soft);
}

.btn-export {
  width: 100%;
  height: 52px;
  border-radius: 10px;
  border: 1px solid var(--color-border);
  background: var(--color-bg-white);
  color: var(--color-text-body);
  font-size: 16px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  transition: all 0.2s;
}

.btn-export:hover:not(:disabled) {
  border-color: var(--color-border-hover);
  background: var(--color-bg-secondary);
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
  padding: 24px 60px 0 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  background: var(--color-bg-secondary);
}

.mode-switch {
  display: flex;
  gap: 4px;
  background: var(--color-bg-secondary);
  border-radius: 10px;
  padding: 4px;
}

.mode-pill {
  height: 38px;
  padding: 0 16px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  border-radius: 8px;
  cursor: pointer;
  font-size: 15px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s;
}

.mode-pill.active {
  background: var(--color-bg-white);
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
}

.mode-pill.active [data-icon] {
  color: rgba(43, 87, 154, 1);
}

.mode-pill:not(.active):hover {
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
  font-size: 14px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  color: var(--color-text-faint);
}

.editor-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 16px 60px 0 60px;
}

.doc-card {
  background: var(--color-bg-white);
  border-radius: 14px;
  padding: 48px 56px;
  max-width: 990px;
  margin: 0 auto;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  min-height: 720px;
  border: 1px solid var(--color-border-soft);
  display: flex;
  flex-direction: column;
}

.docx-editor-wrap {
  flex: 1;
  overflow: auto;
  position: relative;
}

.word-sync-badge {
  position: absolute;
  bottom: 16px;
  right: 24px;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: 14px;
  background: var(--color-primary-soft, #fbe9e7);
  color: var(--color-primary, #c0392b);
  font-size: 12px;
  font-weight: 500;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  z-index: 10;
}

.md-editor-wrap {
  flex: 1;
  overflow: auto;
  padding: 32px;
}

.tiptap-editor-container {
  outline: none;
  min-height: 600px;
}

.tiptap-editor-container :deep(.tiptap) {
  min-height: 600px;
  outline: none;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  color: var(--color-text-primary);
  line-height: 1.2;
  font-size: 17px;
}

.tiptap-editor-container :deep(.tiptap h1) {
  font-family: "SourceHanSans-Bold", var(--font-family-cn);
  font-size: 32px;
  font-weight: 700;
  text-align: left;
  color: var(--color-text-primary);
  margin: 0 0 28px;
  line-height: 1.2;
}

.tiptap-editor-container :deep(.tiptap h2) {
  font-family: "SourceHanSans-Bold", var(--font-family-cn);
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 32px 0 20px;
  padding-left: 0;
  border-left: none;
}

.tiptap-editor-container :deep(.tiptap h3) {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 24px 0 12px;
}

.tiptap-editor-container :deep(.tiptap p) {
  margin: 12px 0;
  font-size: 17px;
  line-height: 1.2;
}

.tiptap-editor-container :deep(.tiptap ul),
.tiptap-editor-container :deep(.tiptap ol) {
  margin: 12px 0;
  padding-left: 40px;
}

.tiptap-editor-container :deep(.tiptap li) {
  margin: 8px 0;
  font-size: 17px;
  line-height: 1.2;
}

.tiptap-editor-container :deep(.tiptap blockquote) {
  margin: 16px 0;
  padding: 12px 20px;
  border-left: 3px solid var(--color-primary);
  background: var(--color-bg-secondary);
  color: var(--color-text-secondary);
  font-style: italic;
  border-radius: 0 6px 6px 0;
}

.tiptap-editor-container :deep(.tiptap code) {
  font-family: 'Menlo', 'Consolas', monospace;
  background: var(--color-bg-secondary);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 14px;
  color: var(--color-primary);
}

.tiptap-editor-container :deep(.tiptap pre) {
  margin: 16px 0;
  padding: 16px 20px;
  background: var(--color-bg-secondary);
  border-radius: 6px;
  overflow-x: auto;
}

.tiptap-editor-container :deep(.tiptap strong) {
  font-weight: 600;
  color: var(--color-text-primary);
}

.editor-footer {
  padding: 24px 60px;
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
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 17px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
}

.btn-outline-md {
  background: var(--color-bg-white);
  color: rgba(176, 141, 58, 1);
  border: 1px solid rgba(198, 164, 92, 1);
}

.btn-outline-md:hover:not(:disabled) {
  background: var(--color-bg-secondary);
  border-color: rgba(198, 164, 92, 1);
  color: rgba(176, 141, 58, 1);
}

.btn-outline-md [data-icon] {
  color: rgba(198, 164, 92, 1);
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
  background: rgba(251, 249, 245, 1);
  border: 1px solid var(--color-border-soft);
  position: relative;
}

.assistant-header {
  padding: 20px 24px;
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
  gap: 12px;
  font-size: 17px;
  font-weight: 700;
  color: var(--color-text-primary);
  font-family: "SourceHanSans-Bold", var(--font-family-cn);
}

.assistant-title [data-icon] {
  color: var(--color-info);
}

.model-select {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 38px;
  padding: 0 16px;
  background: var(--color-bg-secondary);
  border-radius: 10px;
  cursor: pointer;
  font-size: 15px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  color: var(--color-text-body);
  transition: all 0.2s;
  min-width: 0;
}

.model-select:hover {
  background: var(--color-border-soft);
}

.model-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 130px;
}

.model-select > [data-icon] {
  color: var(--color-text-secondary);
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
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
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
  align-items: flex-start;
  max-width: 100%;
}

.bubble {
  max-width: 360px;
  padding: 16px 20px;
  border-radius: 12px;
  font-size: 15px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  line-height: 1.2;
  color: var(--color-text-primary);
  word-break: break-word;
}

.msg-ai .bubble {
  background: rgba(243, 238, 228, 1);
}

.msg-user .bubble {
  background: rgba(247, 238, 220, 1);
}

.chat-input-wrap {
  padding: 8px 20px 24px 20px;
  border-top: 1px solid var(--color-border-soft);
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  position: relative;
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
  padding: 0 8px 0 20px;
  border: 1px solid var(--color-border);
  border-radius: 12px;
  background: var(--color-bg-white);
  outline: none;
  font-size: 15px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  color: var(--color-text-primary);
  transition: all 0.2s;
}

.chat-input:focus {
  border-color: var(--color-primary);
}

.chat-input::placeholder {
  color: var(--color-text-faint);
}

.skill-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--color-primary-soft);
  border: 1px solid var(--color-primary);
  border-radius: 8px;
  font-size: 13px;
  color: var(--color-primary);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  align-self: flex-start;
}

.skill-chip > [data-icon] {
  color: var(--color-primary);
}

.skill-chip-close {
  background: transparent;
  border: none;
  color: var(--color-primary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2px;
  border-radius: 4px;
  margin-left: 4px;
}

.skill-chip-close:hover {
  background: rgba(192, 57, 43, 0.1);
}

.skill-popup {
  position: absolute;
  bottom: 100%;
  left: 20px;
  right: 20px;
  margin-bottom: 4px;
  background: var(--color-bg-white);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.1);
  z-index: 20;
  max-height: 240px;
  overflow-y: auto;
  padding: 6px;
}

.skill-option {
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
}

.skill-option:hover,
.skill-option.active {
  background: var(--color-bg-secondary);
}

.skill-option.active {
  background: var(--color-primary-soft);
}

.skill-option-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-primary);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}

.skill-option-desc {
  font-size: 12px;
  color: var(--color-text-muted);
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.send-btn {
  position: absolute;
  right: 8px;
  width: 38px;
  height: 38px;
  border-radius: 10px;
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
