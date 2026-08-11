import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
})

export interface Section {
  id: string
  title: string
  level: number
  parent_id: string
  content: string
  children: Section[]
  pmPos?: number
}

export const uploadDocument = (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export const getOutline = (docId: string) => api.get(`/document/${docId}/outline`)
export const updateOutline = (docId: string, outline: any) => api.put(`/document/${docId}/outline`, outline)
export const getSection = (docId: string, sectionId: string) => api.get(`/document/${docId}/section/${sectionId}`)
export const saveSection = (docId: string, sectionId: string, content: string) => api.put(`/document/${docId}/section/${sectionId}`, { content })
export const exportDocument = (docId: string, format?: string) => api.post(`/document/${docId}/export`, { format }, {
  responseType: 'blob',
})
export const sendChat = (data: any) => api.post('/chat', data)
export const getModels = () => api.get('/config/models')
export const getTemplates = () => api.get('/templates')
export const getTemplate = (id: string) => api.get(`/templates/${id}`)
export const testApiKey = (data: { provider: string; model: string; key: string; endpoint?: string; format?: string }) => api.post('/config/test-key', data)

export const getLocalSkills = () => api.get('/local-skills')
export const getMarkdown = (docId: string) => api.get(`/document/${docId}/markdown`)
export const saveMarkdown = (docId: string, markdown: string) =>
  api.put(`/document/${docId}/markdown`, { markdown })
export const exportDocx = (docId: string) =>
  api.post(`/document/${docId}/export`, {}, {
    responseType: 'blob',
  })

export const postTemplate = (name: string, file: File) => {
  const formData = new FormData()
  formData.append('name', name)
  formData.append('file', file)
  return api.post('/templates', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export async function generateOutline(data: {
  provider: string
  model: string
  endpoint: string
  apiKey: string
  skill_prompt: string
  message: string
}): Promise<{ outline: Section[] }> {
  const res = await api.post('/generate-outline', data)
  return res.data
}

export async function generateSectionStream(
  data: {
    provider: string
    model: string
    endpoint: string
    apiKey: string
    section_id: string
    section: Section
    outline: Section[]
    user_prompt: string
  },
  onChunk: (sectionId: string, chunk: string) => void,
  onDone: (sectionId: string) => void,
  onError: (sectionId: string, error: string) => void
): Promise<void> {
  const response = await fetch('/api/generate-section', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!response.ok) {
    throw new Error(`generate-section failed: ${response.statusText}`)
  }
  const reader = response.body!.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''
    for (const line of lines) {
      const trimmed = line.trim()
      if (!trimmed.startsWith('data: ')) continue
      try {
        const parsed = JSON.parse(trimmed.slice(6))
        if (parsed.error) {
          onError(parsed.section_id, parsed.error)
        } else if (parsed.done) {
          onDone(parsed.section_id)
        } else if (parsed.chunk) {
          onChunk(parsed.section_id, parsed.chunk)
        }
      } catch { /* skip malformed */ }
    }
  }
}

export default api
