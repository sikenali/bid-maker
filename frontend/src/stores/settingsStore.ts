import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

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

const domesticModels: ModelInfo[] = [
  { id: 'qwen', name: '通义千问', provider: '阿里云', model: 'qwen-turbo' },
  { id: 'wenxin', name: '文心一言', provider: '百度', model: 'ernie-4.0' },
  { id: 'glm', name: '智谱 GLM', provider: '智谱', model: 'glm-4' },
]

const foreignModels: ModelInfo[] = [
  { id: 'gpt4o', name: 'GPT-4o', provider: 'OpenAI', model: 'gpt-4o' },
  { id: 'claude', name: 'Claude 3.5', provider: 'Anthropic', model: 'claude-3-5-sonnet' },
]

export const useSettingsStore = defineStore('settings', () => {
  const allModels = computed(() => [...domesticModels, ...foreignModels])
  const selectedModelId = ref('gpt4o')
  const theme = ref<'light' | 'dark' | 'paper'>('light')
  const exportFormat = ref<'word' | 'md'>('word')
  const activeTplTab = ref('招标模板')

  const selectedModel = computed(() =>
    allModels.value.find(m => m.id === selectedModelId.value) || allModels.value[0]
  )

  const apiKeys = ref<ApiKeyEntry[]>([])
  const apiKeyForm = ref({ provider: '', model: '', key: '', keyVisible: false })

  function setModel(id: string) {
    selectedModelId.value = id
  }

  function setTheme(t: 'light' | 'dark' | 'paper') {
    theme.value = t
  }

  function setExportFormat(f: 'word' | 'md') {
    exportFormat.value = f
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
    selectedModelId, selectedModel, theme, exportFormat, activeTplTab,
    apiKeys, apiKeyForm,
    setModel, setTheme, setExportFormat, addApiKey, removeApiKey, toggleKeyVisibility,
  }
})
