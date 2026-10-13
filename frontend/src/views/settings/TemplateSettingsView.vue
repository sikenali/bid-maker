<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { ElMessage } from 'element-plus';
import apiClient from '@/api/client';
import { UiInput, UiSwitch, UiModal, UiTip } from '@/components/ui';

interface TemplateItem {
  id: string;
  name: string;
  description: string;
  category: string;
  type: string;
  enabled: boolean;
}

const templates = ref<TemplateItem[]>([]);
const loading = ref(false);
const viewMode = ref<'list' | 'grid'>((() => {
  try { return localStorage.getItem('settings.template.viewMode') === 'grid' ? 'grid' : 'list'; }
  catch { return 'list'; }
})());
watch(viewMode, (v) => {
  try { localStorage.setItem('settings.template.viewMode', v); } catch {}
});
const searchQuery = ref('');
const showAddDialog = ref(false);
const newTemplateName = ref('');
const newTemplateDesc = ref('');
const newTemplateCategory = ref('');

async function loadTemplates() {
  loading.value = true;
  try {
    const res = await apiClient.get('/settings/templates');
    templates.value = ((res.data.templates || []) as any[]).map((t: any) => ({
      ...t,
      enabled: t.enabled ?? true,
    }));
  } catch {
    templates.value = [];
    ElMessage.error('模板加载失败，请检查后端服务');
  } finally {
    loading.value = false;
  }
}

const filteredTemplates = ref<TemplateItem[]>([]);

function applyFilter() {
  let list = templates.value;
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase();
    list = list.filter(t => t.name.toLowerCase().includes(q) || t.description.toLowerCase().includes(q));
  }
  filteredTemplates.value = list;
}

async function toggleEnabled(tpl: TemplateItem) {
  const next = !tpl.enabled;
  tpl.enabled = next;
  try {
    await apiClient.put(`/settings/templates/${tpl.id}`, { enabled: next });
    ElMessage.success(next ? `已启用「${tpl.name}」` : `已停用「${tpl.name}」`);
  } catch {
    tpl.enabled = !next;
    ElMessage.error('操作失败');
  }
}

const showDeleteDialog = ref(false);
const deletingTpl = ref<TemplateItem | null>(null);

function openDelete(tpl: TemplateItem) {
  deletingTpl.value = tpl;
  showDeleteDialog.value = true;
}

async function confirmDelete() {
  if (!deletingTpl.value) return;
  try {
    await apiClient.delete(`/settings/templates/${deletingTpl.value.id}`);
    templates.value = templates.value.filter(t => t.id !== deletingTpl.value?.id);
    applyFilter();
    showDeleteDialog.value = false;
    ElMessage.success('已删除');
  } catch {
    ElMessage.error('删除失败');
  }
}

async function handleAdd() {
  if (!newTemplateName.value.trim()) {
    ElMessage.warning('请输入模板名称'); return;
  }
  try {
    const res: any = await apiClient.post('/settings/templates', {
      name: newTemplateName.value.trim(),
      description: newTemplateDesc.value.trim(),
      category: newTemplateCategory.value.trim() || '自定义',
      type: 'custom',
    });
    templates.value.unshift({ ...res.data, enabled: true });
    showAddDialog.value = false;
    newTemplateName.value = '';
    newTemplateDesc.value = '';
    newTemplateCategory.value = '';
    applyFilter();
    ElMessage.success('已添加');
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || '添加失败');
  }
}

const showResetDialog = ref(false);

async function confirmReset() {
  try {
    const res: any = await apiClient.post('/settings/templates/reset');
    templates.value = ((res.data?.templates || []) as any[]).map((t: any) => ({
      ...t,
      enabled: t.enabled ?? true,
    }));
    applyFilter();
    showResetDialog.value = false;
    ElMessage.success('已恢复默认模板');
  } catch {
    ElMessage.error('恢复默认失败');
  }
}

const showEditDialog = ref(false);
const editingTpl = ref<TemplateItem | null>(null);
const editName = ref('');
const editDesc = ref('');
const editCategory = ref('');

function openEdit(tpl: TemplateItem) {
  editingTpl.value = tpl;
  editName.value = tpl.name;
  editDesc.value = tpl.description;
  editCategory.value = tpl.category;
  showEditDialog.value = true;
}

async function confirmEdit() {
  if (!editingTpl.value || !editName.value.trim()) {
    ElMessage.warning('请输入模板名称');
    return;
  }
  try {
    const res: any = await apiClient.put(`/settings/templates/${editingTpl.value.id}`, {
      name: editName.value.trim(),
      description: editDesc.value.trim(),
      category: editCategory.value.trim() || '自定义',
    });
    const idx = templates.value.findIndex(t => t.id === editingTpl.value?.id);
    if (idx >= 0) {
      templates.value[idx] = { ...templates.value[idx], ...res.data, enabled: res.data.enabled ?? true };
    }
    applyFilter();
    showEditDialog.value = false;
    ElMessage.success('模板已更新');
  } catch {
    ElMessage.error('保存失败');
  }
}

onMounted(() => {
  loadTemplates().then(applyFilter);
});
</script>

<template>
  <div class="template-view">
    <!-- 页面头部卡片 -->
    <div class="page-header-card">
      <div class="header-content">
        <div class="header-info">
          <div class="breadcrumb">
            <span>首页</span><span class="sep">/</span>
            <span>设置中心</span><span class="sep">/</span>
            <span class="current">模板管理</span>
          </div>
          <div class="header-title">
            <h1>模板管理</h1>
            <p>维护标书章节模板与排版模板，可在编辑文书时一键套用</p>
          </div>
        </div>
        <button class="reset-button" @click="showResetDialog = true">
          <span class="icon ri-restart-line"></span>
          恢复默认
        </button>
      </div>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <button class="btn-add" @click="showAddDialog = true">
          <span class="icon ri-add-line"></span>
          <span>新增模板</span>
        </button>
      </div>
      <div class="toolbar-right">
        <UiInput
          v-model="searchQuery"
          type="search"
          class="toolbar-search"
          placeholder="搜索模板名称或描述"
          clearable
          @update:modelValue="applyFilter"
        />
        <div class="view-toggle">
          <UiTip text="列表视图">
            <button class="view-btn" :class="{ active: viewMode === 'list' }" @click="viewMode = 'list'">
              <span class="icon ri-layout-column-line"></span>
            </button>
          </UiTip>
          <UiTip text="网格视图">
            <button class="view-btn" :class="{ active: viewMode === 'grid' }" @click="viewMode = 'grid'">
              <span class="icon ri-layout-grid-line"></span>
            </button>
          </UiTip>
        </div>
      </div>
    </div>

    <!-- 列表视图 -->
    <div v-if="viewMode === 'list'" class="list-container">
      <div class="table-card">
        <table class="list-table">
          <thead>
            <tr>
              <th class="col-name">模板名称</th>
              <th class="col-desc">模板描述</th>
              <th class="col-action">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="tpl in filteredTemplates" :key="tpl.id" class="list-row">
              <td class="col-name">
                <span class="row-name">{{ tpl.name }}</span>
                <span class="row-category">{{ tpl.category }}</span>
              </td>
              <td class="col-desc">
                <span class="row-desc">{{ tpl.description }}</span>
              </td>
              <td class="col-action">
                <div class="row-actions">
                  <button class="action-btn edit" @click="openEdit(tpl)">修改</button>
                  <button class="action-btn delete" @click="openDelete(tpl)">删除</button>
                  <UiSwitch :model-value="tpl.enabled" size="sm" @update:model-value="toggleEnabled(tpl)" />
                </div>
              </td>
            </tr>
            <tr v-if="filteredTemplates.length === 0" class="empty-row">
              <td colspan="3" class="empty-cell">暂无模板</td>
            </tr>
          </tbody>
        </table>
        <div class="table-footer">
          共 {{ filteredTemplates.length }} 个模板
        </div>
      </div>
    </div>

    <!-- 网格视图 -->
    <div v-else class="grid-container">
      <div class="grid-cards">
        <div
          v-for="tpl in filteredTemplates"
          :key="tpl.id"
          class="grid-card"
        >
          <div class="grid-card-body">
            <div class="grid-card-icon">{{ tpl.category }}</div>
            <div class="grid-card-title">{{ tpl.name }}</div>
            <div class="grid-card-desc">{{ tpl.description }}</div>
          </div>
          <div class="grid-card-footer">
            <div class="grid-card-actions">
              <button class="action-btn edit" @click="openEdit(tpl)">修改</button>
              <button class="action-btn delete" @click="openDelete(tpl)">删除</button>
              <UiSwitch :model-value="tpl.enabled" size="sm" @update:model-value="toggleEnabled(tpl)" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 新增模板弹窗 -->
    <UiModal
      v-model="showAddDialog"
      title="新增模板"
      subtitle="填写模板基本信息，创建后可在列表中启用"
      icon="ri-add-line"
      :width="560"
      @confirm="handleAdd"
    >
      <div class="form-group">
        <label class="form-label">模板名称 <span class="required">*</span></label>
        <UiInput v-model="newTemplateName" placeholder="请输入模板名称" />
      </div>
      <div class="form-group">
        <label class="form-label">模板描述</label>
        <UiInput v-model="newTemplateDesc" multiline :rows="3" placeholder="请输入模板描述（可选）" />
      </div>
      <div class="form-group">
        <label class="form-label">所属分类</label>
        <UiInput v-model="newTemplateCategory" placeholder="例如：商务标、技术标" />
      </div>
    </UiModal>

    <!-- 编辑模板弹窗 -->
    <UiModal
      v-model="showEditDialog"
      title="编辑模板"
      subtitle="修改模板基本信息"
      icon="ri-edit-line"
      :width="560"
      @confirm="confirmEdit"
    >
      <div class="form-group">
        <label class="form-label">模板名称 <span class="required">*</span></label>
        <UiInput v-model="editName" placeholder="请输入模板名称" />
      </div>
      <div class="form-group">
        <label class="form-label">模板描述</label>
        <UiInput v-model="editDesc" multiline :rows="3" placeholder="请输入模板描述（可选）" />
      </div>
      <div class="form-group">
        <label class="form-label">所属分类</label>
        <UiInput v-model="editCategory" placeholder="例如：商务标、技术标" />
      </div>
    </UiModal>
    <!-- 删除模板确认弹窗 -->
    <UiModal
      v-model="showDeleteDialog"
      title="删除模板"
      subtitle="删除后不可恢复，请谨慎操作"
      icon="ri-delete-bin-line"
      :width="480"
      confirm-text="确认删除"
      @confirm="confirmDelete"
    >
      <p class="confirm-text">确定删除模板「{{ deletingTpl?.name }}」？</p>
    </UiModal>

    <!-- 恢复默认确认弹窗 -->
    <UiModal
      v-model="showResetDialog"
      title="恢复默认模板"
      subtitle="自定义模板将被清除"
      icon="ri-restart-line"
      :width="480"
      confirm-text="确认恢复"
      @confirm="confirmReset"
    >
      <p class="confirm-text">确定恢复默认模板？自定义模板将被清除。</p>
    </UiModal>
  </div>
</template>

<style scoped>
.template-view {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-width: 1620px;
}

/* 页面头部 */
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
.header-info { flex: 1; display: flex; flex-direction: column; gap: 4px; }
.breadcrumb { display: flex; align-items: center; gap: 8px; font-size: 14px; font-family: "SourceHanSans-Regular", var(--font-family-cn); }
.breadcrumb span { color: var(--color-text-faint, #A7A296); }
.breadcrumb span.sep { color: var(--color-text-faint, #A7A296); }
.breadcrumb span.current { color: var(--color-text-body, #3A3A3A); font-weight: 500; }
.header-title h1 { font-size: 30px; font-weight: 700; color: var(--color-text-primary, #1F1D1A); margin: 12px 0 0 0; letter-spacing: -0.5px; font-family: "SourceHanSans-Bold", var(--font-family-cn); line-height: 1.2; }
.header-title p { font-size: 15px; color: var(--color-text-secondary, #6E6A60); margin: 8px 0 0 0; font-family: "SourceHanSans-Regular", var(--font-family-cn); }
.reset-button {
  display: inline-flex; align-items: center; gap: 8px; padding: 0 20px; height: 48px;
  border: 0.7px solid var(--color-border, #E5DFD3); background-color: var(--color-bg-white, #FFFFFF); border-radius: 12px;
  font-size: 15px; color: var(--color-text-body, #3A3A3A); cursor: pointer; transition: all 0.2s; flex-shrink: 0;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
.reset-button:hover { background-color: var(--color-bg-secondary, #F7F3EA); }
.reset-button .icon { font-family: "remixicon", sans-serif; font-style: normal; font-size: 18px; }

/* 工具栏 */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.toolbar-left { display: flex; gap: 12px; align-items: center; }
.toolbar-right { display: flex; align-items: center; gap: 12px; }
.btn-add {
  display: inline-flex; align-items: center; gap: 8px; padding: 0 20px; height: 46px;
  border: none; border-radius: 10px; background: var(--color-primary, #C0392B); color: #fff;
  font-size: 16px; font-weight: 500; cursor: pointer;
  font-family: "SourceHanSans-Medium", var(--font-family-cn); transition: all 0.2s;
}
.btn-add:hover { background: color-mix(in srgb, var(--color-primary, #C0392B) 88%, black); }
.btn-add .icon { font-family: "remixicon", sans-serif; font-style: normal; font-size: 20px; }
.toolbar-search { width: 548px; }
.view-toggle {
  display: flex; border: 0.7px solid var(--color-border, #E5DFD3); border-radius: 12px; padding: 4px;
  background: var(--color-bg-white, #fff); gap: 4px;
}
.view-btn {
  width: 54px; height: 38px; border: none; border-radius: 9px; background: transparent;
  display: flex; align-items: center; justify-content: center; cursor: pointer; transition: all 0.2s;
}
.view-btn .icon {
  font-family: "remixicon", sans-serif; font-style: normal; font-size: 18px; color: var(--color-text-muted, #8A8578);
}
.view-btn.active { background: #F6EEDA; }
.view-btn.active .icon { color: var(--color-primary, #C0392B); }

/* 列表视图 */
.list-container { background-color: var(--color-bg-white, #FFFFFF); border: 0.7px solid var(--color-border-soft, #EDE9E0); border-radius: 16px; overflow: hidden; }
.table-card { background: var(--color-bg-white, #fff); }
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

/* 操作按钮（与技能管理一致） */
.row-actions .action-btn {
  font-size: 15px;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  transition: opacity 0.2s;
}
.row-actions .action-btn.edit {
  color: var(--color-info, #4A72A8);
}
.row-actions .action-btn.delete {
  color: var(--color-error, #E74C3C);
}
.row-actions .action-btn:hover {
  opacity: 0.7;
}

/* 开关已统一为 UiSwitch，样式见 components/ui */
.empty-row td { text-align: center; vertical-align: middle; color: var(--color-text-faint, #A7A296); font-size: 14px; padding: 40px; }
.table-footer {
  padding: 0 40px; height: 54px; display: flex; align-items: center;
  font-size: 14px; color: var(--color-text-faint, #A7A296); font-family: "SourceHanSans-Regular", var(--font-family-cn);
}

/* 网格视图 */
.grid-container { display: flex; flex-direction: column; gap: 20px; }
.grid-cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
.grid-card {
  border: 0.7px solid var(--color-border-soft, #EDE9E0); border-radius: 14px; background: var(--color-bg-white, #fff);
  overflow: hidden; display: flex; flex-direction: column;
  transition: box-shadow 0.2s; min-height: 220px;
}
.grid-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.08); }
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

@media (max-width: 1200px) { .grid-cards { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) { .grid-cards { grid-template-columns: 1fr; } }

/* 弹窗表单（壳已统一为 UiModal） */
.form-group { display: flex; flex-direction: column; gap: 8px; }
.form-label { font-size: 14px; color: var(--color-text-body, #3A3A3A); font-weight: 500; font-family: "SourceHanSans-Regular", var(--font-family-cn); }
.form-label .required { color: var(--color-primary, #C0392B); }
.form-group .ui-input { width: 100%; }

/* 确认类弹窗正文 */
.confirm-text {
  font-size: 15px;
  color: var(--color-text-body, #3A3A3A);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  margin: 0;
  line-height: 1.7;
}
</style>
