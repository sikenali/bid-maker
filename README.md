# Bid-Maker - AI 标书编写系统

基于技术方案的 AI 标书编写系统后端，支持从招标文件解析到投标文件导出的全流程智能化。

## 技术架构

### 四层分离
- **解析层**: officeparser Python 微服务（待接入）
- **语义层**: Skill 注册表 + 大纲状态机
- **生成层**: 模型网关 + 按纲补写
- **格式层**: Markdown → Word 导出

### 核心组件

| 组件 | 说明 |
|------|------|
| LLMGateway | 统一OpenAI兼容网关，支持多提供商路由 |
| SkillRegistry | Skill插件注册表 (Parse/Outline/Generate/Export/RAG) |
| TaskQueue | 异步任务队列，支持断点续跑 |
| SessionStore | 会话存储，持久化标书状态 |

### API 接口

```
POST   /api/sessions              # 创建会话
GET    /api/sessions/:id          # 获取会话
PUT    /api/sessions/:id          # 更新会话
POST   /api/sessions/:id/skills/parse     # 解析招标文件
POST   /api/sessions/:id/skills/outline   # 生成大纲(双通道)
PUT    /api/sessions/:id/outline        # 更新大纲
POST   /api/sessions/:id/skills/generate/section  # 逐章生成
POST   /api/sessions/:id/skills/generate/full     # 全篇生成
GET    /api/sessions/:id/skills/generate/stream   # 流式生成
GET    /api/sessions/:id/export           # 导出Word
GET    /api/sessions/:id/export/markdown  # 获取Markdown
GET    /api/skills                    # 列出Skills
GET    /api/tasks/:taskId             # 查询任务状态
GET    /api/models                    # 列出模型
POST   /api/test-key                  # 测试API Key
```

## 启动方式

### 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 配置 LLM_PROVIDERS
```

### 启动服务

```bash
go run ./cmd/server/
```

## 数据结构

### BidParsingOutput (解析层输出)
```json
{
  "source_meta": { "title", "bid_no", "issuer", "file_hash", ... },
  "template": { "has_preset_structure", "preset_sections", "format_rules" },
  "technical_indicators": [...],
  "commercial_reqs": [...],
  "scoring_points": [...],
  "disqualify_clauses": [...]
}
```

### BidOutline (规划层大纲)
```json
{
  "outline_id": "...",
  "source_mode": "auto | custom | hybrid",
  "locked": false,
  "sections": [{
    "section_id": "sec-01",
    "title": "投标函",
    "level": 1,
    "source": "auto | manual | edited",
    "requirements": [{ "ref": "tech-01", "type": "technical_indicator" }],
    "response_strategy": "satisfy | deviate | exceed",
    "status": "pending | generating | generated | failed"
  }]
}
```

## 后续计划

- [ ] 接入 Python officeparser 微服务
- [ ] 实现 RAG 知识库检索
- [ ] 完善 md → Word 模板导出
- [ ] 添加合规检查集成
- [ ] 前端交互统一调整
