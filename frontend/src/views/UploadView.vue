<template>
  <div class="page">
    <!-- Top Navbar -->
    <header class="navbar">
      <div class="logo-area">
        <div class="logo-icon">
          <Icon name="RadarChartFilled" :size="22" color="#fff" />
        </div>
        <span class="brand-zh">文制猩</span>
        <span class="brand-en">Boomerang</span>
      </div>
      <div class="nav-actions">
        <button class="icon-btn" title="Help">
          <Icon name="InformationCircleFilled" :size="18" color="#8b7355" />
        </button>
        <button class="icon-btn" title="Settings">
          <Icon name="Settings4Filled" :size="18" color="#8b7355" />
        </button>
      </div>
    </header>

    <!-- Main Content -->
    <main class="main">
      <div class="content-area">
        <!-- Upload Zone -->
        <div
          class="upload-zone"
          :class="{ 'upload-zone-hover': isDragOver }"
          @drop.prevent="onDrop"
          @dragover.prevent="isDragOver = true"
          @dragleave.prevent="isDragOver = false"
          @click="triggerUpload"
        >
          <div class="upload-icon-wrap">
            <Icon name="Upload2Filled" :size="36" color="#c43d3d" />
          </div>
          <h2 class="upload-title">上传招标文件以生成标书</h2>
          <p class="upload-desc">系统将自动提取关键信息并生成标书大纲</p>
          <p class="upload-format">支持 DOCX、MD格式，单个文件不超过 50MB</p>
          <div class="upload-actions">
            <button class="btn-primary" @click.stop="triggerUpload">
              <Icon name="Upload2Filled" :size="18" color="#fff" />
              <span>选择文件</span>
            </button>
          </div>
          <p class="upload-hint">或拖拽文件到此处</p>
          <input
            ref="fileInput"
            type="file"
            accept=".docx,.md,.txt"
            hidden
            @change="onFileSelected"
          />
        </div>

        <!-- Bottom Actions -->
        <div class="bottom-actions">
          <button class="btn-outline" disabled>
            <Icon name="BracesFilled" :size="18" color="#8b7355" />
            <span>大纲提取</span>
          </button>
          <button class="btn-gradient" disabled>
            <Icon name="Book2Filled" :size="18" color="#fff" />
            <span>标书生成</span>
          </button>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Icon } from '@remixicon/vue'

const isDragOver = ref(false)
const fileInput = ref<HTMLInputElement>()

const triggerUpload = () => fileInput.value?.click()

const onFileSelected = (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (file) {
    console.log('Selected file:', file.name)
    // TODO: integrate with uploadDocument API
  }
}

const onDrop = (e: DragEvent) => {
  isDragOver.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file) {
    console.log('Dropped file:', file.name)
    // TODO: integrate with uploadDocument API
  }
}
</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: #fbf7f0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Source Han Sans CN', 'Microsoft YaHei', sans-serif;
}

/* Navbar */
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 24px;
  background: #fbf7f0;
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #c43d3d;
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-zh {
  font-size: 22px;
  font-weight: 700;
  color: #3d2b1f;
}

.brand-en {
  font-size: 12px;
  font-weight: 400;
  color: #8b7355;
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.icon-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: #f0e8d8;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.2s;
}

.icon-btn:hover {
  background: #e4d9c4;
}

/* Main */
.main {
  flex: 1;
  display: flex;
  align-items: flex-start;
  justify-content: flex-start;
  padding-top: 32px;
}

.content-area {
  width: 100%;
  padding: 32px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

/* Upload Zone */
.upload-zone {
  width: 100%;
  height: 596px;
  border: 2px solid #d4c5a9;
  border-style: dashed;
  border-radius: 16px;
  background: #f5efe3;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.2s, background 0.2s;
}

.upload-zone-hover {
  border-color: #c43d3d;
  background: #faf5ee;
}

.upload-icon-wrap {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: #f0e8d8;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
}

.upload-title {
  font-size: 20px;
  font-weight: 600;
  color: #3d2b1f;
  margin: 0 0 8px;
}

.upload-desc {
  font-size: 14px;
  color: #9b8c7c;
  margin: 0 0 8px;
}

.upload-format {
  font-size: 12px;
  color: #b8a88a;
  margin: 0 0 24px;
}

.upload-actions {
  margin-bottom: 16px;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 32px;
  border: none;
  border-radius: 12px;
  background: #c43d3d;
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-primary:hover {
  background: #a83232;
}

.upload-hint {
  font-size: 12px;
  color: #b8a88a;
  margin: 0;
}

/* Bottom Actions */
.bottom-actions {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 24px;
}

.btn-outline {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 32px;
  border: 0.7px solid #d4c5a9;
  border-radius: 12px;
  background: #f0e8d8;
  color: #5c4a3a;
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
  background: linear-gradient(to bottom, #c43d3d, #a83232);
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  cursor: not-allowed;
  opacity: 0.7;
}
</style>
