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
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Source Han Sans CN', 'Microsoft YaHei', sans-serif;
  background-color: #fbf7f0;
  color: #3d2b1f;
}

#app {
  min-height: 100vh;
}
</style>
