/**
 * ui/ — 统一基础组件库（主题化、可独立扩展）
 *
 * UiTip        hover 提示气泡
 * UiInput      输入框（text / password / search / textarea）
 * UiRadioGroup 单选框组
 * UiSelect     下拉框（自定义面板）
 * UiSwitch     开关滑块
 * UiCheck      勾选选中图标（选中卡片徽标）
 * UiSelectCard 通用可选选中卡片（主题 / 模板 / 导出 / 个性）
 * UiTabs       标签切换（滑动指示器 + 内容过渡动画）
 * UiModal      统一弹框壳（对齐信用查询平台，仅 width 可调）
 */
export { default as UiTip } from './UiTip.vue';
export { default as UiInput } from './UiInput.vue';
export { default as UiRadioGroup } from './UiRadioGroup.vue';
export type { UiRadioOption } from './UiRadioGroup.vue';
export { default as UiSelect } from './UiSelect.vue';
export type { UiSelectOption } from './UiSelect.vue';
export { default as UiSwitch } from './UiSwitch.vue';
export { default as UiCheck } from './UiCheck.vue';
export { default as UiSelectCard } from './UiSelectCard.vue';
export { default as UiTabs } from './UiTabs.vue';
export type { UiTabItem } from './UiTabs.vue';
export { default as UiModal } from './UiModal.vue';
