<template>
  <div class="md-editor">
    <textarea
      class="md-editor__input"
      :value="modelValue"
      placeholder="在此输入或编辑 Markdown 内容……"
      @input="onInput"
    ></textarea>
  </div>
</template>

<script setup lang="ts">
defineProps<{ modelValue: string }>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
}>()

function onInput(event: Event) {
  const value = (event.target as HTMLTextAreaElement).value
  emit('update:modelValue', value)
  emit('change', value)
}
</script>

<style scoped>
.md-editor {
  height: 100%;
  display: flex;
  flex-direction: column;
}
.md-editor__input {
  flex: 1;
  width: 100%;
  border: none;
  outline: none;
  resize: none;
  padding: 24px;
  font-size: 14px;
  line-height: 1.8;
  color: #3D2B1F;
  background: #fff;
  font-family: inherit;
  box-sizing: border-box;
}
</style>
