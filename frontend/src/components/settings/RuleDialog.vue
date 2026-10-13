<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { UiModal, UiInput, UiRadioGroup } from '@/components/ui';

const RULE_GROUPS = [
  { value: 'info', label: '项目信息' },
  { value: 'business', label: '商务条款' },
  { value: 'tech', label: '技术条款' },
  { value: 'score', label: '评分标准' },
  { value: 'seal', label: '封标信息' },
  { value: 'star', label: '标星条款' },
];

const FIELD_HINTS: Record<string, Record<string, string>> = {
  info: { regex: '如：项目名称、项目编号、投标截止时间、开标地点、预算金额', keyword: '如：采购人、代理机构、资质要求' },
  business: { regex: '如：付款方式、质保期、履约保证金、违约责任', keyword: '如：废标、否决投标、实质性偏离' },
  tech: { regex: '如：技术参数、性能指标、验收标准、配置清单', keyword: '如：技术要求、售后服务、培训方案' },
  score: { regex: '如：价格评分、技术评分、商务评分、满分值', keyword: '如：得分、评分细则、扣分项' },
  seal: { regex: '如：正本份数、副本份数、密封要求、电子标书', keyword: '如：密封、盖章、签字、封装' },
  star: { regex: '如：★关键条款、强制性要求、一票否决项', keyword: '如：★、必须、强制、禁止' },
};

const props = defineProps<{
  visible: boolean;
  mode: 'regex' | 'keyword';
  editRule?: { id: string; fieldName: string; pattern: string; groupName?: string; aliasFor?: string } | null;
}>();

const emit = defineEmits<{
  close: []
  submit: [data: { fieldName: string; pattern: string; category: string; groupName: string; aliasFor: string }]
}>();

const dialogVisible = ref(false);
const fieldName = ref('');
const pattern = ref('');
const groupName = ref('info');
const aliasFor = ref('');
const showAlias = ref(false);

const dialogTitle = computed(() => {
  const action = props.editRule ? '编辑' : '添加';
  return props.mode === 'regex' ? `${action}规则` : `${action}关键字`;
});

watch(() => props.visible, (val) => {
  dialogVisible.value = val;
  if (val) {
    if (props.editRule) {
      fieldName.value = props.editRule.fieldName;
      pattern.value = props.editRule.pattern;
      groupName.value = props.editRule.groupName || 'info';
      aliasFor.value = props.editRule.aliasFor || '';
      showAlias.value = !!props.editRule.aliasFor;
    } else {
      fieldName.value = '';
      pattern.value = '';
      groupName.value = 'info';
      aliasFor.value = '';
      showAlias.value = false;
    }
  }
});

watch(dialogVisible, (val) => {
  if (!val) emit('close');
});

watch([groupName, () => props.mode], () => {
  if (!props.editRule) pattern.value = '';
});

function handleSubmit() {
  if (!fieldName.value.trim() && !pattern.value.trim()) return;
  emit('submit', {
    fieldName: fieldName.value.trim() || '未知字段',
    pattern: pattern.value.trim(),
    category: props.mode,
    groupName: groupName.value,
    aliasFor: aliasFor.value.trim(),
  });
  dialogVisible.value = false;
}
</script>

<template>
  <UiModal
    v-model="dialogVisible"
    :title="dialogTitle"
    :subtitle="props.mode === 'regex' ? '用正则匹配招标文件中的目标字段' : '用关键字标记需要关注的条款'"
    :icon="props.editRule ? 'ri-edit-line' : 'ri-add-line'"
    :width="560"
    :confirm-disabled="!fieldName.trim()"
    @confirm="handleSubmit"
  >
    <div class="form-group">
      <label class="form-label">所属分组</label>
      <UiRadioGroup v-model="groupName" :options="RULE_GROUPS" />
    </div>
    <div class="form-group">
      <label class="form-label">字段名称 <span class="required">*</span></label>
      <UiInput
        v-model="fieldName"
        :placeholder="FIELD_HINTS[groupName]?.[props.mode] || '请输入字段名称'"
        clearable
        @enter="handleSubmit"
      />
      <p class="form-hint">支持通配符 *：如 <code>*编号</code> 可命中 招标编号 / 项目编号 / 标识编号，结果字段名显示实际命中的名称</p>
    </div>
    <div class="form-group">
      <label class="form-label">
        <input type="checkbox" v-model="showAlias" />
        合并到字段（别名归因）
      </label>
      <UiInput
        v-if="showAlias"
        v-model="aliasFor"
        placeholder="合并到已有字段，如：项目名称，留空则独立成行"
        clearable
        @enter="handleSubmit"
      />
    </div>
    <div v-if="props.mode === 'regex'" class="form-group">
      <label class="form-label">正则表达式</label>
      <UiInput
        v-model="pattern"
        placeholder="输入正则表达式（可为空，自动补全）"
        clearable
        @enter="handleSubmit"
      />
    </div>
  </UiModal>
</template>

<style scoped>
.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.form-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-primary, #1F1D1A);
  font-family: "SourceHanSans-Regular", var(--font-family-cn), sans-serif;
  display: flex;
  align-items: center;
  gap: 6px;
}
.required {
  color: var(--color-primary, #C0392B);
}
.form-hint {
  font-weight: 400;
  font-size: 12px;
  color: var(--color-text-muted, #8A8578);
}
.form-hint code {
  background: rgba(245,239,227,0.5);
  padding: 1px 4px;
  border-radius: 4px;
  font-size: 11px;
}
.form-group .ui-input { width: 100%; }
</style>

