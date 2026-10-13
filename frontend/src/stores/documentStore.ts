import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getSession, getMarkdown, updateOutline, updateSession } from '../api/client'

export interface OutlineSection {
  section_id: string
  title: string
  level: number
  order: number
  required: boolean
  parent_id: string | null
  source: string
  requirements: { ref: string; type: string }[]
  response_strategy: string
  response_basis: string
  children: string[]
  content?: string
  status?: string
  generated_at?: string
}

export const useDocumentStore = defineStore('document', () => {
  const outline = ref<OutlineSection[]>([])
  const activeSectionId = ref('')
  const markdown = ref('')
  const session = ref<any>(null)

  const loadSession = async (sessionId: string) => {
    const res = await getSession(sessionId)
    session.value = res.data
    
    if (res.data.outline?.sections) {
      outline.value = res.data.outline.sections
    }
    
    if (res.data.markdown) {
      markdown.value = res.data.markdown
    }
  }

  const loadMarkdown = async (sessionId: string) => {
    try {
      const res = await getMarkdown(sessionId)
      markdown.value = res.data.markdown || ''
    } catch {
      markdown.value = ''
    }
  }

  const saveMarkdown = async (sessionId: string, md: string) => {
    markdown.value = md
    await updateSession(sessionId, { markdown: md })
  }

  const updateOutlineTree = async (sessionId: string, newOutline: any) => {
    await updateOutline(sessionId, newOutline)
    outline.value = newOutline.sections || []
  }

  const getFullOutline = () => {
    return outline.value
  }

  const setActiveSection = (sectionId: string) => {
    activeSectionId.value = sectionId
  }

  const getActiveSection = () => {
    return outline.value.find(s => s.section_id === activeSectionId.value)
  }

  return {
    outline,
    activeSectionId,
    markdown,
    session,
    loadSession,
    loadMarkdown,
    saveMarkdown,
    updateOutlineTree,
    getFullOutline,
    setActiveSection,
    getActiveSection,
  }
})
