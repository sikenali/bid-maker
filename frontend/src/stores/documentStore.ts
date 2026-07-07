import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getOutline, getSection, saveSection, updateOutline } from '../api/client'

export interface Section {
  id: string
  title: string
  level: number
  parent_id: string
  content: string
  children: Section[]
}

export const useDocumentStore = defineStore('document', () => {
  const outline = ref<Section[]>([])
  const sections = ref<Map<string, Section>>(new Map())
  const activeSectionId = ref('')

  const loadOutline = async (docId: string) => {
    const res = await getOutline(docId)
    outline.value = res.data.outline
    res.data.outline.forEach((s: Section) => sections.value.set(s.id, s))
  }

  const loadSection = async (docId: string, sectionId: string) => {
    const res = await getSection(docId, sectionId)
    sections.value.set(sectionId, res.data)
    activeSectionId.value = sectionId
  }

  const saveSectionContent = async (docId: string, sectionId: string, content: string) => {
    await saveSection(docId, sectionId, content)
    const section = sections.value.get(sectionId)
    if (section) section.content = content
  }

  const updateOutlineTree = async (docId: string, newOutline: Section[]) => {
    await updateOutline(docId, newOutline)
    outline.value = newOutline
  }

  return { outline, sections, activeSectionId, loadOutline, loadSection, saveSectionContent, updateOutlineTree }
})
