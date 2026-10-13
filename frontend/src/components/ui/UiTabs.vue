<script setup lang="ts">
/**
 * UiTabs — 统一标签切换（滑动指示器 + 内容淡入滑动动画）
 * 内容用具名 slot `panel-<key>` 按当前 key 切换，自带过渡
 *
 * @example
 * <UiTabs v-model="tab" :tabs="[{ key: 'a', label: '基础' }, { key: 'b', label: '高级', badge: 3 }]">
 *   <template #panel-a>基础内容</template>
 *   <template #panel-b>高级内容</template>
 * </UiTabs>
 */
import { ref, watch, nextTick, onMounted, onUnmounted } from 'vue';

export interface UiTabItem {
  key: string;
  label: string;
  icon?: string;
  badge?: string | number;
  disabled?: boolean;
}

const props = withDefaults(defineProps<{
  modelValue?: string;
  tabs?: UiTabItem[];
  /** pill=胶囊底条 underline=下划线 */
  variant?: 'pill' | 'underline';
  size?: 'sm' | 'md';
}>(), {
  modelValue: '',
  tabs: () => [],
  variant: 'pill',
  size: 'md',
});

const emit = defineEmits<{
  (e: 'update:modelValue', v: string): void;
  (e: 'change', v: string): void;
}>();

const bar = ref<HTMLElement | null>(null);
const indicator = ref({ left: 0, width: 0, visible: false });
let ro: ResizeObserver | null = null;

function moveIndicator() {
  if (!bar.value) return;
  const active = bar.value.querySelector<HTMLElement>('.ui-tab.is-active');
  if (!active) { indicator.value.visible = false; return; }
  indicator.value = {
    left: active.offsetLeft,
    width: active.offsetWidth,
    visible: true,
  };
}

function pick(t: UiTabItem) {
  if (t.disabled || t.key === props.modelValue) return;
  emit('update:modelValue', t.key);
  emit('change', t.key);
}

watch(() => [props.modelValue, props.tabs], () => nextTick(moveIndicator), { flush: 'post' });

onMounted(() => {
  nextTick(moveIndicator);
  if (typeof ResizeObserver !== 'undefined' && bar.value) {
    ro = new ResizeObserver(moveIndicator);
    ro.observe(bar.value);
  }
  window.addEventListener('resize', moveIndicator);
});
onUnmounted(() => {
  ro?.disconnect();
  window.removeEventListener('resize', moveIndicator);
});
</script>

<template>
  <div class="ui-tabs" :class="['is-' + variant, 'is-' + size]">
    <div ref="bar" class="ui-tab-bar" role="tablist">
      <span
        v-show="indicator.visible"
        class="ui-tab-indicator"
        :style="{ left: indicator.left + 'px', width: indicator.width + 'px' }"
      />
      <button
        v-for="t in tabs"
        :key="t.key"
        type="button"
        role="tab"
        :aria-selected="modelValue === t.key"
        class="ui-tab"
        :class="{ 'is-active': modelValue === t.key, 'is-disabled': t.disabled }"
        :disabled="t.disabled"
        @click="pick(t)"
      >
        <span v-if="t.icon" :class="['icon', t.icon]"></span>
        <span class="ui-tab-label">{{ t.label }}</span>
        <span v-if="t.badge !== undefined" class="ui-tab-badge">{{ t.badge }}</span>
      </button>
    </div>
    <div class="ui-tab-panels">
      <transition name="ui-tab-fade" mode="out-in">
        <div :key="modelValue" class="ui-tab-panel">
          <slot :name="'panel-' + modelValue" />
        </div>
      </transition>
    </div>
  </div>
</template>

<style scoped>
.ui-tabs { display: flex; flex-direction: column; gap: 0; width: 100%; }
.ui-tab-bar {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px;
  border-radius: 12px;
  background-color: var(--color-bg-secondary, #F7F3EA);
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  align-self: flex-start;
  max-width: 100%;
  overflow-x: auto;
}
.is-underline .ui-tab-bar { background: transparent; border: none; border-radius: 0; padding: 0; gap: 4px; }
.ui-tab {
  position: relative;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  gap: 7px;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 9px;
  color: var(--color-text-muted, #8A8578);
  transition: color 0.2s;
  font-family: "SourceHanSans-Regular", var(--font-family-cn), sans-serif;
  white-space: nowrap;
}
.is-md .ui-tab { height: 38px; padding: 0 18px; font-size: 14.5px; }
.is-sm .ui-tab { height: 32px; padding: 0 14px; font-size: 13px; }
.ui-tab:hover:not(:disabled):not(.is-active) { color: var(--color-text-primary, #1F1D1A); }
.ui-tab.is-active { color: var(--color-primary, #C0392B); font-weight: 500; }
.ui-tab:disabled { opacity: 0.45; cursor: not-allowed; }
.ui-tab .icon { font-family: "remixicon", sans-serif; font-style: normal; font-size: 16px; }
.ui-tab-badge {
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 9999px;
  background-color: var(--color-primary, #C0392B);
  color: #fff;
  font-size: 11.5px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.ui-tab.is-active .ui-tab-badge { background-color: var(--color-primary, #C0392B); }
.ui-tab:not(.is-active) .ui-tab-badge { background-color: var(--color-text-faint, #A7A296); }

.ui-tab-indicator {
  position: absolute;
  top: 4px;
  bottom: 4px;
  border-radius: 9px;
  background-color: var(--color-bg-white, #fff);
  box-shadow: var(--shadow-sm, 0 2px 8px rgba(0, 0, 0, 0.08));
  transition: left 0.28s cubic-bezier(0.32, 0.72, 0.35, 1), width 0.28s cubic-bezier(0.32, 0.72, 0.35, 1);
}
.is-underline .ui-tab-indicator {
  top: auto;
  bottom: 0;
  height: 2.5px;
  border-radius: 2px;
  background-color: var(--color-primary, #C0392B);
  box-shadow: none;
}

.ui-tab-panels { padding-top: 16px; }
.ui-tab-fade-enter-active { transition: opacity 0.22s ease, transform 0.22s ease; }
.ui-tab-fade-leave-active { transition: opacity 0.14s ease, transform 0.14s ease; }
.ui-tab-fade-enter-from { opacity: 0; transform: translateY(8px); }
.ui-tab-fade-leave-to { opacity: 0; transform: translateY(-5px); }
.icon { font-family: "remixicon", sans-serif; font-style: normal; display: inline-flex; }
</style>
