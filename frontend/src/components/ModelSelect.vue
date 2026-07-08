<template>
  <div class="model-picker" ref="containerRef">
    <button class="picker-trigger" @click="toggleOpen">
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
  gap: 4px;
  padding: 4px 8px;
  background: #F0E8D5;
  border: 0.7px solid #E0D5C0;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s, border-color 0.2s;
  font-size: 11px;
  color: #3D2B1F;
  font-weight: 500;
  white-space: nowrap;
}

.picker-trigger:hover {
  background: #E8DCC8;
  border-color: #D4C4A8;
}

.picker-label {
  min-width: 36px;
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
  min-width: 140px;
  background: #fff;
  border: 0.7px solid #E0D5C0;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.1);
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

.picker-option:hover {
  background: rgba(194, 59, 34, 0.06);
}

.picker-option.selected {
  font-weight: 600;
}

.option-check {
  flex-shrink: 0;
}
</style>