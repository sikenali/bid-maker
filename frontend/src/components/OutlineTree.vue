<template>
  <div class="outline-tree">
    <div class="outline-header">
      <div class="header-left">
        <RiListCheck size="18" color="#C23B22" />
        <span class="header-title">标书大纲</span>
      </div>
      <button class="add-btn" title="添加章节" @click="addTopSection" :disabled="isLocked">
        <RiAddLine size="16" color="#8B7355" />
      </button>
    </div>
    <div class="outline-list" ref="listRef" :class="{ 'menu-open': openMenuId !== null }">
      <div class="outline-indicator" :style="indicatorStyle" />
      <OutlineTreeNode
        v-for="section in outline"
        :key="section.section_id"
        :section="toLegacySection(section)"
        :depth="0"
        :active-section-id="activeSectionId"
        :open-menu-id="openMenuId"
        :section-state="(getSectionState(section.section_id) as any)"
        @select="handleSelectSection"
        @toggle-menu="toggleMenu"
        @promote-level="promoteLevel"
        @demote-level="demoteLevel"
        @add-child="addChild"
        @remove-section="removeSection"
        @generate="handleGenerate"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import type { OutlineSection } from '../stores/documentStore'
import OutlineTreeNode from './OutlineTreeNode.vue'
import {
  RiListCheck,
  RiAddLine,
} from '@remixicon/vue'

const route = useRoute()
const docStore = useDocumentStore()
const _sessionId = route.params.id as string // eslint-disable-line
const outline = computed(() => docStore.outline)
const activeSectionId = computed(() => docStore.activeSectionId)
const isLocked = computed(() => docStore.session?.outline?.locked || false)

const openMenuId = ref<string | null>(null)
const listRef = ref<HTMLElement>()
const indicatorTop = ref(0)
const indicatorHeight = ref(0)

const emit = defineEmits<{ select: [id: string]; generate: [id: string] }>()

// 转换为旧版 Section 格式供 OutlineTreeNode 使用
function toLegacySection(section: OutlineSection): any {
  return {
    id: section.section_id,
    title: section.title,
    level: section.level,
    parent_id: section.parent_id || '',
    content: section.content || '',
    children: (section.children || []).map(childId => {
      const child = findSectionById(outline.value, childId)
      return child ? toLegacySection(child) : null
    }).filter(Boolean),
    required: section.required,
    source: section.source,
    status: section.status,
  }
}

function getSectionState(sectionId: string): string {
  const section = findSectionById(outline.value, sectionId)
  return section?.status || 'pending'
}

const closeMenu = () => { openMenuId.value = null }

const toggleMenu = (id: string) => {
  if (openMenuId.value === id) {
    openMenuId.value = null
    return
  }
  openMenuId.value = id
}

const handleSelectSection = (id: string) => {
  emit('select', id)
  docStore.setActiveSection(id)
  nextTick(() => updateIndicator(id))
  closeMenu()
}

const handleGenerate = (id: string) => {
  emit('generate', id)
}

const updateIndicator = (id: string) => {
  const list = listRef.value
  if (!list) return
  const target = list.querySelector(`[data-id="${id}"]`) as HTMLElement | null
  if (!target) return
  const listRect = list.getBoundingClientRect()
  const itemRect = target.getBoundingClientRect()
  indicatorTop.value = itemRect.top - listRect.top + list.scrollTop
  indicatorHeight.value = itemRect.height
}

const indicatorStyle = computed(() => ({
  top: `${indicatorTop.value}px`,
  height: `${indicatorHeight.value}px`,
}))

const addTopSection = () => {
  const newSection: OutlineSection = {
    section_id: 'sec-' + Date.now(),
    title: '新章节',
    level: 1,
    order: outline.value.length + 1,
    required: false,
    parent_id: null,
    source: 'manual',
    requirements: [],
    response_strategy: 'satisfy',
    response_basis: '',
    children: [],
    status: 'pending',
  }
  outline.value.push(newSection)
  handleSelectSection(newSection.section_id)
  closeMenu()
}

const addChild = (parentId: string) => {
  const parent = findSectionById(outline.value, parentId)
  if (!parent) return
  const newSection: OutlineSection = {
    section_id: 'sec-' + Date.now(),
    title: '新子章节',
    level: parent.level + 1,
    order: (parent.children?.length || 0) + 1,
    required: false,
    parent_id: parentId,
    source: 'manual',
    requirements: [],
    response_strategy: 'satisfy',
    response_basis: '',
    children: [],
    status: 'pending',
  }
  if (!parent.children) parent.children = []
  parent.children.push(newSection.section_id)
  outline.value.push(newSection)
  handleSelectSection(newSection.section_id)
  closeMenu()
}

const removeSection = (sectionId: string) => {
  const idx = outline.value.findIndex(s => s.section_id === sectionId)
  if (idx >= 0) {
    outline.value.splice(idx, 1)
    closeMenu()
  }
}

const promoteLevel = (sectionId: string) => {
  const section = findSectionById(outline.value, sectionId)
  if (!section || section.level <= 1) return
  section.level--
  closeMenu()
}

const demoteLevel = (sectionId: string) => {
  const section = findSectionById(outline.value, sectionId)
  if (!section || section.level >= 3) return
  section.level++
  closeMenu()
}

function findSectionById(sections: OutlineSection[], id: string): OutlineSection | null {
  for (const s of sections) {
    if (s.section_id === id) return s
    if (s.children) {
      for (const childId of s.children) {
        const child = findSectionById(sections, childId)
        if (child) return child
      }
    }
  }
  return null
}

const handleClickOutside = (e: MouseEvent) => {
  if (openMenuId.value) {
    const list = listRef.value
    if (list && list.contains(e.target as Node)) return
    closeMenu()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  if (activeSectionId.value) nextTick(() => updateIndicator(activeSectionId.value))
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
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

.add-btn:hover:not(:disabled) {
  background: #DCC8B0;
}

.add-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.outline-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
  position: relative;
}

.outline-indicator {
  position: absolute;
  left: 12px;
  right: 12px;
  background: rgba(194, 59, 34, 0.15);
  border-radius: 10px;
  transition: top 0.35s ease-out, height 0.35s ease-out;
  pointer-events: none;
  z-index: 0;
}
</style>
