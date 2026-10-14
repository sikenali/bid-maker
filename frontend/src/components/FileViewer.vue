<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { OpenFileViewer } from '@open-file-viewer/vue'
import { imagePlugin, officePlugin, pdfPlugin, textPlugin } from '@open-file-viewer/core'
import '@open-file-viewer/core/style.css'
import pdfWorkerSrc from 'pdfjs-dist/build/pdf.worker.mjs?url'

const props = defineProps<{
  file?: File | null
  width?: string
  height?: string
}>()

const viewerRef = ref<InstanceType<typeof OpenFileViewer> | null>(null)
const error = ref('')
const loading = ref(false)

const plugins = [
  imagePlugin(),
  textPlugin(),
  pdfPlugin({ workerSrc: pdfWorkerSrc }),
  officePlugin(),
]

const handleLoad = () => {
  loading.value = false
  error.value = ''
}

const handleError = (e: any) => {
  loading.value = false
  error.value = e?.message || '文件预览失败'
}

const handleUnsupported = () => {
  loading.value = false
  error.value = '不支持的文件格式'
}

onMounted(() => {
  if (props.file) {
    loading.value = true
  }
})
</script>

<template>
  <div class="file-viewer">
    <div v-if="loading" class="viewer-loading">
      <span class="spinner" />
      <span>加载中...</span>
    </div>
    <div v-else-if="error" class="viewer-error">
      <span>{{ error }}</span>
    </div>
    <OpenFileViewer
      v-else-if="file"
      ref="viewerRef"
      :file="file"
      :file-name="file.name"
      :width="width || '100%'"
      :height="height || '600px'"
      fit="contain"
      :toolbar="true"
      :plugins="plugins"
      @load="handleLoad"
      @error="handleError"
      @unsupported="handleUnsupported"
    />
    <div v-else class="viewer-empty">
      <span>暂无文件</span>
    </div>
  </div>
</template>

<style scoped>
.file-viewer {
  width: 100%;
  height: 100%;
  position: relative;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
}

.viewer-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 400px;
  gap: 12px;
  color: #8B7355;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #F0E8D5;
  border-top-color: #C23B22;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.viewer-error {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 400px;
  color: #C23B22;
  background: #FFF5F5;
  border-radius: 8px;
}

.viewer-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 400px;
  color: #8B7355;
  background: #FDF6E3;
  border-radius: 8px;
}
</style>
