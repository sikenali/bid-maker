<template>
  <div class="skill-chip" :class="{ 'skill-chip-active': !disabled }">
    <span class="chip-slash">/</span><span class="chip-name">{{ skillName }}</span>
    <button class="chip-remove" @click="emit('remove')" v-if="!disabled">
      <RiCloseLine size="12" color="#fff" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { RiCloseLine } from '@remixicon/vue'
defineProps<{
  skillName: string
  disabled?: boolean
}>()
const emit = defineEmits<{ remove: [] }>()
</script>

<style scoped>
.skill-chip {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  background: #C23B22;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  border-radius: 4px;
  padding: 2px 4px;
  margin-right: 6px;
  white-space: nowrap;
  line-height: 1.5;
  flex-shrink: 0;
  user-select: none;
}

.skill-chip-active:hover {
  cursor: pointer;
}

.chip-slash {
  opacity: 0.85;
}

.chip-remove {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  padding: 0;
  margin-left: 2px;
  cursor: pointer;
  border-radius: 0;
  transition: opacity 0.1s;
}

.chip-remove:hover {
  opacity: 0.7;
}
</style>
