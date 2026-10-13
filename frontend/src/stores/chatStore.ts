import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import { sendChat } from '../api/client'
import { useSettingsStore } from './settingsStore'

export interface ChatMessage {
  role: 'user' | 'ai'
  content: string
}

export const useChatStore = defineStore('chat', () => {
  const messages = reactive<ChatMessage[]>([])
  const mode = ref<'free' | 'context'>('free')
  const modelId = ref('')
  const isSending = ref(false)
  const settingsStore = useSettingsStore()
  let lastUserMsgIdx = -1

  const sendMessage = async (text: string, sectionId?: string, docId?: string, selectedModelId?: string) => {
    if (isSending.value) return
    isSending.value = true
    
    // Track user message index so we can replace it on failure
    lastUserMsgIdx = messages.length
    messages.push({ role: 'user', content: text })

    try {
      let apiKeyEntry: any = null
      if (selectedModelId) {
        apiKeyEntry = settingsStore.apiKeys.find(k => k.id === selectedModelId)
      }
      
      if (!apiKeyEntry && !selectedModelId) {
        messages.pop()
        messages.push({ role: 'ai', content: '请先在设置中配置 API Key' })
        isSending.value = false
        return
      }
      if (apiKeyEntry && !apiKeyEntry.key) {
        messages.pop()
        messages.push({ role: 'ai', content: '所选模型的 API Key 未填写，请在设置中补充' })
        isSending.value = false
        return
      }
      
      const modelName = apiKeyEntry ? (apiKeyEntry.modelName || apiKeyEntry.model) : modelId.value
      const provider = apiKeyEntry?.provider || ''
      const endpoint = apiKeyEntry?.endpoint || ''
      const apiKey = apiKeyEntry?.key || ''
      const format = apiKeyEntry?.format || 'openai'
      
      if (!endpoint && !apiKey) {
        messages.pop()
        messages.push({ role: 'ai', content: '所选模型缺少 API Key，请在设置中配置' })
        isSending.value = false
        return
      }

      // Don't include the current message in history — send separately
      const currentHistory = [...messages]
      const res = await sendChat({
        document_id: docId || '',
        message: text,
        mode: mode.value,
        section_id: sectionId || '',
        history: currentHistory.filter((_, i) => i !== lastUserMsgIdx).map(m => ({ role: m.role, content: m.content })),
        provider: provider,
        model: modelName,
        endpoint: endpoint,
        format: format,
        apiKey: apiKey,
      })
      messages.push({ role: 'ai', content: res.data.reply || res.data.response || '' })
    } catch (err: any) {
      console.error('Chat failed:', err)
      messages[lastUserMsgIdx].content = `[请求失败] ${messages[lastUserMsgIdx].content}`
      const msg = err?.response?.data?.error || '抱歉，出了点问题。'
      messages.push({ role: 'ai', content: msg })
    } finally {
      isSending.value = false
    }
  }

  const setModel = (m: string) => { modelId.value = m }
  const setMode = (m: 'free' | 'context') => { mode.value = m }
  const clearMessages = () => { 
    messages.length = 0
    lastUserMsgIdx = -1
  }

  return { messages, mode, modelId, isSending, sendMessage, setMode, setModel, clearMessages }
})
