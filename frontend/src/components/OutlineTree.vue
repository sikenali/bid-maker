<template>
  <div class="outline-tree">
    <div class="outline-header">
      <div class="header-left">
        <RiListCheck size="18" color="#C43D3D" />
        <span class="header-title">标书大纲</span>
      </div>
      <button class="add-btn" title="添加章节">
        <RiAddLine size="16" color="#8B7355" />
      </button>
    </div>
    <div class="outline-list">
      <template v-for="section in outline" :key="section.id">
        <div
          class="outline-item"
          :class="{ active: section.id === activeSectionId }"
          @click="$emit('select', section.id)"
        >
          <RiFileTextFill v-if="section.id === activeSectionId" size="16" color="#C43D3D" />
          <RiFileTextLine v-else size="16" color="#8B7355" />
          <span class="item-title">{{ section.title }}</span>
        </div>
        <div
          v-for="child in section.children"
          :key="child.id"
          class="outline-subitem"
          :class="{ active: child.id === activeSectionId }"
          @click="$emit('select', child.id)"
        >
          <RiArrowRightSLine size="14" color="#9B8C7C" />
          <span class="subitem-title">{{ child.title }}</span>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useDocumentStore } from '../stores/documentStore'
import {
  RiListCheck,
  RiAddLine,
  RiFileTextFill,
  RiFileTextLine,
  RiArrowRightSLine,
} from '@remixicon/vue'

const docStore = useDocumentStore()
const outline = computed(() => docStore.outline)
const activeSectionId = computed(() => docStore.activeSectionId)
defineEmits<{ select: [id: string] }>()
</script>

<style scoped>
.outline-tree {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.outline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-title {
  font-size: 14px;
  font-weight: 600;
  color: #3D2B1F;
}

.add-btn {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: none;
  background: #E8DCC8;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.add-btn:hover {
  background: #DCC8B0;
}

.outline-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.outline-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  cursor: pointer;
  transition: background 0.15s;
}

.outline-item:hover {
  background: rgba(196, 61, 61, 0.04);
}

.outline-item.active {
  background: rgba(196, 61, 61, 0.08);
}

.item-title {
  font-size: 13px;
  color: #3D2B1F;
  font-weight: 500;
}

.outline-subitem {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 16px 6px 40px;
  cursor: pointer;
  transition: background 0.15s;
}

.outline-subitem:hover {
  background: rgba(196, 61, 61, 0.04);
}

.outline-subitem.active {
  background: rgba(196, 61, 61, 0.08);
}

.subitem-title {
  font-size: 12px;
  color: #3D2B1F;
}
</style>