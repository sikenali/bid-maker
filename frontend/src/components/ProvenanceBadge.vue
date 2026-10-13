<template>
  <el-tag v-if="hasProvenance" size="small" type="info" class="prov-badge" @click="expanded = !expanded">
    <span v-if="page != null">第{{ page }}页 · </span>{{ chapter }}
    <el-icon><ArrowDown /></el-icon>
  </el-tag>
  <div v-if="expanded && sourceText" class="prov-source">{{ sourceText }}</div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { ArrowDown } from '@element-plus/icons-vue';

const props = defineProps<{
  page?: number | null;
  chapter?: string;
  sourceText?: string;
}>();
const expanded = ref(false);
const hasProvenance = computed(() => props.page != null || !!props.chapter);
</script>

<style scoped>
.prov-badge { cursor: pointer; margin-left: 8px; }
.prov-source { margin-top: 4px; color: var(--color-text-faint, #A7A296); font-size: 12px; }
</style>
