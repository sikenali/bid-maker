<template>
  <div class="upload-view">
    <h1>Bid Maker</h1>
    <p>Upload a document to get started</p>
    <input type="file" accept=".docx,.pdf,.txt" @change="handleFileUpload" />
    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="docId" class="success">
      <p>Document uploaded: {{ docId }}</p>
      <router-link :to="`/editor/${docId}`">Open Editor</router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { uploadDocument } from '../api/client'

const router = useRouter()
const docId = ref('')
const error = ref('')

const handleFileUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  try {
    error.value = ''
    const res = await uploadDocument(file)
    docId.value = res.data.doc_id || res.data.id || ''
    if (docId.value) {
      router.push(`/editor/${docId.value}`)
    }
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Upload failed'
  }
}
</script>

<style scoped>
.upload-view {
  max-width: 600px;
  margin: 80px auto;
  padding: 40px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  text-align: center;
}

h1 {
  margin-bottom: 16px;
  color: #333;
}

input[type='file'] {
  margin: 20px 0;
  padding: 10px;
  border: 2px dashed #ccc;
  border-radius: 4px;
  width: 100%;
}

.error {
  color: red;
  margin-top: 16px;
}

.success {
  color: green;
  margin-top: 16px;
}

a {
  display: inline-block;
  margin-top: 12px;
  padding: 8px 20px;
  background: #42b883;
  color: white;
  border-radius: 4px;
  text-decoration: none;
}
</style>
