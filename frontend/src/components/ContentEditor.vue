<template>
  <div class="content-editor">
    <div class="editor-header">
      <span class="current-section">{{ currentSectionTitle }}</span>
      <div class="toolbar">
        <button title="AI Assist">&#10024;&#65039;</button>
        <button title="Save">&#128190;</button>
      </div>
    </div>
    <div class="editor-body">
      <div ref="editorRef" class="editor-container"></div>
    </div>
    <div class="editor-footer">
      <button class="outline-btn">Extract Outline</button>
      <button class="export-btn">Generate Bid</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const currentSectionTitle = computed(() => {
  return 'Editing...'
})
const editorRef = ref<HTMLElement>()
</script>

<style scoped>
.content-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #eee;
}
.toolbar {
  display: flex;
  gap: 8px;
}
.editor-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 0;
}
.editor-footer {
  display: flex;
  justify-content: center;
  gap: 16px;
  padding: 16px 0;
}
.outline-btn, .export-btn {
  padding: 12px 24px;
  border-radius: 6px;
  border: 1px solid #ddd;
  background: white;
  cursor: pointer;
}
.export-btn {
  background: #1677ff;
  color: white;
  border: none;
}
</style>
