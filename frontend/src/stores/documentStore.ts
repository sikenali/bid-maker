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
  pmPos?: number
}

export const useDocumentStore = defineStore('document', () => {
  const outline = ref<Section[]>([])
  const sections = ref<Map<string, Section>>(new Map())
  const activeSectionId = ref('')
  const docxBuffer = ref<ArrayBuffer | null>(null)
  const headingPositions = ref<Map<string, { pmPos: number; timestamp: number }>>(new Map())
  const CACHE_TTL = 3000

  const loadOutline = async (docId: string) => {
    const res = await getOutline(docId)
    const items = res.data.outline || []
    outline.value = items.map((s: any) => ({ ...s, pmPos: s.pmPos ?? undefined }))
    items.forEach((s: Section) => sections.value.set(s.id, s))
  }

  const loadSection = async (docId: string, sectionId: string) => {
    try {
      const res = await getSection(docId, sectionId)
      sections.value.set(sectionId, res.data)
    } catch {
      // Section not found in backend — keep local data (e.g. newly added sections)
    }
    activeSectionId.value = sectionId
  }

  const saveSectionContent = async (docId: string, sectionId: string, content: string) => {
    try {
      await saveSection(docId, sectionId, content)
    } catch {
      // Section may not exist in backend yet — just update local state
    }
    const section = sections.value.get(sectionId)
    if (section) section.content = content
  }

  const updateOutlineTree = async (docId: string, newOutline: Section[]) => {
    await updateOutline(docId, newOutline)
    outline.value = newOutline
  }

  const getFullOutline = () => {
    return outline.value
  }

  const setDocxBuffer = (buffer: ArrayBuffer) => {
    docxBuffer.value = buffer
  }

  const syncHeadingInfo = (headings: Array<{text: string; level: number; pmPos: number}>) => {
    const tree = outline.value
    for (const h of headings) {
      let target: Section | null = null
      for (let i = 0; i < tree.length; i++) {
        if (tree[i].title === h.text) {
          target = tree[i]
          break
        }
        if (tree[i].children?.length) {
          for (const child of tree[i].children) {
            if (child.title === h.text) {
              target = child
              break
            }
          }
        }
        if (target) break
      }
      if (target && target.pmPos === undefined) {
        target.pmPos = h.pmPos
      }
    }
  }

  const setHeadingPosition = (sectionId: string, pmPos: number) => {
    headingPositions.value.set(sectionId, { pmPos, timestamp: Date.now() })
  }

  const clearHeadingPositions = () => {
    headingPositions.value.clear()
  }

  const syncHeadingFromEditor = (headings: Array<{ text: string; pmPos: number }>) => {
    const search = (sections: Section[]): void => {
      for (const s of sections) {
        for (const h of headings) {
          if (s.title === h.text || h.text.includes(s.title)) {
            s.pmPos = h.pmPos
            headingPositions.value.set(s.id, { pmPos: h.pmPos, timestamp: Date.now() })
          }
        }
        if (s.children?.length) search(s.children)
      }
    }
    search(outline.value)
  }

  const syncTitleFromEditor = (pmPos: number, newTitle: string) => {
    const search = (sections: Section[]): boolean => {
      for (const s of sections) {
        if (s.pmPos === pmPos) {
          s.title = newTitle
          return true
        }
        if (s.children?.length && search(s.children)) return true
      }
      return false
    }
    search(outline.value)
  }

  const getSectionPmPos = (sectionId: string): number | undefined => {
    const cached = headingPositions.value.get(sectionId)
    if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
      return cached.pmPos
    }
    const search = (sections: Section[]): number | undefined => {
      for (const s of sections) {
        if (s.id === sectionId) return s.pmPos
        if (s.children?.length) {
          const found = search(s.children)
          if (found !== undefined) return found
        }
      }
      return undefined
    }
    return search(outline.value)
  }

  return { outline, sections, activeSectionId, docxBuffer, headingPositions, loadOutline, loadSection, saveSectionContent, updateOutlineTree, getFullOutline, setDocxBuffer, syncHeadingInfo, setHeadingPosition, clearHeadingPositions, syncHeadingFromEditor, syncTitleFromEditor, getSectionPmPos }
})
