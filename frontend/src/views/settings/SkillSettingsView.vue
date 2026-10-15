<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { ElMessage, ElMessageBox } from '@/utils/toast';
import { getSkillList, updateSkill, deleteSkill, resetSkills, syncSkills } from '@/api/settings';
import { apiClient } from '@/api/client';
import { UiInput, UiSwitch, UiModal, UiTip } from '@/components/ui';

interface RuleItem {
  id: string;
  fieldName: string;
  pattern: string;
  enabled: boolean;
  category: string;
  groupName: string;
}

interface SkillItem {
  id?: string;
  name: string;
  description: string;
  group: string;
  type: 'rules' | 'prompt';
  fields: string[];
  content?: string;
  enabled: boolean;
  importedAt: string;
  rules: RuleItem[];
}

const skills = ref<SkillItem[]>([]);
const loading = ref(false);
const importLoading = ref(false);
const syncLoading = ref(false);
const expandedSkill = ref<string | null>(null);
const selectedFileName = ref('');
const viewMode = ref<'list' | 'grid'>((() => {
  try { return localStorage.getItem('settings.skill.viewMode') === 'grid' ? 'grid' : 'list'; }
  catch { return 'list'; }
})());
watch(viewMode, (v) => {
  try { localStorage.setItem('settings.skill.viewMode', v); } catch {}
});
const searchTerm = ref('');
const showEditDialog = ref(false);
const editingSkill = ref<SkillItem | null>(null);
const editName = ref('');
const editDesc = ref('');

function openEdit(skill: SkillItem) {
  editingSkill.value = skill;
  editName.value = skill.name;
  editDesc.value = skill.description;
  showEditDialog.value = true;
}

async function confirmEdit() {
  if (!editingSkill.value || !editName.value.trim()) {
    ElMessage.warning('请输入技能名称');
    return;
  }
  try {
    await updateSkill(editingSkill.value.name, {
      name: editName.value.trim(),
      description: editDesc.value.trim(),
    });
    await loadSkills();
    showEditDialog.value = false;
    ElMessage.success('技能已更新');
  } catch {
    ElMessage.error('保存失败');
  }
}

const groupLabels: Record<string, string> = {
  info: '项目信息',
  business: '商务条款',
  tech: '技术条款',
  score: '评分标准',
  seal: '封标标准',
  star: '标星信息',
};

const filteredSkills = ref<SkillItem[]>([]);

watch([skills, searchTerm], ([s, t]) => {
  if (!t) {
    filteredSkills.value = s;
    return;
  }
  const term = t.toLowerCase();
  filteredSkills.value = s.filter(sk =>
    sk.name.toLowerCase().includes(term) ||
    sk.description.toLowerCase().includes(term)
  );
});

const MOCK_SKILLS: SkillItem[] = [
  {
    id: '1',
    name: '招标文件要点提取',
    description: '自动抽取资质门槛、工期与评分办法',
    group: 'info',
    type: 'rules',
    fields: ['项目名称', '招标编号', '投标截止时间'],
    enabled: true,
    importedAt: new Date().toISOString(),
    rules: [
      { id: 'r1', fieldName: '项目名称', pattern: '项目名称[：:]\\s*(.+)', enabled: true, category: 'regex', groupName: 'info' },
      { id: 'r2', fieldName: '招标编号', pattern: '招标编号', enabled: true, category: 'keyword', groupName: 'info' },
      { id: 'r3', fieldName: '投标截止时间', pattern: '投标截止.*?(\\d{4}[年-/]\\d{1,2}[月-/]\\d{1,2})', enabled: true, category: 'regex', groupName: 'info' },
    ],
  },
  {
    id: '2',
    name: '废标条款风险识别',
    description: '比对废标红线，输出风险清单与预警',
    group: 'business',
    type: 'rules',
    fields: ['废标条款', '风险等级'],
    enabled: true,
    importedAt: new Date().toISOString(),
    rules: [
      { id: 'r4', fieldName: '废标条款', pattern: '废标|否决投标|无效标', enabled: true, category: 'keyword', groupName: 'business' },
      { id: 'r5', fieldName: '风险等级', pattern: '重大偏差|实质性偏离', enabled: true, category: 'regex', groupName: 'business' },
    ],
  },
  {
    id: '3',
    name: '工程量清单核对',
    description: '比对编码、单位与数量口径差异',
    group: 'tech',
    type: 'rules',
    fields: ['清单编码', '工程量', '单位'],
    enabled: true,
    importedAt: new Date().toISOString(),
    rules: [
      { id: 'r6', fieldName: '清单编码', pattern: '[A-Z]{2,4}-\\d{3,6}', enabled: true, category: 'regex', groupName: 'tech' },
      { id: 'r7', fieldName: '工程量', pattern: '工程量[:：]\\s*([\\d,\\.]+)', enabled: true, category: 'regex', groupName: 'tech' },
    ],
  },
  {
    id: '4',
    name: '报价合理性分析',
    description: '结合市场价区间识别异常报价项',
    group: 'score',
    type: 'rules',
    fields: ['报价', '市场价'],
    enabled: false,
    importedAt: new Date().toISOString(),
    rules: [
      { id: 'r8', fieldName: '报价', pattern: '报价[:：]\\s*([\\d,\\.]+)\\s*万元', enabled: true, category: 'regex', groupName: 'score' },
    ],
  },
  {
    id: '5',
    name: '方案相似度检测',
    description: '检测自有方案与外部素材重合片段',
    group: 'info',
    type: 'rules',
    fields: ['相似度', '重合片段'],
    enabled: false,
    importedAt: new Date().toISOString(),
    rules: [
      { id: 'r9', fieldName: '相似度', pattern: '相似度[:：]\\s*([\\d.]+)%', enabled: true, category: 'regex', groupName: 'info' },
    ],
  },
  {
    id: '6',
    name: '公文红头自动套用',
    description: '按 GB/T 9704 一键套合红头与版记',
    group: 'tech',
    type: 'prompt',
    fields: ['公文格式'],
    content: '请按照GB/T 9704-2012标准，对文档进行红头、版记和页码格式化。',
    enabled: true,
    importedAt: new Date().toISOString(),
    rules: [],
  },
];

async function handleResetSkills() {
  try {
    await ElMessageBox.confirm('确定恢复默认技能？已导入的自定义技能将被清除。', '确认恢复默认');
    await resetSkills();
    skills.value = MOCK_SKILLS;
    filteredSkills.value = MOCK_SKILLS;
    expandedSkill.value = null;
    ElMessage.success('已恢复默认技能');
  } catch { /* cancelled */ }
}

async function handleSyncSkills() {
  syncLoading.value = true;
  try {
    const result = await syncSkills();
    if (result.synced > 0) {
      ElMessage.success(`同步成功：${result.synced} 个本地技能已导入`);
    } else {
      ElMessage.info(result.message || '本地没有可同步的技能文件');
    }
    await loadSkills();
  } catch (err: any) {
    ElMessage.error(err?.message || '同步失败');
  } finally {
    syncLoading.value = false;
  }
}

async function loadSkills() {
  loading.value = true;
  try {
    const data = getSkillList();
    const skillsList = (data || []).map((s: any) => ({
      ...s,
      enabled: s.enabled ?? true,
      rules: Array.isArray(s.rules) ? s.rules : [],
    }));
    if (skillsList.length > 0) {
      skills.value = skillsList;
    } else {
      skills.value = MOCK_SKILLS;
    }
    filteredSkills.value = skills.value;
  } catch {
    skills.value = MOCK_SKILLS;
    filteredSkills.value = MOCK_SKILLS;
  } finally {
    loading.value = false;
  }
}

async function doImport(file: File) {
  importLoading.value = true;
  try {
    // Try backend first
    try {
      const formData = new FormData();
      formData.append('file', file);
      const res = await apiClient.post('/rules/import-skill', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      });
      const result = res.data;
      ElMessage.success(`导入技能「${result.name}」成功：${result.count} 条规则`);
      selectedFileName.value = '';
      await loadSkills();
      return;
    } catch (_backendErr) { /* fall through to localStorage */ }
    // Fallback: read file and parse as JSON/YAML skill definition
    const text = await file.text();
    let importedSkill: any;
    try { importedSkill = JSON.parse(text); } catch { importedSkill = null; }
    if (!importedSkill) {
      ElMessage.warning('仅支持 JSON 格式的技能定义文件');
      return;
    }
    const newSkill = {
      id: `custom_${Date.now()}`,
      name: importedSkill.name || file.name.replace(/\.\w+$/, ''),
      description: importedSkill.description || '',
      group: importedSkill.group || 'info',
      type: importedSkill.type || 'prompt',
      fields: importedSkill.fields || [],
      content: importedSkill.content || '',
      enabled: true,
      importedAt: new Date().toISOString().slice(0, 10),
      rules: importedSkill.rules || [],
    };
    const list = [...skills.value, newSkill];
    skills.value = list;
    filteredSkills.value = list;
    ElMessage.success(`已导入技能「${newSkill.name}」`);
    selectedFileName.value = '';
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || '导入失败');
  } finally {
    importLoading.value = false;
  }
}

function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  selectedFileName.value = file.name;
  doImport(file);
  input.value = '';
}

function triggerFileInput() {
  document.getElementById('skillFileInput')?.click();
}

async function toggleEnabled(skill: SkillItem) {
  const next = !skill.enabled;
  skill.enabled = next;
  try {
    await updateSkill(skill.name, { enabled: next ? 1 : 0 });
    ElMessage.success(next ? `已启用「${skill.name}」` : `已禁用「${skill.name}」`);
  } catch {
    skill.enabled = !next;
    ElMessage.error('操作失败');
  }
}

async function handleDelete(skill: SkillItem) {
  try {
    await ElMessageBox.confirm(`确定删除技能「${skill.name}」及其 ${(skill.rules || []).length} 条规则？`, '确认删除');
    await deleteSkill(skill.name);
    ElMessage.success('已删除');
    if (expandedSkill.value === skill.name) expandedSkill.value = null;
    await loadSkills();
  } catch {
    // cancelled
  }
}

onMounted(loadSkills);
</script>

<template>
  <div class="skill-view">
    <!-- 页面头部卡片 -->
    <div class="page-header-card">
      <div class="header-content">
        <div class="header-info">
          <div class="breadcrumb">
            <span>首页</span>
            <span class="sep">/</span>
            <span>设置中心</span>
            <span class="sep">/</span>
            <span class="current">技能管理</span>
          </div>
          <div class="header-title">
            <h1>技能管理</h1>
            <p>导入或同步能力插件，控制各技能在创作流程中的启用状态</p>
          </div>
        </div>
        <button class="reset-button" @click="handleResetSkills">
          <RiRestartLine class="icon" />
          恢复默认
        </button>
      </div>
    </div>

    <!-- 工具栏 -->
      <div class="toolbar">
      <div class="toolbar-left">
        <button class="btn-import" @click="triggerFileInput">
          <RiFileUploadLine class="icon" />
          <span>导入技能</span>
        </button>
        <button class="btn-sync" :disabled="syncLoading" @click="handleSyncSkills">
          <RiLoader4Line v-if="syncLoading" class="icon spinning" /><RiRefreshLine v-else class="icon" />
          <span>{{ syncLoading ? '同步中...' : '同步技能' }}</span>
        </button>
      </div>
      <div class="toolbar-right">
        <UiInput
          v-model="searchTerm"
          type="search"
          class="toolbar-search"
          placeholder="搜索技能名称或描述"
          clearable
        />
        <div class="view-toggle">
          <UiTip text="列表视图">
            <button
              class="view-btn"
              :class="{ active: viewMode === 'list' }"
              @click="viewMode = 'list'"
            >
              <RiLayoutColumnLine class="icon" />
            </button>
          </UiTip>
          <UiTip text="网格视图">
            <button
              class="view-btn"
              :class="{ active: viewMode === 'grid' }"
              @click="viewMode = 'grid'"
            >
              <RiLayoutGridLine class="icon" />
            </button>
          </UiTip>
        </div>
      </div>
    </div>

      <div class="skill-content">
      <!-- 加载状态 -->
      <div v-if="loading" class="loading-state">
        <RiLoader4Line class="icon spinning" />
        <span>加载中...</span>
      </div>

      <!-- 列表视图 -->
      <div v-else-if="viewMode === 'list'" class="list-container">
        <table class="list-table">
          <thead>
            <tr>
              <th class="col-name">技能名称</th>
              <th class="col-desc">技能描述</th>
              <th class="col-action">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="skill in filteredSkills"
              :key="skill.name"
              class="list-row"
              :class="{ expanded: expandedSkill === skill.name }"
            >
              <td class="col-name">
                <span class="row-name">{{ skill.name }}</span>
                <span class="row-category">{{ groupLabels[skill.group] || skill.group }}</span>
              </td>
              <td class="col-desc">
                <span class="row-desc">{{ skill.description }}</span>
              </td>
              <td class="col-action">
                <div class="row-actions">
                  <button class="action-btn edit" @click.stop="openEdit(skill)">修改</button>
                  <button class="action-btn delete" @click.stop="handleDelete(skill)">删除</button>
                  <UiSwitch :model-value="skill.enabled" size="sm" @update:model-value="toggleEnabled(skill)" />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        <div class="table-footer">
          共 {{ filteredSkills.length }} 个技能 · 已开启 {{ filteredSkills.filter(s => s.enabled).length }}
        </div>

        <!-- 展开的技能详情 -->
        <div
          v-for="skill in filteredSkills"
          :key="skill.name + '-detail'"
          v-show="expandedSkill === skill.name"
          class="skill-detail"
          @click.stop
        >
          <div class="detail-header">
            <span class="detail-name">{{ skill.name }}</span>
            <span class="detail-close" @click="expandedSkill = null">
              <RiCloseLine class="icon" />
            </span>
          </div>
          <div v-if="skill.type === 'prompt' && skill.content" class="detail-prompt">
            <pre class="prompt-text">{{ skill.content }}</pre>
          </div>
          <template v-else>
            <div class="rules-divider"></div>
            <table class="rule-table">
              <thead>
                <tr>
                  <th>字段名称</th>
                  <th>类型</th>
                  <th>规则内容</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="rule in skill.rules" :key="rule.id">
                  <td class="rule-field">{{ rule.fieldName }}</td>
                  <td>
                    <span class="type-tag" :class="rule.category">
                      {{ rule.category === 'keyword' ? '关键字' : '正则' }}
                    </span>
                  </td>
                  <td class="rule-pattern-cell">
                    <code v-if="rule.pattern" class="rule-pattern">{{ rule.pattern }}</code>
                    <span v-else class="empty-pattern">—</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </template>
        </div>
      </div>

      <!-- 卡片视图 -->
      <div v-else class="grid-container">
        <div class="grid-cards">
          <div
            v-for="skill in filteredSkills"
            :key="skill.name"
            class="grid-card"
          >
            <div class="grid-card-body">
              <div class="grid-card-icon">{{ groupLabels[skill.group] || skill.group }}</div>
              <div class="grid-card-title">{{ skill.name }}</div>
              <div class="grid-card-desc">{{ skill.description }}</div>
            </div>
            <div class="grid-card-footer">
              <div class="grid-card-actions">
                <button class="action-btn edit" @click="openEdit(skill)">修改</button>
                <button class="action-btn delete" @click="handleDelete(skill)">删除</button>
                <UiSwitch :model-value="skill.enabled" size="sm" @click.stop="toggleEnabled(skill)" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <input
      id="skillFileInput"
      type="file"
      class="file-input"
      accept=".md,.yaml,.yml"
      @change="handleFileSelect"
    />

    <div v-if="importLoading" class="loading-overlay">
      <RiLoader4Line class="icon spinning" />
      <span>正在导入: {{ selectedFileName }}</span>
    </div>

    <!-- 编辑技能弹窗 -->
    <UiModal
      v-model="showEditDialog"
      title="编辑技能"
      subtitle="修改技能名称与描述（规则明细不受影响）"
      icon="ri-edit-line"
      :width="560"
      @confirm="confirmEdit"
    >
      <div class="form-group">
        <label class="form-label">技能名称 <span class="required">*</span></label>
        <UiInput v-model="editName" placeholder="请输入技能名称" />
      </div>
      <div class="form-group">
        <label class="form-label">技能描述</label>
        <UiInput v-model="editDesc" multiline :rows="3" placeholder="请输入技能描述（可选）" />
      </div>
    </UiModal>
  </div>
</template>

<style scoped>
.skill-view {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-width: 1620px;
}

/* 页面头部卡片 */
.page-header-card {
  background-color: var(--color-bg-white, #FFFFFF);
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  border-radius: 16px;
  padding: 20px 24px;
}
.header-content {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
}
.header-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
.breadcrumb span {
  color: var(--color-text-faint, #A7A296);
}
.breadcrumb span.sep {
  color: var(--color-text-faint, #A7A296);
}
.breadcrumb span.current {
  color: var(--color-text-body, #3A3A3A);
  font-weight: 500;
}
.header-title h1 {
  font-size: 30px;
  font-weight: 700;
  color: var(--color-text-primary, #1F1D1A);
  margin: 12px 0 0 0;
  letter-spacing: -0.5px;
  font-family: "SourceHanSans-Bold", var(--font-family-cn);
  line-height: 1.2;
}
.header-title p {
  font-size: 15px;
  color: var(--color-text-secondary, #6E6A60);
  margin: 8px 0 0 0;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}

/* 工具栏 */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.toolbar-left {
  display: flex;
  gap: 12px;
  align-items: center;
}
.btn-import {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0 20px;
  height: 46px;
  border: none;
  border-radius: 10px;
  background: var(--color-primary, #C0392B);
  color: #fff;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
  transition: all 0.2s;
}
.btn-import:hover {
  background: color-mix(in srgb, var(--color-primary, #C0392B) 88%, black);
}
.btn-sync {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0 20px;
  height: 46px;
  border: 0.7px solid var(--color-border, #E5DFD3);
  border-radius: 10px;
  background: var(--color-bg-white, #fff);
  color: var(--color-text-body, #3A3A3A);
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
  transition: all 0.2s;
}
.btn-sync:hover {
  background: var(--color-bg-secondary, #F7F3EA);
}
.btn-sync .icon {
  color: var(--color-primary, #C0392B);
}
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.toolbar-search { width: 548px; }
.view-toggle {
  display: flex;
  border: 0.7px solid var(--color-border, #E5DFD3);
  border-radius: 12px;
  padding: 4px;
  background: var(--color-bg-white, #fff);
  gap: 4px;
}
.view-btn {
  width: 54px;
  height: 38px;
  border: none;
  border-radius: 9px;
  background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}
.view-btn .icon {
  font-family: "remixicon", sans-serif;
  font-style: normal;
  font-size: 18px;
  color: var(--color-text-muted, #8A8578);
}
.view-btn.active {
  background: #F6EEDA;
}
.view-btn.active .icon {
  color: var(--color-primary, #C0392B);
}

.reset-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0 20px;
  height: 48px;
  border: 0.7px solid var(--color-border, #E5DFD3);
  background-color: var(--color-bg-white, #FFFFFF);
  border-radius: 12px;
  font-size: 15px;
  color: var(--color-text-body, #3A3A3A);
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
.reset-button:hover {
  background-color: var(--color-bg-secondary, #F7F3EA);
}
.reset-button .icon {
  font-family: "remixicon", sans-serif;
  font-style: normal;
  font-size: 18px;
}

/* 内容区 */
.skill-content {
  flex: 1;
}

/* 加载状态 */
.loading-state {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 40px;
  color: var(--color-text-muted, #8A8578);
  font-size: 14px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
.spinning {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 列表视图 */
.list-container { background-color: var(--color-bg-white, #FFFFFF); border: 0.7px solid var(--color-border-soft, #EDE9E0); border-radius: 16px; overflow: hidden; }
.list-table { width: 100%; border-collapse: collapse; }
.list-table thead { background-color: var(--color-bg-secondary, #F7F3EA); }
.list-table th {
  text-align: left; font-size: 15px; font-weight: 400; color: var(--color-text-secondary, #6E6A60);
  padding: 0 40px; height: 52px; font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
.list-table th:last-child { text-align: right; }
.list-table td { padding: 0 40px; height: 76px; vertical-align: middle; border-bottom: 0.7px solid var(--color-border-soft, #EDE9E0); }
.list-table td:last-child { text-align: right; }
.col-name { width: 430px; flex-shrink: 0; }
.col-desc { flex: 1; }
.col-action { width: 260px; flex-shrink: 0; }
.row-name {
  display: block; font-size: 16px; font-weight: 500; color: var(--color-text-primary, #1F1D1A);
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
}
.row-category {
  display: block; font-size: 12px; color: var(--color-text-muted, #8A8578); margin-top: 2px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
.row-desc { font-size: 15px; color: var(--color-text-secondary, #6E6A60); font-family: "SourceHanSans-Regular", var(--font-family-cn); }
.row-actions { display: flex; align-items: center; justify-content: flex-end; gap: 24px; }
.row-actions .action-btn {
  font-size: 15px; font-family: "SourceHanSans-Regular", var(--font-family-cn);
  background: none; border: none; cursor: pointer; padding: 0; transition: opacity 0.2s;
}
.row-actions .action-btn.edit { color: var(--color-info, #4A72A8); }
.row-actions .action-btn.delete { color: var(--color-error, #E74C3C); }
.row-actions .action-btn:hover { opacity: 0.7; }
.table-footer {
  padding: 0 40px; height: 54px; display: flex; align-items: center;
  font-size: 14px; color: var(--color-text-faint, #A7A296); font-family: "SourceHanSans-Regular", var(--font-family-cn);
}

/* 开关已统一为 UiSwitch，样式见 components/ui */

/* 展开详情 */
.skill-detail {
  background: var(--color-bg-white, #fff);
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  border-top: none;
  border-radius: 0 0 16px 16px;
  margin-top: -1px;
  padding: 20px 40px;
  animation: slideDown 0.2s ease;
}
@keyframes slideDown {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: translateY(0); }
}
.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}
.detail-name {
  font-size: 17px;
  font-weight: 600;
  color: var(--color-text-primary, #1F1D1A);
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
}
.detail-close {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.2s;
}
.detail-close:hover {
  background: var(--color-bg-secondary, #F7F3EA);
}
.detail-close .icon {
  font-family: "remixicon", sans-serif;
  font-style: normal;
  font-size: 18px;
  color: var(--color-text-secondary, #6E6A60);
}
.rules-divider {
  height: 1px;
  background: var(--color-border-soft, #EDE9E0);
  margin-bottom: 12px;
}
.rule-table {
  width: 100%;
  border-collapse: collapse;
}
.rule-table th {
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 600;
  text-align: left;
  color: var(--color-text-muted, #8A8578);
  background: rgba(245,239,227,0.5);
}
.rule-table td {
  padding: 8px 12px;
  font-size: 12px;
  border-bottom: 1px solid rgba(245,239,227,0.8);
}
.rule-table th:first-child,
.rule-table td:first-child { width: 25%; }
.rule-table th:nth-child(2),
.rule-table td:nth-child(2) { width: 10%; }
.rule-table th:nth-child(3),
.rule-table td:nth-child(3) { width: 65%; }
.rule-field { font-weight: 500; color: var(--color-text-primary, #1F1D1A); }
.type-tag { font-size: 11px; padding: 2px 8px; border-radius: 9999px; }
.type-tag.regex { background: var(--color-info-soft, #E8F0F8); color: var(--color-info, #4A72A8); }
.type-tag.keyword { background: #F0E8D8; color: var(--color-text-muted, #8A8578); }
.rule-pattern { font-family: monospace; font-size: 11px; color: var(--color-info, #4A72A8); background: rgba(245,239,227,0.4); padding: 2px 6px; border-radius: 4px; word-break: break-all; }
.empty-pattern { color: var(--color-text-muted, #8A8578); }
.prompt-content {
  max-height: 400px;
  overflow-y: auto;
  padding: 4px 0;
}
.prompt-text {
  font-family: monospace;
  font-size: 12px;
  line-height: 1.6;
  color: var(--color-text-primary, #1F1D1A);
  white-space: pre-wrap;
  word-break: break-word;
  margin: 0;
}

/* 卡片视图 */
.grid-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.grid-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}
.grid-row {
  display: flex;
  gap: 20px;
}
.grid-card {
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  border-radius: 14px;
  background: var(--color-bg-white, #fff);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: box-shadow 0.2s;
  cursor: default;
  min-height: 220px;
}
.grid-card:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
}
.grid-card-body { padding: 24px; display: flex; flex: 1; flex-direction: column; gap: 8px; }
.grid-card-icon {
  display: inline-block; padding: 4px 12px; border-radius: 20px;
  background: var(--color-bg-secondary, #F7F3EA); font-size: 12px; color: var(--color-text-muted, #8A8578);
  font-family: "SourceHanSans-Regular", var(--font-family-cn); width: fit-content;
}
.grid-card-title { font-size: 17px; font-weight: 500; color: var(--color-text-primary, #1F1D1A); font-family: "SourceHanSans-Medium", var(--font-family-cn); margin: 0; }
.grid-card-desc { font-size: 14px; color: var(--color-text-secondary, #6E6A60); font-family: "SourceHanSans-Regular", var(--font-family-cn); flex: 1; }
.grid-card-footer { padding: 16px 24px; border-top: 1px solid var(--color-border-soft, #EDE9E0); display: flex; justify-content: flex-end; }
.grid-card-actions { display: flex; align-items: center; gap: 12px; }
.grid-card-actions .action-btn {
  font-size: 14px; font-family: "SourceHanSans-Regular", var(--font-family-cn);
  background: none; border: none; cursor: pointer; padding: 4px 8px; border-radius: 6px;
  transition: background 0.15s;
}
.grid-card-actions .action-btn.edit { color: var(--color-info, #4A72A8); }
.grid-card-actions .action-btn.edit:hover { background: rgba(74, 114, 168, 0.08); }
.grid-card-actions .action-btn.delete { color: var(--color-error, #E74C3C); }
.grid-card-actions .action-btn.delete:hover { background: rgba(201, 80, 60, 0.08); }

/* 加载遮罩 */
.loading-overlay {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px;
  color: var(--color-text-muted, #8A8578);
  font-size: 14px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}

.file-input {
  position: fixed;
  top: -1000px;
  left: -1000px;
}

@media (max-width: 1620px) {
  .grid-cards { grid-template-columns: repeat(2, 1fr); }
}
/* 编辑弹窗表单（壳为 UiModal） */
.form-group { display: flex; flex-direction: column; gap: 8px; }
.form-group .ui-input { width: 100%; }
.form-label { font-size: 14px; color: var(--color-text-body, #3A3A3A); font-weight: 500; font-family: "SourceHanSans-Regular", var(--font-family-cn); }
.form-label .required { color: var(--color-primary, #C0392B); }
@media (max-width: 768px) {
  .grid-cards { grid-template-columns: 1fr; }
}
</style>
