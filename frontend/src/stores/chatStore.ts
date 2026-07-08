import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import { sendChat } from '../api/client'

export interface ChatMessage {
  role: 'user' | 'ai'
  content: string
}

export const useChatStore = defineStore('chat', () => {
  const messages = reactive<ChatMessage[]>([])
  const mode = ref<'free' | 'context'>('free')
  const model = ref('')
  const isSending = ref(false)

  const sendMessage = async (text: string, sectionId?: string, docId?: string) => {
    if (isSending.value) return
    isSending.value = true
    messages.push({ role: 'user', content: text })

    try {
      const res = await sendChat({
        document_id: docId || null,
        message: text,
        mode: mode.value,
        section_id: sectionId || null,
        history: messages.slice(0, -1).map(m => ({ role: m.role, content: m.content })),
        model: model.value,
      })
      messages.push({ role: 'ai', content: res.data.reply || res.data.response || '' })
    } catch (err) {
      console.error('Chat failed:', err)
      messages.push({ role: 'ai', content: '抱歉，出了点问题。' })
    } finally {
      isSending.value = false
    }
  }

  const setMode = (m: 'free' | 'context') => { mode.value = m }
  const setModel = (m: string) => { model.value = m }
  const clearMessages = () => { messages.length = 0 }

  return { messages, mode, model, isSending, sendMessage, setMode, setModel, clearMessages }
})
