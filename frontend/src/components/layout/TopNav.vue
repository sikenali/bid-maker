<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import SettingsModal from '@/components/layout/SettingsModal.vue';
import { UiTip } from '@/components/ui';
import { BRAND_NAME } from '@/constants/brand';

const props = withDefaults(defineProps<{
  title?: string;
  brand?: string;
  showBack?: boolean;
}>(), {
  showBack: true,
});

const router = useRouter();
const showSettingsModal = ref(false);
</script>

<template>
  <header class="top-nav">
    <div class="nav-left" @click="router.push('/upload')">
      <div class="tool-mark">
        <RiTargetLine class="tool-logo" />
      </div>
      <span class="brand-cn">{{ brand || BRAND_NAME }}</span>
      <span class="divider"></span>
      <span class="page-title">{{ title || '' }}</span>
    </div>
    <div class="nav-right">
      <UiTip v-if="showBack" text="返回" placement="bottom">
        <button class="nav-btn" @click="router.push('/upload')">
          <RiArrowLeftSLine class="icon" />
        </button>
      </UiTip>
      <UiTip text="设置" placement="bottom">
        <button class="nav-btn" @click="showSettingsModal = true">
          <RiSettings3Line class="icon" />
        </button>
      </UiTip>
    </div>
    <SettingsModal v-model:visible="showSettingsModal" />
  </header>
</template>

<style scoped>
.top-nav {
  height: 72px;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: var(--color-bg-white, #FFFFFF);
  border-bottom: 0.7px solid var(--color-border-soft, #EDE9E0);
}
.nav-left {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}
.nav-left:hover .brand-cn {
  color: var(--color-primary, #C0392B);
}
.tool-mark {
  width: auto;
  height: auto;
  display: flex;
  align-items: center;
  justify-content: center;
}
.tool-logo {
  font-family: "remixicon", sans-serif;
  font-style: normal;
  font-size: 26px;
  line-height: 1;
  color: var(--color-primary, #C0392B);
}
.brand-cn {
  font-size: 20px;
  font-weight: 700;
  color: var(--color-text-primary, #1F1D1A);
  font-family: "SourceHanSans-Bold", var(--font-family-cn);
  line-height: 1.2;
}
.divider {
  width: 2px;
  height: 20px;
  background-color: var(--color-border, #E5DFD3);
}
.page-title {
  font-size: 16px;
  color: var(--color-text-secondary, #6E6A60);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  line-height: 1.2;
}
.nav-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.nav-btn {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: none;
  background-color: var(--color-bg-white, #FFFFFF);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 0.25s ease, color 0.25s ease, transform 0.25s ease, box-shadow 0.25s ease;
}
.nav-btn:hover {
  background-color: var(--color-primary, #C0392B);
  color: #fff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px color-mix(in srgb, var(--color-primary, #C0392B) 30%, transparent);
}
.nav-btn .icon {
  font-family: "remixicon", sans-serif;
  font-style: normal;
  font-size: 16px;
  color: var(--color-text-secondary, #6E6A60);
}
.nav-btn:hover .icon {
  color: #fff;
}
</style>