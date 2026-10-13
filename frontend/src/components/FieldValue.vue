<template>
  <div class="field-value">
    <span class="field-value-text">{{ displayValue || '—' }}</span>
    <ProvenanceBadge
      :page="page"
      :chapter="chapter"
      :source-text="sourceText"
    />
    <UiTip text="在预览中定位">
      <button
        v-if="page != null"
        type="button"
        class="locate-btn"
        @click="handleLocate"
      >
        <span class="icon ri-crosshair-2-line"></span>
        <span>定位</span>
      </button>
    </UiTip>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import ProvenanceBadge from './ProvenanceBadge.vue';
import { UiTip } from '@/components/ui';

const props = defineProps<{
  field: string;
  value: unknown;
  page?: number | null;
  chapter?: string;
  sourceText?: string;
}>();

const emit = defineEmits<{
  locate: [payload: { field: string; page: number | null }];
}>();

const displayValue = computed(() => {
  const v = props.value;
  if (v == null) return '';
  if (typeof v === 'string' || typeof v === 'number') return String(v);
  return JSON.stringify(v);
});

function handleLocate() {
  emit('locate', { field: props.field, page: props.page ?? null });
}
</script>

<style scoped>
.field-value {
  display: inline-flex;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 4px;
  max-width: 100%;
}
.field-value-text {
  word-break: break-all;
  white-space: pre-wrap;
}
.locate-btn {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 0 8px;
  height: 22px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-bg-card);
  font-size: 12px;
  color: var(--color-primary);
  cursor: pointer;
  vertical-align: middle;
  margin-left: 8px;
}
.locate-btn:hover {
  border-color: var(--color-primary);
  background: var(--color-bg-secondary);
}
.locate-btn .icon {
  font-family: "remixicon", sans-serif;
  font-style: normal;
  font-size: 12px;
}
</style>
