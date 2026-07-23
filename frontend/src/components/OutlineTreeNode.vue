<template>
  <div>
    <div
      :data-id="section.id"
      class="outline-node"
      :class="{ active: section.id === activeSectionId }"
      :style="{ paddingLeft: `${16 + depth * 30}px` }"
      @click="emit('select', section.id)"
    >
      <RiFileTextFill v-if="section.id === activeSectionId" size="20" color="#C23B22" />
      <RiFileTextLine v-else size="20" color="#8B7355" />
      <span class="node-title">{{ section.title }}</span>
      <span class="menu-wrapper">
        <button
          class="more-btn"
          :class="{ active: openMenuId === section.id }"
          @click.stop="emit('toggle-menu', section.id)"
        >
          <RiMore2Fill size="16" color="#8B7355" />
        </button>
        <div v-if="openMenuId === section.id" class="context-menu" @click.stop>
          <div class="menu-arrow" />
          <button class="menu-item" :class="{ disabled: section.level <= 1 }" @click.stop="emit('promote-level', section.id)">升级</button>
          <button class="menu-item" @click.stop="emit('demote-level', section.id)">降级</button>
          <div class="menu-divider" />
          <button class="menu-item" @click.stop="emit('add-child', section.id)">新增子章节</button>
          <div class="menu-divider" />
          <button class="menu-item menu-item-danger" @click.stop="emit('remove-section', section.id)">删除</button>
        </div>
      </span>
    </div>
    <OutlineTreeNode
      v-for="child in section.children"
      :key="child.id"
      :section="child"
      :depth="depth + 1"
      :active-section-id="activeSectionId"
      :open-menu-id="openMenuId"
      @select="emit('select', $event)"
      @toggle-menu="emit('toggle-menu', $event)"
      @demote-level="emit('demote-level', $event)"
      @add-child="emit('add-child', $event)"
      @remove-section="emit('remove-section', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import type { Section } from '../stores/documentStore'
import {
  RiMore2Fill,
  RiFileTextFill,
  RiFileTextLine,
} from '@remixicon/vue'
defineOptions({ name: 'OutlineTreeNode' })

defineProps<{
  section: Section
  depth: number
  activeSectionId: string
  openMenuId: string | null
}>()

const emit = defineEmits<{
  select: [id: string]
  'toggle-menu': [id: string]
  'promote-level': [id: string]
  'demote-level': [id: string]
  'add-child': [id: string]
  'remove-section': [id: string]
}>()
</script>

<style scoped>
.outline-node {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  cursor: pointer;
  transition: background 0.15s;
  position: relative;
  z-index: 1;
}

.outline-node:hover {
  background: rgba(194, 59, 34, 0.06);
}

.outline-node.active {
  background: transparent;
}

.node-title {
  font-size: 14px;
  color: #3D2B1F;
  font-weight: 500;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.more-btn {
  width: 24px;
  height: 24px;
  border: 0.7px solid #E0D5C0;
  border-radius: 50%;
  background: #FBF7EF;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.15s, background 0.15s, border-color 0.15s;
  flex-shrink: 0;
  position: relative;
  z-index: 2;
}

.outline-node:hover .more-btn {
  opacity: 1;
}

.more-btn:hover {
  background: #F5EFE0;
  border-color: #D4C4A8;
}

.more-btn:focus-visible {
  opacity: 1;
}

.more-btn.active {
  opacity: 1;
  background: #F5EFE0;
  border-color: #C23B22;
}

.context-menu {
  position: absolute;
  right: 0;
  top: calc(100% + 6px);
  background: #fff;
  border: 0.7px solid #E0D5C0;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.12);
  z-index: 200;
  min-width: 100px;
  padding: 4px;
  display: flex;
  flex-direction: column;
}

.menu-wrapper {
  position: relative;
  display: inline-flex;
}

.menu-arrow {
  position: absolute;
  top: -5px;
  right: 14px;
  left: auto;
  width: 8px;
  height: 8px;
  background: #fff;
  border-left: 0.7px solid #E0D5C0;
  border-top: 0.7px solid #E0D5C0;
  transform: rotate(45deg);
  z-index: -1;
}

.menu-item {
  padding: 8px 12px;
  border: none;
  border-radius: 6px;
  background: transparent;
  font-size: 12px;
  color: #3D2B1F;
  cursor: pointer;
  text-align: left;
  white-space: nowrap;
  transition: background 0.1s;
}

.menu-item:hover {
  background: rgba(194, 59, 34, 0.06);
}

.menu-item-danger {
  color: #C43A31;
}

.menu-item-danger:hover {
  background: rgba(196, 58, 49, 0.08);
}

.menu-item.disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.menu-item.disabled:hover {
  background: transparent;
}

.menu-divider {
  height: 1px;
  background: #E0D5C0;
  margin: 4px 0;
}
</style>
