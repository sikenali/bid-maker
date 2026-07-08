<template>
  <div id="app" :class="'theme-' + settingsStore.theme">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { useSettingsStore } from './stores/settingsStore'

const settingsStore = useSettingsStore()
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
}

.theme-light body,
.theme-light {
  background-color: #FDF6E3;
}

.theme-dark body,
.theme-dark {
  background-color: #2C2416;
  color: #E8DCC8;
}

.theme-paper body,
.theme-paper {
  background-color: #FFFFFF;
}

#app {
  min-height: 100vh;
}
</style>
