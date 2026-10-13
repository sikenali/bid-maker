<script setup lang="ts">
import { BRAND_NAME, BRAND_TITLE } from '@/constants/brand';

const currentYear = new Date().getFullYear();
// 应用版本号：当前为常量，后续打包时从软件版本提取显示
const appVersion = 'V1.0.0';
</script>

<template>
  <div class="about-view">
    <!-- 页面头部卡片 -->
    <div class="page-header-card">
      <div class="header-content">
        <div class="header-info">
          <div class="breadcrumb">
            <span>首页</span>
            <span class="sep">/</span>
            <span>设置中心</span>
            <span class="sep">/</span>
            <span class="current">关于</span>
          </div>
          <div class="header-title">
            <h1>关于</h1>
            <p>版本信息、开源仓库与版权声明</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 品牌信息卡片 -->
    <div class="about-card">
      <div class="about-badge">
        <RiTargetLine class="badge-logo" />
      </div>

      <div class="about-name">{{ BRAND_TITLE }}</div>

      <div class="about-slogan">智能化文件提取工具</div>

      <div class="about-version">版本: {{ appVersion }}</div>

      <a class="about-repo" href="https://github.com/sikenali/bid-extractor" target="_blank" rel="noopener noreferrer" title="访问 GitHub 仓库">
        <span class="repo-right">
          <RiGithubLine class="icon" />
        </span>
      </a>

      <div class="about-copy">© {{ currentYear }} {{ BRAND_NAME }} · sikenali</div>
    </div>
  </div>
</template>

<style scoped>
.about-view {
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

/* 品牌信息卡片 */
.about-card {
  background-color: var(--color-bg-white, #FFFFFF);
  border: 0.7px solid var(--color-border-soft, #EDE9E0);
  border-radius: 16px;
  height: 525px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}
.about-badge {
  width: 118px;
  height: 118px;
  border-radius: 50%;
  border: 0.7px solid #E8D9B0;
  background-color: #FBF5E3;
  display: flex;
  align-items: center;
  justify-content: center;
}
.badge-logo {
  font-family: "remixicon", sans-serif;
  font-style: normal;
  font-size: 52px;
  line-height: 1;
  color: #BE3B2E;
}
.about-name {
  margin-top: 32px;
  font-size: 24px;
  font-weight: 700;
  color: var(--color-text-primary, #1F1D1A);
  font-family: "SourceHanSans-Bold", var(--font-family-cn);
  line-height: 1.2;
}
.about-slogan {
  margin-top: 16px;
  font-size: 16px;
  color: var(--color-text-secondary, #6E6A60);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  line-height: 1.4;
}
.about-version {
  margin-top: 8px;
  font-size: 14px;
  color: var(--color-text-muted, #8A8578);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
  line-height: 1.4;
}
.about-repo {
  margin-top: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  text-decoration: none;
  color: inherit;
}
.about-repo:hover {
  opacity: 0.75;
}
.repo-right {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  gap: 8px;
}
.repo-right .icon {
  font-family: "remixicon", sans-serif;
  font-style: normal;
  font-size: 34px;
  color: var(--color-text-secondary, #6E6A60);
}
.about-copy {
  margin-top: 32px;
  font-size: 14px;
  color: var(--color-text-faint, #A7A296);
  font-family: "SourceHanSans-Regular", var(--font-family-cn);
}
</style>