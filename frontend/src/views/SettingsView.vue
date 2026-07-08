<template>
  <div class="page">
    <header class="navbar">
      <div class="navbar-left" @click="goHome">
        <div class="logo">
          <RiRadarFill size="22" color="#fff" />
        </div>
        <span class="brand-cn">文制星</span>
        <span class="brand-en">Boomerang</span>
      </div>
      <div class="navbar-right">
        <button class="back-btn" @click="goBack">
          <RiArrowLeftLine size="16" color="#8B7355" />
          <span>返回</span>
        </button>
      </div>
    </header>

    <div class="body">
      <aside class="sidebar">
        <span class="sidebar-label">系统设置</span>
        <nav class="nav-list">
          <button
            v-for="item in navItems"
            :key="item.id"
            class="nav-item"
            :class="{ 'nav-item-active': activeNav === item.id }"
            @click="activeNav = item.id"
          >
            <component :is="item.icon" :size="'20'" />
            <span class="nav-label">{{ item.label }}</span>
          </button>
        </nav>
      </aside>

      <main class="content">
        <div class="content-title-group">
          <h3 class="content-title">{{ currentNav.title }}</h3>
          <span class="content-subtitle">{{ currentNav.desc }}</span>
        </div>

        <div class="content-scroll">
          <!-- 主题设置 -->
          <div v-if="activeNav === 'theme'" class="panel">
            <div class="theme-grid">
              <div
                v-for="t in themes"
                :key="t.id"
                class="theme-card"
                :class="{ 'theme-card-active': selectedTheme === t.id }"
                @click="selectedTheme = t.id"
              >
                <div class="theme-preview" :style="{ background: t.preview }" />
                <div class="theme-info">
                  <span class="theme-name">{{ t.name }}</span>
                  <span class="theme-desc">{{ t.desc }}</span>
                </div>
                <div class="theme-check" :class="{ 'theme-checked': selectedTheme === t.id }">
                  <RiCheckLine v-if="selectedTheme === t.id" size="10" color="#fff" />
                </div>
              </div>
            </div>
          </div>

          <!-- 模板设置 (Calicat Template Shelf) -->
          <div v-if="activeNav === 'template'" class="panel">
            <div class="tpl-tabs">
              <button
                v-for="tab in tplTabs"
                :key="tab"
                class="tpl-tab"
                :class="{ 'tpl-tab-active': activeTplTab === tab }"
                @click="activeTplTab = tab"
              >{{ tab }}</button>
            </div>
            <div class="tpl-shelf">
              <div
                v-for="(tpl, idx) in tplCards"
                :key="idx"
                class="tpl-card"
                @click="selectedTplCard = idx"
              >
                <div class="tpl-card-cover">
                  <div class="tpl-card-icon" :style="{ background: tpl.iconBg }">
                    <component :is="tpl.iconComp" :size="'24'" :color="tpl.iconColor" />
                  </div>
                  <span class="tpl-card-cat">{{ tpl.category }}</span>
                  <span class="tpl-card-label">标准模板</span>
                </div>
                <div class="tpl-card-info">
                  <span class="tpl-card-name">{{ tpl.name }}</span>
                  <span class="tpl-card-desc">{{ tpl.desc }}</span>
                </div>
                <div v-if="selectedTplCard === idx" class="tpl-card-check">
                  <RiCheckLine size="12" color="#fff" />
                </div>
              </div>
              <div class="tpl-card tpl-card-add">
                <div class="tpl-add-icon">
                  <RiAddLine size="22" color="#8B7355" />
                </div>
                <span class="tpl-add-text">添加模板</span>
              </div>
            </div>
          </div>

          <!-- 规则设置 (placeholder) -->
          <div v-if="activeNav === 'rules'" class="panel">
            <div class="rules-placeholder">
              <RiFileListLine size="48" color="#D4C4A8" />
              <span class="rules-placeholder-text">规则设置功能开发中</span>
            </div>
          </div>

          <!-- 导出设置 (Calicat Export Format Cards) -->
          <div v-if="activeNav === 'export'" class="panel export-panel">
            <div class="export-cards">
              <div class="export-card">
                <div class="export-card-header">
                  <div class="export-card-icon" style="background: #E8F0F8">
                    <RiFileWord2Line size="28" color="#2D6A9F" />
                  </div>
                  <div class="export-card-titles">
                    <span class="export-card-name">Word 格式</span>
                    <span class="export-card-ext">.docx</span>
                  </div>
                  <div class="export-check" :class="{ 'export-checked': exportFormat === 'word' }" @click="exportFormat = 'word'">
                    <RiCheckLine v-if="exportFormat === 'word'" size="14" color="#fff" />
                  </div>
                </div>
                <div class="export-features">
                  <div v-for="f in wordFeatures" :key="f" class="export-feature">
                    <RiCheckLine size="16" color="#2D8A4E" />
                    <span>{{ f }}</span>
                  </div>
                </div>
              </div>
              <div class="export-card">
                <div class="export-card-header">
                  <div class="export-card-icon" style="background: #F0E8D8">
                    <RiMarkdownLine size="28" color="#8B7355" />
                  </div>
                  <div class="export-card-titles">
                    <span class="export-card-name">Markdown 格式</span>
                    <span class="export-card-ext">.md</span>
                  </div>
                  <div class="export-check" :class="{ 'export-checked': exportFormat === 'md' }" @click="exportFormat = 'md'">
                    <RiCheckLine v-if="exportFormat === 'md'" size="14" color="#fff" />
                  </div>
                </div>
                <div class="export-features">
                  <div v-for="f in mdFeatures" :key="f" class="export-feature">
                    <RiCheckLine size="16" color="#2D8A4E" />
                    <span>{{ f }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- API Key (Calicat Model List + Config Form) -->
          <div v-if="activeNav === 'apikey'" class="panel api-full-panel">
            <div class="model-list">
              <span class="model-section">国内模型</span>
              <div
                v-for="m in domesticModels"
                :key="m.id"
                class="model-item"
                :class="{ 'model-item-active': settingsStore.selectedModelId === m.id }"
                @click="settingsStore.selectedModelId = m.id"
              >
                <div class="model-icon-wrap" :class="{ 'model-icon-active': settingsStore.selectedModelId === m.id }">
                  <component :is="m.icon" :size="'16'" :color="settingsStore.selectedModelId === m.id ? '#fff' : '#8B7355'" />
                </div>
                <span class="model-name" :class="{ 'model-name-active': settingsStore.selectedModelId === m.id }">{{ m.name }}</span>
              </div>

              <div class="model-divider" />

              <span class="model-section">国外模型</span>
              <div
                v-for="m in foreignModels"
                :key="m.id"
                class="model-item"
                :class="{ 'model-item-active': settingsStore.selectedModelId === m.id }"
                @click="settingsStore.selectedModelId = m.id"
              >
                <div class="model-icon-wrap" :class="{ 'model-icon-active': settingsStore.selectedModelId === m.id }">
                  <component :is="m.icon" :size="'16'" :color="settingsStore.selectedModelId === m.id ? '#fff' : '#8B7355'" />
                </div>
                <span class="model-name" :class="{ 'model-name-active': settingsStore.selectedModelId === m.id }">{{ m.name }}</span>
              </div>
            </div>

            <div class="config-panel">
              <div class="config-tabs">
                <button class="config-tab-selected">模型制造商</button>
                <button class="config-tab">自定义配置</button>
              </div>
              <div class="config-form">
                <div class="form-field">
                  <label class="form-label">服务商</label>
                  <div class="form-input">{{ currentModelConfig.provider }}</div>
                </div>
                <div class="form-field">
                  <label class="form-label">模型</label>
                  <div class="form-input">{{ currentModelConfig.model }}</div>
                </div>
                <div class="form-field">
                  <label class="form-label">API Key</label>
                  <div class="form-input form-input-row">
                    <span class="form-key-masked">sk-••••••••••••••••••••••••</span>
                    <button class="form-key-toggle">
                      <RiEyeOffLine size="16" color="#8B7355" />
                    </button>
                  </div>
                </div>
                <div class="form-actions">
                  <button class="form-btn-cancel">取消</button>
                  <button class="form-btn-add">添加</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useSettingsStore } from '../stores/settingsStore'
import {
  RiRadarFill,
  RiArrowLeftLine,
  RiPaletteLine,
  RiBookmarkLine,
  RiFileListLine,
  RiFileDownloadLine,
  RiKeyLine,
  RiCheckLine,
  RiAddLine,
  RiFileTextLine,
  RiBuildingLine,
  RiServerLine,
  RiCustomerServiceLine,
  RiFileWord2Line,
  RiMarkdownLine,
  RiEyeOffLine,
  RiRobotLine,
  RiOpenaiFill,
} from '@remixicon/vue'

const router = useRouter()

const goHome = () => router.push('/')
function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/')
  }
}

interface NavItem {
  id: string
  label: string
  title: string
  desc: string
  icon: any
}

const navItems: NavItem[] = [
  { id: 'theme', label: '主题设置', title: '主题设置', desc: '选择你喜欢的界面风格', icon: RiPaletteLine },
  { id: 'template', label: '模板设置', title: '模板设置', desc: '管理标书模板', icon: RiBookmarkLine },
  { id: 'rules', label: '规则设置', title: '规则设置', desc: '配置标书生成规则', icon: RiFileListLine },
  { id: 'export', label: '导出设置', title: '导出设置', desc: '配置标书导出的默认格式', icon: RiFileDownloadLine },
  { id: 'apikey', label: 'API Key', title: 'API Key', desc: '管理 AI 模型密钥', icon: RiKeyLine },
]

const activeNav = ref('theme')
const selectedTheme = ref('light')
const exportFormat = ref('word')
const settingsStore = useSettingsStore()

const themes = [
  { id: 'light', name: '浅色主题', desc: '经典羊皮纸底色', preview: '#FDF6E3' },
  { id: 'dark', name: '深色主题', desc: '深色护眼模式', preview: '#2C2416' },
  { id: 'paper', name: '纯白纸', desc: '清爽干净', preview: '#FFFFFF' },
]

const tplTabs = ['招标模板', '投标模板', '自定义模板']
const activeTplTab = ref('招标模板')
const selectedTplCard = ref(-1)

interface TplCard {
  name: string
  desc: string
  category: string
  iconComp: any
  iconBg: string
  iconColor: string
}

const tplCards: TplCard[] = [
  { name: '政府采购货物类', desc: '适用于货物类采购项目', category: '政府采购', iconComp: RiFileTextLine, iconBg: '#C23B22', iconColor: '#fff' },
  { name: '工程施工类', desc: '适用于工程施工招标', category: '工程建设', iconComp: RiBuildingLine, iconBg: '#2D6A9F', iconColor: '#fff' },
  { name: '信息化服务类', desc: '适用于IT服务采购', category: 'IT服务', iconComp: RiServerLine, iconBg: '#2D8A4E', iconColor: '#fff' },
  { name: '咨询服务类', desc: '适用于咨询类采购', category: '咨询服务', iconComp: RiCustomerServiceLine, iconBg: '#D4A017', iconColor: '#fff' },
]

const wordFeatures = [
  '保留完整格式与排版样式',
  '支持表格、图片、页眉页脚',
  '兼容 Microsoft Word / WPS',
  '支持目录自动生成',
]

const mdFeatures = [
  '纯文本格式，轻量易读',
  '适合版本管理与协作',
  '可快速转换为 HTML/PDF',
  '兼容各类 Markdown 编辑器',
]

interface ModelItem {
  id: string
  name: string
  icon: any
  provider: string
  model: string
}

const domesticModels: ModelItem[] = [
  { id: 'qwen', name: '通义千问', icon: RiRobotLine, provider: '阿里云', model: 'qwen-turbo' },
  { id: 'wenxin', name: '文心一言', icon: RiRobotLine, provider: '百度', model: 'ernie-4.0' },
  { id: 'glm', name: '智谱 GLM', icon: RiRobotLine, provider: '智谱', model: 'glm-4' },
]

const foreignModels: ModelItem[] = [
  { id: 'gpt4o', name: 'GPT-4o', icon: RiOpenaiFill, provider: 'OpenAI', model: 'gpt-4o' },
  { id: 'claude', name: 'Claude 3.5', icon: RiRobotLine, provider: 'Anthropic', model: 'claude-3-5-sonnet' },
]

const currentModelConfig = computed(() => settingsStore.selectedModel)

const currentNav = computed(() => navItems.find(i => i.id === activeNav.value)!)

</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #FDF6E3;
  display: flex;
  flex-direction: column;
  font-family: 'Noto Sans SC', -apple-system, BlinkMacSystemFont, sans-serif;
}

/* ── Navbar ── */
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 24px;
  background: #FBF7F0;
  flex-shrink: 0;
}

.navbar-left {
  display: flex;
  align-items: center;
  gap: 0;
  cursor: pointer;
}

.logo {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #C23B22;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.brand-cn {
  font-size: 22px;
  font-weight: 700;
  font-family: 'Noto Serif SC', serif;
  color: #3D2B1F;
  margin-left: 12px;
}

.brand-en {
  font-size: 12px;
  color: #8B7355;
  margin-left: 6px;
  line-height: 1;
  align-self: flex-end;
  padding-bottom: 3px;
}

.navbar-right {
  display: flex;
  align-items: center;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 40px;
  padding: 0 16px;
  background: #F0E8D5;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: #8B7355;
  transition: background 0.2s;
}

.back-btn:hover {
  background: #E0D5C0;
}

/* ── Body ── */
.body {
  display: flex;
  flex: 1;
  min-height: 0;
}

/* ── Sidebar ── */
.sidebar {
  width: 220px;
  flex-shrink: 0;
  background: #F5EFE0;
  padding: 24px 24px;
  display: flex;
  flex-direction: column;
}

.sidebar-label {
  font-size: 11px;
  font-weight: 600;
  color: #8B7355;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 8px;
}

.nav-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: none;
  border-radius: 12px;
  background: transparent;
  cursor: pointer;
  font-size: 14px;
  color: #8B7355;
  transition: all 0.2s;
  text-align: left;
  width: 100%;
}

.nav-item:hover {
  background: rgba(0,0,0,0.03);
}

.nav-item-active {
  background: #C23B22;
  color: #fff;
}

.nav-item-active .nav-label {
  color: #fff;
  font-weight: 600;
}

/* ── Content ── */
.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  padding: 32px;
  background: #FBF7EF;
}

.content-title-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 24px;
}

.content-title {
  font-size: 24px;
  font-weight: 700;
  color: #3D2B1F;
  margin: 0;
}

.content-subtitle {
  font-size: 14px;
  color: #8B7355;
}

.content-scroll {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}

.panel {
  background: transparent;
}

/* ── Theme Grid ── */
.theme-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.theme-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  text-align: center;
  cursor: pointer;
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  transition: box-shadow 0.2s;
}

.theme-card:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
}

.theme-card-active {
  box-shadow: 0 8px 24px rgba(194,59,34,0.18);
}

.theme-preview {
  width: 64px;
  height: 64px;
  border-radius: 8px;
  border: 1px solid #E0D5C0;
  margin-bottom: 10px;
}

.theme-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.theme-name {
  font-size: 13px;
  font-weight: 700;
  color: #3D2B1F;
}

.theme-desc {
  font-size: 11px;
  color: #8B7355;
}

.theme-check {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #D4C4A8;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.theme-checked {
  background: #C23B22;
}

/* ── Template Tabs ── */
.tpl-tabs {
  display: flex;
  gap: 4px;
  background: #F0E8D5;
  border-radius: 12px;
  padding: 4px;
  width: fit-content;
  margin-bottom: 24px;
}

.tpl-tab {
  padding: 8px 24px;
  border: none;
  border-radius: 8px;
  font-size: 13px;
  cursor: pointer;
  background: transparent;
  color: #8B7355;
  font-weight: 500;
}

.tpl-tab-active {
  background: #C23B22;
  color: #fff;
  font-weight: 600;
}

/* ── Template Shelf ── */
.tpl-shelf {
  display: flex;
  gap: 20px;
}

.tpl-card {
  width: 200px;
  border-radius: 12px;
  background: #fff;
  position: relative;
  cursor: pointer;
  overflow: hidden;
  transition: box-shadow 0.2s;
}

.tpl-card:hover {
  box-shadow: 0 4px 16px rgba(0,0,0,0.1);
}

.tpl-card-cover {
  height: 240px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.tpl-card-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 4px;
}

.tpl-card-cat {
  font-size: 14px;
  font-weight: 600;
  color: #3D2B1F;
}

.tpl-card-label {
  font-size: 12px;
  color: #8B7355;
}

.tpl-card-info {
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.tpl-card-name {
  font-size: 13px;
  font-weight: 500;
  color: #3D2B1F;
}

.tpl-card-desc {
  font-size: 11px;
  color: #8B7355;
}

.tpl-card-check {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: #C23B22;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ── Add Template Card ── */
.tpl-card-add {
  background: #F5EFE3;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 340px;
  gap: 8px;
}

.tpl-add-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: #F0E8D5;
  display: flex;
  align-items: center;
  justify-content: center;
}

.tpl-add-text {
  font-size: 14px;
  font-weight: 500;
  color: #8B7355;
}

/* ── Rules Placeholder ── */
.rules-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 0;
  gap: 16px;
}

.rules-placeholder-text {
  font-size: 16px;
  color: #8B7355;
}

/* ── Export Settings ── */
.export-cards {
  display: flex;
  gap: 24px;
}

.export-card {
  flex: 1;
  background: #fff;
  border-radius: 16px;
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.export-card-header {
  display: flex;
  align-items: center;
  gap: 16px;
}

.export-card-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.export-card-titles {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
}

.export-card-name {
  font-size: 20px;
  font-weight: 700;
  color: #3D2B1F;
}

.export-card-ext {
  font-size: 13px;
  color: #8B7355;
}

.export-check {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #D4C4A8;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.2s;
  flex-shrink: 0;
}

.export-checked {
  background: #C23B22;
}

.export-features {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.export-feature {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #3D2B1F;
}

/* ── API Key (Model List + Config Form) ── */
.api-full-panel {
  display: flex;
  gap: 32px;
  min-height: 0;
}

.model-list {
  width: 260px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.model-section {
  font-size: 11px;
  font-weight: 600;
  color: #8B7355;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 8px 0 4px;
}

.model-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.model-item:hover {
  background: rgba(0,0,0,0.03);
}

.model-item-active {
  background: #C23B22;
}

.model-icon-wrap {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #E8DCC8;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background 0.2s;
}

.model-icon-active {
  background: rgba(255,255,255,0.2);
}

.model-name {
  font-size: 14px;
  color: #8B7355;
  font-weight: 500;
}

.model-name-active {
  color: #fff;
  font-weight: 600;
}

.model-divider {
  height: 1px;
  background: #E0D5C0;
  margin: 8px 0;
}

/* ── Config Panel ── */
.config-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 24px;
  min-width: 0;
}

.config-tabs {
  display: flex;
  gap: 4px;
  background: #F0E8D5;
  border-radius: 12px;
  padding: 4px;
  width: fit-content;
}

.config-tab-selected {
  padding: 8px 20px;
  border: none;
  border-radius: 8px;
  background: #C23B22;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.config-tab {
  padding: 8px 20px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #8B7355;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.config-form {
  background: #fff;
  border-radius: 16px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-field {
  display: flex;
  align-items: center;
  gap: 16px;
}

.form-label {
  width: 100px;
  font-size: 14px;
  font-weight: 600;
  color: #3D2B1F;
  flex-shrink: 0;
}

.form-input {
  flex: 1;
  background: #F5EFE3;
  border-radius: 8px;
  padding: 12px 16px;
  font-size: 14px;
  color: #3D2B1F;
}

.form-input-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.form-key-masked {
  font-size: 14px;
  color: #3D2B1F;
  font-family: monospace;
}

.form-key-toggle {
  border: none;
  background: transparent;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.form-btn-cancel {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  background: #F0E8D5;
  color: #8B7355;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.form-btn-add {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  background: #C23B22;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.form-btn-add:hover {
  opacity: 0.9;
}
</style>
