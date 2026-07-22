# Outline Tree 递归层级展示设计

**日期:** 2026-07-22
**状态:** 待审核

## 背景

当前 OutlineTree 组件仅渲染 2 层（顶层 section + 一层 children），而招标/投标文件的大纲结构通常包含 3-4 层及以上（如 章→节→条→款）。后端 `extractSections` 已通过 parentStack 支持无限嵌套，但前端渲染能力不足。

## 改动范围

仅前端 `OutlineTree.vue`，后端无变化。

## 实现方案

### 1. 递归组件

将当前固定双层模板改为递归组件。Vue 3 支持组件在自身模板中引用自身（需设置 `name` 或使用 `__name`）。

**组件结构：**
```
OutlineTree.vue
  └── OutlineTreeNode (递归)
        ├── <div.outline-item>   (当前节点)
        │     ├── icon (按层级显示不同图标)
        │     ├── title
        │     └── context-menu
        └── OutlineTreeNode (递归 children)
```

### 2. 层级视觉设计

| Level | 缩进 | 字号 | 字重 | 左侧装饰 |
|-------|------|------|------|---------|
| 1 (章) | 16px | 14px | 600 | 无 |
| 2 (节) | 46px | 13px | 500 | 灰色竖线 |
| 3 (条) | 76px | 13px | 400 | 灰色竖线 |
| 4+ (款) | 106px+ | 12px | 400 | 灰色竖线 |

每层缩进按 `16 + (level - 1) * 30` px 计算。

### 3. 功能保留

- 滑动激活指示器（适配任意层级）
- 右键菜单：降级 / 新增 / 删除（操作自身节点）
- 键盘导航：Tab 切换、Escape 关闭
- 点击外部关闭菜单
- 顶部"添加章节"按钮（添加顶层章节）

### 4. 激活指示器适配

当前基于固定 DOM 结构计算位置。改为递归后，`updateIndicator` 仍通过 `data-id` 属性查找目标元素，无需大改。

## 数据流

```
后端 extractSections 返回嵌套 Section[]
  → documentStore.outline (Section[])
    → OutlineTree 递归渲染
      → 点击选中 -> emit select(id) -> documentStore.loadSection()
```

## 不做的事情

- 不改后端代码
- 不改 Section 数据模型
- 不引入新依赖
- 不改变父组件 EditorView
