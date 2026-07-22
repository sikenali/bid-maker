<template>
  <div class="ai-chat">
    <div class="chat-header">
      <div class="header-left">
        <div class="ai-icon-box">
          <RiSparklingFill size="16" color="#fff" />
        </div>
        <span class="header-title">AI 助手</span>
      </div>
      <div class="model-select-wrap">
        <ModelSelect
          v-model="selectedModelId"
          :items="settingsStore.allModels"
          @update:model-value="onModelChange"
        />
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
      <div v-if="settingsStore.allSkills.length > 0" class="skill-bar">
        <button
          class="skill-btn"
          :class="{ 'skill-btn-active': !settingsStore.selectedSkillId }"
          @click="settingsStore.setSelectedSkill('')"
        >自由对话</button>
        <button
          v-for="skill in settingsStore.allSkills"
          :key="skill.id"
          class="skill-btn"
          :class="{ 'skill-btn-active': settingsStore.selectedSkillId === skill.id }"
          @click="settingsStore.setSelectedSkill(skill.id === settingsStore.selectedSkillId ? '' : skill.id)"
        >{{ skill.name }}</button>
      </div>
      <div class="input-container">
        <textarea
          v-model="inputText"
          @keydown.enter.prevent="handleSend"
          placeholder="输入您的问题..."
          class="chat-input"
          rows="2"
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
import { useRoute } from 'vue-router'
import { useChatStore } from '../stores/chatStore'
import { useSettingsStore } from '../stores/settingsStore'
import { useDocumentStore } from '../stores/documentStore'
import ModelSelect from './ModelSelect.vue'
import {
  RiSparklingFill,
  RiSendPlaneFill,
} from '@remixicon/vue'

const chatStore = useChatStore()
const settingsStore = useSettingsStore()
const docStore = useDocumentStore()
const route = useRoute()
const docId = route.params.id as string
const inputText = ref('')
const messagesRef = ref<HTMLElement>()

const selectedModelId = ref(settingsStore.selectedModelId)

const onModelChange = (id: string) => {
  settingsStore.setModel(id)
  chatStore.setModel(settingsStore.selectedModel.model)
}

chatStore.setModel(settingsStore.selectedModel.model)
watch(() => settingsStore.selectedModelId, () => {
  selectedModelId.value = settingsStore.selectedModelId
  chatStore.setModel(settingsStore.selectedModel.model)
})

const handleSend = () => {
  if (!inputText.value.trim()) return
  let text = inputText.value
  inputText.value = ''
  const sectionId = chatStore.mode === 'context' ? docStore.activeSectionId : undefined

  if (settingsStore.selectedSkillId) {
    const skill = settingsStore.allSkills.find(s => s.id === settingsStore.selectedSkillId)
    if (skill) {
      text = skill.prompt + '\n\n' + text
    }
  }

  chatStore.sendMessage(text, sectionId, docId)
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
  background: #C23B22;
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

.model-select-wrap {
  display: flex;
  align-items: center;
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
  background: #C23B22;
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
  background: #C23B22;
  color: #fff;
}

.chat-input-area {
  padding: 8px 16px 12px;
  flex-shrink: 0;
}

.skill-bar {
  display: flex;
  gap: 4px;
  margin-bottom: 8px;
  overflow-x: auto;
  flex-shrink: 0;
}

.skill-btn {
  padding: 4px 10px;
  border: 0.7px solid #E0D5C0;
  border-radius: 6px;
  background: #F5EFE0;
  font-size: 11px;
  color: #8B7355;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s;
  flex-shrink: 0;
}

.skill-btn:hover {
  border-color: #C23B22;
  color: #C23B22;
}

.skill-btn-active {
  background: #C23B22;
  border-color: #C23B22;
  color: #fff;
}

.skill-btn-active:hover {
  color: #fff;
}

.input-container {
  display: flex;
  align-items: flex-end;
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
  padding: 4px 0;
  resize: none;
  line-height: 1.5;
  font-family: inherit;
}

.chat-input::placeholder {
  color: #B8A88A;
}

.chat-input::-webkit-scrollbar {
  width: 4px;
}

.chat-input::-webkit-scrollbar-thumb {
  background: #D4C4A8;
  border-radius: 2px;
}

.send-btn {
  width: 28px;
  height: 28px;
  background: #C23B22;
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
  background: #A83028;
}

.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>