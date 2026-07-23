import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { getModels, getTemplates } from '../api/client'

export interface ApiKeyEntry {
  id: string
  provider: string
  model: string
  modelName: string
  key: string
  enabled: boolean
}

export interface ModelInfo {
  id: string
  name: string
  provider: string
  model: string
}

export interface TemplateInfo {
  id: string
  name: string
  description: string
  category: string
  icon: string
  outline: any[]
}

const fallbackDomestic: ModelInfo[] = [
  { id: 'qwen', name: '通义千问', provider: '阿里云', model: 'qwen-turbo' },
  { id: 'wenxin', name: '文心一言', provider: '百度', model: 'ernie-4.0' },
  { id: 'glm', name: '智谱 GLM', provider: '智谱', model: 'glm-4' },
]

const fallbackForeign: ModelInfo[] = [
  { id: 'gpt4o', name: 'GPT-4o', provider: 'OpenAI', model: 'gpt-4o' },
  { id: 'claude', name: 'Claude 3.5', provider: 'Anthropic', model: 'claude-3-5-sonnet' },
]

export interface SkillInfo {
  id: string
  name: string
  description: string
  category: string
  prompt: string
  enabled?: boolean
  path?: string
}

const builtinSkills: SkillInfo[] = [
  { id: 'outline', name: '大纲生成', description: '根据招标文件自动生成标书大纲', category: '智能写作', prompt: '请根据以下内容生成一份完整的标书大纲，包含章节标题和层级结构：' },
  { id: 'expand', name: '内容扩写', description: '基于大纲要点自动扩写章节内容', category: '智能写作', prompt: '请扩写以下内容，使其更加详细和专业：' },
  { id: 'summarize', name: '摘要总结', description: '提取文档关键信息生成简洁摘要', category: '智能处理', prompt: '请对以下内容进行摘要总结，提取关键信息：' },
  { id: 'format', name: '格式优化', description: '统一文档格式、段落和排版风格', category: '智能处理', prompt: '请优化以下内容的格式和排版，使其更加规范统一：' },
]

export const useSettingsStore = defineStore('settings', () => {
  const modelsFromApi = ref<ModelInfo[]>([])
  const domesticModels = ref<ModelInfo[]>(fallbackDomestic)
  const foreignModels = ref<ModelInfo[]>(fallbackForeign)

  const allModels = computed(() => {
    if (modelsFromApi.value.length > 0) return modelsFromApi.value
    return [...domesticModels.value, ...foreignModels.value]
  })

  const selectedModelId = ref('gpt4o')
  const theme = ref<'light' | 'dark' | 'paper'>('light')
  try {
    const saved = localStorage.getItem('settings_theme')
    if (saved === 'dark' || saved === 'paper') theme.value = saved
  } catch {}
  watch(
    theme,
    (t) => { try { localStorage.setItem('settings_theme', t) } catch {} },
  )
  const exportFormat = ref<'word' | 'md'>('word')

  const selectedModel = computed(() => {
    const fromApiKey = apiKeys.value.find(k => k.id === selectedModelId.value)
    if (fromApiKey) {
      return { id: fromApiKey.id, name: fromApiKey.modelName, provider: fromApiKey.provider, model: fromApiKey.model }
    }
    return allModels.value.find(m => m.id === selectedModelId.value) || allModels.value[0]
  })

  // ── API Keys ──
  const apiKeys = ref<ApiKeyEntry[]>([])
  const apiKeyForm = ref({ provider: '', model: '', key: '', keyVisible: false })

  try {
    const saved = localStorage.getItem('settings_api_keys')
    if (saved) apiKeys.value = JSON.parse(saved)
  } catch {}
  watch(
    () => apiKeys.value,
    (keys) => { try { localStorage.setItem('settings_api_keys', JSON.stringify(keys)) } catch {} },
    { deep: true }
  )

  function addApiKey(key: ApiKeyEntry) {
    apiKeys.value.push({ ...key, enabled: true })
    apiKeyForm.value = { provider: '', model: '', key: '', keyVisible: false }
  }
  function removeApiKey(id: string) {
    apiKeys.value = apiKeys.value.filter(k => k.id !== id)
  }
  function toggleApiKey(id: string) {
    const key = apiKeys.value.find(k => k.id === id)
    if (key) key.enabled = !key.enabled
  }
  function toggleKeyVisibility() {
    apiKeyForm.value.keyVisible = !apiKeyForm.value.keyVisible
  }

  // ── Templates ──
  const templates = ref<TemplateInfo[]>([])
  const selectedTemplateId = ref('')

  try {
    const saved = localStorage.getItem('user_templates')
    if (saved) templates.value = JSON.parse(saved)
  } catch {}
  watch(
    () => templates.value,
    (tpls) => { try { localStorage.setItem('user_templates', JSON.stringify(tpls)) } catch {} },
    { deep: true }
  )

  function fetchTemplates() {
    getTemplates().then((res: any) => {
      const tpls: TemplateInfo[] = res?.data?.templates || []
      const existingIds = new Set(templates.value.map(t => t.id))
      for (const t of tpls) {
        if (!existingIds.has(t.id)) {
          templates.value.push(t)
        }
      }
    }).catch(() => {})
  }

  function addTemplate(name: string, file?: File) {
    templates.value.push({
      id: `tpl_${Date.now()}`,
      name: name.trim(),
      description: file ? `${file.name} (${(file.size / 1024).toFixed(1)}KB)` : '',
      category: '自定义',
      icon: 'RiFileTextLine',
      outline: [],
    })
    selectedTemplateId.value = templates.value[templates.value.length - 1].id
  }

  function removeTemplate(id: string) {
    templates.value = templates.value.filter(t => t.id !== id)
  }

  // ── Skills ──
  const localSkills = ref<SkillInfo[]>([
    { id: 'calicat', name: 'calicat-cli-operator', description: 'Calicat CLI 操作技能', category: '本地技能', prompt: '', enabled: true },
    { id: 'grill-me', name: 'grill-me', description: 'A relentless interview to sharpen a plan or design.', category: '本地技能', prompt: '', enabled: true },
    { id: 'brainstorming', name: 'brainstorming', description: 'Helps you explore ideas and requirements before building.', category: '本地技能', prompt: '', enabled: true },
    { id: 'huashu-design', name: 'huashu-design', description: '花叔Design——用HTML做高保真原型、幻灯片、动画、可视化。', category: '本地技能', prompt: '', enabled: true },
    { id: 'find-skills', name: 'find-skills', description: 'Helps users discover and install agent skills.', category: '本地技能', prompt: '', enabled: true },
  ])
  const customSkills = ref<SkillInfo[]>([])
  const allSkills = computed(() => [...builtinSkills, ...localSkills.value, ...customSkills.value])
  const selectedSkillId = ref('')

  function fetchLocalSkills() {
    fetch('/api/local-skills').then(async (res) => {
      const data = await res.json()
      localSkills.value = (data.skills || []).map((s: any) => ({
        id: `local_${s.path}`, name: s.name, description: s.description,
        category: '本地技能', prompt: '', enabled: true, path: s.path,
      }))
    }).catch(() => {})
  }

  function addCustomSkill(skill: Omit<SkillInfo, 'id'>) {
    const newSkill = { ...skill, id: `custom_${Date.now()}` }
    customSkills.value.push(newSkill)
    localStorage.setItem('custom_skills', JSON.stringify(customSkills.value))
  }
  function removeCustomSkill(id: string) {
    customSkills.value = customSkills.value.filter(s => s.id !== id)
    localStorage.setItem('custom_skills', JSON.stringify(customSkills.value))
  }
  function toggleSkillEnabled(id: string) {
    const skill = customSkills.value.find(s => s.id === id)
    if (skill) { skill.enabled = !skill.enabled; localStorage.setItem('custom_skills', JSON.stringify(customSkills.value)) }
  }
  function loadCustomSkillsFromLocalStorage() {
    try {
      const stored = localStorage.getItem('custom_skills')
      if (stored) customSkills.value = JSON.parse(stored)
    } catch { customSkills.value = [] }
  }

  // ── General ──
  async function fetchModels() {
    try {
      const res = await getModels()
      const list: string[] = res.data.models || []
      if (list.length > 0) {
        modelsFromApi.value = list.map((m) => ({ id: m, name: m, provider: m, model: m }))
        if (!modelsFromApi.value.find(m => m.id === selectedModelId.value)) {
          selectedModelId.value = modelsFromApi.value[0]?.id || 'gpt4o'
        }
      }
    } catch { console.warn('Failed to fetch models from API, using fallback list') }
  }

  function setModel(id: string) { selectedModelId.value = id }
  function setTheme(t: 'light' | 'dark' | 'paper') { theme.value = t }
  function setExportFormat(f: 'word' | 'md') { exportFormat.value = f }
  function setSelectedTemplate(id: string) { selectedTemplateId.value = id }
  function setSelectedSkill(id: string) { selectedSkillId.value = id }

  return {
    allModels, domesticModels, foreignModels,
    selectedModelId, selectedModel, theme, exportFormat,
    apiKeys, apiKeyForm,
    templates, selectedTemplateId,
    localSkills, customSkills, allSkills, selectedSkillId,
    fetchModels, fetchTemplates, fetchLocalSkills,
    setModel, setTheme, setExportFormat,
    setSelectedTemplate, setSelectedSkill,
    addApiKey, removeApiKey, toggleKeyVisibility, toggleApiKey,
    addTemplate, removeTemplate,
    addCustomSkill, removeCustomSkill, toggleSkillEnabled, loadCustomSkillsFromLocalStorage,
  }
})
