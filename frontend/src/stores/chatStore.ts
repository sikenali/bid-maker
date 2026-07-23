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

  const sendMessage = async (text: string, sectionId?: string, docId?: string, selectedModelId?: string) => {
    if (isSending.value) return
    isSending.value = true
    messages.push({ role: 'user', content: text })

    try {
      let apiKeyEntry: any = null
      if (selectedModelId) {
        apiKeyEntry = settingsStore.apiKeys.find(k => k.id === selectedModelId)
      }
      const modelName = apiKeyEntry ? (apiKeyEntry.modelName || apiKeyEntry.model) : modelId.value
      const res = await sendChat({
        document_id: docId || null,
        message: text,
        mode: mode.value,
        section_id: sectionId || null,
        history: messages.slice(0, -1).map(m => ({ role: m.role, content: m.content })),
        model: modelName,
        endpoint: apiKeyEntry?.endpoint || '',
        format: apiKeyEntry?.format || 'openai',
        apiKey: apiKeyEntry?.key || '',
      })
      messages.push({ role: 'ai', content: res.data.reply || res.data.response || '' })
    } catch (err: any) {
      console.error('Chat failed:', err)
      const msg = err?.response?.data?.error || '抱歉，出了点问题。'
      messages.push({ role: 'ai', content: msg })
    } finally {
      isSending.value = false
    }
  }

  const setModel = (m: string) => { modelId.value = m }
  const setMode = (m: 'free' | 'context') => { mode.value = m }
  const clearMessages = () => { messages.length = 0 }

  return { messages, mode, modelId, isSending, sendMessage, setMode, setModel, clearMessages }
})
