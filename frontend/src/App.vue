<script setup lang="ts">
import { onMounted } from 'vue';
import { useThemeStore } from '@/stores/theme';

const themeStore = useThemeStore();

onMounted(() => {
  themeStore.loadTheme();
});
</script>

<template>
  <router-view />
</template>
