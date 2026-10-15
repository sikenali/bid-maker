<template>
  <teleport to="body">
    <transition name="modal">
      <div v-if="visible" class="modal-overlay" @click.self="close">
        <div class="settings-panel">
          <div class="panel-header">
            <div class="header-left">
              <div class="header-icon">
                <RiSettings3Line />
              </div>
              <div class="header-text">
                <h2 class="panel-title">设置中心</h2>
                <p class="panel-subtitle">选择要进入的设置界面</p>
              </div>
            </div>
            <button class="close-btn" @click="close">
              <RiCloseLine />
            </button>
          </div>

          <div class="panel-body">
            <nav class="nav-list">
              <div
                v-for="item in menuItems"
                :key="item.key"
                class="nav-item"
                :class="{ active: activeKey === item.key }"
                @click="selectTab(item.key)"
              >
                <component :is="riToCompName(item.icon)" class="icon" />
                <span class="label">{{ item.label }}</span>
              </div>
            </nav>
            <div class="panel-divider"></div>
            <div class="panel-content">
              <component :is="activeComponent" :key="activeKey" />
            </div>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import ThemeSettingsView from '@/views/settings/ThemeSettingsView.vue';
import TemplateSettingsView from '@/views/settings/TemplateSettingsView.vue';
import SkillSettingsView from '@/views/settings/SkillSettingsView.vue';
import ApiKeySettingsView from '@/views/settings/ApiKeySettingsView.vue';
import ExportSettingsView from '@/views/settings/ExportSettingsView.vue';
import AboutSettingsView from '@/views/settings/AboutSettingsView.vue';
import { riToCompName } from '@/utils/icon';

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['update:visible']);

function close() {
  emit('update:visible', false);
}

const menuItems = [
  { key: 'theme', label: '主题管理', icon: 'ri-palette-line', component: ThemeSettingsView },
  { key: 'template', label: '模板管理', icon: 'ri-file-list-2-line', component: TemplateSettingsView },
  { key: 'skill', label: '技能管理', icon: 'ri-git-branch-line', component: SkillSettingsView },
  { key: 'model', label: '模型管理', icon: 'ri-cpu-line', component: ApiKeySettingsView },
  { key: 'export', label: '导出管理', icon: 'ri-download-2-line', component: ExportSettingsView },
  { key: 'about', label: '关于', icon: 'ri-information-line', component: AboutSettingsView },
];

const activeKey = ref('theme');

// 每次打开弹窗都聚焦主题管理
watch(() => props.visible, (v) => {
  if (v) activeKey.value = 'theme';
});

const activeComponent = computed(() => {
  return menuItems.find(i => i.key === activeKey.value)?.component;
});

function selectTab(key: string) {
  activeKey.value = key;
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.settings-panel {
  width: 1320px;
  max-width: 94vw;
  height: 86vh;
  max-height: 900px;
  background: var(--color-bg-secondary, #F7F3EA);
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  border-radius: 16px;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.1);
  padding: 28px 32px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding-bottom: 20px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.header-icon {
  width: 48px;
  height: 48px;
  background: var(--color-primary-soft, #FBE9E7);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  color: var(--color-primary, #C0392B);
  flex-shrink: 0;
}

.header-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.panel-title {
  font-size: 22px;
  font-weight: 700;
  color: var(--color-text-primary, #1F1D1A);
  margin: 0;
  line-height: 28px;
}

.panel-subtitle {
  font-size: 13px;
  color: var(--color-text-muted, #8A8578);
  margin: 0;
  line-height: 18px;
}

.close-btn {
  width: 40px;
  height: 40px;
  background: var(--color-bg-secondary, #F7F3EA);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: var(--color-text-secondary, #6E6A60);
  cursor: pointer;
  border: none;
  flex-shrink: 0;
}

.close-btn:hover {
  background: var(--color-border-soft, #EDE9E0);
}

.panel-body {
  flex: 1;
  min-height: 0;
  display: flex;
  gap: 20px;
}

.nav-list {
  width: 240px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
  background: var(--color-bg-white, #FFFFFF);
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  border-radius: 16px;
  padding: 16px;
}

.nav-item {
  height: 52px;
  padding: 0 20px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 14px;
  cursor: pointer;
  transition: background-color 0.2s;
  flex-shrink: 0;
}

.nav-item:hover {
  background-color: var(--color-bg-secondary, #F7F3EA);
}

.nav-item .icon {
  font-family: "remixicon", sans-serif;
  font-style: normal;
  font-size: 20px;
  color: var(--color-text-body, #3A3A3A);
  width: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-item .label {
  font-size: 15px;
  color: var(--color-text-body, #3A3A3A);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  white-space: nowrap;
}

.nav-item.active {
  background-color: var(--color-primary-soft);
}

.nav-item.active .icon,
.nav-item.active .label {
  color: var(--color-primary, #C0392B);
  font-weight: 500;
}

.panel-divider {
  display: none;
}

.panel-content {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
  padding-right: 4px;
}

/* 隐藏嵌入视图自带的页面头部卡片，避免与弹窗头重复 */
.panel-content :deep(.page-header-card) {
  display: none;
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
