import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { runOutlineSkill, getGenerateStream } from '../api/client'
import type { OutlineSection } from '../api/client'

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
  const outline = ref<OutlineSection[]>([])
  const sectionStates = ref<Map<string, SectionState>>(new Map())
  const currentSectionId = ref<string | null>(null)
  const error = ref<string | null>(null)
  const modelConfig = ref<ModelConfig | null>(null)
  const sessionId = ref<string>('')
  const userMessage = ref<string>('')
  const skillPrompt = ref<string>('')
  let sses: EventSource[] = []

  const totalSections = computed(() => {
    let count = 0
    const countRecursive = (secs: OutlineSection[]) => {
      for (const s of secs) { count++; if ((s as any).children) countRecursive((s as any).children as OutlineSection[]) }
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
    sessionId.value = id
    userMessage.value = message
    skillPrompt.value = skillPromptText
    modelConfig.value = config
    error.value = null

    try {
      const mode = message ? 'hybrid' : 'auto'
      const res = await runOutlineSkill(id, {
        mode,
        user_input: message || undefined,
        provider: config.provider,
        model: config.model,
      })
      outline.value = res.data.sections || []
      phase.value = 'preview'
    } catch (err: any) {
      error.value = err?.response?.data?.error || err.message || '生成大纲失败'
      phase.value = 'error'
    }
  }

  function confirmGeneration() {
    phase.value = 'generating'
    const states = new Map<string, SectionState>()
    const initRecursive = (secs: OutlineSection[]) => {
      for (const s of secs) {
        states.set(s.section_id, 'pending')
        if ((s as any).children) initRecursive((s as any).children as OutlineSection[])
      }
    }
    initRecursive(outline.value)
    sectionStates.value = states
  }

  async function generateSection(sectionId: string, _section: OutlineSection, _sectionPath: string[], _outlineContext: OutlineSection[]) {
    if (!modelConfig.value || !sessionId.value) { error.value = 'no model config'; return }

    sectionStates.value.set(sectionId, 'generating')
    currentSectionId.value = sectionId

    try {
      const es = getGenerateStream(sessionId.value, sectionId, modelConfig.value.provider, modelConfig.value.model)
      sses.push(es)
      await new Promise<void>((resolve, reject) => {
        es.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data)
            if (data.chunk) {
              window.dispatchEvent(new CustomEvent('gen-chunk', { detail: { sectionId, chunk: data.chunk } }))
            } else if (data.done) {
              es.close()
              sses = sses.filter(x => x !== es)
              sectionStates.value.set(sectionId, 'done')
              currentSectionId.value = null
              window.dispatchEvent(new CustomEvent('gen-done', { detail: { sectionId } }))
              let allDone = true
              sectionStates.value.forEach(s => { if (s !== 'done') allDone = false })
              if (allDone) phase.value = 'done'
              resolve()
            } else if (data.error) {
              reject(new Error(data.error))
            }
          } catch { /* skip */ }
        }
        es.onerror = () => {
          es.close()
          sses = sses.filter(x => x !== es)
          reject(new Error('stream error'))
        }
      })
    } catch (err: any) {
      sectionStates.value.set(sectionId, 'error')
      currentSectionId.value = null
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
    sessionId.value = ''
    userMessage.value = ''
    skillPrompt.value = ''
    sses.forEach(es => es.close())
    sses = []
  }

  function getSectionState(sectionId: string): SectionState {
    return sectionStates.value.get(sectionId) || 'pending'
  }

  return {
    phase, outline, sectionStates, currentSectionId, error,
    modelConfig, sessionId, userMessage, skillPrompt,
    totalSections, completedSections, progressPercent,
    generateOutline: generateOutlineAction,
    confirmGeneration, generateSection, retrySection, reset, getSectionState,
  }
})
