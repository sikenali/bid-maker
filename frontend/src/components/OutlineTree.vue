<template>
  <div class="outline-tree">
    <div class="outline-header">
      <div class="header-left">
        <RiListCheck size="18" color="#C23B22" />
        <span class="header-title">标书大纲</span>
      </div>
      <button class="add-btn" title="添加章节" @click="addTopSection">
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
          <RiFileTextFill v-if="section.id === activeSectionId" size="16" color="#C23B22" />
          <RiFileTextLine v-else size="16" color="#8B7355" />
          <span class="item-title">{{ section.title }}</span>
          <span class="item-actions">
            <RiAddBoxLine size="14" color="#8B7355" class="action-icon" @click.stop="showAddMenu(section.id)" />
            <RiDeleteBinLine size="14" color="#C43A31" class="action-icon" @click.stop="removeSection(section.id)" />
          </span>
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
          <span class="item-actions">
            <RiAddBoxLine size="14" color="#8B7355" class="action-icon" @click.stop="showAddMenu(child.id)" />
            <RiDeleteBinLine size="14" color="#C43A31" class="action-icon" @click.stop="removeSection(child.id)" />
          </span>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import type { Section } from '../stores/documentStore'
import {
  RiListCheck,
  RiAddLine,
  RiAddBoxLine,
  RiDeleteBinLine,
  RiFileTextFill,
  RiFileTextLine,
  RiArrowRightSLine,
} from '@remixicon/vue'

const route = useRoute()
const docStore = useDocumentStore()
const docId = route.params.id as string
const outline = computed(() => docStore.outline)
const activeSectionId = computed(() => docStore.activeSectionId)

const emit = defineEmits<{ select: [id: string] }>()

const saveOutline = () => {
  docStore.updateOutlineTree(docId, docStore.outline)
}

const addTopSection = () => {
  const newSection: Section = {
    id: Date.now().toString(),
    title: '新章节',
    level: 1,
    parent_id: '',
    content: '',
    children: [],
  }
  docStore.outline.push(newSection)
  emit('select', newSection.id)
  saveOutline()
}

const addChild = (parentId: string) => {
  const parent = findSection(docStore.outline, parentId)
  if (!parent) return
  const newSection: Section = {
    id: Date.now().toString(),
    title: '新子章节',
    level: parent.level + 1,
    parent_id: parentId,
    content: '',
    children: [],
  }
  parent.children.push(newSection)
  emit('select', newSection.id)
  saveOutline()
}

const addSibling = (sectionId: string) => {
  const parent = findParent(docStore.outline, sectionId)
  const newSection: Section = {
    id: Date.now().toString(),
    title: '新章节',
    level: parent ? parent.level : 1,
    parent_id: parent ? parent.id : '',
    content: '',
    children: [],
  }
  if (parent) {
    parent.children.push(newSection)
  } else {
    docStore.outline.push(newSection)
  }
  emit('select', newSection.id)
  saveOutline()
}

const removeSection = (sectionId: string) => {
  const removed = removeFromTree(docStore.outline, sectionId)
  if (removed) saveOutline()
}

const showAddMenu = (sectionId: string) => {
  const action = window.confirm('添加子大纲点"确定"，添加同级大纲点"取消"')
  if (action) {
    addChild(sectionId)
  } else {
    addSibling(sectionId)
  }
}

function findSection(sections: Section[], id: string): Section | null {
  for (const s of sections) {
    if (s.id === id) return s
    if (s.children.length > 0) {
      const found = findSection(s.children, id)
      if (found) return found
    }
  }
  return null
}

function findParent(sections: Section[], id: string): Section | null {
  for (const s of sections) {
    if (s.children.some(c => c.id === id)) return s
    if (s.children.length > 0) {
      const found = findParent(s.children, id)
      if (found) return found
    }
  }
  return null
}

function removeFromTree(sections: Section[], id: string): boolean {
  for (let i = 0; i < sections.length; i++) {
    if (sections[i].id === id) {
      sections.splice(i, 1)
      return true
    }
    if (sections[i].children.length > 0) {
      if (removeFromTree(sections[i].children, id)) return true
    }
  }
  return false
}
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
  background: #F0E8D5;
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
  background: rgba(194, 59, 34, 0.04);
}

.outline-item.active {
  background: rgba(194, 59, 34, 0.08);
}

.item-title {
  font-size: 13px;
  color: #3D2B1F;
  font-weight: 500;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-actions {
  display: flex;
  align-items: center;
  gap: 2px;
  opacity: 0;
  transition: opacity 0.15s;
}

.outline-item:hover .item-actions,
.outline-subitem:hover .item-actions {
  opacity: 1;
}

.action-icon {
  cursor: pointer;
  padding: 2px;
  border-radius: 4px;
  transition: background 0.15s;
  flex-shrink: 0;
}

.action-icon:hover {
  background: rgba(0, 0, 0, 0.06);
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
  background: rgba(194, 59, 34, 0.04);
}

.outline-subitem.active {
  background: rgba(194, 59, 34, 0.08);
}

.subitem-title {
  font-size: 12px;
  color: #3D2B1F;
}
</style>