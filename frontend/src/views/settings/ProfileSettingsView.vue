<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue';
import { ElMessage } from '@/utils/toast';
import { UiSwitch, UiInput, UiModal, UiSelect } from '@/components/ui';
import RuleDialog from '@/components/settings/RuleDialog.vue';
import apiClient from '@/api/client';
import { getLlmSkills, createLlmSkill, updateLlmSkill, deleteLlmSkill, getLlmPrompts } from '@/api/llm';
import { getRules, addRule, updateRule, deleteRule, seedUniversalRules, type ExtractionRule } from '@/api/rules';
import { riToCompName } from '@/utils/icon';

// ── 主标签切换 ──────────────────────────────────────────────────
const mainTab = ref<'llm' | 'prompts' | 'rules'>('llm');

const TABS = [
  { key: 'llm' as const,   label: 'LLM 提取',    icon: 'ri-brain-line' },
  { key: 'prompts' as const, label: '提示词模板',  icon: 'ri-chat-3-line' },
  { key: 'rules' as const,  label: '规则管理',    icon: 'ri-list-settings-line' },
];

// ── Tab 2: LLM 提取 ─────────────────────────────────────────────
interface LlmSkill { id: string; skill_name: string; skill_code: string; biz_type: string; prompt_template: string; temperature: number; top_p: number; max_tokens: number; enable: number; enable_llm: number; bind_prompt_template_id: string; version: number; create_time: string; }
const llmSkills = ref<LlmSkill[]>([]);
const llmLoading = ref(false);
const llmSearch = ref('');
const llmShowCreate = ref(false);
const llmShowEdit = ref(false);
const editingSkill = ref<LlmSkill | null>(null);
const eName = ref(''), eCode = ref(''), eBiz = ref('info'), ePrompt = ref('');
const eTemp = ref(0.1), eTopP = ref(0.9), eMaxTok = ref(2048);
const eEnableLlm = ref(false);
const eBindPrompt = ref('');
const GROUP_LABELS: Record<string, string> = { info:'项目信息', business:'商务条款', tech:'技术条款', score:'评分标准', seal:'封标信息', star:'标星条款' };
const BIZ_OPTIONS = Object.entries(GROUP_LABELS).map(([value, label]) => ({ value, label }));
const PROMPT_OPTIONS = computed(() => [
  { value: '', label: '不绑定（使用技能自带提示词）' },
  ...templates.value.filter(t => t.enable === 1).map(t => ({ value: t.id, label: `${t.name}（${GROUP_LABELS[t.biz_type] || t.biz_type}）` })),
]);

async function loadLlmSkills() {
  llmLoading.value = true;
  try { llmSkills.value = await getLlmSkills(); } catch { ElMessage.error('加载技能失败'); }
  finally { llmLoading.value = false; }
}
function openLlmCreate() { editingSkill.value=null; eName.value=''; eCode.value=''; eBiz.value='info'; ePrompt.value=''; eTemp.value=0.1; eTopP.value=0.9; eMaxTok.value=2048; eEnableLlm.value=false; eBindPrompt.value=''; llmShowCreate.value=true; }
function openLlmEdit(s: LlmSkill) { editingSkill.value=s; eName.value=s.skill_name; eCode.value=s.skill_code; eBiz.value=s.biz_type; ePrompt.value=s.prompt_template; eTemp.value=s.temperature; eTopP.value=s.top_p; eMaxTok.value=s.max_tokens; eEnableLlm.value=!!s.enable_llm; eBindPrompt.value=s.bind_prompt_template_id||''; llmShowEdit.value=true; }
async function saveLlmCreate() {
  if (!eName.value.trim()||!eCode.value.trim()||!ePrompt.value.trim()) { ElMessage.warning('请填写名称、代码和提示词'); return; }
  try { await createLlmSkill({skill_name:eName.value.trim(),skill_code:eCode.value.trim(),biz_type:eBiz.value,prompt_template:ePrompt.value,temperature:eTemp.value,top_p:eTopP.value,max_tokens:eMaxTok.value,enable_llm:eEnableLlm.value?1:0,bind_prompt_template_id:eBindPrompt.value}); ElMessage.success('技能已创建'); llmShowCreate.value=false; await loadLlmSkills(); } catch { ElMessage.error('创建失败'); }
}
async function saveLlmEdit() {
  if (!editingSkill.value||!eName.value.trim()||!ePrompt.value.trim()) { ElMessage.warning('请填写名称和提示词'); return; }
  try { await updateLlmSkill(editingSkill.value.id,{skill_name:eName.value.trim(),prompt_template:ePrompt.value,biz_type:eBiz.value,temperature:eTemp.value,top_p:eTopP.value,max_tokens:eMaxTok.value,enable_llm:eEnableLlm.value?1:0,bind_prompt_template_id:eBindPrompt.value}); ElMessage.success('已更新'); llmShowEdit.value=false; await loadLlmSkills(); } catch { ElMessage.error('保存失败'); }
}
async function toggleLlmSkill(s: LlmSkill) { s.enable = s.enable===0?1:0; await updateLlmSkill(s.id,{enable:s.enable}); ElMessage.success(s.enable===1?'已启用':'已禁用'); }
async function toggleLlmFeature(s: LlmSkill) { const next = s.enable_llm===1?0:1; s.enable_llm = next; try { await updateLlmSkill(s.id,{enable_llm:next}); ElMessage.success(next===1?'LLM增强已启用':'LLM增强已禁用'); } catch { s.enable_llm = next===1?0:1; ElMessage.error('操作失败'); } }
const showDelSkillDialog = ref(false);
const deletingSkill = ref<LlmSkill | null>(null);
function openDelSkill(s: LlmSkill) { deletingSkill.value = s; showDelSkillDialog.value = true; }
async function confirmDelSkill() {
  if (!deletingSkill.value) return;
  try {
    await deleteLlmSkill(deletingSkill.value.id);
    showDelSkillDialog.value = false;
    ElMessage.success('已删除');
    await loadLlmSkills();
  } catch { ElMessage.error('删除失败'); }
}

// ── Tab 3: 提示词模板 ───────────────────────────────────────────
interface PromptTpl { id: string; name: string; description: string; biz_type: string; template_text: string; enable: number; version: number; }
const templates = ref<PromptTpl[]>([]);
const tplLoading = ref(false);
const tplSearch = ref('');
const tplShowEdit = ref(false);
const editingTpl = ref<PromptTpl | null>(null);
const tName = ref(''), tDesc = ref(''), tBiz = ref('info'), tText = ref('');

async function loadTemplates() {
  tplLoading.value = true;
  try { templates.value = await getLlmPrompts(); } catch { ElMessage.error('加载模板失败'); }
  finally { tplLoading.value = false; }
}
function openTplCreate() { editingTpl.value=null; tName.value=''; tDesc.value=''; tBiz.value='info'; tText.value=''; tplShowEdit.value=true; }
function openTplEdit(t: PromptTpl) { editingTpl.value=t; tName.value=t.name; tDesc.value=t.description; tBiz.value=t.biz_type; tText.value=t.template_text; tplShowEdit.value=true; }
async function saveTpl() {
  if (!tName.value.trim()||!tText.value.trim()) { ElMessage.warning('请填写名称和内容'); return; }
  try {
    if (editingTpl.value) { await apiClient.put(`/llm-prompts/${editingTpl.value.id}`,{name:tName.value.trim(),description:tDesc.value.trim(),template_text:tText.value,biz_type:tBiz.value}); }
    else { await apiClient.post('/llm-prompts',{name:tName.value.trim(),description:tDesc.value.trim(),template_text:tText.value,biz_type:tBiz.value}); }
    ElMessage.success('已保存'); tplShowEdit.value=false; await loadTemplates();
  } catch { ElMessage.error('保存失败'); }
}
async function toggleTpl(t: PromptTpl) { t.enable = t.enable===1?0:1; await apiClient.put(`/llm-prompts/${t.id}`,{enable:t.enable}); }
const showDelTplDialog = ref(false);
const deletingTpl = ref<PromptTpl | null>(null);
function openDelTpl(t: PromptTpl) { deletingTpl.value = t; showDelTplDialog.value = true; }
async function confirmDelTpl() {
  if (!deletingTpl.value) return;
  try {
    await apiClient.delete(`/llm-prompts/${deletingTpl.value.id}`);
    showDelTplDialog.value = false;
    ElMessage.success('已删除');
    await loadTemplates();
  } catch { ElMessage.error('删除失败'); }
}

watch(mainTab, (tab) => {
  if (tab === 'llm') loadLlmSkills();
  if (tab === 'prompts') loadTemplates();
  if (tab === 'rules') loadRules();
});

onMounted(() => { loadLlmSkills(); loadTemplates(); });

// ── Tab 3: 规则管理 ─────────────────────────────────────────────
interface RuleItem { id: string; fieldName: string; pattern: string; groupName: string; aliasFor: string; category: string; }
const rules = ref<RuleItem[]>([]);
const rLoading = ref(false);
const rSearch = ref('');
const rGroup = ref('info');
const showRuleDialog = ref(false);
const rDialogMode = ref<'regex' | 'keyword'>('regex');
const editingRule = ref<{ id: string; fieldName: string; pattern: string; groupName: string; aliasFor: string } | null>(null);
const GROUP_META = [
  { key: 'info', label: '项目信息', icon: 'ri-file-list-3-line' },
  { key: 'business', label: '商务条款', icon: 'ri-file-paper-2-line' },
  { key: 'tech', label: '技术条款', icon: 'ri-search-eye-line' },
  { key: 'score', label: '评分标准', icon: 'ri-layout-line' },
  { key: 'seal', label: '封标信息', icon: 'ri-mail-send-line' },
  { key: 'star', label: '标星条款', icon: 'ri-star-s-line' },
];

function toItem(r: ExtractionRule): RuleItem { return { id: r.id, fieldName: r.fieldName, pattern: r.pattern, groupName: r.groupName, aliasFor: r.aliasFor || '', category: r.category }; }
const filteredRules = ref<RuleItem[]>([]);
watch([rules, rGroup, rSearch], () => {
  const q = rSearch.value.trim().toLowerCase();
  filteredRules.value = rules.value.filter(r =>
    (!rGroup.value || r.groupName === rGroup.value) &&
    (!q || r.fieldName.toLowerCase().includes(q) || (r.pattern || '').toLowerCase().includes(q))
  );
});

async function loadRules() {
  rLoading.value = true;
  try { rules.value = (await getRules()).map(toItem); } catch { ElMessage.error('加载规则失败'); }
  finally { rLoading.value = false; }
}
function openAdd(mode: 'regex' | 'keyword') { rDialogMode.value = mode; editingRule.value = null; showRuleDialog.value = true; }
function openEdit(r: RuleItem) { rDialogMode.value = 'regex'; editingRule.value = { id: r.id, fieldName: r.fieldName, pattern: r.pattern, groupName: r.groupName, aliasFor: r.aliasFor }; showRuleDialog.value = true; }
async function saveRule(data: { fieldName: string; pattern: string; category: string; groupName: string; aliasFor?: string }) {
  if (!data || !data.fieldName.trim()) { ElMessage.warning('请填写字段名称'); return; }
  try {
    if (editingRule.value?.id) { await updateRule(editingRule.value.id, { fieldName: data.fieldName, pattern: data.pattern, category: data.category, groupName: data.groupName, aliasFor: data.aliasFor }); }
    else { await addRule({ fieldName: data.fieldName, pattern: data.pattern, category: data.category || rDialogMode.value, groupName: data.groupName, aliasFor: data.aliasFor, enabled: true }); }
    ElMessage.success('已保存'); showRuleDialog.value = false; await loadRules();
  } catch { ElMessage.error('保存失败'); }
}
const showDelRuleDialog = ref(false);
const deletingRuleId = ref('');
function openDelRule(id: string) { deletingRuleId.value = id; showDelRuleDialog.value = true; }
async function confirmDelRule() {
  if (!deletingRuleId.value) return;
  try {
    await deleteRule(deletingRuleId.value);
    showDelRuleDialog.value = false;
    ElMessage.success('已删除');
    await loadRules();
  } catch { ElMessage.error('删除失败'); }
}
async function handleSeedUniversal() {
  const res = await seedUniversalRules();
  await loadRules();
  ElMessage.success(`已补全 ${res.added} 条通用规则，当前共 ${res.total} 条`);
}
</script>

<template>
  <div class="profile-view">
    <!-- 标签导航 -->
    <div class="tab-nav">
      <button v-for="t in TABS" :key="t.key" class="tab-btn" :class="{active: mainTab===t.key}" @click="mainTab=t.key">
        <component :is="riToCompName(t.icon)" class="icon" /><span>{{ t.label }}</span>
      </button>
    </div>

    <!-- ═══════════════════════════════════════════════════════════ -->
    <!-- Tab 1: LLM 提取                                              -->
    <!-- ═══════════════════════════════════════════════════════════ -->
    <template v-if="mainTab==='llm'">
      <div class="toolbar">
        <button class="btn-primary" @click="openLlmCreate"><RiAddLine class="icon" />提取技能</button>
        <UiInput v-model="llmSearch" type="search" class="search-input" placeholder="搜索技能名称" clearable />
      </div>
      <div class="content-area">
        <div v-if="llmLoading" class="loading">正在加载…</div>
        <div v-else-if="llmSkills.length===0" class="empty">暂无LLM技能，点击「提取技能」创建</div>
        <div v-else class="skill-grid">
          <div v-for="s in llmSkills.filter(sk=>!llmSearch||sk.skill_name.toLowerCase().includes(llmSearch.toLowerCase()))" :key="s.id" class="skill-card" :class="{disabled:s.enable===0}">
            <div class="card-top">
              <div class="card-title">
                <span class="name">{{ s.skill_name }}</span>
                <span class="code">{{ s.skill_code }}</span>
                <span class="biz-tag">{{ GROUP_LABELS[s.biz_type]||s.biz_type }}</span>
              </div>
              <UiSwitch :model-value="s.enable===1" size="sm" @update:model-value="toggleLlmSkill(s)" />
            </div>
            <div class="card-prompt-preview"><pre>{{ s.prompt_template }}</pre></div>
            <div class="card-meta">
              <span class="meta clickable" :class="{on:s.enable_llm===1}" @click="toggleLlmFeature(s)" :title="s.enable_llm===1?'点击关闭LLM增强':'点击开启LLM增强'"><RiBrainLine />LLM增强</span>
              <span v-if="s.bind_prompt_template_id" class="meta"><RiFileTextLine />已绑定模板</span>
              <span class="meta"><RiSettings3Line />t={{s.temperature}}</span>
              <span class="meta"><RiHistoryLine />v{{s.version}}</span>
            </div>
            <div class="card-actions">
              <button class="act-btn edit" @click="openLlmEdit(s)">编辑</button>
              <button class="act-btn del" @click="openDelSkill(s)">删除</button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- ═══════════════════════════════════════════════════════════ -->
    <!-- Tab 3: 提示词模板                                              -->
    <!-- ═══════════════════════════════════════════════════════════ -->
    <template v-else-if="mainTab==='prompts'">
      <div class="toolbar">
        <button class="btn-primary" @click="openTplCreate"><RiAddLine class="icon" />新建提示词</button>
        <UiInput v-model="tplSearch" type="search" class="search-input" placeholder="搜索模板名称" clearable />
      </div>
      <div class="content-area">
        <div v-if="tplLoading" class="loading">正在加载…</div>
        <div v-else-if="templates.length===0" class="empty">暂无提示词模板</div>
        <div v-else class="tpl-grid">
          <div v-for="t in templates.filter(tp=>!tplSearch||tp.name.toLowerCase().includes(tplSearch.toLowerCase()))" :key="t.id" class="tpl-card" :class="{disabled:t.enable===0}">
            <div class="tpl-top">
              <div><div class="tpl-name">{{ t.name }}</div><div class="tpl-biz">{{ GROUP_LABELS[t.biz_type]||t.biz_type }} · v{{t.version}}</div></div>
            </div>
            <div class="tpl-desc">{{ t.description }}</div>
            <div class="tpl-preview"><pre>{{ t.template_text }}</pre></div>
            <div class="tpl-actions">
              <UiSwitch :model-value="t.enable===1" size="sm" @update:model-value="toggleTpl(t)" />
              <button class="act-btn edit" @click="openTplEdit(t)">编辑</button>
              <button class="act-btn del" @click="openDelTpl(t)">删除</button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- ═══════════════════════════════════════════════════════════ -->
    <!-- Tab 3: 规则管理                                              -->
    <!-- ═══════════════════════════════════════════════════════════ -->
    <template v-else>
      <div class="rule-toolbar">
        <div class="rule-toolbar-left">
          <button class="add-rule-button" @click="openAdd('regex')"><RiAddLine class="icon" /><span>新增规则</span></button>
          <button class="seed-universal-button" @click="handleSeedUniversal"><RiGlobalLine class="icon" /><span>补全通用规则</span></button>
        </div>
        <div class="rule-toolbar-right">
          <UiInput v-model="rSearch" type="search" class="toolbar-search" placeholder="搜索规则字段或描述" clearable />
          <div class="group-filter">
            <button v-for="g in GROUP_META" :key="g.key" class="group-icon-btn" :class="{ active: rGroup === g.key }" @click="rGroup = rGroup === g.key ? '' : g.key" :title="g.label">
              <component :is="riToCompName(g.icon)" class="icon" />
            </button>
          </div>
        </div>
      </div>
      <div class="rule-table-card">
        <table class="rule-table">
          <thead>
            <tr>
              <th>规则描述</th>
              <th>分组</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rule in filteredRules" :key="rule.id">
              <td class="rule-desc">
                <span class="rule-field-name">{{ rule.fieldName }}</span>
                <span v-if="rule.id.startsWith('univ_')" class="universal-badge">通用</span>
                <span v-if="rule.aliasFor" class="alias-tag">→ {{ rule.aliasFor }}</span>
                <br><span class="rule-pattern-text">{{ rule.pattern || '—' }}</span>
              </td>
              <td class="rule-group"><span class="group-tag">{{ GROUP_META.find(g => g.key === rule.groupName)?.label || rule.groupName }}</span></td>
              <td class="rule-actions">
                <div class="row-actions">
                  <button class="action-btn edit" @click="openEdit(rule)">修改</button>
                  <button class="action-btn delete" @click="openDelRule(rule.id)">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="filteredRules.length === 0" class="empty-row">
              <td colspan="3" class="empty-cell">暂无规则，可点击新增规则添加</td>
            </tr>
          </tbody>
        </table>
        <div class="table-footer">
          共 {{ filteredRules.length }} 条规则<span v-if="rGroup"> · 当前分组：{{ GROUP_META.find(g => g.key === rGroup)?.label }}</span>
        </div>
      </div>
    </template>

    <!-- ── Dialogs ── -->
    <UiModal v-model="llmShowCreate" title="新建LLM抽取技能" subtitle="配置大模型抽取技能" icon="ri-brain-line" :width="640" @confirm="saveLlmCreate">
      <div class="fg"><label class="fl">技能名称 <span class="req">*</span></label><UiInput v-model="eName" placeholder="如：项目信息抽取" /></div>
      <div class="fg"><label class="fl">技能代码 <span class="req">*</span></label><UiInput v-model="eCode" placeholder="唯一英文标识" /></div>
      <div class="fg"><label class="fl">关联分类</label><UiSelect v-model="eBiz" :options="BIZ_OPTIONS" size="sm" /></div>
      <div class="fg"><label class="fl">绑定提示词模板</label><UiSelect v-model="eBindPrompt" :options="PROMPT_OPTIONS" size="sm" /></div>
      <div class="fg"><label class="fl">提示词内容 <span class="req">*</span></label><textarea v-model="ePrompt" class="fta" rows="5" :placeholder="'输入提示词，支持 ' + '{{' + 'chunkText' + '}}' + ' 占位符'"></textarea></div>
      <div class="frow">
        <div class="fg"><label class="fl">Temperature</label><UiInput v-model.number="eTemp" type="number" step="0.1" min="0" max="2" /></div>
        <div class="fg"><label class="fl">Top P</label><UiInput v-model.number="eTopP" type="number" step="0.1" min="0" max="1" /></div>
        <div class="fg"><label class="fl">Max Tokens</label><UiInput v-model.number="eMaxTok" type="number" /></div>
      </div>
    </UiModal>

    <UiModal v-model="llmShowEdit" title="编辑LLM抽取技能" subtitle="修改技能配置" icon="ri-edit-line" :width="640" @confirm="saveLlmEdit">
      <div class="fg"><label class="fl">技能名称 <span class="req">*</span></label><UiInput v-model="eName" /></div>
      <div class="fg"><label class="fl">关联分类</label><UiSelect v-model="eBiz" :options="BIZ_OPTIONS" size="sm" /></div>
      <div class="fg"><label class="fl">绑定提示词模板</label><UiSelect v-model="eBindPrompt" :options="PROMPT_OPTIONS" size="sm" /></div>
      <div class="fg"><label class="fl">提示词内容 <span class="req">*</span></label><textarea v-model="ePrompt" class="fta" rows="5"></textarea></div>
      <div class="frow">
        <div class="fg"><label class="fl">Temperature</label><UiInput v-model.number="eTemp" type="number" step="0.1" min="0" max="2" /></div>
        <div class="fg"><label class="fl">Top P</label><UiInput v-model.number="eTopP" type="number" step="0.1" min="0" max="1" /></div>
        <div class="fg"><label class="fl">Max Tokens</label><UiInput v-model.number="eMaxTok" type="number" /></div>
      </div>
    </UiModal>

    <UiModal v-model="tplShowEdit" :title="editingTpl?'编辑提示词模板':'新建提示词模板'" subtitle="维护可复用的LLM抽取提示词" icon="ri-chat-3-line" :width="640" @confirm="saveTpl">
      <div class="fg"><label class="fl">模板名称 <span class="req">*</span></label><UiInput v-model="tName" placeholder="如：项目信息抽取提示词" /></div>
      <div class="fg"><label class="fl">关联分类</label><UiSelect v-model="tBiz" :options="BIZ_OPTIONS" size="sm" /></div>
      <div class="fg"><label class="fl">描述</label><UiInput v-model="tDesc" multiline :rows="2" /></div>
      <div class="fg"><label class="fl">提示词内容 <span class="req">*</span></label><textarea v-model="tText" class="fta" rows="7" :placeholder="'输入提示词，可使用 ' + '{{' + 'chunkText' + '}}' + ' 占位符'"></textarea></div>
    </UiModal>

    <RuleDialog
      v-model:visible="showRuleDialog"
      :mode="rDialogMode"
      :edit-rule="editingRule"
      @close="showRuleDialog = false"
      @submit="saveRule"
    />

    <!-- 删除技能确认弹窗 -->
    <UiModal
      v-model="showDelSkillDialog"
      title="删除技能"
      subtitle="删除后不可恢复，请谨慎操作"
      icon="ri-delete-bin-line"
      :width="480"
      confirm-text="确认删除"
      @confirm="confirmDelSkill"
    >
      <p class="confirm-text">确定删除技能「{{ deletingSkill?.skill_name }}」？</p>
    </UiModal>

    <!-- 删除模板确认弹窗 -->
    <UiModal
      v-model="showDelTplDialog"
      title="删除模板"
      subtitle="删除后不可恢复，请谨慎操作"
      icon="ri-delete-bin-line"
      :width="480"
      confirm-text="确认删除"
      @confirm="confirmDelTpl"
    >
      <p class="confirm-text">确定删除模板「{{ deletingTpl?.name }}」？</p>
    </UiModal>

    <!-- 删除规则确认弹窗 -->
    <UiModal
      v-model="showDelRuleDialog"
      title="删除规则"
      subtitle="删除后不可恢复，请谨慎操作"
      icon="ri-delete-bin-line"
      :width="480"
      confirm-text="确认删除"
      @confirm="confirmDelRule"
    >
      <p class="confirm-text">确定删除此规则？</p>
    </UiModal>
  </div>
</template>

<style scoped>
.profile-view { display:flex; flex-direction:column; gap:20px; max-width:1620px; }

/* Tab nav */
.tab-nav { display:flex; gap:8px; background:var(--color-bg-white,#fff); border:0.7px solid var(--color-border-soft,#EDE9E0); border-radius:12px; padding:6px; width:fit-content; }
.tab-btn { display:inline-flex; align-items:center; gap:7px; padding:9px 20px; border:none; border-radius:8px; background:transparent; font-size:14px; font-weight:500; color:var(--color-text-secondary,#6E6A60); cursor:pointer; font-family:"SourceHanSans-Regular",var(--font-family-cn); transition:all .2s; }
.tab-btn:hover { background:var(--color-bg-secondary,#F7F3EA); }
.tab-btn.active { background:var(--color-primary-soft,#FBE9E7); color:var(--color-primary,#C0392B); }
.tab-btn .icon { font-family:"remixicon",sans-serif; font-style:normal; font-size:16px; }

/* Section card */
.section-card { background:var(--color-bg-white,#fff); border:0.7px solid var(--color-border-soft,#EDE9E0); border-radius:16px; padding:24px 32px; }
.card-title { font-size:16px; font-weight:600; color:var(--color-text-primary,#1F1D1A); font-family:"SourceHanSans-Medium",var(--font-family-cn); margin-bottom:16px; }

/* Format row */
.format-row { display:flex; gap:16px; }
.format-cell { flex:1; display:flex; align-items:center; gap:14px; padding:14px 18px; border:1.5px solid var(--color-border-soft,#EDE9E0); border-radius:12px; cursor:pointer; transition:all .15s; }
.format-cell:hover { border-color:var(--color-primary,#C0392B); background:rgba(192,57,43,.02); }
.format-cell.selected { border-color:var(--color-primary,#C0392B); background:var(--color-primary-soft,#FBE9E7); }
.format-icon { width:40px; height:40px; border-radius:10px; display:flex; align-items:center; justify-content:center; font-weight:700; font-size:14px; color:#fff; font-family:"SourceHanSans-Bold",var(--font-family-cn); flex-shrink:0; }
.format-info { display:flex; flex-direction:column; gap:2px; }
.fi-name { font-size:15px; font-weight:500; color:var(--color-text-primary,#1F1D1A); font-family:"SourceHanSans-Medium",var(--font-family-cn); }
.fi-desc { font-size:13px; color:var(--color-text-muted,#8A8578); font-family:"SourceHanSans-Regular",var(--font-family-cn); }

/* Setting item */
.setting-item { display:flex; align-items:center; justify-content:space-between; height:64px; }
.setting-info { display:flex; flex-direction:column; gap:4px; }
.setting-name { font-size:15px; font-weight:500; color:var(--color-text-primary,#1F1D1A); font-family:"SourceHanSans-Medium",var(--font-family-cn); }
.setting-desc { font-size:13px; color:var(--color-text-muted,#8A8578); font-family:"SourceHanSans-Regular",var(--font-family-cn); }

/* Toolbar */
.toolbar { display:flex; justify-content:space-between; align-items:center; flex-wrap:wrap; gap:10px; }
.btn-primary { display:inline-flex; align-items:center; gap:6px; padding:0 16px; height:40px; border:none; border-radius:9px; background:var(--color-primary,#C0392B); color:#fff; font-size:14px; font-weight:500; cursor:pointer; font-family:"SourceHanSans-Medium",var(--font-family-cn); transition:all .2s; }
.btn-primary:hover { background:color-mix(in srgb,var(--color-primary,#C0392B) 88%,black); }
.btn-secondary { display:inline-flex; align-items:center; gap:6px; padding:0 16px; height:40px; border:0.7px solid var(--color-border,#E5DFD3); border-radius:9px; background:var(--color-bg-white,#fff); color:var(--color-text-body,#3A3A3A); font-size:14px; font-weight:500; cursor:pointer; font-family:"SourceHanSans-Medium",var(--font-family-cn); transition:all .2s; }
.btn-secondary:hover { background:var(--color-bg-secondary,#F7F3EA); }
.search-input { width:240px; }

/* Content */
.content-area { flex:1; }
.loading, .empty { display:flex; align-items:center; justify-content:center; gap:8px; padding:48px; color:var(--color-text-muted,#8A8578); font-size:14px; font-family:"SourceHanSans-Regular",var(--font-family-cn); }
.spinning { animation:spin 1s linear infinite; }
@keyframes spin { from{transform:rotate(0deg)} to{transform:rotate(360deg)} }

/* ── Skill cards ── */
.skill-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(310px,1fr)); gap:14px; }
.skill-card { background:var(--color-bg-white,#fff); border:0.7px solid var(--color-border-soft,#EDE9E0); border-radius:14px; padding:16px; display:flex; flex-direction:column; gap:10px; transition:box-shadow .2s; }
.skill-card:hover { box-shadow:0 4px 12px rgba(0,0,0,.08); }
.skill-card.disabled { opacity:.55; }
.card-top { display:flex; justify-content:space-between; align-items:flex-start; gap:10px; }
.card-title { display:flex; flex-direction:column; gap:3px; min-width:0; }
.card-title .name { font-size:15px; font-weight:600; color:var(--color-text-primary,#1F1D1A); font-family:"SourceHanSans-Medium",var(--font-family-cn); white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
.card-title .code { font-size:11px; color:var(--color-text-muted,#8A8578); font-family:monospace; }
.card-title .biz-tag { font-size:11px; padding:1px 7px; border-radius:9999px; background:var(--color-bg-secondary,#F7F3EA); color:var(--color-text-muted,#8A8578); width:fit-content; }
.card-prompt-preview { background:var(--color-bg-secondary,#F7F3EA); border-radius:7px; padding:8px 10px; max-height:200px; overflow-y:auto; }
.card-prompt-preview pre { font-family:monospace; font-size:11px; line-height:1.5; color:var(--color-text-body,#3A3A3A); white-space:pre-wrap; word-break:break-all; margin:0; }
.card-meta { display:flex; gap:10px; flex-wrap:wrap; }
.meta { display:inline-flex; align-items:center; gap:3px; font-size:11px; padding:2px 8px; border-radius:9999px; background:var(--color-bg-secondary,#F7F3EA); color:var(--color-text-muted,#8A8578); font-family:"SourceHanSans-Regular",var(--font-family-cn); border:0.7px solid transparent; }
.meta.on { color:var(--color-info,#4A72A8); background:#E8F0F8; border-color:rgba(74,114,168,.3); font-weight:600; }
.meta.clickable { cursor:pointer; user-select:none; transition:all .15s; }
.meta.clickable:hover { border-color:var(--color-info,#4A72A8); }
.meta i { font-family:"remixicon",sans-serif; font-style:normal; font-size:12px; }
.card-actions { display:flex; align-items:center; gap:8px; padding-top:6px; border-top:1px solid var(--color-border-soft,#EDE9E0); }
.ml-auto { margin-left:auto; }
.act-btn { font-size:13px; font-family:"SourceHanSans-Regular",var(--font-family-cn); background:none; border:none; cursor:pointer; padding:2px 7px; border-radius:5px; transition:background .15s; }
.act-btn.edit { color:var(--color-info,#4A72A8); }
.act-btn.edit:hover { background:rgba(74,114,168,.08); }
.act-btn.del { color:var(--color-error,#E74C3C); }
.act-btn.del:hover { background:rgba(201,80,60,.08); }

/* ── Prompt template cards ── */
.tpl-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(280px,1fr)); gap:14px; }
.tpl-card { background:var(--color-bg-white,#fff); border:0.7px solid var(--color-border-soft,#EDE9E0); border-radius:14px; padding:16px; display:flex; flex-direction:column; gap:8px; transition:box-shadow .2s; }
.tpl-card:hover { box-shadow:0 4px 12px rgba(0,0,0,.08); }
.tpl-card.disabled { opacity:.55; }
.tpl-top { display:flex; justify-content:space-between; align-items:flex-start; gap:10px; }
.tpl-name { font-size:15px; font-weight:600; color:var(--color-text-primary,#1F1D1A); font-family:"SourceHanSans-Medium",var(--font-family-cn); }
.tpl-biz { font-size:11px; color:var(--color-text-muted,#8A8578); font-family:"SourceHanSans-Regular",var(--font-family-cn); }
.tpl-desc { font-size:13px; color:var(--color-text-secondary,#6E6A60); font-family:"SourceHanSans-Regular",var(--font-family-cn); line-height:1.5; }
.tpl-preview { background:var(--color-bg-secondary,#F7F3EA); border-radius:7px; padding:8px 10px; max-height:200px; overflow-y:auto; }
.tpl-preview pre { font-family:monospace; font-size:11px; line-height:1.5; color:var(--color-text-body,#3A3A3A); white-space:pre-wrap; word-break:break-all; margin:0; }
.tpl-actions { display:flex; justify-content:flex-end; gap:8px; padding-top:4px; border-top:1px solid var(--color-border-soft,#EDE9E0); }

/* ── Form controls ── */
.fg { display:flex; flex-direction:column; gap:5px; margin-bottom:12px; }

/* 确认类弹窗正文 */
.confirm-text {
  font-size: 15px;
  color: var(--color-text-body, #3A3A3A);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  margin: 0;
  line-height: 1.7;
}
.fl { font-size:13px; color:var(--color-text-body,#3A3A3A); font-weight:500; font-family:"SourceHanSans-Regular",var(--font-family-cn); }
.fl .req { color:var(--color-primary,#C0392B); }
.fta { width:100%; padding:10px; border:1px solid var(--color-border,#E5DFD3); border-radius:8px; background:var(--color-bg-white,#fff); font-size:13px; color:var(--color-text-body,#3A3A3A); font-family:monospace; resize:vertical; outline:none; line-height:1.6; }
.fta:focus { border-color:var(--color-primary,#C0392B); }
.frow { display:grid; grid-template-columns:repeat(3,1fr); gap:12px; }

/* ── Rule table (from RulesSettingsView) ── */
.rule-toolbar { display:flex; justify-content:space-between; align-items:center; gap:16px; flex-wrap:wrap; }
.rule-toolbar-left { display:flex; gap:12px; align-items:center; }
.rule-toolbar-right { display:flex; align-items:center; gap:12px; flex-wrap:wrap; }
.add-rule-button { display:inline-flex; align-items:center; gap:8px; padding:0 20px; height:46px; background:var(--color-primary,#C0392B); color:#fff; border:none; border-radius:10px; font-size:16px; font-weight:500; cursor:pointer; font-family:"SourceHanSans-Medium",var(--font-family-cn); transition:all .2s; }
.add-rule-button:hover { filter:brightness(1.05); }
.add-rule-button .icon { font-family:"remixicon",sans-serif; font-style:normal; font-size:20px; }
.seed-universal-button { display:inline-flex; align-items:center; gap:6px; padding:0 16px; height:40px; border:0.7px solid var(--color-info,#4A72A8); border-radius:9px; background:rgba(74,114,168,.06); color:var(--color-info,#4A72A8); font-size:14px; font-weight:500; cursor:pointer; font-family:"SourceHanSans-Medium",var(--font-family-cn); transition:all .2s; }
.seed-universal-button:hover { background:rgba(74,114,168,.12); }
.universal-badge { display:inline-block; margin-left:6px; padding:1px 6px; border-radius:9999px; font-size:10px; font-weight:600; color:var(--color-info,#4A72A8); background:rgba(74,114,168,.1); white-space:nowrap; vertical-align:1px; }
.toolbar-search { width:300px; }
.group-filter { display:flex; align-items:center; gap:4px; padding:4px; border:0.7px solid var(--color-border,#E5DFD3); background:var(--color-bg-white,#fff); border-radius:12px; }
.group-icon-btn { width:38px; height:38px; border:none; border-radius:9px; background:transparent; cursor:pointer; display:flex; align-items:center; justify-content:center; color:var(--color-text-muted,#8A8578); transition:background .15s, color .15s; }
.group-icon-btn .icon { font-family:"remixicon",sans-serif; font-style:normal; font-size:19px; }
.group-icon-btn:hover { background:var(--color-bg-secondary,#F7F3EA); color:var(--color-text-primary,#1F1D1A); }
.group-icon-btn.active { background:var(--color-primary-soft,#FBE9E7); color:var(--color-primary,#C0392B); }
.rule-table-card { background:var(--color-bg-white,#fff); border:0.7px solid var(--color-border-soft,#EDE9E0); border-radius:16px; overflow:hidden; }
.rule-table { width:100%; border-collapse:collapse; }
.rule-table thead { background:var(--color-bg-secondary,#F7F3EA); }
.rule-table thead tr th { text-align:left; font-size:15px; font-weight:400; color:var(--color-text-secondary,#6E6A60); padding:0 40px; height:52px; font-family:"SourceHanSans-Regular",var(--font-family-cn); }
.rule-table thead tr th:last-child { text-align:right; }
.rule-table tbody tr td { padding:0 40px; height:76px; vertical-align:middle; border-bottom:0.7px solid var(--color-border-soft,#EDE9E0); }
.rule-table tbody tr td:last-child { text-align:right; }
.rule-table tbody tr:hover td { background:var(--color-bg-card,#FCFBF8); }
.rule-desc { font-size:14px; color:var(--color-text-body,#3A3A3A); font-family:"SourceHanSans-Regular",var(--font-family-cn); }
.rule-field-name { font-size:15px; font-weight:500; color:var(--color-text-primary,#1F1D1A); font-family:"SourceHanSans-Medium",var(--font-family-cn); }
.rule-pattern-text { font-size:12px; color:var(--color-text-muted,#8A8578); font-family:monospace; white-space:pre-wrap; word-break:break-all; margin-top:4px; display:block; }
.rule-group { display:flex; align-items:center; }
.group-tag { display:inline-block; padding:4px 12px; border-radius:9999px; font-size:13px; color:var(--color-primary,#C0392B); background:var(--color-primary-soft,#FBE9E7); font-family:"SourceHanSans-Regular",var(--font-family-cn); white-space:nowrap; }
.alias-tag { display:inline-block; margin-left:8px; padding:2px 10px; border-radius:9999px; font-size:12px; font-weight:400; color:var(--color-success-strong,#3E8A6B); background:var(--color-success-soft,#E8F5EF); white-space:nowrap; vertical-align:1px; }
.rule-actions { width:100%; }
.row-actions { display:flex; align-items:center; justify-content:flex-end; gap:24px; }
.action-btn { font-size:15px; font-family:"SourceHanSans-Regular",var(--font-family-cn); background:none; border:none; cursor:pointer; padding:0; transition:opacity .2s; }
.action-btn.edit { color:var(--color-info,#4A72A8); }
.action-btn.delete { color:var(--color-error,#E74C3C); }
.action-btn:hover { opacity:0.7; }
.table-footer { padding:0 40px; height:54px; display:flex; align-items:center; font-size:14px; color:var(--color-text-faint,#A7A296); font-family:"SourceHanSans-Regular",var(--font-family-cn); }
.empty-row .empty-cell { text-align:center; vertical-align:middle; height:auto; padding:40px 0; color:var(--color-text-muted,#8A8578); font-size:14px; }
</style>
