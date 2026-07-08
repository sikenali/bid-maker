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
        <button class="nav-btn" title="帮助" @click="showHelp">
          <span class="nav-btn-content">
            <RiQuestionLine size="20" />
            <span class="nav-btn-label">帮助</span>
          </span>
        </button>
        <button class="nav-btn" title="设置" @click="goSettings">
          <span class="nav-btn-content">
            <RiSettingsLine size="20" />
            <span class="nav-btn-label">设置</span>
          </span>
        </button>
      </div>
    </header>

    <main class="main">
      <div class="main-inner">
        <h2 class="upload-title">上传招标文件以生成标书</h2>
        <p class="upload-desc">系统将自动提取关键信息并生成标书大纲</p>

        <div
          class="upload-zone"
          :class="{ 'upload-zone-hover': isDragOver }"
          @drop.prevent="onDrop"
          @dragover.prevent="isDragOver = true"
          @dragleave.prevent="isDragOver = false"
          @click="triggerUpload"
        >
          <div class="upload-icon-wrap" :class="{ uploading: loading }">
            <RiUploadCloudLine v-if="!loading" :size="'36'" color="#C23B22" />
            <RiLoaderLine v-else :size="'36'" color="#5B8C5A" class="spin-icon" />
          </div>
          <p class="upload-hint">拖拽文件到此处，或点击选择文件</p>
          <p class="upload-format">支持 DOCX、MD 格式，单个文件不超过 50MB</p>
          <button class="btn-primary" @click.stop="triggerUpload">
            <RiAddFill size="18" color="#fff" />
            <span>选择文件</span>
          </button>
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
          <span class="progress-text">{{ Math.round(uploadProgress) }}%</span>
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
} from '@remixicon/vue'

const router = useRouter()
const isDragOver = ref(false)
const fileInput = ref<HTMLInputElement>()
const loading = ref(false)
const uploadProgress = ref(0)

const goHome = () => router.push('/')
const goSettings = () => router.push('/settings')
const showHelp = () => alert('文制星 - 标书智能生成工具\n\n上传招标文件 → 编辑大纲内容 → AI辅助写作 → 导出标书')

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
  background: #FDF6E3;
}

.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 64px;
  padding: 0 32px;
  background: #FDF6E3;
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
  width: 44px;
  height: 44px;
  background: #C23B22;
  border-radius: 8px;
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
  gap: 8px;
}

.nav-btn {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: #F5EFE0;
  border: 0.7px solid #E0D5C0;
  cursor: pointer;
  transition: all 0.2s;
  overflow: hidden;
  white-space: nowrap;
  color: #5C4033;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-btn:hover {
  width: 90px;
  background: #C23B22;
  border-color: transparent;
  color: #fff;
}

.nav-btn-content {
  display: flex;
  align-items: center;
  gap: 6px;
}

.nav-btn-label {
  font-size: 13px;
  font-weight: 500;
  color: inherit;
  display: none;
}

.nav-btn:hover .nav-btn-label {
  display: inline;
}

.main {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 64px;
  padding-left: 32px;
  padding-right: 32px;
}

.main-inner {
  width: 100%;
  max-width: 800px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.upload-title {
  font-size: 28px;
  font-weight: 700;
  color: #3D2B1F;
  margin: 0 0 8px;
  font-family: 'Noto Serif SC', serif;
}

.upload-desc {
  font-size: 16px;
  color: #8B7355;
  margin: 0 0 40px;
}

.upload-zone {
  width: 100%;
  border: 2px dashed #D4C4A8;
  border-radius: 12px;
  background: #FBF7EF;
  padding: 60px 40px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.2s, background 0.2s;
}

.upload-zone-hover {
  border-color: #C23B22;
  background: rgba(194, 59, 34, 0.05);
}

.upload-icon-wrap {
  width: 80px;
  height: 80px;
  background: #F0E8D5;
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

.upload-hint {
  font-size: 14px;
  color: #8B7355;
  margin-bottom: 8px;
}

.upload-format {
  font-size: 12px;
  color: #B8A88A;
  margin-bottom: 24px;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  background: #C23B22;
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-primary:hover {
  background: #A83028;
}

.progress-area {
  width: 100%;
  max-width: 400px;
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 40px;
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
</style>
