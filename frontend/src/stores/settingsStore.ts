import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getModels } from '../api/client'
import {
  getApiKeyList, addApiKey as apiAddKey, deleteApiKey as apiDelKey,
  toggleApiKey as apiToggleKey, updateApiKey as apiUpdateKey,
  getSkillList,
  type ApiKeyData,
} from '../api/settings'

export interface ApiKeyEntry {
  id: string
  provider: string
  model: string
  modelName: string
  key: string
  enabled: boolean
  endpoint?: string
  format?: string
}

export interface ModelInfo {
  id: string
  name: string
  provider: string
  model: string
}

export interface SkillInfo {
  id: string
  name: string
  description: string
  category: string
  prompt: string
  enabled?: boolean
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

function adaptKey(k: ApiKeyData): ApiKeyEntry {
  return {
    id: k.id,
    provider: k.provider,
    model: k.model,
    modelName: k.model,
    key: k.api_key || '',
    enabled: k.enabled !== 0,
    endpoint: k.base_url || '',
    format: k.format || 'openai',
  }
}

// 单一数据源:全部读写走 api/settings.ts (localStorage['bid-maker-settings']),
// 设置页与编辑页/对话共用同一份数据,新增模型/技能/模板即时生效
export const useSettingsStore = defineStore('settings', () => {
  // localStorage 非响应式,用版本号驱动 computed 刷新
  const rev = ref(0)
  const refresh = () => { rev.value++ }

  const modelsFromApi = ref<ModelInfo[]>([])

  const allModels = computed(() => {
    if (modelsFromApi.value.length > 0) return modelsFromApi.value
    return [...fallbackDomestic, ...fallbackForeign]
  })

  const selectedModelId = ref('gpt4o')
  try {
    const saved = localStorage.getItem('settings_selected_model')
    if (saved) selectedModelId.value = saved
  } catch {}
  function setModel(id: string) {
    selectedModelId.value = id
    try { localStorage.setItem('settings_selected_model', id) } catch {}
  }

  const selectedModel = computed(() => {
    const fromApiKey = apiKeys.value.find(k => k.id === selectedModelId.value)
    if (fromApiKey) {
      return { id: fromApiKey.id, name: fromApiKey.modelName, provider: fromApiKey.provider, model: fromApiKey.model }
    }
    return allModels.value.find(m => m.id === selectedModelId.value) || allModels.value[0]
  })

  // ── API Keys(设置页 ApiKeySettingsView 同源) ──
  const apiKeys = computed<ApiKeyEntry[]>(() => {
    void rev.value
    try { return getApiKeyList().map(adaptKey) } catch { return [] }
  })

  async function addApiKey(entry: { provider: string; model: string; key: string; endpoint?: string; format?: string }) {
    await apiAddKey({ provider: entry.provider, model: entry.model, api_key: entry.key, base_url: entry.endpoint, format: entry.format })
    refresh()
  }
  async function removeApiKey(id: string) {
    await apiDelKey(id)
    refresh()
  }
  async function toggleApiKey(id: string) {
    await apiToggleKey(id)
    refresh()
  }
  async function updateApiKeyEntry(id: string, patch: { provider?: string; model?: string; key?: string; endpoint?: string; format?: string }) {
    await apiUpdateKey(id, {
      provider: patch.provider,
      model: patch.model,
      api_key: patch.key,
      base_url: patch.endpoint,
      format: patch.format,
    })
    refresh()
  }

  // ── Skills(设置页 SkillSettingsView 同源,/ 弹窗直接显示管理列表) ──
  const allSkills = computed<SkillInfo[]>(() => {
    void rev.value
    try {
      return getSkillList().map(s => ({
        id: String((s as any).id ?? s.name),
        name: s.name,
        description: s.description || '',
        category: (s as any).group || '',
        prompt: (s as any).content || '',
        enabled: (s as any).enabled !== false,
      }))
    } catch { return [] }
  })

  // ── Models ──
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

  return {
    allModels,
    selectedModelId, selectedModel,
    apiKeys,
    allSkills,
    fetchModels,
    setModel, refresh,
    addApiKey, removeApiKey, toggleApiKey, updateApiKeyEntry,
  }
})
