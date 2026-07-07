import { defineStore } from 'pinia'

export const useDocStore = defineStore('doc', {
  state: () => ({
    docId: '',
    outline: [],
    sections: [] as any[],
    activeSection: null as string | null,
  }),
})
