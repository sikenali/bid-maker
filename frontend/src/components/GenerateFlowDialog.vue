<template>
  <div class="dialog-overlay" @click.self="$emit('cancel')">
    <div class="dialog">
      <div class="dialog-header">
        <RiFileList3Line size="20" />
        <span>标书生成预览</span>
      </div>
      <div class="dialog-body">
        <p class="dialog-query">"{{ genStore.userMessage }}"</p>
        <div class="outline-preview">
          <div v-for="sec in genStore.outline" :key="sec.id" class="preview-node">
            <PreviewNode :section="sec" :depth="0" />
          </div>
        </div>
      </div>
      <div class="dialog-footer">
        <button class="btn-cancel" @click="$emit('cancel')">取消</button>
        <button class="btn-confirm" @click="$emit('confirm')">确认生成</button>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { useGenerateStore } from '../stores/generateStore'
import { RiFileList3Line } from '@remixicon/vue'
import PreviewNode from './PreviewNode.vue'
defineEmits<{ confirm: []; cancel: [] }>()
const genStore = useGenerateStore()
</script>
<style scoped>
.dialog-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.dialog { background: #fff; border-radius: 16px; width: 640px; max-height: 80vh; display: flex; flex-direction: column; box-shadow: 0 8px 32px rgba(0,0,0,0.15); }
.dialog-header { display: flex; align-items: center; gap: 10px; padding: 20px 24px 0; font-size: 18px; font-weight: 600; color: #3D2B1F; }
.dialog-body { flex: 1; overflow-y: auto; padding: 16px 24px; }
.dialog-query { font-size: 14px; color: #8B7355; padding: 12px; background: #FBF7EF; border-radius: 8px; margin-bottom: 16px; }
.outline-preview { font-size: 14px; color: #3D2B1F; }
.dialog-footer { display: flex; justify-content: flex-end; gap: 12px; padding: 16px 24px; border-top: 1px solid #E0D5C0; }
.btn-cancel { padding: 8px 24px; border-radius: 8px; border: 1px solid #E0D5C0; background: #fff; color: #5C4033; cursor: pointer; font-size: 14px; }
.btn-confirm { padding: 8px 24px; border-radius: 8px; border: none; background: #C23B22; color: #fff; cursor: pointer; font-size: 14px; }
.btn-confirm:hover { background: #A8321D; }
</style>