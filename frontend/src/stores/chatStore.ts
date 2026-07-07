import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface ChatMessage {
  role: 'user' | 'ai'
  content: string
}

export const useChatStore = defineStore('chat', () => {
  const messages = ref<ChatMessage[]>([])
  const mode = ref<'free' | 'context'>('free')
  const model = ref('')
  const isSending = ref(false)

  return { messages, mode, model, isSending }
})
