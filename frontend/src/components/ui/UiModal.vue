<script setup lang="ts">
/**
 * UiModal — 统一弹框壳（对齐信用查询平台弹框，仅 width 可调）
 * 头部：48px 图标 + 标题 + 副标题 + 圆形关闭钮；底部标准取消/确定按钮（可隐藏改用 footer slot）
 *
 * @example
 * <UiModal v-model="show" title="新增模板" subtitle="创建后可在列表中启用" icon="ri-add-line" :width="560"
 *   confirm-text="确定" :confirm-disabled="!name" @confirm="handleAdd">
 *   <UiInput v-model="name" placeholder="请输入" />
 * </UiModal>
 */
import { watch, onMounted, onUnmounted } from 'vue';
import { riToCompName } from '@/utils/icon';

const props = withDefaults(defineProps<{
  modelValue?: boolean;
  title?: string;
  subtitle?: string;
  /** remix icon class，如 ri-add-line */
  icon?: string;
  /** 面板宽度（px 数字或 css 字符串），默认 560 */
  width?: number | string;
  /** 点击遮罩关闭 */
  closeOnClickOverlay?: boolean;
  /** ESC 关闭 */
  closeOnEsc?: boolean;
  /** 显示右上关闭钮 */
  showClose?: boolean;
  /** 显示底部标准按钮组 */
  showFooter?: boolean;
  cancelText?: string;
  confirmText?: string;
  confirmDisabled?: boolean;
  confirmLoading?: boolean;
}>(), {
  modelValue: false,
  title: '',
  subtitle: '',
  icon: 'ri-file-edit-line',
  width: 560,
  closeOnClickOverlay: true,
  closeOnEsc: true,
  showClose: true,
  showFooter: true,
  cancelText: '取消',
  confirmText: '确定',
  confirmDisabled: false,
  confirmLoading: false,
});

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void;
  (e: 'close'): void;
  (e: 'confirm'): void;
  (e: 'cancel'): void;
}>();

function close() {
  emit('update:modelValue', false);
  emit('close');
}
function onCancel() {
  emit('cancel');
  close();
}
function onConfirm() {
  if (props.confirmDisabled || props.confirmLoading) return;
  emit('confirm');
}
// eslint-disable-next-line @typescript-eslint/no-unused-vars
function _onKey(ev: KeyboardEvent) {
  if (ev.key === 'Escape' && props.modelValue && props.closeOnEsc) close();
}
// body 滚动锁定（多弹框嵌套时以计数方式处理）
let lockCount = 0;
function lockScroll(lock: boolean) {
  lockCount = Math.max(0, lockCount + (lock ? 1 : -1));
  document.body.style.overflow = lockCount > 0 ? 'hidden' : '';
}
watch(() => props.modelValue, v => lockScroll(v), { immediate: false });
onMounted(() => { if (props.modelValue) lockScroll(true); });
onUnmounted(() => { if (props.modelValue) lockScroll(false); });
</script>

<template>
  <teleport to="body">
    <transition name="ui-modal">
      <div v-if="modelValue" class="ui-modal-overlay" @click.self="closeOnClickOverlay && close()">
        <div
          class="ui-modal-panel"
          :style="{ width: typeof width === 'number' ? width + 'px' : width }"
          role="dialog"
          aria-modal="true"
        >
          <div class="ui-modal-header">
            <div class="ui-modal-header-left">
              <div class="ui-modal-icon">
                <component :is="riToCompName(icon)" />
              </div>
              <div class="ui-modal-header-text">
                <h2 class="ui-modal-title">{{ title }}</h2>
                <p v-if="subtitle || $slots.subtitle" class="ui-modal-subtitle">
                  <slot name="subtitle">{{ subtitle }}</slot>
                </p>
              </div>
            </div>
            <button v-if="showClose" class="ui-modal-close" @click="close" aria-label="关闭">
              <RiCloseLine />
            </button>
          </div>

          <div class="ui-modal-body">
            <slot />
          </div>

          <div v-if="showFooter || $slots.footer" class="ui-modal-footer">
            <slot name="footer">
              <span class="ui-modal-footer-hint">
                <slot name="footer-hint" />
              </span>
              <span class="ui-modal-footer-btns">
                <button class="ui-modal-cancel" @click="onCancel">{{ cancelText }}</button>
                <button
                  class="ui-modal-confirm"
                  :disabled="confirmDisabled || confirmLoading"
                  @click="onConfirm"
                >
                  <RiLoader4Line v-if="confirmLoading" class="icon spinning" />
                  {{ confirmText }}
                </button>
              </span>
            </slot>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<style scoped>
.ui-modal-overlay {
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
  padding: 24px;
}
.ui-modal-panel {
  max-width: calc(100vw - 48px);
  max-height: 90vh;
  background: var(--color-bg-white, #FFFFFF);
  border-radius: 16px;
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.1);
  padding: 32px 36px;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}
.ui-modal-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 22px;
}
.ui-modal-header-left {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  min-width: 0;
}
.ui-modal-icon {
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
.ui-modal-icon i { font-style: normal; }
.ui-modal-header-text { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.ui-modal-title {
  font-size: 22px;
  font-weight: 700;
  color: var(--color-text-primary, #1F1D1A);
  margin: 0;
  line-height: 28px;
  font-family: "SourceHanSans-Bold", var(--font-family-cn), sans-serif;
}
.ui-modal-subtitle {
  font-size: 13px;
  color: var(--color-text-muted, #8A8578);
  margin: 0;
  line-height: 18px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn), sans-serif;
}
.ui-modal-close {
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
  transition: background-color 0.2s, transform 0.15s;
}
.ui-modal-close:hover { background: var(--color-border-soft, #EDE9E0); }
.ui-modal-close:active { transform: scale(0.94); }
.ui-modal-close i { font-style: normal; }
.ui-modal-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 0;
}
.ui-modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-top: 20px;
  margin-top: 4px;
  border-top: 1px solid var(--color-border-soft, #EDE9E0);
}
.ui-modal-footer-hint:empty { display: none; }
.ui-modal-footer-btns {
  display: flex;
  gap: 12px;
  margin-left: auto;
}
.ui-modal-cancel {
  height: 42px;
  padding: 0 24px;
  border: 0.7px solid var(--color-border, #E5DFD3);
  border-radius: 10px;
  background-color: var(--color-bg-card, #FCFBF8);
  color: var(--color-text-secondary, #6E6A60);
  font-size: 15px;
  cursor: pointer;
  transition: background-color 0.2s;
  font-family: "SourceHanSans-Regular", var(--font-family-cn), sans-serif;
}
.ui-modal-cancel:hover { background-color: var(--color-bg-secondary, #F7F3EA); }
.ui-modal-confirm {
  height: 42px;
  padding: 0 28px;
  border: none;
  border-radius: 10px;
  background-color: var(--color-primary, #C0392B);
  color: #fff;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  transition: opacity 0.2s, transform 0.15s;
  font-family: "SourceHanSans-Medium", var(--font-family-cn), sans-serif;
}
.ui-modal-confirm:hover:not(:disabled) { opacity: 0.92; }
.ui-modal-confirm:disabled { opacity: 0.5; cursor: not-allowed; }
.ui-modal-confirm .icon { font-style: normal; }
.spinning { animation: ui-modal-spin 1s linear infinite; display: inline-flex; }
@keyframes ui-modal-spin { to { transform: rotate(360deg); } }

.ui-modal-enter-active, .ui-modal-leave-active { transition: opacity 0.25s ease; }
.ui-modal-enter-active .ui-modal-panel, .ui-modal-leave-active .ui-modal-panel {
  transition: opacity 0.25s ease, transform 0.25s cubic-bezier(0.32, 0.72, 0.35, 1);
}
.ui-modal-enter-from, .ui-modal-leave-to { opacity: 0; }
.ui-modal-enter-from .ui-modal-panel, .ui-modal-leave-to .ui-modal-panel {
  opacity: 0;
  transform: translateY(14px) scale(0.98);
}
</style>
