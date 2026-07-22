import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getModels, getTemplates } from '../api/client'

export interface ApiKeyEntry {
  id: string
  provider: string
  model: string
  modelName: string
  key: string
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
  const exportFormat = ref<'word' | 'md'>('word')

  const selectedModel = computed(() =>
    allModels.value.find(m => m.id === selectedModelId.value) || allModels.value[0]
  )

  const apiKeys = ref<ApiKeyEntry[]>([])
  const apiKeyForm = ref({ provider: '', model: '', key: '', keyVisible: false })

  const templates = ref<TemplateInfo[]>([])
  const selectedTemplateId = ref('')

  const customSkills = ref<SkillInfo[]>([])
  const allSkills = computed(() => [...builtinSkills, ...customSkills.value])
  const selectedSkillId = ref('')

  async function fetchModels() {
    try {
      const res = await getModels()
      const list: string[] = res.data.models || []
      if (list.length > 0) {
        modelsFromApi.value = list.map((m) => ({
          id: m,
          name: m,
          provider: m,
          model: m,
        }))
        if (!modelsFromApi.value.find(m => m.id === selectedModelId.value)) {
          selectedModelId.value = modelsFromApi.value[0]?.id || 'gpt4o'
        }
      }
    } catch {
      console.warn('Failed to fetch models from API, using fallback list')
    }
  }

  async function fetchTemplates() {
    try {
      const res = await getTemplates()
      templates.value = res.data.templates || []
    } catch {
      console.warn('Failed to fetch templates')
    }
  }

  function setModel(id: string) {
    selectedModelId.value = id
  }

  function setTheme(t: 'light' | 'dark' | 'paper') {
    theme.value = t
  }

  function setExportFormat(f: 'word' | 'md') {
    exportFormat.value = f
  }

  function setSelectedTemplate(id: string) {
    selectedTemplateId.value = id
  }

  function setSelectedSkill(id: string) {
    selectedSkillId.value = id
  }

  function addApiKey(key: ApiKeyEntry) {
    apiKeys.value.push(key)
    apiKeyForm.value = { provider: '', model: '', key: '', keyVisible: false }
  }

  function removeApiKey(id: string) {
    apiKeys.value = apiKeys.value.filter(k => k.id !== id)
  }

  function toggleKeyVisibility() {
    apiKeyForm.value.keyVisible = !apiKeyForm.value.keyVisible
  }

  return {
    allModels, domesticModels, foreignModels,
    selectedModelId, selectedModel, theme, exportFormat,
    apiKeys, apiKeyForm,
    templates, selectedTemplateId,
    allSkills, customSkills, selectedSkillId,
    fetchModels, fetchTemplates, setModel, setTheme, setExportFormat,
    setSelectedTemplate, setSelectedSkill, addApiKey, removeApiKey, toggleKeyVisibility,
  }
})
