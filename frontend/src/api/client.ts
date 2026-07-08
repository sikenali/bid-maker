import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
})

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
export const exportDocument = (docId: string) => api.post(`/document/${docId}/export`, null, {
  responseType: 'blob',
})
export const sendChat = (data: any) => api.post('/chat', data)
export const getModels = () => api.get('/config/models')

export default api
