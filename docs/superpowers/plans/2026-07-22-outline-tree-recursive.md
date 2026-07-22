# Outline Tree 递归层级展示实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 将 OutlineTree 组件从固定 2 层渲染改为递归渲染，支持任意深度的章节层级展示。

**架构：** 新增 `OutlineTreeNode.vue` 递归组件，OutlineTree.vue 作为容器保留头部和激活指示器，将节点渲染委托给 OutlineTreeNode。

**技术栈：** Vue 3, TypeScript, @remixicon/vue

---

### 任务 1：创建 OutlineTreeNode 递归组件

**文件：**
- 创建：`frontend/src/components/OutlineTreeNode.vue`
- 修改：`frontend/src/components/OutlineTree.vue`

- [ ] **步骤 1：创建 OutlineTreeNode.vue**

```vue
<template>
  <div>
    <div
      :data-id="section.id"
      class="outline-node"
      :class="{ active: section.id === activeSectionId }"
      :style="{ paddingLeft: `${16 + depth * 30}px` }"
      @click="selectSection(section.id)"
    >
      <RiFileTextFill v-if="section.id === activeSectionId" size="20" color="#C23B22" />
      <RiFileTextLine v-else size="20" color="#8B7355" />
      <span class="node-title">{{ section.title }}</span>
      <span class="menu-wrapper">
        <button
          class="more-btn"
          :class="{ active: openMenuId === section.id }"
          @click.stop="toggleMenu(section.id)"
        >
          <RiMore2Fill size="16" color="#8B7355" />
        </button>
        <div v-if="openMenuId === section.id" class="context-menu" @click.stop>
          <div class="menu-arrow" />
          <button class="menu-item" :class="{ disabled: section.level >= 9 }" @click.stop="demoteLevel(section.id)">降级</button>
          <button class="menu-item" @click.stop="addChild(section.id)">新增</button>
          <div class="menu-divider" />
          <button class="menu-item menu-item-danger" @click.stop="removeSection(section.id)">删除</button>
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
      @select="(id: string) => emit('select', id)"
      @toggle-menu="(id: string) => emit('toggle-menu', id)"
      @demote-level="(id: string) => emit('demote-level', id)"
      @add-child="(id: string) => emit('add-child', id)"
      @remove-section="(id: string) => emit('remove-section', id)"
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
import { nextTick } from 'vue'

defineOptions({ name: 'OutlineTreeNode' })

const props = defineProps<{
  section: Section
  depth: number
  activeSectionId: string
  openMenuId: string | null
}>()

const emit = defineEmits<{
  select: [id: string]
  'toggle-menu': [id: string]
  'demote-level': [id: string]
  'add-child': [id: string]
  'remove-section': [id: string]
}>()

const selectSection = (id: string) => {
  emit('select', id)
}

const toggleMenu = (id: string) => {
  emit('toggle-menu', id)
}

const demoteLevel = (id: string) => {
  emit('demote-level', id)
}

const addChild = (id: string) => {
  emit('add-child', id)
}

const removeSection = (id: string) => {
  emit('remove-section', id)
}
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
```

- [ ] **步骤 2：修改 OutlineTree.vue 使用递归组件**

```vue
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
```

- [ ] **步骤 3：验证编译**

运行：
```bash
cd frontend && npx vue-tsc --noEmit
```

预期：无类型错误

- [ ] **步骤 4：验证构建**

运行：
```bash
cd frontend && npm run build
```

预期：构建成功

- [ ] **步骤 5：Commit**

```bash
git add frontend/src/components/OutlineTree.vue frontend/src/components/OutlineTreeNode.vue docs/superpowers/specs/2026-07-22-outline-tree-recursive.md docs/superpowers/plans/2026-07-22-outline-tree-recursive.md
git commit -m "feat: 支持递归渲染任意层级的大纲树"
```
