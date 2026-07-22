<template>
  <div class="model-picker" ref="containerRef">
    <button class="picker-trigger" :class="{ 'picker-trigger-disabled': items.length === 0 }" @click="toggleOpen" :disabled="items.length === 0">
      <span class="picker-label">{{ selectedLabel }}</span>
      <RiArrowDownSLine size="14" color="#8B7355" class="picker-arrow" :class="{ open: isOpen }" />
    </button>
    <div v-if="isOpen" class="picker-dropdown" @click.stop>
      <button
        v-for="item in items"
        :key="item.id"
        class="picker-option"
        :class="{ selected: item.id === modelValue }"
        @click="select(item.id)"
      >
        <span class="option-name">{{ item.name }}</span>
        <RiCheckLine v-if="item.id === modelValue" size="14" color="#C23B22" class="option-check" />
      </button>
      <div v-if="items.length === 0" class="picker-empty">请先添加 API Key</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { RiArrowDownSLine, RiCheckLine } from '@remixicon/vue'

export interface ModelItem {
  id: string
  name: string
}

const props = defineProps<{
  modelValue: string
  items: ModelItem[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const isOpen = ref(false)
const containerRef = ref<HTMLElement>()

const selectedLabel = computed(() => {
  if (props.items.length === 0) return '未配置模型'
  const item = props.items.find(i => i.id === props.modelValue)
  return item ? item.name : '选择模型'
})

const toggleOpen = () => { isOpen.value = !isOpen.value }

const select = (id: string) => {
  emit('update:modelValue', id)
  isOpen.value = false
}

const handleClickOutside = (e: MouseEvent) => {
  if (containerRef.value && !containerRef.value.contains(e.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.model-picker {
  position: relative;
}

.picker-trigger {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px 12px;
  background: #FAFAF5;
  border: 0.7px solid #E0D5C0;
  border-radius: 8px;
  cursor: pointer;
  transition: border-color 0.2s;
  font-size: 13px;
  color: #3D2B1F;
  font-weight: 500;
  white-space: nowrap;
  min-width: 160px;
}

.picker-trigger:hover {
  border-color: #D4C4A8;
}

.picker-trigger:focus-visible {
  border-color: #C23B22;
  box-shadow: 0 0 0 1px rgba(194, 59, 34, 0.15);
  outline: none;
}

.picker-trigger-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.picker-trigger-disabled:hover {
  border-color: #E0D5C0;
}

.picker-label {
  flex: 1;
  text-align: left;
}

.picker-arrow {
  transition: transform 0.2s;
  flex-shrink: 0;
}

.picker-arrow.open {
  transform: rotate(180deg);
}

.picker-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  min-width: 200px;
  background: #fff;
  border: 0.7px solid #E0D5C0;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.12);
  z-index: 200;
  padding: 4px;
  display: flex;
  flex-direction: column;
}

.picker-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px 12px;
  border: none;
  border-radius: 6px;
  background: transparent;
  font-size: 13px;
  color: #3D2B1F;
  cursor: pointer;
  text-align: left;
  white-space: nowrap;
  transition: background 0.1s;
}

.picker-option:hover {
  background: rgba(194, 59, 34, 0.06);
}

.picker-option.selected {
  font-weight: 600;
}

.option-check {
  flex-shrink: 0;
}

.picker-empty {
  padding: 12px 16px;
  font-size: 12px;
  color: #8B7355;
  text-align: center;
}
</style>