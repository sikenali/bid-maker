# Bid-Maker — AI 智能标书编写系统

基于技术方案的 AI 标书编写系统，支持从招标文件解析、大纲生成、逐章补写到 Word 导出的全流程智能化。

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 API | Go · Gin · UUID |
| 解析服务 | Node.js · express · officeparser |
| 前端 | Vue 3 · TypeScript · Vite · Pinia · Tiptap |
| 导出引擎 | docxtemplater · pizzip · file-saver |

## 架构

```
┌─────────────┐    HTTP      ┌──────────────┐    HTTP      ┌──────────────┐
│   Frontend   │ ◄─────────► │  Go Backend   │ ◄─────────► │ Parse Service │
│  Vue 3 + TS  │  :8080/proxy│  (Gin)        │  :8000      │ (Node.js)     │
│             │             │              │             │ officeparser  │
└─────────────┘             └──────────────┘             └──────────────┘
      │                           │                            │
      │ localStorage              │ SessionStore               │
      │ (theme / apiKeys /        │ (in-memory)                │
      │  templates / skills)      │                            │
```

### 核心流程

1. **上传** — 用户上传招标文件（.docx/.pdf/.xlsx/.md/.txt）
2. **解析** — Node.js `parse-service` 调用 `officeparser` 提取标题、技术指标、商务要求、评分点、废标条款
3. **大纲** — 根据解析结果或用户输入，生成结构化大纲（auto/custom/hybrid 三通道）
4. **生成** — 按章节调用 LLM 流式生成标书内容
5. **编辑** — 支持 Markdown 模式（Tiptap）和 Word 模式（DocxEditor）双编辑器
6. **导出** — 将 Markdown 内容渲染为专业 Word 文档（docxtemplater 模板引擎）

## API 接口

```
POST   /api/sessions                  # 创建会话
GET    /api/sessions/:id              # 获取会话
PUT    /api/sessions/:id              # 更新会话
POST   /api/sessions/:id/skills/parse # 解析招标文件
POST   /api/sessions/:id/skills/outline # 生成大纲
PUT    /api/sessions/:id/outline      # 更新大纲
POST   /api/sessions/:id/outline/lock # 锁定大纲
POST   /api/sessions/:id/skills/generate/section # 逐章生成
POST   /api/sessions/:id/skills/generate/full    # 全篇生成
GET    /api/sessions/:id/skills/generate/stream  # 流式生成
GET    /api/sessions/:id/export             # 导出 Word
GET    /api/sessions/:id/export/markdown    # 获取 Markdown
GET    /api/skills                        # 列出 Skills
GET    /api/tasks/:taskId                 # 查询任务状态
GET    /api/models                        # 列出模型
POST   /api/test-key                      # 测试 API Key
POST   /api/chat                          # AI 对话
GET    /api/local-skills                  # 扫描本地技能
```

## 启动方式

### 安装依赖

```bash
# 根目录
npm i

# 解析服务（Node.js，端口 8000）
cd parse-service && npm i && cd ..

# 前端（Vite，端口 3000，可用 FRONTEND_PORT 覆盖）
cd frontend && npm i && cd ..

# 后端（Go，端口 8080）
cd backend && go mod download && cd ..
```

### 配置环境变量

```bash
cp .env.example .env
# 编辑 .env：
#   PARSE_SERVICE=http://localhost:8000   解析服务地址
#   LLM_PROVIDERS=[...]                   LLM 提供商列表（JSON）
```

### 启动服务

```bash
# 一键启动全部（开发模式，含热更新）
npm run dev

# 单独启动
npm run dev:parse     # Node.js 解析服务
npm run dev:backend   # Go 后端
npm run dev:frontend  # Vue 前端
```

### 生产部署说明

| 服务 | 端口(默认) | 说明 |
|------|-----------|------|
| 解析服务 | 8000 | `PORT=8001 node server.js` 可改端口 |
| Go 后端 | 8080 | `PORT=8081` + `PARSE_SERVICE=http://localhost:8000` |
| 前端开发 | 3000 | `FRONTEND_PORT=3001 npm run dev`；后端地址 `BACKEND_URL=http://localhost:8080` |

注意：
- `vite preview` / 静态托管**没有** `/api` 代理，生产环境须经网关或反代把 `/api/*` 指到 Go 后端（8080），否则前端请求 404。
- 本机 3000 端口如被其他项目占用，用 `FRONTEND_PORT` 另起端口。

```nginx
# 反代示例
location /api/ { proxy_pass http://127.0.0.1:8080; }
location / { root /var/www/bid-maker/dist; try_files $uri $uri/ /index.html; }
```

## 数据结构

### 解析输出 (BidParsingOutput)

```json
{
  "source_meta": { "title", "bid_no", "issuer", "file_hash", "total_pages" },
  "template": { "has_preset_structure", "preset_sections", "format_rules" },
  "technical_indicators": [
    { "indicator_id", "name", "requirement", "is_mandatory", "unit", "min_value" }
  ],
  "commercial_reqs": [
    { "req_id", "category", "title", "description", "deliverable" }
  ],
  "scoring_points": [
    { "scoring_id", "scoring_item", "full_score", "grading_tiers" }
  ],
  "disqualify_clauses": [
    { "clause_id", "clause_text", "severity", "check_type" }
  ]
}
```

### 大纲 (BidOutline)

```json
{
  "outline_id": "...",
  "source_mode": "auto | custom | hybrid",
  "locked": false,
  "sections": [{
    "section_id": "sec-01",
    "title": "投标函",
    "level": 1,
    "parent_id": null,
    "source": "auto | manual | edited",
    "requirements": [{ "ref": "tech-01", "type": "technical_indicator" }],
    "response_strategy": "satisfy | deviate | exceed",
    "status": "pending | generating | generated | failed",
    "children": ["sec-02"]
  }]
}
```

## 前端目录结构

```
frontend/src/
├── api/                 # API 客户端
│   ├── client.ts        # 会话/大纲/生成/导出 API
│   ├── llm.ts           # LLM 网关
│   ├── settings.ts      # 设置 API（localStorage 持久化）
│   └── upload.ts        # 文件上传
├── components/
│   ├── layout/          # TopNav / SettingsModal / PageFooter
│   ├── upload/          # UploadCard
│   ├── settings/        # RuleDialog
│   ├── ui/              # UiInput / UiModal / UiSwitch 等基础组件
│   ├── EditorView.vue   # 主编辑器（大纲 + 双模式编辑器 + AI 助手）
│   ├── OutlineTree.vue  # 大纲树组件
│   └── GenerateFlowDialog.vue
├── stores/
│   ├── documentStore.ts # 文档状态
│   ├── generateStore.ts # 生成流程状态
│   ├── chatStore.ts     # AI 对话
│   └── settingsStore.ts # 主题 / 模型 / 技能 / 模板
├── views/
│   ├── UploadView.vue   # 上传页
│   ├── EditorView.vue   # 编辑页
│   ├── SettingsView.vue # 设置页入口
│   └── settings/        # 各设置子页面
└── utils/
    ├── markdownDocx.ts  # Markdown → DOCX 转换（docxtemplater 实现）
    └── toast.ts         # 消息提示
```

## 核心功能

| 功能 | 说明 |
|------|------|
| 文件解析 | 支持 docx/pdf/xlsx/pptx/md/txt/html，提取技术指标、商务要求、评分点、废标条款 |
| 大纲管理 | auto（纯 AI）/ custom（纯用户输入）/ hybrid（结合解析结果）三种生成模式 |
| 逐章生成 | SSE 流式输出，实时预览，进度追踪 |
| 全文生成 | 自动按大纲顺序逐章生成，支持断点续跑 |
| 双模式编辑 | Markdown（Tiptap）/ Word（DocxEditor）实时切换，双向同步 |
| AI 助手 | 支持多模型路由，带上下文感知的智能问答 |
| 技能系统 | 本地技能扫描 + 自定义技能导入，支持 `/` 触发命令面板 |
| 模板管理 | 内置标准投标/技术方案/商务标模板，支持自定义上传 |
| 导出 | Markdown / Word 双格式，专业排版样式 |

## 后续计划

- [ ] 实现 RAG 知识库检索
- [ ] 添加合规检查集成
- [ ] 前端交互统一调整
- [ ] 多语言支持
