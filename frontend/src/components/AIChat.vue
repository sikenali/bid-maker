<template>
  <div class="ai-chat">
    <div class="chat-header">
      <div class="header-left">
        <div class="ai-icon-box">
          <RiSparklingFill size="16" color="#fff" />
        </div>
        <span class="header-title">AI 助手</span>
      </div>
      <div class="model-display">
        <span class="model-label">{{ settingsStore.selectedModel.name }}</span>
      </div>
    </div>
    <div class="chat-messages" ref="messagesRef">
      <div
        v-for="(msg, i) in chatStore.messages"
        :key="i"
        class="message"
        :class="msg.role"
      >
        <template v-if="msg.role === 'ai'">
          <div class="ai-avatar">
            <RiSparklingFill size="14" color="#fff" />
          </div>
          <div class="bubble ai-bubble">{{ msg.content }}</div>
        </template>
        <template v-else>
          <div class="bubble user-bubble">{{ msg.content }}</div>
        </template>
      </div>
      <div v-if="chatStore.isSending" class="message ai">
        <div class="ai-avatar">
          <RiSparklingFill size="14" color="#fff" />
        </div>
        <div class="bubble ai-bubble">思考中...</div>
      </div>
    </div>
    <div class="chat-input-area">
      <div class="input-container">
        <input
          v-model="inputText"
          @keyup.enter="handleSend"
          placeholder="输入您的问题..."
          class="chat-input"
        />
        <button @click="handleSend" :disabled="chatStore.isSending" class="send-btn">
          <RiSendPlaneFill size="14" color="#fff" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useChatStore } from '../stores/chatStore'
import { useSettingsStore } from '../stores/settingsStore'
import { useDocumentStore } from '../stores/documentStore'
import {
  RiSparklingFill,
  RiSendPlaneFill,
} from '@remixicon/vue'

const chatStore = useChatStore()
const settingsStore = useSettingsStore()
const docStore = useDocumentStore()
const inputText = ref('')
const messagesRef = ref<HTMLElement>()

chatStore.setModel(settingsStore.selectedModel.model)
watch(() => settingsStore.selectedModelId, () => {
  chatStore.setModel(settingsStore.selectedModel.model)
})

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
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ai-icon-box {
  width: 28px;
  height: 28px;
  background: #C43D3D;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-title {
  font-size: 14px;
  font-weight: 600;
  color: #3D2B1F;
}

.model-display {
  display: flex;
  align-items: center;
  background: #E8DCC8;
  border-radius: 8px;
  padding: 4px 8px;
  cursor: default;
}

.model-label {
  font-size: 11px;
  color: #3D2B1F;
  font-weight: 500;
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
  justify-content: flex-end;
}

.ai-avatar {
  width: 28px;
  height: 28px;
  background: #C43D3D;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.bubble {
  max-width: 80%;
  padding: 12px;
  border-radius: 12px;
  font-size: 12px;
  line-height: 1.6;
}

.ai-bubble {
  background: #fff;
  color: #3D2B1F;
}

.user-bubble {
  background: #C43D3D;
  color: #fff;
}

.chat-input-area {
  padding: 12px 16px;
  flex-shrink: 0;
}

.input-container {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fff;
  border-radius: 12px;
  padding: 8px;
}

.chat-input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 12px;
  color: #3D2B1F;
  background: transparent;
  padding: 8px 0;
}

.chat-input::placeholder {
  color: #B8A88A;
}

.send-btn {
  width: 28px;
  height: 28px;
  background: #C43D3D;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background 0.2s;
}

.send-btn:hover {
  background: #A83232;
}

.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>