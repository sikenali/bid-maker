<script setup lang="ts">
/**
 * UiInput — 统一输入框（text / password / search / textarea）
 * 视觉对齐现有 .form-input（44px / #FAF8F0 / 10px 圆角），焦点色统一走 --color-primary
 *
 * @example
 * <UiInput v-model="name" placeholder="请输入" clearable />
 * <UiInput v-model="pwd" type="password" />
 * <UiInput v-model="q" type="search" @search="onSearch" />
 * <UiInput v-model="desc" multiline :rows="3" />
 */
import { computed, useAttrs } from 'vue';
import { riToCompName } from '@/utils/icon';

const props = withDefaults(defineProps<{
  modelValue?: string | number;
  type?: 'text' | 'password' | 'search' | 'number';
  placeholder?: string;
  disabled?: boolean;
  readonly?: boolean;
  clearable?: boolean;
  multiline?: boolean;
  rows?: number;
  maxlength?: number;
  /** sm=36px md=44px lg=52px */
  size?: 'sm' | 'md' | 'lg';
  /** remix icon class，渲染在左侧，如 ri-search-line */
  prefixIcon?: string;
}>(), {
  modelValue: '',
  type: 'text',
  placeholder: '',
  disabled: false,
  readonly: false,
  clearable: false,
  multiline: false,
  rows: 3,
  size: 'md',
  prefixIcon: '',
});

const emit = defineEmits<{
  (e: 'update:modelValue', v: string | number): void;
  (e: 'search', v: string | number): void;
  (e: 'enter', ev: KeyboardEvent): void;
  (e: 'clear'): void;
  (e: 'focus', ev: FocusEvent): void;
  (e: 'blur', ev: FocusEvent): void;
}>();

const attrs = useAttrs();
const showClear = computed(() =>
  props.clearable && !props.disabled && !props.readonly && String(props.modelValue ?? '').length > 0,
);

function onInput(ev: Event) {
  const el = ev.target as HTMLInputElement | HTMLTextAreaElement;
  emit('update:modelValue', props.type === 'number' ? Number(el.value) : el.value);
}
function onKeydown(ev: KeyboardEvent) {
  if (ev.key === 'Enter') {
    emit('enter', ev);
    if (props.type === 'search') emit('search', props.modelValue ?? '');
  }
}
function clear() {
  emit('update:modelValue', '');
  emit('clear');
}
</script>

<template>
  <span
    class="ui-input"
    :class="['is-' + size, { 'has-prefix': prefixIcon || type === 'search', 'has-clear': showClear, 'has-suffix': $slots.suffix, 'is-disabled': disabled }]"
  >
    <component v-if="prefixIcon || type === 'search'" :is="riToCompName(prefixIcon || 'ri-search-line')" class="ui-input-prefix icon" />
    <textarea
      v-if="multiline"
      class="ui-input-el"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :readonly="readonly"
      :rows="rows"
      :maxlength="maxlength"
      v-bind="attrs"
      @input="onInput"
      @keydown="onKeydown"
      @focus="emit('focus', $event)"
      @blur="emit('blur', $event)"
    />
    <input
      v-else
      class="ui-input-el"
      :type="type === 'search' ? 'text' : type"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :readonly="readonly"
      :maxlength="maxlength"
      v-bind="attrs"
      @input="onInput"
      @keydown="onKeydown"
      @focus="emit('focus', $event)"
      @blur="emit('blur', $event)"
    />
    <button v-if="showClear" class="ui-input-clear" type="button" title="清空" @click.stop="clear">
      <RiCloseCircleFill class="icon" />
    </button>
    <span v-if="$slots.suffix" class="ui-input-suffix">
      <slot name="suffix" />
    </span>
  </span>
</template>

<style scoped>
.ui-input {
  position: relative;
  display: inline-flex;
  align-items: center;
  width: 100%;
}
.ui-input-el {
  flex: 1;
  width: 100%;
  min-width: 0;
  background-color: var(--color-bg-secondary, #F7F3EA);
  border: 0.7px solid var(--color-border, #E5DFD3);
  border-radius: 10px;
  font-size: 14px;
  color: var(--color-text-primary, #1F1D1A);
  outline: none;
  transition: border-color 0.2s, background-color 0.2s, box-shadow 0.2s;
  font-family: "SourceHanSans-Regular", var(--font-family-cn), sans-serif;
}
.is-sm .ui-input-el { height: 36px; padding: 0 12px; font-size: 13px; }
.is-md .ui-input-el { height: 42px; padding: 0 16px; }
.is-lg .ui-input-el { height: 52px; padding: 0 18px; font-size: 15px; }
textarea.ui-input-el { height: auto; padding: 12px 16px; resize: vertical; line-height: 1.6; }
.ui-input-el::placeholder { color: var(--color-text-faint, #A7A296); }
.ui-input-el:hover:not(:disabled) { border-color: var(--color-border-hover, #D9D2C2); }
.ui-input-el:focus {
  border-color: var(--color-primary, #C0392B);
  background-color: var(--color-bg-white, #fff);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-primary, #C0392B) 12%, transparent);
}
.ui-input-el:disabled {
  background-color: var(--color-bg-secondary, #F7F3EA);
  color: var(--color-text-muted, #8A8578);
  cursor: not-allowed;
}
.has-prefix .ui-input-el { padding-left: 38px; }
.is-sm.has-prefix .ui-input-el { padding-left: 32px; }
.ui-input-prefix {
  position: absolute;
  left: 13px;
  font-size: 16px;
  color: var(--color-text-faint, #A7A296);
  pointer-events: none;
}
.is-sm .ui-input-prefix { left: 10px; font-size: 14px; }
.has-clear .ui-input-el { padding-right: 36px; }
.has-suffix .ui-input-el { padding-right: 44px; }
.has-suffix.has-clear .ui-input-el { padding-right: 68px; }
.has-suffix.has-clear .ui-input-clear { right: 36px; }
.ui-input-suffix {
  position: absolute;
  right: 8px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.ui-input-clear {
  position: absolute;
  right: 8px;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  color: var(--color-text-faint, #A7A296);
}
.ui-input-clear:hover { color: var(--color-text-secondary, #6E6A60); }
.ui-input-clear .icon { font-family: "remixicon", sans-serif; font-style: normal; font-size: 16px; }
.icon { font-family: "remixicon", sans-serif; font-style: normal; display: inline-flex; }
</style>
