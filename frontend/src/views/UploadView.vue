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
        <button class="nav-btn" title="上传" @click="goHome">
          <span class="nav-btn-content">
            <RiQuestionLine size="20" />
            <span class="nav-btn-label">上传</span>
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
          v-if="!parsedFile"
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
          <p class="upload-format">支持 DOCX/XLSX/PPTX/PDF 格式，单个文件不超过 50MB</p>
          <button class="btn-primary" @click.stop="triggerUpload">
            <RiAddFill size="18" color="#fff" />
            <span>选择文件</span>
          </button>
          <input
            ref="fileInput"
            type="file"
            accept=".docx,.xlsx,.pptx,.pdf"
            hidden
            @change="onFileSelected"
          />
        </div>

        <div v-if="parsedFile" class="preview-area">
          <div class="preview-header">
            <span class="preview-title">{{ parsedFile.name }}</span>
            <button class="btn-change" @click="resetUpload">
              <RiRefreshLine size="16" />
              重新上传
            </button>
          </div>
          <FileViewer :file="parsedFile" height="400px" />
        </div>

        <div v-if="uploadProgress > 0" class="progress-area">
          <div class="progress-bar">
            <div class="progress-fill" :style="{ width: uploadProgress + '%' }" />
          </div>
          <span class="progress-text">{{ Math.round(uploadProgress) }}%</span>
        </div>

        <div v-if="parseResult" class="result-area">
          <div class="result-header">
            <span class="result-title">解析完成</span>
            <span class="result-meta">{{ parseResult.source_meta.title }} | {{ parseResult.source_meta.bid_no }}</span>
          </div>
          <div class="result-stats">
            <div class="stat-item">
              <span class="stat-value">{{ parseResult.template.preset_sections.length }}</span>
              <span class="stat-label">章节</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ parseResult.technical_indicators.length }}</span>
              <span class="stat-label">技术指标</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ parseResult.scoring_points.length }}</span>
              <span class="stat-label">评分点</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ parseResult.disqualify_clauses.length }}</span>
              <span class="stat-label">废标条款</span>
            </div>
          </div>
          <button class="btn-primary" @click="goToEditor">
            <RiArrowRightLine size="18" color="#fff" />
            <span>进入编辑</span>
          </button>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { createSession, runParseSkill } from '../api/client'
import FileViewer from '../components/FileViewer.vue'
import {
  RiRadarFill,
  RiQuestionLine,
  RiSettingsLine,
  RiUploadCloudLine,
  RiLoaderLine,
  RiAddFill,
  RiRefreshLine,
  RiArrowRightLine,
} from '@remixicon/vue'

const router = useRouter()
const isDragOver = ref(false)
const fileInput = ref<HTMLInputElement>()
const loading = ref(false)
const uploadProgress = ref(0)
const parsedFile = ref<File | null>(null)
const parseResult = ref<any>(null)
let currentSessionId = ''

const goHome = () => router.push('/')
const goSettings = () => router.push('/settings')
const triggerUpload = () => fileInput.value?.click()
const resetUpload = () => {
  parsedFile.value = null
  parseResult.value = null
  uploadProgress.value = 0
  currentSessionId = ''
}

const goToEditor = () => {
  if (currentSessionId) {
    router.push(`/editor/${currentSessionId}`)
  }
}

const handleFile = async (file: File) => {
  loading.value = true
  uploadProgress.value = 0
  parsedFile.value = file
  
  // 模拟进度条
  const interval = setInterval(() => {
    if (uploadProgress.value < 30) {
      uploadProgress.value += Math.random() * 10
    }
  }, 200)

  try {
    // 创建会话
    const sessionRes = await createSession({
      title: file.name.replace(/\.[^.]+$/, ''),
      bid_no: '',
    })
    currentSessionId = sessionRes.data.session_id
    
    // 运行解析 Skill
    const parseRes = await runParseSkill(currentSessionId, file)
    const taskId = parseRes.data.task_id
    
    // 轮询任务状态
    const pollInterval = setInterval(async () => {
      try {
        // 获取会话状态查看解析结果
        const sessionRes = await createSession({ title: '', bid_no: '' }) // 这里需要getSession，暂时用轮询
        uploadProgress.value = Math.min(90, uploadProgress.value + 5)
      } catch {
        // 忽略
      }
    }, 1000)

    // 等待解析完成（简化：直接轮询）
    await new Promise<void>((resolve) => {
      const checkTask = setInterval(async () => {
        try {
          // 直接通过会话获取结果
          const res = await fetch(`/api/sessions/${currentSessionId}`)
          const data = await res.json()
          if (data.status === 'completed' || data.parsing_output) {
            clearInterval(checkTask)
            clearInterval(interval)
            uploadProgress.value = 100
            parseResult.value = data.parsing_output || data
            setTimeout(() => resolve(), 500)
          }
        } catch {
          // 继续轮询
        }
      }, 1000)
      
      // 超时保护
      setTimeout(() => {
        clearInterval(checkTask)
        resolve()
      }, 30000)
    })

    clearInterval(interval)
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
  max-width: 900px;
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

.preview-area {
  width: 100%;
  margin-top: 24px;
}

.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.preview-title {
  font-size: 16px;
  font-weight: 600;
  color: #3D2B1F;
}

.btn-change {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: 1px solid #D4C4A8;
  border-radius: 6px;
  background: #fff;
  color: #5C4033;
  font-size: 13px;
  cursor: pointer;
}

.btn-change:hover {
  background: #F5EFE0;
}

.result-area {
  width: 100%;
  margin-top: 32px;
  padding: 24px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

.result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.result-title {
  font-size: 18px;
  font-weight: 600;
  color: #5B8C5A;
}

.result-meta {
  font-size: 13px;
  color: #8B7355;
}

.result-stats {
  display: flex;
  gap: 32px;
  margin-bottom: 24px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #C23B22;
}

.stat-label {
  font-size: 13px;
  color: #8B7355;
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
