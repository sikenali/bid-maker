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
        <button class="nav-btn" title="帮助">
          <RiQuestionLine size="18" color="#8B7355" />
        </button>
        <button class="nav-btn" title="设置" @click="goSettings">
          <RiSettingsLine size="18" color="#8B7355" />
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

onMounted(() => {
  docStore.loadOutline(props.id)
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
  background: #FBF7F0;
}

.navbar {
  height: 64px;
  padding: 0 24px;
  background: #FBF7F0;
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
  width: 40px;
  height: 40px;
  background: #C43D3D;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.brand-texts {
  display: flex;
  align-items: center;
  gap: 4px;
}

.brand-zh {
  font-size: 22px;
  font-weight: 700;
  color: #3D2B1F;
  line-height: 1.3;
}

.brand-en {
  font-size: 12px;
  color: #8B7355;
  line-height: 1.3;
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.nav-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: #F0E8D8;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
  flex-shrink: 0;
}

.nav-btn:hover {
  background: #E8DCC8;
}

.editor-body {
  flex: 1;
  display: flex;
  gap: 16px;
  padding: 0 24px 24px;
  overflow: hidden;
}

.left-panel {
  width: 260px;
  flex-shrink: 0;
  background: #F5EFE3;
  border-radius: 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.center-panel {
  flex: 1;
  background: #fff;
  border-radius: 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.right-panel {
  width: 300px;
  flex-shrink: 0;
  background: #F5EFE3;
  border-radius: 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>