import { createRouter, createWebHistory } from 'vue-router'
import UploadView from '../views/UploadView.vue'
import EditorView from '../views/EditorView.vue'
import SettingsView from '../views/SettingsView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: UploadView },
    { path: '/editor/:id', component: EditorView, props: true },
    { path: '/settings', component: SettingsView },
  ],
})

export default router
