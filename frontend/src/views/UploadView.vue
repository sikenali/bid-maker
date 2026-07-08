<template>
  <div class="page">
    <header class="navbar">
      <div class="logo-area" @click="goHome">
        <div class="logo-icon">
          <RiRadarFill size="22" color="#fff" />
        </div>
        <div class="logo-texts">
          <span class="brand-zh">文制星</span>
          <span class="brand-en">Boomerang</span>
        </div>
      </div>
      <div class="nav-actions">
        <button class="nav-btn" title="帮助">
          <RiQuestionLine size="18" color="#8B7355" />
        </button>
        <button class="nav-btn" title="设置" @click="goSettings">
          <RiSettingsLine size="18" color="#8B7355" />
        </button>
      </div>
    </header>

    <main class="main">
      <div class="main-inner">
        <div
          class="upload-zone"
          :class="{ 'upload-zone-hover': isDragOver }"
          @drop.prevent="onDrop"
          @dragover.prevent="isDragOver = true"
          @dragleave.prevent="isDragOver = false"
          @click="triggerUpload"
        >
          <div class="upload-icon-wrap" :class="{ uploading: loading }">
            <RiUploadCloudLine v-if="!loading" :size="'36'" color="#C43D3D" />
            <RiLoaderLine v-else :size="'36'" color="#5B8C5A" class="spin-icon" />
          </div>
          <h2 class="upload-title">上传招标文件以生成标书</h2>
          <p class="upload-desc">系统将自动提取关键信息并生成标书大纲</p>
          <p class="upload-format">支持 DOCX、MD格式，单个文件不超过 50MB</p>
          <button class="btn-primary" @click.stop="triggerUpload">
            <RiAddFill size="18" color="#fff" />
            <span>选择文件</span>
          </button>
          <p class="upload-hint">或拖拽文件到此处</p>
          <input
            ref="fileInput"
            type="file"
            accept=".docx,.md,.txt"
            hidden
            @change="onFileSelected"
          />
        </div>

        <div v-if="uploadProgress > 0" class="progress-area">
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: uploadProgress + '%' }" />
          </div>
          <span class="progress-text">{{ uploadProgress }}%</span>
        </div>

        <div class="bottom-actions">
          <button class="btn-outline" disabled>
            <RiBracesFill size="18" color="#8B7355" />
            <span>大纲提取</span>
          </button>
          <button class="btn-gradient" disabled>
            <RiSparklingFill size="18" color="#fff" />
            <span>标书生成</span>
          </button>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { uploadDocument } from '../api/client'
import {
  RiRadarFill,
  RiQuestionLine,
  RiSettingsLine,
  RiUploadCloudLine,
  RiLoaderLine,
  RiAddFill,
  RiBracesFill,
  RiSparklingFill,
} from '@remixicon/vue'

const router = useRouter()
const isDragOver = ref(false)
const fileInput = ref<HTMLInputElement>()
const loading = ref(false)
const uploadProgress = ref(0)

const goHome = () => router.push('/')
const goSettings = () => router.push('/settings')

const triggerUpload = () => fileInput.value?.click()

const handleFile = async (file: File) => {
  loading.value = true
  uploadProgress.value = 0
  const interval = setInterval(() => {
    if (uploadProgress.value < 90) {
      uploadProgress.value += Math.random() * 15
    }
  }, 300)
  try {
    const res = await uploadDocument(file)
    clearInterval(interval)
    uploadProgress.value = 100
    setTimeout(() => router.push(`/editor/${res.data.id}`), 300)
  } catch (err) {
    clearInterval(interval)
    uploadProgress.value = 0
    console.error('Upload failed:', err)
    alert('上传失败，请重试。')
  } finally {
    loading.value = false
  }
}

const onFileSelected = (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (file) handleFile(file)
}

const onDrop = (e: DragEvent) => {
  isDragOver.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file) handleFile(file)
}
</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: #FBF7F0;
}

.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 64px;
  padding: 0 24px;
  background: #FBF7F0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  z-index: 50;
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.logo-icon {
  width: 40px;
  height: 40px;
  background: #C43D3D;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.logo-texts {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.brand-zh {
  font-size: 22px;
  font-weight: 700;
  color: #3D2B1F;
  line-height: 1.2;
}

.brand-en {
  font-size: 12px;
  color: #8B7355;
  line-height: 1.3;
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.nav-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #F0E8D8;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.2s;
}

.nav-btn:hover {
  background: #E4D9C4;
}

.main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 64px;
}

.main-inner {
  width: 100%;
  max-width: 800px;
  padding: 32px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.upload-zone {
  width: 100%;
  height: 596px;
  border: 2px dashed #D4C5A9;
  border-radius: 16px;
  background: #F5EFE3;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.2s, background 0.2s;
}

.upload-zone-hover {
  border-color: #C43D3D;
  background: #FAF5EE;
}

.upload-icon-wrap {
  width: 80px;
  height: 80px;
  background: #F0E8D8;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
  transition: background 0.3s;
}

.upload-icon-wrap.uploading {
  background: rgba(91, 140, 90, 0.1);
}

.spin-icon {
  animation: spin 1.5s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.upload-title {
  font-size: 20px;
  font-weight: 600;
  color: #3D2B1F;
  margin: 0 0 8px;
}

.upload-desc {
  font-size: 14px;
  color: #9B8C7C;
  margin: 0 0 8px;
}

.upload-format {
  font-size: 12px;
  color: #B8A88A;
  margin: 0 0 24px;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 32px;
  border: none;
  border-radius: 12px;
  background: #C43D3D;
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-primary:hover {
  background: #A83232;
}

.upload-hint {
  font-size: 12px;
  color: #B8A88A;
  margin: 16px 0 0;
}

.progress-area {
  width: 100%;
  max-width: 400px;
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 24px;
}

.progress-bar {
  flex: 1;
  height: 8px;
  background: #F0E8D5;
  border-radius: 9999px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: #5B8C5A;
  border-radius: 9999px;
  transition: all 0.3s ease-out;
}

.progress-text {
  font-size: 12px;
  color: #5B8C5A;
  font-weight: 600;
  white-space: nowrap;
}

.bottom-actions {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-top: 24px;
}

.btn-outline {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 32px;
  border: 0.7px solid #D4C5A9;
  border-radius: 12px;
  background: #F0E8D8;
  color: #5C4A3A;
  font-size: 15px;
  font-weight: 600;
  cursor: not-allowed;
  opacity: 0.7;
}

.btn-gradient {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 32px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(to bottom, #C43D3D, #A83232);
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  cursor: not-allowed;
  opacity: 0.7;
}
</style>
