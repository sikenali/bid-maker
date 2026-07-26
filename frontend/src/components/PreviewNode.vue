<template>
  <div class="preview-node" :style="{ paddingLeft: depth * 20 + 'px' }">
    <div class="preview-title">
      <RiFileLine size="16" />
      <span>{{ section.title }}</span>
    </div>
    <PreviewNode v-for="child in section.children" :key="child.id"
      :section="child" :depth="depth + 1" />
  </div>
</template>
<script setup lang="ts">
import { RiFileLine } from '@remixicon/vue'
defineProps<{ section: any; depth: number }>()
</script>
<style scoped>
.preview-node { margin: 4px 0; }
.preview-title { display: flex; align-items: center; gap: 8px; padding: 6px 8px; border-radius: 6px; font-size: 14px; color: #3D2B1F; }
.preview-title:hover { background: #F5EFE0; }
</style>