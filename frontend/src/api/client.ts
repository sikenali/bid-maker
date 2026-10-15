import axios from 'axios'
import { saveAs } from 'file-saver'

const api = axios.create({
  baseURL: '/api',
})

// ========== 新后端 API ==========

export interface Session {
  session_id: string
  title: string
  bid_no: string
  parsing_output?: ParsingOutput
  outline?: Outline
  markdown: string
  current_section?: string
  created_at: string
  updated_at: string
}

export interface ParsingOutput {
  source_meta: SourceMeta
  template: Template
  technical_indicators: TechnicalIndicator[]
  commercial_reqs: CommercialReq[]
  scoring_points: ScoringPoint[]
  disqualify_clauses: DisqualifyClause[]
}

export interface SourceMeta {
  title: string
  bid_no: string
  issuer: string
  parse_time: string
  file_hash: string
  total_pages: number
}

export interface Template {
  has_preset_structure: boolean
  preset_sections: SectionNode[]
  format_rules: FormatRules
}

export interface SectionNode {
  section_id: string
  title: string
  level: number
  order: number
  required: boolean
  parent_id: string | null
}

export interface FormatRules {
  font: string
  font_size: string
  has_cover: boolean
  has_toc: boolean
  seal_pages: string[]
  other: Record<string, any>
}

export interface TechnicalIndicator {
  indicator_id: string
  name: string
  requirement: string
  clause_ref: string
  is_mandatory: boolean
  unit?: string
  min_value?: number
  max_value?: number
  response_type: string
}

export interface CommercialReq {
  req_id: string
  category: string
  title: string
  description: string
  clause_ref: string
  deliverable: string[]
}

export interface ScoringPoint {
  scoring_id: string
  scoring_item: string
  full_score: number
  description: string
  clause_ref: string
  grading_tiers: GradingTier[]
  required_docs: string[]
}

export interface GradingTier {
  tier: string
  range: string
  criteria: string
}

export interface DisqualifyClause {
  clause_id: string
  clause_text: string
  clause_ref: string
  severity: string
  check_type: string
  required_docs: string[]
  deadline?: string
}

export interface Outline {
  outline_id: string
  source_mode: string
  locked: boolean
  sections: OutlineSection[]
}

export interface OutlineSection {
  section_id: string
  title: string
  level: number
  order: number
  required: boolean
  parent_id: string | null
  source: string
  requirements: ReqRef[]
  response_strategy: string
  response_basis: string
  children: string[]
  content?: string
  status?: string
  generated_at?: string
}

export interface ReqRef {
  ref: string
  type: string
}

export interface Task {
  task_id: string
  session_id: string
  type: string
  status: string
  progress: number
  section_id?: string
  result?: Record<string, any>
  error?: string
  created_at: string
  updated_at: string
}

export const createSession = (data: { title: string; bid_no?: string }) =>
  api.post<Session>('/sessions', data)

export const getSession = (sessionId: string) =>
  api.get<Session>(`/sessions/${sessionId}`)

export const updateSession = (sessionId: string, data: Partial<Session>) =>
  api.put<Session>(`/sessions/${sessionId}`, data)

export const runParseSkill = (sessionId: string, file: File) => {
  const form = new FormData()
  form.append('file', file)
  return api.post<{ task_id: string; status: string }>(
    `/sessions/${sessionId}/skills/parse`,
    form,
    { headers: { 'Content-Type': 'multipart/form-data' } }
  )
}

export const runOutlineSkill = (sessionId: string, data: {
  mode: 'auto' | 'custom' | 'hybrid'
  user_input?: string
  provider?: string
  model?: string
}) =>
  api.post<Outline>(`/sessions/${sessionId}/skills/outline`, data)

export const updateOutline = (sessionId: string, outline: Outline) =>
  api.put(`/sessions/${sessionId}/outline`, outline)

export const lockOutline = (sessionId: string) =>
  api.post<{ message: string; outline_id: string }>(`/sessions/${sessionId}/outline/lock`)

export const generateSection = (sessionId: string, data: {
  section_id: string
  provider?: string
  model?: string
}) =>
  api.post<{ task_id: string; status: string }>(
    `/sessions/${sessionId}/skills/generate/section`,
    data
  )

export const generateFull = (sessionId: string, data: {
  provider?: string
  model?: string
}) =>
  api.post<{ task_id: string; status: string }>(
    `/sessions/${sessionId}/skills/generate/full`,
    data
  )

export const getGenerateFullStream = (sessionId: string, provider?: string, model?: string): EventSource => {
  const params = new URLSearchParams()
  if (provider) params.set('provider', provider)
  if (model) params.set('model', model)
  const url = `/sessions/${sessionId}/skills/generate/full/stream?${params}`
  return new EventSource(url)
}

export const getGenerateStream = (sessionId: string, sectionId: string, provider?: string, model?: string): EventSource => {
  const params = new URLSearchParams({ section_id: sectionId })
  if (provider) params.set('provider', provider)
  if (model) params.set('model', model)
  const url = `/sessions/${sessionId}/skills/generate/stream?${params}`
  return new EventSource(url)
}

export const getMarkdown = (sessionId: string) =>
  api.get<{ markdown: string }>(`/sessions/${sessionId}/export/markdown`)

export const exportDocx = (sessionId: string): Promise<Blob> =>
  api.get(`/sessions/${sessionId}/export`, { responseType: 'blob' }).then(r => r.data)

export const exportPdf = (sessionId: string, options?: {
  high_res?: number; bookmarks?: number; watermark?: number; watermark_text?: string
}): Promise<Blob> => {
  const params = new URLSearchParams()
  if (options?.high_res !== undefined) params.set('high_res', String(options.high_res))
  if (options?.bookmarks !== undefined) params.set('bookmarks', String(options.bookmarks))
  if (options?.watermark !== undefined) params.set('watermark', String(options.watermark))
  if (options?.watermark_text !== undefined) params.set('watermark_text', options.watermark_text)
  const qs = params.toString()
  return api.get(`/sessions/${sessionId}/export/pdf${qs ? `?${qs}` : ''}`, { responseType: 'blob' }).then(r => r.data)
}

export const getTaskStatus = (taskId: string) =>
  api.get<Task>(`/tasks/${taskId}`)

export const listSkills = () =>
  api.get<{ skills: { name: string; description: string }[] }>('/skills')

export const listModels = () =>
  api.get<{ models: string[] }>('/models')

export const testApiKey = (data: {
  provider: string
  model: string
  key: string
  endpoint?: string
  format?: string
}) =>
  api.post<{ available: boolean; error?: string }>('/test-key', data)

// ========== 旧 API 保留（兼容过渡期）==========

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
export const updateOutlineLegacy = (docId: string, outline: any) => api.put(`/document/${docId}/outline`, outline)
export const getSection = (docId: string, sectionId: string) => api.get(`/document/${docId}/section/${sectionId}`)
export const saveSection = (docId: string, sectionId: string, content: string) => api.put(`/document/${docId}/section/${sectionId}`, { content })
export const exportDocument = (docId: string, format?: string) => api.post(`/document/${docId}/export`, { format }, {
  responseType: 'blob',
})
export const sendChat = (data: any) => api.post('/chat', data)
export const getModels = () => api.get('/config/models')
export const getTemplates = () => api.get('/templates')
export const getTemplate = (id: string) => api.get(`/templates/${id}`)
export const testApiKeyLegacy = (data: { provider: string; model: string; key: string; endpoint?: string; format?: string }) => api.post('/config/test-key', data)
export const getLocalSkills = () => api.get('/local-skills')
export const getMarkdownLegacy = (docId: string) => api.get(`/document/${docId}/markdown`)
export const saveMarkdown = (docId: string, markdown: string) => api.put(`/document/${docId}/markdown`, { markdown })
export const exportDocxLegacy = (docId: string) => api.post(`/document/${docId}/export`, {}, { responseType: 'blob' })
export const postTemplate = (name: string, file: File) => {
  const formData = new FormData()
  formData.append('name', name)
  formData.append('file', file)
  return api.post('/templates', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
}

export async function generateOutlineLegacy(data: {
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

export async function generateSectionStreamLegacy(
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

// Legacy aliases for backward compatibility
export { generateOutlineLegacy as generateOutline }
export { generateSectionStreamLegacy as generateSectionStream }

// file-saver 直接导出
export { saveAs }

export { api as apiClient }
export default api
