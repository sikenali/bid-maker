<template>
  <div class="progress-tracker">
    <div class="progress-header">
      <span class="progress-title">生成进度</span>
      <span class="progress-count">{{ genStore.completedSections }}/{{ genStore.totalSections }}</span>
    </div>
    <div class="progress-bar">
      <div class="progress-fill" :style="{ width: genStore.progressPercent + '%' }" />
    </div>
    <div class="section-states">
      <div v-for="sec in flatSections" :key="sec.id" class="state-row" :class="genStore.getSectionState(sec.id)">
        <span class="state-icon">{{ iconFor(genStore.getSectionState(sec.id)) }}</span>
        <span class="state-title" :style="{ paddingLeft: (sec._level || 0) * 16 + 'px' }">{{ sec.title }}</span>
        <button v-if="genStore.getSectionState(sec.id) === 'error'" class="retry-btn" @click="genStore.retrySection(sec.id)">重试</button>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useGenerateStore } from '../stores/generateStore'
import type { Section } from '../api/client'
import type { SectionState } from '../stores/generateStore'

type FlatSection = Section & { _level?: number }

const genStore = useGenerateStore()

const flatSections = computed(() => {
  const flat: FlatSection[] = []
  const flatten = (secs: Section[], level: number) => {
    for (const s of secs) { flat.push({ ...s, _level: level }); flatten(s.children, level + 1) }
  }
  flatten(genStore.outline, 0)
  return flat
})

function iconFor(state: SectionState) {
  switch (state) {
    case 'done': return '✅'
    case 'generating': return '⏳'
    case 'error': return '❌'
    default: return '⬜'
  }
}
</script>
<style scoped>
.progress-tracker { padding: 8px 16px; background: #FBF7EF; border-top: 1px solid #E0D5C0; }
.progress-header { display: flex; justify-content: space-between; margin-bottom: 8px; }
.progress-title { font-size: 13px; font-weight: 600; color: #3D2B1F; }
.progress-count { font-size: 13px; color: #8B7355; }
.progress-bar { height: 6px; background: #E0D5C0; border-radius: 3px; margin-bottom: 12px; overflow: hidden; }
.progress-fill { height: 100%; background: #C23B22; border-radius: 3px; transition: width 0.3s; }
.section-states { max-height: 200px; overflow-y: auto; }
.state-row { display: flex; align-items: center; gap: 8px; padding: 6px 4px; font-size: 13px; }
.state-icon { width: 18px; text-align: center; }
.state-title { flex: 1; color: #3D2B1F; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.retry-btn { padding: 2px 8px; font-size: 11px; border: 1px solid #C23B22; border-radius: 4px; background: #fff; color: #C23B22; cursor: pointer; }
.state-row.generating .state-title { color: #C23B22; }
.state-row.done .state-title { color: #8B7355; }
</style>