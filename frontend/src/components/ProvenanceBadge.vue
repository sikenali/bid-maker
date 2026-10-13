<template>
  <span v-if="hasProvenance" class="prov-badge" @click="expanded = !expanded">
    <span v-if="page != null">第{{ page }}页 · </span>{{ chapter }}
    <span class="prov-icon" :class="{ expanded }">▾</span>
  </span>
  <div v-if="expanded && sourceText" class="prov-source">{{ sourceText }}</div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const props = defineProps<{
  page?: number | null;
  chapter?: string;
  sourceText?: string;
}>();
const expanded = ref(false);
const hasProvenance = computed(() => props.page != null || !!props.chapter);
</script>

<style scoped>
.prov-badge {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--color-bg-secondary, #F5EFE0);
  color: var(--color-text-muted, #8A8578);
  font-size: 12px;
  cursor: pointer;
  margin-left: 6px;
  vertical-align: middle;
  white-space: nowrap;
}
.prov-badge:hover {
  background: var(--color-border, #E5DFD3);
}
.prov-icon {
  font-size: 10px;
  transition: transform 0.2s;
}
.prov-icon.expanded {
  transform: rotate(180deg);
}
.prov-source {
  margin-top: 4px;
  padding: 8px 12px;
  border-radius: 6px;
  background: var(--color-bg-secondary, #F5EFE0);
  color: var(--color-text-secondary, #6E6A60);
  font-size: 12px;
  line-height: 1.5;
  max-width: 400px;
  word-break: break-word;
}
</style>
