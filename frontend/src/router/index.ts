import { createRouter, createWebHistory } from 'vue-router';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: () => import('@/views/UploadView.vue') },
    { path: '/editor/:id', component: () => import('@/views/EditorView.vue'), props: true },
    { path: '/settings', component: () => import('@/views/SettingsView.vue') },
    { path: '/settings/theme', component: () => import('@/views/settings/ThemeSettingsView.vue') },
    { path: '/settings/template', component: () => import('@/views/settings/TemplateSettingsView.vue') },
    { path: '/settings/skill', component: () => import('@/views/settings/SkillSettingsView.vue') },
    { path: '/settings/api-key', component: () => import('@/views/settings/ApiKeySettingsView.vue') },
    { path: '/settings/export', component: () => import('@/views/settings/ExportSettingsView.vue') },
    { path: '/settings/about', component: () => import('@/views/settings/AboutSettingsView.vue') },
  ],
});

export default router;
