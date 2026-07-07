import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUiStore = defineStore('ui', () => {
  const loading = ref(false)
  const activePanel = ref('')

  const setActivePanel = (panel: string) => {
    activePanel.value = panel
  }

  return { loading, activePanel, setActivePanel }
})
