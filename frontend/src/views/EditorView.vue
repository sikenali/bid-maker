<template>
  <div class="editor-view">
    <header class="top-bar">
      <div class="logo-area">
        <div class="logo-icon">📋</div>
        <span class="brand-name">Bid-Maker</span>
      </div>
      <div class="right-actions">
        <button class="icon-btn" title="Help">?</button>
        <button class="icon-btn" title="Settings">⚙</button>
      </div>
    </header>
    <main class="editor-body">
      <aside class="left-panel">
        <OutlineTree @select="handleSelectSection" />
      </aside>
      <section class="center-panel">
        <ContentEditor />
      </section>
      <aside class="right-panel">
        <AIChat />
      </aside>
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, defineProps } from 'vue'
import { useDocumentStore } from '../stores/documentStore'
import OutlineTree from '../components/OutlineTree.vue'
import ContentEditor from '../components/ContentEditor.vue'
import AIChat from '../components/AIChat.vue'

const props = defineProps<{ id: string }>()
const docStore = useDocumentStore()

onMounted(() => {
  docStore.loadOutline(props.id)
})

const handleSelectSection = (sectionId: string) => {
  docStore.loadSection(props.id, sectionId)
}
</script>

<style scoped>
.editor-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
}
.top-bar {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  border-bottom: 1px solid #eee;
}
.logo-area {
  display: flex;
  align-items: center;
  gap: 10px;
}
.logo-icon {
  font-size: 22px;
}
.brand-name {
  font-weight: bold;
  font-size: 18px;
}
.right-actions {
  display: flex;
  gap: 8px;
}
.icon-btn {
  width: 36px;
  height: 36px;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: border-color 0.2s, background 0.2s;
}
.icon-btn:hover {
  border-color: #999;
  background: #fafafa;
}
.editor-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}
.left-panel {
  width: 260px;
  border-right: 1px solid #eee;
  overflow-y: auto;
}
.center-panel {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}
.right-panel {
  width: 300px;
  border-left: 1px solid #eee;
  display: flex;
  flex-direction: column;
}
</style>
