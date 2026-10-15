<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ElMessage } from '@/utils/toast';
import { useThemeStore } from '@/stores/theme';
import { UiCheck } from '@/components/ui';
import { setTheme as saveTheme } from '@/api/settings';

const themeStore = useThemeStore();

interface ThemeOption {
  type: string;
  name: string;
  desc: string;
  skin: string;
  topbar: string;
  content: string;
  previewBg: string;
}

const themes: ThemeOption[] = [
  {
    type: 'parchment',
    name: '浅色主题 · 羊皮纸',
    desc: '当前默认，长时间阅读标书不易疲劳',
    skin: '#F7EED4',
    topbar: '#FFFFFF',
    content: '#E1DCD1',
    previewBg: '#F7EED4',
  },
  {
    type: 'dark',
    name: '深色主题 · 墨夜',
    desc: '夜间或暗光环境投标作业更护眼',
    skin: '#333A45',
    topbar: '#404955',
    content: '#8E939A',
    previewBg: '#333A45',
  },
];

interface AccentOption {
  key: string;
  name: string;
  color: string;
  varName: string;
}

const accents: AccentOption[] = [
  { key: 'vermilion', name: '朱红 · 默认', color: '#BE3B2E', varName: '--color-primary' },
  { key: 'gold', name: '鎏金', color: '#D3A93F', varName: '--color-accent-gold' },
  { key: 'jade', name: '玉绿', color: '#3E8C6E', varName: '--color-accent-jade' },
  { key: 'cloud', name: '云蓝', color: '#4A72A8', varName: '--color-accent-cloud' },
  { key: 'ink', name: '黛墨', color: '#2F3640', varName: '--color-accent-ink' },
];

interface LayoutOption {
  key: string;
  name: string;
  desc: string;
}

const layouts: LayoutOption[] = [
  { key: 'topbar', name: '顶栏模式 · 居中导航', desc: '六大模块在顶栏居中排布，右上角统一为返回与设置' },
  { key: 'sidebar', name: '侧边栏模式 · 左侧导航', desc: '六大模块在左侧竖列居中，顶栏不显示图标文字' },
];

const selectedTheme = ref('parchment');
const selectedAccent = ref('vermilion');
const selectedLayout = ref('topbar');

onMounted(() => {
  const savedTheme = localStorage.getItem('theme');
  if (savedTheme && themes.some(t => t.type === savedTheme)) {
    selectedTheme.value = savedTheme;
  }
  selectedAccent.value = localStorage.getItem('accent') || 'vermilion';
  selectedLayout.value = localStorage.getItem('layout') || 'topbar';
});

async function selectTheme(type: string) {
  selectedTheme.value = type;
  await themeStore.setTheme(type as any);
  try {
    await saveTheme(type);
  } catch {}
}

function selectAccent(key: string) {
  selectedAccent.value = key;
  const accent = accents.find(a => a.key === key);
  if (!accent) return;
  localStorage.setItem('accent', key);
  themeStore.applyAccent(accent.color);
}

function selectLayout(key: string) {
  selectedLayout.value = key;
  localStorage.setItem('layout', key);
  ElMessage.success(`已切换布局：${layouts.find(l => l.key === key)?.name || ''}`);
}

function resetAll() {
  selectedTheme.value = 'parchment';
  selectedAccent.value = 'vermilion';
  selectedLayout.value = 'topbar';
  themeStore.setTheme('parchment');
  saveTheme('parchment').catch(() => {});
  localStorage.setItem('accent', 'vermilion');
  themeStore.applyAccent('#BE3B2E');
  localStorage.setItem('layout', 'sidebar');
  ElMessage.success('已恢复默认');
}
</script>

<template>
  <div class="theme-view">
    <!-- 页面头部卡片 -->
    <div class="page-header-card">
      <div class="header-content">
        <div class="header-info">
          <div class="breadcrumb">
            <span>首页</span>
            <span class="sep">/</span>
            <span>设置中心</span>
            <span class="sep">/</span>
            <span class="current">主题管理</span>
          </div>
          <div class="header-title">
            <h1>主题管理</h1>
            <p>统一切换外观底色、强调色与导航布局模式，修改即时生效</p>
          </div>
        </div>
        <button class="reset-button" @click="resetAll">
          <RiRestartLine class="icon" />
          恢复默认
        </button>
      </div>
    </div>

    <!-- 外观主题卡片 -->
    <div class="section-card">
      <div class="card-title">外观主题</div>
      <div class="theme-group">
        <div
          v-for="theme in themes"
          :key="theme.type"
          class="theme-preview-card"
          :class="{ active: selectedTheme === theme.type }"
          @click="selectTheme(theme.type)"
        >
          <UiCheck :checked="selectedTheme === theme.type" class="card-check" :size="22" />
          <div class="preview-thumb" :style="{ backgroundColor: theme.previewBg }">
            <div class="thumb-topbar" :style="{ backgroundColor: theme.topbar }"></div>
            <div class="thumb-content" :style="{ backgroundColor: theme.content }"></div>
          </div>
          <div class="preview-info">
            <div class="preview-name">{{ theme.name }}</div>
            <div class="preview-desc">{{ theme.desc }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 强调色卡片 -->
    <div class="section-card">
      <div class="card-title">强调色</div>
      <div class="accent-group">
        <div
          v-for="accent in accents"
          :key="accent.key"
          class="accent-item"
          @click="selectAccent(accent.key)"
        >
          <div class="accent-circle" :style="{ backgroundColor: accent.color }">
            <RiCheckLine v-if="selectedAccent === accent.key" class="icon" />
          </div>
          <div class="accent-name" :class="{ current: selectedAccent === accent.key }">{{ accent.name }}</div>
        </div>
      </div>
    </div>

    <!-- 布局切换卡片 -->
    <div class="section-card">
      <div class="card-title">布局切换</div>
      <div class="layout-group">
        <div
          v-for="layout in layouts"
          :key="layout.key"
          class="layout-card"
          :class="{ active: selectedLayout === layout.key }"
          @click="selectLayout(layout.key)"
        >
          <UiCheck :checked="selectedLayout === layout.key" class="card-check" :size="22" />
          <div class="layout-preview">
            <template v-if="layout.key === 'topbar'">
              <div class="preview-mockup topbar-mockup">
                <div class="mockup-header">
                  <div class="mockup-brand">
                    <span class="brand-dot"></span>
                    <span class="brand-text"></span>
                    <span class="brand-text muted"></span>
                  </div>
                  <div class="mockup-modules">
                    <span class="module-block" style="background:#BE3B2E"></span>
                    <span class="module-block" style="background:#D3A93F"></span>
                    <span class="module-block" style="background:#3E8C6E"></span>
                    <span class="module-block" style="background:#4A72A8"></span>
                    <span class="module-block" style="background:#BE3B2E"></span>
                    <span class="module-block" style="background:#D3A93F"></span>
                  </div>
                  <div class="mockup-actions">
                    <span class="action-dot"></span>
                    <span class="action-dot"></span>
                  </div>
                </div>
                <div class="mockup-body">
                  <div class="mockup-side"></div>
                  <div class="mockup-main"></div>
                </div>
              </div>
            </template>
            <template v-else>
              <div class="preview-mockup sidebar-mockup">
                <div class="mockup-sidebar">
                  <span class="sidebar-mod" style="background:#D3A93F"></span>
                  <span class="sidebar-mod" style="background:#BE3B2E"></span>
                  <span class="sidebar-mod" style="background:#3E8C6E"></span>
                  <span class="sidebar-mod" style="background:#4A72A8"></span>
                  <span class="sidebar-mod" style="background:#2F3640"></span>
                </div>
                <div class="mockup-content">
                  <div class="mockup-header">
                    <div class="mockup-brand">
                      <span class="brand-dot"></span>
                      <span class="brand-text"></span>
                      <span class="brand-text muted"></span>
                    </div>
                    <div class="mockup-actions">
                      <span class="action-dot"></span>
                      <span class="action-dot"></span>
                    </div>
                  </div>
                  <div class="mockup-body">
                    <div class="mockup-main"></div>
                  </div>
                </div>
              </div>
            </template>
          </div>
          <div class="layout-info">
            <div class="layout-name">{{ layout.name }}</div>
            <div class="layout-desc">{{ layout.desc }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.theme-view {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-width: 1620px;
}

/* 页面头部卡片 */
.page-header-card {
  background-color: var(--color-bg-white, #FFFFFF);
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  border-radius: 16px;
  padding: 20px 24px;
}
.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
}
.header-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
.breadcrumb span {
  color: var(--color-text-faint, #A7A296);
}
.breadcrumb span.sep {
  color: var(--color-text-faint, #A7A296);
}
.breadcrumb span.current {
  color: var(--color-text-body, #3A3A3A);
  font-weight: 500;
}
.header-title h1 {
  font-size: 30px;
  font-weight: 700;
  color: var(--color-text-primary, #1F1D1A);
  margin: 12px 0 0 0;
  letter-spacing: -0.5px;
  font-family: "SourceHanSans-Bold", var(--font-family-cn);
  line-height: 1.2;
}
.header-title p {
  font-size: 15px;
  color: var(--color-text-secondary, #6E6A60);
  margin: 8px 0 0 0;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
.reset-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0 20px;
  height: 48px;
  border: 0.7px solid var(--color-border, #E5DFD3);
  background-color: var(--color-bg-white, #FFFFFF);
  border-radius: 12px;
  font-size: 15px;
  color: var(--color-text-body, #3A3A3A);
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
.reset-button:hover {
  background-color: var(--color-bg-secondary, #F7F3EA);
}
.reset-button .icon {
  font-family: "remixicon", sans-serif;
  font-style: normal;
  font-size: 18px;
}

/* 通用卡片 */
.section-card {
  background-color: var(--color-bg-white, #FFFFFF);
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  border-radius: 16px;
  padding: 24px 28px;
}
.card-title {
  font-size: 17px;
  font-weight: 500;
  color: var(--color-text-primary, #1F1D1A);
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
}

/* 统一勾选徽标（UiCheck） */
.card-check {
  position: absolute;
  top: 14px;
  right: 14px;
  z-index: 2;
}

/* 外观主题 */
.theme-group {
  display: flex;
  gap: 24px;
  margin-top: 20px;
}
.theme-preview-card {
  flex: 1;
  position: relative;
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  background-color: var(--color-bg-white, #FFFFFF);
  border-radius: 14px;
  padding: 24px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 24px;
  transition: all 0.2s;
}
.theme-preview-card.active {
  background-color: var(--color-primary-soft);
  border-color: #E4CD8E;
}
.preview-thumb {
  width: 100%;
  height: 128px;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}
.thumb-topbar {
  width: 100%;
  height: 28px;
  border-radius: 5px;
}
.thumb-content {
  width: 100%;
  height: 54px;
  border-radius: 6px;
}
.preview-info .preview-name {
  font-size: 16px;
  font-weight: 500;
  color: var(--color-text-primary, #1F1D1A);
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
}
.preview-info .preview-desc {
  font-size: 14px;
  color: var(--color-text-muted, #8A8578);
  margin-top: 8px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}

/* 强调色 */
.accent-group {
  display: flex;
  gap: 40px;
  margin-top: 24px;
}
.accent-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  cursor: pointer;
}
.accent-circle {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.accent-circle .icon {
  font-family: "remixicon", sans-serif;
  font-style: normal;
  font-size: 30px;
  color: #fff;
}
.accent-name {
  font-size: 15px;
  color: var(--color-text-body, #3A3A3A);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
.accent-name.current {
  color: var(--color-text-primary, #1F1D1A);
  font-weight: 500;
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
}

/* 布局切换 */
.layout-group {
  display: flex;
  gap: 24px;
  margin-top: 20px;
}
.layout-card {
  flex: 1;
  position: relative;
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  background-color: var(--color-bg-white, #FFFFFF);
  border-radius: 14px;
  padding: 24px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 24px;
  transition: all 0.2s;
}
.layout-card.active {
  background-color: var(--color-primary-soft);
  border-color: #E4CD8E;
}
.layout-preview {
  width: 100%;
  height: 150px;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
}
.preview-mockup {
  flex: 1;
  display: flex;
  flex-direction: column;
}
.topbar-mockup {
  background-color: rgba(255,255,255,1);
  border-radius: 8px;
}
.mockup-header {
  height: 32px;
  background-color: #fff;
  border-radius: 6px;
  padding: 0 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.mockup-brand {
  display: flex;
  align-items: center;
  gap: 8px;
}
.brand-dot {
  width: 11px;
  height: 10px;
  border-radius: 50%;
  background-color: #BE3B2E;
}
.brand-text {
  width: 25px;
  height: 6px;
  border-radius: 50%;
  background-color: #2F3640;
}
.brand-text.muted {
  width: 17px;
  background-color: #A7A296;
}
.mockup-modules {
  display: flex;
  gap: 8px;
}
.module-block {
  width: 31px;
  height: 10px;
  border-radius: 50%;
}
.mockup-actions {
  margin-left: auto;
  display: flex;
  gap: 4px;
}
.action-dot {
  width: 9px;
  height: 8px;
  border-radius: 50%;
  background-color: #D6CDC8;
}
.mockup-body {
  flex: 1;
  margin-top: 12px;
  display: flex;
  gap: 12px;
}
.mockup-side {
  width: 85px;
  border-radius: 6px;
  background-color: #E2DDD2;
}
.mockup-main {
  flex: 1;
  border-radius: 6px;
  background-color: #E6E2D8;
}
.sidebar-mockup {
  background-color: rgba(245,241,230,1);
  border-radius: 8px;
  flex-direction: row;
}
.mockup-sidebar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  align-items: center;
  justify-content: center;
  padding: 12px 0;
  width: 56px;
}
.sidebar-mod {
  width: 44px;
  height: 9px;
  border-radius: 50%;
}
.mockup-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding-left: 12px;
}
.sidebar-mockup .mockup-header {
  margin-bottom: 12px;
}
.sidebar-mockup .mockup-body {
  flex: 1;
}
.sidebar-mockup .mockup-main {
  width: 100%;
  height: 100%;
  border-radius: 6px;
  background-color: #E6E2D8;
}
.layout-info .layout-name {
  font-size: 16px;
  font-weight: 500;
  color: var(--color-text-primary, #1F1D1A);
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
}
.layout-info .layout-desc {
  font-size: 14px;
  color: var(--color-text-muted, #8A8578);
  margin: 8px 0 0 0;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
</style>