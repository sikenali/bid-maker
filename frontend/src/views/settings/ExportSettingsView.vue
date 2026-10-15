<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { ElMessage } from '@/utils/toast';
import { getExportSettings, setExportSettings } from '@/api/settings';
import { UiSwitch, UiSelectCard } from '@/components/ui';
import { BRAND_NAME } from '@/constants/brand';

interface ExportItem {
  key: string;
  name: string;
  desc: string;
  iconText: string;
  iconColor: string;
}

const formats: ExportItem[] = [
  { key: 'docx', name: 'Word 文档 ·.docx', desc: '保留可编辑正文，商用文库首选', iconText: 'W', iconColor: '#2B579A' },
  { key: 'pdf', name: 'PDF 文档 ·.pdf', desc: '矢量高清，适用于打印与送审', iconText: 'PDF', iconColor: '#C0392B' },
  { key: 'markdown', name: 'Markdown ·.md', desc: '轻盈纯文本，便于二次编辑', iconText: 'MD', iconColor: '#3A4A63' },
];

const wordSettings = ref({
  preserveTemplate: true,
  autoTable: true,
  embedFont: false,
});

const pdfSettings = ref({
  vectorHD: true,
  keepBookmarks: true,
  addWatermark: false,
});

const mdSettings = ref({
  preserveTable: true,
  downgradeTitle: false,
});

const selectedFormat = ref('docx');

let ready = false;
const num = (v: boolean) => (v ? 1 : 0);

async function persist() {
  if (!ready) return;
  try {
    await setExportSettings({
      format: selectedFormat.value,
      word_keep_style: num(wordSettings.value.preserveTemplate),
      word_toc: num(wordSettings.value.autoTable),
      word_embed_font: num(wordSettings.value.embedFont),
      pdf_high_res: num(pdfSettings.value.vectorHD),
      pdf_bookmarks: num(pdfSettings.value.keepBookmarks),
      pdf_watermark: num(pdfSettings.value.addWatermark),
      md_keep_tables: num(mdSettings.value.preserveTable),
      md_demote_heading: num(mdSettings.value.downgradeTitle),
    });
  } catch {
    ElMessage.error('保存导出设置失败');
  }
}

function applyLoaded(s: Record<string, any>) {
  selectedFormat.value = s.format || 'docx';
  wordSettings.value = {
    preserveTemplate: (s.word_keep_style ?? 1) === 1,
    autoTable: (s.word_toc ?? 1) === 1,
    embedFont: (s.word_embed_font ?? 0) === 1,
  };
  pdfSettings.value = {
    vectorHD: (s.pdf_high_res ?? 1) === 1,
    keepBookmarks: (s.pdf_bookmarks ?? 1) === 1,
    addWatermark: (s.pdf_watermark ?? 0) === 1,
  };
  mdSettings.value = {
    preserveTable: (s.md_keep_tables ?? 1) === 1,
    downgradeTitle: (s.md_demote_heading ?? 0) === 1,
  };
}

onMounted(async () => {
  try {
    applyLoaded(await getExportSettings());
  } catch {
    // 后端不可用时保持本地默认值
    ElMessage.warning('后端服务不可用，导出设置暂用本地默认');
  } finally {
    ready = true;
  }
});

watch([selectedFormat, wordSettings, pdfSettings, mdSettings], persist, { deep: true });

async function resetAll() {
  applyLoaded({
    format: 'docx',
    word_keep_style: 1, word_toc: 1, word_embed_font: 0,
    pdf_high_res: 1, pdf_bookmarks: 1, pdf_watermark: 0,
    md_keep_tables: 1, md_demote_heading: 0,
  });
  ready = true;
  await persist();
  ElMessage.success('已恢复默认导出设置');
}
</script>

<template>
  <div class="export-view">
    <!-- 页面头部卡片 -->
    <div class="page-header-card">
      <div class="header-content">
        <div class="header-info">
          <div class="breadcrumb">
            <span>首页</span>
            <span class="sep">/</span>
            <span>设置中心</span>
            <span class="sep">/</span>
            <span class="current">导出管理</span>
          </div>
          <div class="header-title">
            <h1>导出管理</h1>
            <p>配置 Word、PDF、Markdown 三类成品的输出规格与默认行为</p>
          </div>
        </div>
        <button class="reset-button" @click="resetAll">
          <RiRestartLine class="icon" />
          恢复默认
        </button>
      </div>
    </div>

    <!-- 默认导出格式 -->
    <div class="section-card">
      <div class="card-title">默认导出格式</div>
      <div class="format-group">
        <div v-for="fmt in formats" :key="fmt.key" class="format-cell">
          <UiSelectCard
            :selected="selectedFormat === fmt.key"
            :title="fmt.name"
            :desc="fmt.desc"
            :icon-text="fmt.iconText"
            :icon-color="fmt.iconColor"
            @select="selectedFormat = fmt.key"
          />
        </div>
      </div>
    </div>

    <!-- Word 导出设置 -->
    <div class="section-card">
      <div class="card-title">Word 导出设置</div>
      <div class="setting-item">
        <div class="setting-info">
          <div class="setting-name">保留原始模板样式</div>
          <div class="setting-desc">导出时沿用模板的页眉页脚与标题层级</div>
        </div>
        <UiSwitch v-model="wordSettings.preserveTemplate" />
      </div>
      <div class="setting-item">
        <div class="setting-info">
          <div class="setting-name">自动生成目录与页码</div>
          <div class="setting-desc">按二级标题生成可跳转目录并标注页码</div>
        </div>
        <UiSwitch v-model="wordSettings.autoTable" />
      </div>
      <div class="setting-item">
        <div class="setting-info">
          <div class="setting-name">嵌入正文所用字体</div>
          <div class="setting-desc">避免他人电脑打开时出现字体替换</div>
        </div>
        <UiSwitch v-model="wordSettings.embedFont" />
      </div>
    </div>

    <!-- PDF 导出设置 -->
    <div class="section-card">
      <div class="card-title">PDF 导出设置</div>
      <div class="setting-item">
        <div class="setting-info">
          <div class="setting-name">矢量高清输出 300dpi</div>
          <div class="setting-desc">文字与图片均不压缩，适合打印装订后送审</div>
        </div>
        <UiSwitch v-model="pdfSettings.vectorHD" />
      </div>
      <div class="setting-item">
        <div class="setting-info">
          <div class="setting-name">保留书签导航树</div>
          <div class="setting-desc">按章节层级生成左侧书签，评委可快速跳转</div>
        </div>
        <UiSwitch v-model="pdfSettings.keepBookmarks" />
      </div>
      <div class="setting-item">
        <div class="setting-info">
          <div class="setting-name">添加内部稿水印</div>
          <div class="setting-desc">叠加「{{ BRAND_NAME }} · 内部稿」对角淡色水印</div>
        </div>
        <UiSwitch v-model="pdfSettings.addWatermark" />
      </div>
    </div>

    <!-- Markdown 导出设置 -->
    <div class="section-card">
      <div class="card-title">Markdown 导出设置</div>
      <div class="setting-item">
        <div class="setting-info">
          <div class="setting-name">保留表格与图片链接</div>
          <div class="setting-desc">表格转为 Markdown 语法，图片保留相对路径</div>
        </div>
        <UiSwitch v-model="mdSettings.preserveTable" />
      </div>
      <div class="setting-item">
        <div class="setting-info">
          <div class="setting-name">一级标题自动降级</div>
          <div class="setting-desc">便于直接粘贴到团队 Wiki 的文档树中</div>
        </div>
        <UiSwitch v-model="mdSettings.downgradeTitle" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.export-view {
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

/* 通用卡片 */
.section-card {
  background-color: var(--color-bg-white, #FFFFFF);
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  border-radius: 16px;
  padding: 24px 40px;
}
.card-title {
  font-size: 17px;
  font-weight: 500;
  color: var(--color-text-primary, #1F1D1A);
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
  display: flex;
  align-items: center;
  gap: 10px;
}
.coming-badge {
  font-size: 12px;
  font-weight: 400;
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  color: var(--color-text-muted, #8A8578);
  background: var(--color-bg-secondary, #F7F3EA);
  border-radius: 9999px;
  padding: 3px 10px;
  white-space: nowrap;
}

/* 默认导出格式 */
.format-group {
  display: flex;
  gap: 24px;
  margin-top: 20px;
  align-items: stretch;
}
.format-cell { flex: 1; display: flex; }
.format-cell > * { flex: 1; }

/* 设置项 */
.setting-item {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  height: 76px;
  gap: 20px;
}
.setting-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 8px;
}
.setting-name {
  font-size: 16px;
  font-weight: 500;
  color: var(--color-text-primary, #1F1D1A);
  font-family: "SourceHanSans-Medium", var(--font-family-cn);
}
.setting-desc {
  font-size: 15px;
  color: var(--color-text-muted, #8A8578);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
</style>