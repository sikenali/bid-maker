package service

import "fmt"

// LLMProviderConfig 模型提供商配置
type LLMProviderConfig struct {
	Name      string `json:"name"`
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	ModelName string `json:"model_name"`
	Type      string `json:"type"`
}

// Skill 接口定义
type Skill interface {
	Name() string
	Description() string
	Execute(ctx Context, input map[string]any) (map[string]any, error)
}

// Context Skill执行上下文
type Context struct {
	SessionID string
	Provider  *LLMGateway
	Input     map[string]any
}

// ParseSkill 解析Skill - 转发Python微服务
type ParseSkill struct {
	PythonServiceURL string
}

func (s *ParseSkill) Name() string { return "parse" }

func (s *ParseSkill) Description() string {
	return "解析招标文件，提取模板、技术指标、商务要求、评分点、废标条款"
}

func (s *ParseSkill) Execute(ctx Context, input map[string]any) (map[string]any, error) {
	filePath, ok := input["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("file_path is required")
	}

	client := NewParseClient(s.PythonServiceURL)
	out, err := client.ParseFile(filePath)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"status": "parsed",
		"output": out,
	}, nil
}

// OutlineSkill 大纲Skill - 双通道生成大纲
type OutlineSkill struct {
	LLMGateway *LLMGateway
}

func (s *OutlineSkill) Name() string { return "outline" }

func (s *OutlineSkill) Description() string {
	return "根据招标文件或用户输入生成标书大纲（支持双通道）"
}

func (s *OutlineSkill) Execute(ctx Context, input map[string]any) (map[string]any, error) {
	mode, _ := input["mode"].(string)
	if mode == "" {
		mode = "auto"
	}
	
	switch mode {
	case "extract":
		return s.generateFromExtraction(ctx, input)
	case "custom":
		return s.generateCustom(ctx, input)
	default:
		return s.generateFromExtraction(ctx, input)
	}
}

func (s *OutlineSkill) generateFromExtraction(ctx Context, input map[string]any) (map[string]any, error) {
	parsingOutput, ok := input["parsing_output"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("parsing_output is required")
	}
	
	// TODO: 从解析输出中提取大纲结构
	return map[string]any{
		"status": "generated",
		"mode": "extract",
		"outline": parsingOutput,
	}, nil
}

func (s *OutlineSkill) generateCustom(ctx Context, input map[string]any) (map[string]any, error) {
	userInput, _ := input["user_input"].(string)
	
	// TODO: 使用LLM根据用户输入生成大纲
	return map[string]any{
		"status": "generated",
		"mode": "custom",
		"user_input": userInput,
	}, nil
}

// GenerateSkill 生成Skill - 按纲补写
type GenerateSkill struct {
	LLMGateway *LLMGateway
}

func (s *GenerateSkill) Name() string { return "generate" }

func (s *GenerateSkill) Description() string {
	return "按大纲章节生成标书内容（支持逐章/全篇生成）"
}

func (s *GenerateSkill) Execute(ctx Context, input map[string]any) (map[string]any, error) {
	outline, ok := input["outline"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("outline is required")
	}
	
	mode, _ := input["mode"].(string)
	if mode == "" {
		mode = "section"
	}
	
	switch mode {
	case "section":
		return s.generateSection(ctx, input, outline)
	case "full":
		return s.generateFull(ctx, input, outline)
	default:
		return s.generateSection(ctx, input, outline)
	}
}

func (s *GenerateSkill) generateSection(ctx Context, input map[string]any, outline map[string]any) (map[string]any, error) {
	sectionID, _ := input["section_id"].(string)
	if sectionID == "" {
		return nil, fmt.Errorf("section_id is required")
	}
	
	// TODO: 使用LLM生成单个章节内容
	return map[string]any{
		"status": "generated",
		"section_id": sectionID,
	}, nil
}

func (s *GenerateSkill) generateFull(ctx Context, input map[string]any, outline map[string]any) (map[string]any, error) {
	// TODO: 顺序生成所有章节
	return map[string]any{
		"status": "generated",
		"mode": "full",
	}, nil
}

// ExportSkill 导出Skill - md转Word
type ExportSkill struct{}

func (s *ExportSkill) Name() string { return "export" }

func (s *ExportSkill) Description() string {
	return "将Markdown内容导出为Word文档"
}

func (s *ExportSkill) Execute(ctx Context, input map[string]any) (map[string]any, error) {
	_, ok := input["markdown"].(string)
	if !ok {
		return nil, fmt.Errorf("markdown is required")
	}
	
	// TODO: 实现md→Word转换
	return map[string]any{
		"status": "exported",
		"format": "docx",
	}, nil
}

// RAGSkill RAG检索Skill
type RAGSkill struct {
	Enabled bool
}

func (s *RAGSkill) Name() string { return "rag" }

func (s *RAGSkill) Description() string {
	return "从企业知识库检索相关素材，注入到生成Prompt中"
}

func (s *RAGSkill) Execute(ctx Context, input map[string]any) (map[string]any, error) {
	query, _ := input["query"].(string)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	
	// TODO: 实现向量检索
	return map[string]any{
		"status": "searched",
		"query": query,
		"results": []map[string]any{},
	}, nil
}
