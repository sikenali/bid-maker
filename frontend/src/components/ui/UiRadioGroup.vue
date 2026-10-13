<script setup lang="ts">
/**
 * UiRadioGroup — 统一单选框组（radiobox）
 *
 * @example
 * <UiRadioGroup v-model="pay" :options="[
 *   { value: 'alipay', label: '支付宝', desc: '推荐' },
 *   { value: 'wechat', label: '微信支付', disabled: true },
 * ]" />
 * <UiRadioGroup v-model="v" :options="opts" direction="vertical" />
 */
export interface UiRadioOption {
  value: string | number;
  label: string;
  desc?: string;
  disabled?: boolean;
}

withDefaults(defineProps<{
  modelValue?: string | number;
  options?: UiRadioOption[];
  /** horizontal | vertical */
  direction?: 'horizontal' | 'vertical';
  disabled?: boolean;
  name?: string;
}>(), {
  modelValue: '',
  options: () => [],
  direction: 'horizontal',
  disabled: false,
  name: undefined,
});

const emit = defineEmits<{
  (e: 'update:modelValue', v: string | number): void;
  (e: 'change', v: string | number): void;
}>();

function pick(opt: UiRadioOption, disabled?: boolean) {
  if (opt.disabled || disabled) return;
  emit('update:modelValue', opt.value);
  emit('change', opt.value);
}
</script>

<template>
  <div class="ui-radio-group" :class="['is-' + direction]" role="radiogroup">
    <div
      v-for="opt in options"
      :key="String(opt.value)"
      class="ui-radio"
      :class="{ 'is-checked': modelValue === opt.value, 'is-disabled': disabled || opt.disabled }"
      role="radio"
      :aria-checked="modelValue === opt.value"
      :tabindex="disabled || opt.disabled ? -1 : 0"
      @click="pick(opt, disabled)"
      @keydown.enter.space.prevent="pick(opt, disabled)"
    >
      <span class="ui-radio-circle">
        <span class="ui-radio-dot"></span>
      </span>
      <span class="ui-radio-text">
        <span class="ui-radio-label">{{ opt.label }}</span>
        <span v-if="opt.desc" class="ui-radio-desc">{{ opt.desc }}</span>
      </span>
    </div>
  </div>
</template>

<style scoped>
.ui-radio-group { display: flex; gap: 8px 24px; flex-wrap: wrap; }
.ui-radio-group.is-vertical { flex-direction: column; align-items: flex-start; gap: 12px; }
.ui-radio {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  cursor: pointer;
  padding: 6px 8px;
  margin: -6px -8px;
  border-radius: 8px;
  transition: background-color 0.15s;
  outline: none;
}
.ui-radio:hover:not(.is-disabled) { background-color: color-mix(in srgb, var(--color-primary, #C0392B) 7%, transparent); }
.ui-radio:focus-visible { box-shadow: 0 0 0 2px color-mix(in srgb, var(--color-primary, #C0392B) 35%, transparent); }
.ui-radio-circle {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 1.5px solid var(--color-border, #E5DFD3);
  background-color: var(--color-bg-white, #fff);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: border-color 0.18s;
}
.ui-radio-dot {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background-color: var(--color-primary, #C0392B);
  transform: scale(0);
  transition: transform 0.18s cubic-bezier(0.34, 1.4, 0.64, 1);
}
.ui-radio.is-checked .ui-radio-circle { border-color: var(--color-primary, #C0392B); }
.ui-radio.is-checked .ui-radio-dot { transform: scale(1); }
.ui-radio-text { display: flex; align-items: baseline; gap: 8px; }
.ui-radio-label {
  font-size: 14.5px;
  color: var(--color-text-primary, #1F1D1A);
  font-family: "SourceHanSans-Regular", var(--font-family-cn), sans-serif;
}
.ui-radio.is-checked .ui-radio-label { font-weight: 500; }
.ui-radio-desc { font-size: 12.5px; color: var(--color-text-muted, #8A8578); }
.ui-radio.is-disabled { opacity: 0.5; cursor: not-allowed; }
</style>
