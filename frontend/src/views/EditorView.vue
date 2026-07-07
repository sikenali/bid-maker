<template>
  <div class="editor-view">
    <h1>Document Editor</h1>
    <p>Document ID: {{ docId }}</p>
    <div class="editor-container">
      <p>Editor view placeholder — section editing and chat will go here.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const docId = computed(() => route.params.id as string)
</script>

<style scoped>
.editor-view {
  max-width: 900px;
  margin: 40px auto;
  padding: 20px;
}

.editor-container {
  margin-top: 20px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  min-height: 400px;
}
</style>
