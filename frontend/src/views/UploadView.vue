<template>
  <div class="upload-view">
    <div class="upload-area" @drop.prevent="onDrop" @dragover.prevent @click.self="triggerUpload">
      <div class="upload-icon">
        <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="17 8 12 3 7 8"/>
          <line x1="12" y1="3" x2="12" y2="15"/>
        </svg>
      </div>
      <p class="upload-text">Drop your .docx file here, or</p>
      <button class="upload-button" @click.stop="triggerUpload">Upload</button>
      <input
        ref="fileInput"
        type="file"
        accept=".docx"
        hidden
        @change="onFileSelected"
      />
    </div>
    <div v-if="loading" class="loading">Processing document...</div>
    <div v-if="error" class="error">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { uploadDocument } from '../api/client'

const router = useRouter()
const fileInput = ref<HTMLInputElement>()
const loading = ref(false)
const error = ref('')

const triggerUpload = () => fileInput.value?.click()

const onFileSelected = async (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  await handleFile(file)
}

const onDrop = async (e: DragEvent) => {
  const file = e.dataTransfer?.files?.[0]
  if (!file) return
  await handleFile(file)
}

const handleFile = async (file: File) => {
  error.value = ''
  loading.value = true
  try {
    const res = await uploadDocument(file)
    router.push(`/editor/${res.data.id}`)
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Upload failed. Please try again.'
    console.error('Upload failed:', err)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.upload-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 40px;
  background: #f5f5f5;
}

.upload-area {
  border: 2px dashed #ccc;
  border-radius: 12px;
  padding: 60px;
  text-align: center;
  cursor: pointer;
  width: 600px;
  background: white;
  transition: border-color 0.2s, background 0.2s;
}

.upload-area:hover {
  border-color: #666;
  background: #fafafa;
}

.upload-icon {
  margin-bottom: 16px;
  color: #999;
}

.upload-text {
  margin: 0 0 8px;
  font-size: 16px;
  color: #666;
}

.upload-button {
  margin-top: 16px;
  padding: 12px 32px;
  background: #1677ff;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  transition: background 0.2s;
}

.upload-button:hover {
  background: #0958d9;
}

.loading {
  margin-top: 16px;
  color: #666;
  font-size: 14px;
}

.error {
  margin-top: 16px;
  color: #ff4d4f;
  font-size: 14px;
}
</style>
