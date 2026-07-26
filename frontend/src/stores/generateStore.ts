import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { generateOutline, generateSectionStream } from '../api/client'
import type { Section } from '../api/client'

export type GenerationPhase = 'idle' | 'preview' | 'generating' | 'done' | 'error'
export type SectionState = 'pending' | 'generating' | 'done' | 'error'

interface ModelConfig {
  provider: string
  model: string
  endpoint: string
  format: string
  apiKey: string
}

export const useGenerateStore = defineStore('generate', () => {
  const phase = ref<GenerationPhase>('idle')
  const outline = ref<Section[]>([])
  const sectionStates = ref<Map<string, SectionState>>(new Map())
  const currentSectionId = ref<string | null>(null)
  const error = ref<string | null>(null)
  const modelConfig = ref<ModelConfig | null>(null)
  const docId = ref<string>('')
  const userMessage = ref<string>('')
  const skillPrompt = ref<string>('')

  const totalSections = computed(() => {
    let count = 0
    const countRecursive = (secs: Section[]) => {
      for (const s of secs) { count++; countRecursive(s.children) }
    }
    countRecursive(outline.value)
    return count
  })

  const completedSections = computed(() => {
    let count = 0
    sectionStates.value.forEach(s => { if (s === 'done') count++ })
    return count
  })

  const progressPercent = computed(() => {
    if (totalSections.value === 0) return 0
    return Math.round((completedSections.value / totalSections.value) * 100)
  })

  async function generateOutlineAction(
    id: string,
    message: string,
    skillPromptText: string,
    config: ModelConfig
  ) {
    docId.value = id
    userMessage.value = message
    skillPrompt.value = skillPromptText
    modelConfig.value = config
    error.value = null

    try {
      const result = await generateOutline({
        provider: config.provider,
        model: config.model,
        endpoint: config.endpoint,
        apiKey: config.apiKey,
        skill_prompt: skillPromptText,
        message,
      })
      outline.value = result.outline
      phase.value = 'preview'
    } catch (err: any) {
      error.value = err.message || '生成大纲失败'
      phase.value = 'error'
    }
  }

  function confirmGeneration() {
    phase.value = 'generating'
    const states = new Map<string, SectionState>()
    const initRecursive = (secs: Section[]) => {
      for (const s of secs) {
        states.set(s.id, 'pending')
        initRecursive(s.children)
      }
    }
    initRecursive(outline.value)
    sectionStates.value = states
  }

  async function generateSection(sectionId: string, section: Section, sectionPath: string[], outlineContext: Section[]) {
    if (!modelConfig.value) { error.value = 'no model config'; return }

    sectionStates.value.set(sectionId, 'generating')
    currentSectionId.value = sectionId

    try {
      await generateSectionStream(
        {
          provider: modelConfig.value.provider,
          model: modelConfig.value.model,
          endpoint: modelConfig.value.endpoint,
          apiKey: modelConfig.value.apiKey,
          section_id: sectionId,
          section,
          outline: outlineContext,
          user_prompt: `请编写章节「${section.title}」的详细内容`,
        },
        (sid, chunk) => {
          window.dispatchEvent(new CustomEvent('gen-chunk', { detail: { sectionId: sid, chunk } }))
        },
        (sid) => {
          sectionStates.value.set(sid, 'done')
          currentSectionId.value = null
          window.dispatchEvent(new CustomEvent('gen-done', { detail: { sectionId: sid } }))
          let allDone = true
          sectionStates.value.forEach(s => { if (s !== 'done') allDone = false })
          if (allDone) phase.value = 'done'
        },
        (sid, err) => {
          sectionStates.value.set(sid, 'error')
          currentSectionId.value = null
          error.value = err
        }
      )
    } catch (err: any) {
      sectionStates.value.set(sectionId, 'error')
      error.value = err.message || '生成失败'
    }
  }

  function retrySection(sectionId: string) {
    sectionStates.value.set(sectionId, 'pending')
  }

  function reset() {
    phase.value = 'idle'
    outline.value = []
    sectionStates.value = new Map()
    currentSectionId.value = null
    error.value = null
    modelConfig.value = null
    docId.value = ''
    userMessage.value = ''
    skillPrompt.value = ''
  }

  function getSectionState(sectionId: string): SectionState {
    return sectionStates.value.get(sectionId) || 'pending'
  }

  return {
    phase, outline, sectionStates, currentSectionId, error,
    modelConfig, docId, userMessage, skillPrompt,
    totalSections, completedSections, progressPercent,
    generateOutline: generateOutlineAction,
    confirmGeneration, generateSection, retrySection, reset, getSectionState,
  }
})
