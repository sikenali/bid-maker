<template>
  <div class="ai-chat">
    <div class="chat-header">
      <span>AI Assistant</span>
      <div class="header-controls">
        <select v-model="chatStore.model" class="model-select">
          <option value="">Select Model</option>
        </select>
        <button
          :class="{ active: chatStore.mode === 'context' }"
          @click="chatStore.setMode(chatStore.mode === 'context' ? 'free' : 'context')"
          title="Toggle context mode"
        >
          {{ chatStore.mode === 'context' ? 'Context' : 'Free' }}
        </button>
      </div>
    </div>
    <div class="chat-messages" ref="messagesRef">
      <div
        v-for="(msg, i) in chatStore.messages"
        :key="i"
        class="message"
        :class="msg.role"
      >
        <div v-if="msg.role === 'ai'" class="avatar">&#129302;</div>
        <div class="bubble">{{ msg.content }}</div>
      </div>
      <div v-if="chatStore.isSending" class="message ai">
        <div class="avatar">&#129302;</div>
        <div class="bubble">Thinking...</div>
      </div>
    </div>
    <div class="chat-input">
      <input
        v-model="inputText"
        @keyup.enter="handleSend"
        placeholder="Type a message..."
      />
      <button @click="handleSend" :disabled="chatStore.isSending">
        {{ chatStore.isSending ? '...' : 'Send' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { useChatStore } from '../stores/chatStore'
import { useDocumentStore } from '../stores/documentStore'

const chatStore = useChatStore()
const docStore = useDocumentStore()
const inputText = ref('')
const messagesRef = ref<HTMLElement>()

const handleSend = () => {
  if (!inputText.value.trim()) return
  const text = inputText.value
  inputText.value = ''
  const sectionId = chatStore.mode === 'context' ? docStore.activeSectionId : undefined
  chatStore.sendMessage(text, sectionId)
  scrollToBottom()
}

const scrollToBottom = async () => {
  await nextTick()
  if (messagesRef.value) {
    messagesRef.value.scrollTop = messagesRef.value.scrollHeight
  }
}
</script>

<style scoped>
.ai-chat {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
  font-weight: bold;
}
.header-controls {
  display: flex;
  gap: 8px;
  align-items: center;
}
.model-select {
  font-size: 12px;
  padding: 2px 4px;
}
.header-controls button {
  font-size: 12px;
  padding: 4px 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  cursor: pointer;
}
.header-controls button.active {
  background: #1677ff;
  color: white;
  border-color: #1677ff;
}
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}
.message {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.message.user {
  flex-direction: row-reverse;
}
.bubble {
  max-width: 80%;
  padding: 8px 12px;
  border-radius: 8px;
  background: #f0f0f0;
  font-size: 14px;
}
.message.user .bubble {
  background: #1677ff;
  color: white;
}
.chat-input {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid #eee;
}
.chat-input input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
}
.chat-input button {
  padding: 8px 16px;
  background: #1677ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
.chat-input button:disabled {
  background: #ccc;
  cursor: not-allowed;
}
</style>
