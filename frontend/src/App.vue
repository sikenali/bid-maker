<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useSettingsStore } from './stores/settingsStore'

const settingsStore = useSettingsStore()

watch(() => settingsStore.theme, (t) => {
  document.body.className = 'theme-' + t
}, { immediate: true })
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Noto Sans SC', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Source Han Sans CN', 'Microsoft YaHei', sans-serif;
  color: #3D2B1F;
  transition: background-color 0.3s, color 0.3s;
}

body.theme-light {
  background-color: #FDF6E3;
}

body.theme-dark {
  background-color: #2C2416;
  color: #E8DCC8;
}

body.theme-paper {
  background-color: #FFFFFF;
}

#app {
  min-height: 100vh;
}
</style>