<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup lang="ts">
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Noto Sans SC', -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Source Han Sans CN', 'Microsoft YaHei', sans-serif;
  background-color: #FDF6E3;
  color: #3D2B1F;
}

#app {
  min-height: 100vh;
}
</style>
