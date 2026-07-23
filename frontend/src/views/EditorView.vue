<template>
  <div class="page">
    <header class="navbar">
      <div class="logo-area" @click="goHome">
        <div class="logo-icon">
          <RiRadarFill size="22" color="#fff" />
        </div>
        <div class="brand-texts">
          <span class="brand-zh">文制星</span>
          <span class="brand-en">Boomerang</span>
        </div>
      </div>
      <div class="nav-actions">
        <button class="nav-btn" title="帮助" @click="showHelp">
          <span class="nav-btn-content">
            <RiQuestionLine size="20" />
            <span class="nav-btn-label">帮助</span>
          </span>
        </button>
        <button class="nav-btn" title="设置" @click="goSettings">
          <span class="nav-btn-content">
            <RiSettingsLine size="20" />
            <span class="nav-btn-label">设置</span>
          </span>
        </button>
      </div>
    </header>

    <main class="editor-body">
      <aside class="left-panel">
        <OutlineTree @select="handleSelectSection" />
      </aside>
      <section class="center-panel">
        <ContentEditor />
      </section>
      <aside class="right-panel">
        <AIChat />
      </aside>
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import OutlineTree from '../components/OutlineTree.vue'
import ContentEditor from '../components/ContentEditor.vue'
import AIChat from '../components/AIChat.vue'
import {
  RiRadarFill,
  RiQuestionLine,
  RiSettingsLine,
} from '@remixicon/vue'

const props = defineProps<{ id: string }>()
const router = useRouter()
const docStore = useDocumentStore()

const goSettings = () => router.push('/settings')
const goHome = () => router.push('/')
const showHelp = () => alert('文制星 - 标书智能生成工具\n\n编辑大纲 → AI辅助写作 → 导出标书')

onMounted(() => {
  try {
    docStore.loadOutline(props.id)
  } catch (err) {
    console.error('Failed to load outline:', err)
  }
})

const handleSelectSection = (sectionId: string) => {
  docStore.loadSection(props.id, sectionId)
}
</script>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #FDF6E3;
}

.navbar {
  height: 64px;
  padding: 0 32px;
  background: #FDF6E3;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.logo-area {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.logo-icon {
  width: 44px;
  height: 44px;
  background: #C23B22;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.brand-texts {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.brand-zh {
  font-size: 22px;
  font-weight: 700;
  color: #3D2B1F;
  line-height: 1.2;
}

.brand-en {
  font-size: 12px;
  color: #8B7355;
  line-height: 1.3;
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.nav-btn {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: #F5EFE0;
  border: 0.7px solid #E0D5C0;
  cursor: pointer;
  transition: all 0.2s;
  overflow: hidden;
  white-space: nowrap;
  color: #5C4033;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-btn:hover {
  width: 90px;
  background: #C23B22;
  border-color: transparent;
  color: #fff;
}

.nav-btn-content {
  display: flex;
  align-items: center;
  gap: 6px;
}

.nav-btn-label {
  font-size: 13px;
  font-weight: 500;
  color: inherit;
  display: none;
}

.nav-btn:hover .nav-btn-label {
  display: inline;
}

.editor-body {
  flex: 1;
  display: flex;
  gap: 16px;
  padding: 0 32px 24px;
  overflow: hidden;
}

.left-panel {
  width: 260px;
  flex-shrink: 0;
  background: #FBF7EF;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.center-panel {
  flex: 0 0 50%;
  max-width: 55%;
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.right-panel {
  flex: 1;
  min-width: 300px;
  background: #FBF7EF;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>