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
          :items="modelItems"
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
      <div class="input-container" ref="inputContainerRef">
        <!-- Skill chip overlay -->
        <div v-if="selectedSkill" class="skill-chip-overlay" @click="clearSelectedSkill">
          <span class="chip-label">/ {{ selectedSkill.name }}</span>
          <span class="chip-close"><RiCloseLine size="14" /></span>
        </div>
        <div v-if="showSkillPopup" class="skill-popup" ref="popupRef" :style="{ paddingBottom: selectedSkill ? '28px' : '4px' }">
          <div
            v-for="(skill, idx) in filteredSkills"
            :key="skill.id"
            class="skill-popup-item"
            :class="{ 'skill-popup-item-active': idx === activeSkillIdx }"
            @click="selectSkill(skill)"
            @mouseenter="activeSkillIdx = idx"
          >
            <span class="skill-popup-name">{{ skill.name }}</span>
            <span class="skill-popup-desc">{{ skill.description }}</span>
          </div>
          <div v-if="filteredSkills.length === 0" class="skill-popup-empty">无匹配技能</div>
        </div>
        <textarea
          ref="textareaRef"
          v-model="inputText"
          @keydown="onInputKeydown"
          @input="onInput"
          placeholder="输入您的问题...（输入 / 调用技能）"
          class="chat-input"
          rows="1"
        />
        <button @click="handleSend" :disabled="chatStore.isSending" class="send-btn">
          <RiSendPlaneFill size="14" color="#fff" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useChatStore } from '../stores/chatStore'
import { useSettingsStore } from '../stores/settingsStore'
import { useDocumentStore } from '../stores/documentStore'
import ModelSelect from './ModelSelect.vue'
import {
  RiSparklingFill,
  RiSendPlaneFill,
  RiCloseLine,
} from '@remixicon/vue'

const chatStore = useChatStore()
const settingsStore = useSettingsStore()
const docStore = useDocumentStore()
const route = useRoute()
const docId = route.params.id as string
const inputText = ref('')
const messagesRef = ref<HTMLElement>()

const selectedModelId = ref(
  settingsStore.apiKeys.length > 0
    ? settingsStore.apiKeys[0].id
    : settingsStore.selectedModelId
)

const modelItems = computed(() => {
  return settingsStore.apiKeys.map(k => ({
    id: k.id,
    name: k.modelName,
  }))
})

const onModelChange = (id: string) => {
  settingsStore.setModel(id)
  const selected = settingsStore.selectedModel
  if (selected) {
    chatStore.setModel(selected.model)
  }
}

chatStore.setModel(settingsStore.selectedModel?.model || '')
watch(() => settingsStore.selectedModelId, () => {
  selectedModelId.value = settingsStore.selectedModelId
  const selected = settingsStore.selectedModel
  if (selected) {
    chatStore.setModel(selected.model)
  }
})

const textareaRef = ref<HTMLTextAreaElement>()
const inputContainerRef = ref<HTMLElement>()
const popupRef = ref<HTMLElement>()
const showSkillPopup = ref(false)
const activeSkillIdx = ref(0)
const skillQuery = ref('')
// Track the currently selected skill object (replaces string-based selectedSkillId)
const activeSkillObj = ref<{ id: string; name: string; description: string; prompt: string } | null>(null)

// Hidden skills set — shared with SettingsView
const tryGetHiddenSkills = () => {
  try {
    const saved = localStorage.getItem('hidden_skills')
    if (saved) return new Set(JSON.parse(saved))
  } catch {}
  return new Set<string>()
}

// Build cmd-ready skills: only show skills that exist in skill management and are enabled
const cmdSkills = computed(() => {
  const hidden = tryGetHiddenSkills()
  const localIds = new Set(settingsStore.localSkills.map(s => s.id))
  const customIds = new Set(settingsStore.customSkills.map(s => s.id))
  return settingsStore.allSkills
    .filter(s => {
      if (!localIds.has(s.id) && !customIds.has(s.id)) return false
      if (hidden.has(s.id)) return false
      if (customIds.has(s.id)) {
        const customSkill = settingsStore.customSkills.find(cs => cs.id === s.id)
        if (!customSkill || customSkill.enabled === false) return false
      }
      if (localIds.has(s.id)) {
        const localSkill = settingsStore.localSkills.find(ls => ls.id === s.id)
        if (!localSkill || localSkill.enabled === false) return false
      }
      return true
    })
    .map(s => ({ ...s, cmd: '/' + s.id }))
})

const filteredSkills = computed(() => {
  if (!skillQuery.value) return cmdSkills.value
  const q = skillQuery.value.toLowerCase()
  return cmdSkills.value.filter(
    s => s.id.toLowerCase().includes(q) || s.name.toLowerCase().includes(q) || s.cmd.includes(q)
  )
})

const onInput = () => {
  const match = inputText.value.match(/\/(\S*)$/)
  if (match) {
    skillQuery.value = match[1]
    showSkillPopup.value = true
    activeSkillIdx.value = 0
  } else {
    showSkillPopup.value = false
    skillQuery.value = ''
  }
}

const onInputKeydown = (e: KeyboardEvent) => {
  if (showSkillPopup.value) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      activeSkillIdx.value = Math.min(activeSkillIdx.value + 1, filteredSkills.value.length - 1)
      scrollPopupItemIntoView()
      return
    }
    if (e.key === 'ArrowUp') {
      e.preventDefault()
      activeSkillIdx.value = Math.max(activeSkillIdx.value - 1, 0)
      scrollPopupItemIntoView()
      return
    }
    if (e.key === 'Enter' || e.key === 'Tab') {
      if (filteredSkills.value[activeSkillIdx.value]) {
        e.preventDefault()
        selectSkill(filteredSkills.value[activeSkillIdx.value])
        return
      }
    }
    if (e.key === 'Escape') {
      showSkillPopup.value = false
      skillQuery.value = ''
      return
    }
  }
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

const scrollPopupItemIntoView = () => {
  nextTick(() => {
    const popup = popupRef.value
    if (!popup) return
    const item = popup.children[activeSkillIdx.value] as HTMLElement
    if (item) item.scrollIntoView({ block: 'nearest' })
  })
}

const selectSkill = (skill: { id: string; name: string; description: string; prompt: string }) => {
  inputText.value = inputText.value.replace(/\/(\S*)$/, '')
  showSkillPopup.value = false
  skillQuery.value = ''
  activeSkillObj.value = skill
  nextTick(() => textareaRef.value?.focus())
}

const clearSelectedSkill = () => {
  activeSkillObj.value = null
}

const selectedSkill = computed(() => activeSkillObj.value)

const handleSend = () => {
  if (!inputText.value.trim()) return
  let text = inputText.value
  inputText.value = ''
  const sectionId = chatStore.mode === 'context' ? docStore.activeSectionId : undefined

  if (selectedSkill.value) {
    const skill = selectedSkill.value
    text = skill.prompt + '\n\n' + text
  }

  chatStore.sendMessage(text, sectionId, docId, selectedModelId.value)
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

.input-container {
  position: relative;
  display: flex;
  align-items: flex-end;
  gap: 8px;
  background: #fff;
  border-radius: 12px;
  padding: 8px;
}

.skill-chip-overlay {
  position: absolute;
  bottom: 12px;
  left: 8px;
  background: #C23B22;
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  border-radius: 4px;
  padding: 1px 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  z-index: 5;
  user-select: none;
  white-space: nowrap;
  box-shadow: 0 1px 4px rgba(194, 59, 34, 0.2);
}

.skill-chip-overlay:hover {
  opacity: 0.85;
}

.chip-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  opacity: 0.8;
  transition: opacity 0.1s;
}

.chip-close:hover {
  opacity: 1;
}

.skill-popup {
  position: absolute;
  bottom: 100%;
  left: 0;
  right: 0;
  margin-bottom: 4px;
  background: #fff;
  border: 0.7px solid #E0D5C0;
  border-radius: 10px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  max-height: 200px;
  overflow-y: auto;
  z-index: 100;
  padding: 4px;
}

.skill-popup-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.1s;
}

.skill-popup-item:hover,
.skill-popup-item-active {
  background: rgba(194, 59, 34, 0.08);
}

.skill-popup-name {
  font-size: 12px;
  font-weight: 600;
  color: #3D2B1F;
  white-space: nowrap;
}

.skill-popup-desc {
  font-size: 11px;
  color: #8B7355;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.skill-popup-empty {
  padding: 12px;
  text-align: center;
  font-size: 12px;
  color: #8B7355;
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
  min-height: 140px;
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
