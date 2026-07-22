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
    <div class="outline-list" ref="listRef" :class="{ 'menu-open': openMenuId !== null }">
      <div class="outline-indicator" :style="indicatorStyle" />
      <OutlineTreeNode
        v-for="section in outline"
        :key="section.id"
        :section="section"
        :depth="0"
        :active-section-id="activeSectionId"
        :open-menu-id="openMenuId"
        @select="selectSection"
        @toggle-menu="toggleMenu"
        @demote-level="demoteLevel"
        @add-child="addChild"
        @remove-section="removeSection"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import type { Section } from '../stores/documentStore'
import OutlineTreeNode from './OutlineTreeNode.vue'
import {
  RiListCheck,
  RiAddLine,
} from '@remixicon/vue'

const route = useRoute()
const docStore = useDocumentStore()
const docId = route.params.id as string
const outline = computed(() => docStore.outline)
const activeSectionId = computed(() => docStore.activeSectionId)
const openMenuId = ref<string | null>(null)
const listRef = ref<HTMLElement>()
const indicatorTop = ref(0)
const indicatorHeight = ref(0)

const emit = defineEmits<{ select: [id: string] }>()

const closeMenu = () => { openMenuId.value = null }

const toggleMenu = (id: string) => {
  if (openMenuId.value === id) {
    openMenuId.value = null
    return
  }
  openMenuId.value = id
  nextTick(() => {
    const list = listRef.value
    if (!list) return
    const item = list.querySelector(`[data-id="${id}"]`)
    if (!item) return
    const firstBtn = item.querySelector('.menu-item') as HTMLElement | null
    if (firstBtn) firstBtn.focus()
  })
}

const handleMenuKeydown = (e: KeyboardEvent) => {
  if (!openMenuId.value) return
  if (e.key === 'Escape') {
    closeMenu()
    return
  }
  if (e.key === 'Tab') {
    e.preventDefault()
    const list = listRef.value
    if (!list) return
    const item = list.querySelector(`[data-id="${openMenuId.value}"]`)
    if (!item) return
    const buttons = item.querySelectorAll('.menu-item') as NodeListOf<HTMLElement>
    if (buttons.length === 0) return
    const active = document.activeElement
    let idx = Array.from(buttons).findIndex(b => b === active)
    if (e.shiftKey) {
      idx = idx <= 0 ? buttons.length - 1 : idx - 1
    } else {
      idx = idx >= buttons.length - 1 ? 0 : idx + 1
    }
    buttons[idx].focus()
  }
}

const selectSection = (id: string) => {
  emit('select', id)
  nextTick(() => updateIndicator(id))
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
  selectSection(newSection.id)
  closeMenu()
  saveOutline()
}

const demoteLevel = (sectionId: string) => {
  const section = findSection(docStore.outline, sectionId)
  if (!section || section.level >= 9) return
  section.level++
  saveOutline()
  closeMenu()
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
  selectSection(newSection.id)
  closeMenu()
  saveOutline()
}

const removeSection = (sectionId: string) => {
  const removed = removeFromTree(docStore.outline, sectionId)
  if (removed) {
    closeMenu()
    saveOutline()
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

const handleClickOutside = (e: MouseEvent) => {
  if (openMenuId.value) {
    const list = listRef.value
    if (list && list.contains(e.target as Node)) return
    closeMenu()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleMenuKeydown)
  if (activeSectionId.value) nextTick(() => updateIndicator(activeSectionId.value))
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleMenuKeydown)
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

.add-btn:hover {
  background: #DCC8B0;
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

.outline-list.menu-open .more-btn:not(.active) {
  opacity: 0 !important;
}
</style>
