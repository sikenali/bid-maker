<template>
  <div class="outline-tree">
    <div class="tree-header">
      <span>Outline</span>
      <button title="Add section">+</button>
    </div>
    <div class="tree-list">
      <div
        v-for="section in outline"
        :key="section.id"
        class="tree-item"
        :class="{ active: section.id === activeSectionId }"
        @click="$emit('select', section.id)"
      >
        <span class="icon">&#128196;</span>
        <span class="title">{{ section.title }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useDocumentStore } from '../stores/documentStore'

const docStore = useDocumentStore()
const outline = computed(() => docStore.outline)
const activeSectionId = computed(() => docStore.activeSectionId)
defineEmits<{ select: [id: string] }>()
</script>

<style scoped>
.outline-tree {
  padding: 16px;
}
.tree-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: bold;
}
.tree-item {
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
}
.tree-item.active {
  background: #e6f0ff;
}
</style>
