<template>
  <div class="ai-chat">
    <div class="chat-header">
      <span>AI</span>
      <select v-model="model" class="model-select">
        <option value="">Default</option>
      </select>
    </div>
    <div class="chat-messages" ref="messagesRef">
      <div v-for="(msg, i) in messages" :key="i" class="message" :class="msg.role">
        <div v-if="msg.role === 'ai'" class="avatar">&#129302;</div>
        <div class="bubble">{{ msg.content }}</div>
      </div>
    </div>
    <div class="chat-input">
      <input v-model="inputText" @keyup.enter="sendMessage" placeholder="Ask AI..." />
      <button @click="sendMessage">Send</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

const messages = reactive<Array<{ role: string; content: string }>>([])
const inputText = ref('')
const model = ref('')
const messagesRef = ref<HTMLElement>()

const sendMessage = () => {
  if (!inputText.value.trim()) return
  messages.push({ role: 'user', content: inputText.value })
  inputText.value = ''
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
.model-select {
  font-size: 12px;
  padding: 2px 4px;
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
</style>
