<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { UiSwitch, UiSelect } from '@/components/ui';
import { ElMessage } from '@/utils/toast';
import { RiInformationLine } from '@remixicon/vue';

interface ProfileSettings {
  autoRedirectAfterUpload: boolean;
  defaultPageSize: number;
  extractPreviewMode: 'table' | 'card';
  autoDownloadOnComplete: boolean;
}

const settings = ref<ProfileSettings>({
  autoRedirectAfterUpload: true,
  defaultPageSize: 20,
  extractPreviewMode: 'table',
  autoDownloadOnComplete: false,
});

const STORAGE_KEY = 'bid-maker-profile-settings';

function loadSettings() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) settings.value = { ...settings.value, ...JSON.parse(raw) };
  } catch { /* keep defaults */ }
}

function saveSettings() {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings.value));
    ElMessage.success('个性设置已保存');
  } catch {
    ElMessage.error('保存失败');
  }
}

function update<K extends keyof ProfileSettings>(key: K, value: ProfileSettings[K]) {
  settings.value[key] = value;
  saveSettings();
}

onMounted(loadSettings);
</script>

<template>
  <div class="profile-view">
    <div class="section-card">
      <div class="card-header">
        <div class="card-title">通用设置</div>
        <div class="card-desc">上传与展示行为的个性化配置</div>
      </div>

      <div class="setting-row">
        <div class="row-left">
          <div class="row-title">上传完成后自动跳转</div>
          <div class="row-hint">上传成功后立即进入结果页</div>
        </div>
        <UiSwitch :model-value="settings.autoRedirectAfterUpload" @update:model-value="update('autoRedirectAfterUpload', $event)" />
      </div>

      <div class="setting-row">
        <div class="row-left">
          <div class="row-title">结果页默认每页条数</div>
          <div class="row-hint">影响列表类页面的分页粒度</div>
        </div>
        <UiSelect
          :model-value="settings.defaultPageSize"
          :options="[{ value: 10, label: '10' }, { value: 20, label: '20' }, { value: 50, label: '50' }, { value: 100, label: '100' }]"
          @update:model-value="update('defaultPageSize', $event as number)"
        />
      </div>
    </div>

    <div class="section-card">
      <div class="card-header">
        <div class="card-title">视图偏好</div>
        <div class="card-desc">默认视图模式与导出行为</div>
      </div>

      <div class="setting-row">
        <div class="row-left">
          <div class="row-title">提取结果默认视图</div>
          <div class="row-hint">表格视图 / 卡片视图</div>
        </div>
        <UiSelect
          :model-value="settings.extractPreviewMode"
          :options="[{ value: 'table', label: '表格' }, { value: 'card', label: '卡片' }]"
          @update:model-value="update('extractPreviewMode', $event as 'table' | 'card')"
        />
      </div>

      <div class="setting-row">
        <div class="row-left">
          <div class="row-title">处理完成自动下载</div>
          <div class="row-hint">任务完成时自动下载导出结果</div>
        </div>
        <UiSwitch :model-value="settings.autoDownloadOnComplete" @update:model-value="update('autoDownloadOnComplete', $event)" />
      </div>
    </div>

    <div class="footer-tip">
      <RiInformationLine size="16" />
      <span>所有设置本地持久化，切换主题不丢失。</span>
    </div>
  </div>
</template>

<style scoped>
.profile-view {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.section-card {
  background: var(--color-bg-white, #FFFFFF);
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  border-radius: 14px;
  padding: 20px 24px;
}

.card-header {
  margin-bottom: 16px;
  padding-bottom: 14px;
  border-bottom: 0.7px solid var(--color-border-soft, #EDE9E0);
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text-primary, #1F1D1A);
  margin-bottom: 4px;
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
}

.card-desc {
  font-size: 13px;
  color: var(--color-text-muted, #8A8578);
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
  border-bottom: 0.7px solid var(--color-border-soft, #EDE9E0);
  gap: 20px;
}

.setting-row:last-child {
  border-bottom: none;
}

.row-left {
  flex: 1;
  min-width: 0;
}

.row-title {
  font-size: 14px;
  color: var(--color-text-body, #3A3A3A);
  margin-bottom: 3px;
}

.row-hint {
  font-size: 12px;
  color: var(--color-text-muted, #8A8578);
  line-height: 1.4;
}

.footer-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: var(--color-bg-secondary, #F7F3EA);
  border-radius: 10px;
  color: var(--color-text-muted, #8A8578);
  font-size: 13px;
}
</style>
