<script setup lang="ts">
/**
 * UiCheck — 统一勾选选中图标（主题 / 模板 / 导出等选中卡片右上角徽标）
 * 选中时缩放弹出，未选中可完全隐藏（ghost 模式只占位不渲染，cover 模式渲染空心圈）
 *
 * @example
 * <UiCheck :checked="selected === item.key" />
 * <UiCheck :checked="on" mode="cover" :size="22" />
 */
withDefaults(defineProps<{
  checked?: boolean;
  /** ghost=未选中不渲染 cover=未选中显示空心圈 */
  mode?: 'ghost' | 'cover';
  size?: number;
}>(), {
  checked: false,
  mode: 'ghost',
  size: 24,
});
</script>

<template>
  <transition name="ui-check-pop">
    <span
      v-if="checked || mode === 'cover'"
      class="ui-check"
      :class="{ 'is-checked': checked }"
      :style="{ width: size + 'px', height: size + 'px' }"
    >
      <svg
        v-if="checked"
        :width="Math.round(size * 0.58)"
        :height="Math.round(size * 0.58)"
        viewBox="0 0 14 14"
        fill="none"
      >
        <path d="M2.5 7L5.5 10L11.5 3.5" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
    </span>
  </transition>
</template>

<style scoped>
.ui-check {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  flex-shrink: 0;
  border: 1.5px solid var(--color-border, #E5DFD3);
  background-color: var(--color-bg-white, #fff);
  color: transparent;
  transition: background-color 0.18s, border-color 0.18s, color 0.18s;
}
.ui-check.is-checked {
  background-color: var(--color-primary, #C0392B);
  border-color: var(--color-primary, #C0392B);
  color: #fff;
  box-shadow: 0 2px 8px color-mix(in srgb, var(--color-primary, #C0392B) 40%, transparent);
}
.ui-check-pop-enter-active { transition: transform 0.22s cubic-bezier(0.34, 1.5, 0.64, 1), opacity 0.15s; }
.ui-check-pop-leave-active { transition: transform 0.14s ease-in, opacity 0.14s; }
.ui-check-pop-enter-from, .ui-check-pop-leave-to { transform: scale(0.4); opacity: 0; }
</style>
