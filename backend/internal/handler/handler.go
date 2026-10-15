package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/example/bid-maker-backend/internal/config"
	"github.com/example/bid-maker-backend/internal/model"
	"github.com/example/bid-maker-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	gateway      *service.LLMGateway
	sessionStore *service.SessionStore
	taskQueue    *service.TaskQueue
	config       *config.Config
}

func New(cfg *config.Config) *Handler {
	gateway := service.NewLLMGateway(toServiceProviders(cfg.LLMProviders), "")
	return &Handler{
		gateway:      gateway,
		sessionStore: service.GetSessionStore(),
		taskQueue:    service.NewTaskQueue(),
		config:       cfg,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/sessions", h.CreateSession)
		api.GET("/sessions/:id", h.GetSession)
		api.PUT("/sessions/:id", h.UpdateSession)
		api.POST("/sessions/:id/skills/parse", h.RunParseSkill)
		api.POST("/sessions/:id/skills/outline", h.RunOutlineSkill)
		api.PUT("/sessions/:id/outline", h.UpdateOutline)
		api.POST("/sessions/:id/outline/lock", h.LockOutline)
		api.POST("/sessions/:id/skills/generate/section", h.RunGenerateSection)
		api.POST("/sessions/:id/skills/generate/full", h.RunGenerateFull)
		api.GET("/sessions/:id/skills/generate/stream", h.GenerateSectionStream)
		api.GET("/sessions/:id/skills/generate/full/stream", h.GenerateFullStream)
		api.GET("/sessions/:id/export", h.ExportDocument)
		api.GET("/sessions/:id/export/pdf", h.ExportPDF)
		api.GET("/sessions/:id/export/markdown", h.GetMarkdown)
		api.GET("/skills", h.ListSkills)
		api.GET("/tasks/:taskId", h.GetTaskStatus)
		api.GET("/models", h.ListModels)
		api.POST("/test-key", h.TestApiKey)
		api.POST("/chat", h.Chat)
		api.GET("/local-skills", h.ListLocalSkills)
	}
}

func (h *Handler) CreateSession(c *gin.Context) {
	var req struct {
		Title string `json:"title"`
		BidNo string `json:"bid_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sessionID := uuid.New().String()
	session := h.sessionStore.CreateSession(sessionID, req.Title, req.BidNo)
	c.JSON(http.StatusOK, session)
}

func (h *Handler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	session, ok := h.sessionStore.GetSession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *Handler) UpdateSession(c *gin.Context) {
	sessionID := c.Param("id")
	session, ok := h.sessionStore.GetSession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	var req model.Session
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 部分更新:空标题/编号视为未提供,避免自动保存(markdown only)清空标题;
	// Markdown 与 CurrentSection 始终全量替换(清空文档也是合法状态)
	if req.Title != "" {
		session.Title = req.Title
	}
	if req.BidNo != "" {
		session.BidNo = req.BidNo
	}
	session.Markdown = req.Markdown
	if req.CurrentSection != "" {
		session.CurrentSection = req.CurrentSection
	}
	h.sessionStore.UpdateSession(session)
	c.JSON(http.StatusOK, session)
}

func (h *Handler) RunParseSkill(c *gin.Context) {
	sessionID := c.Param("id")
	if _, ok := h.sessionStore.GetSession(sessionID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 双输入:multipart 文件直传,或 JSON {file_path} 服务端本地路径
	var filename string
	var fileBytes []byte
	var filePath string
	if fh, err := c.FormFile("file"); err == nil {
		f, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "打开上传文件失败: " + err.Error()})
			return
		}
		defer f.Close()
		fileBytes, err = io.ReadAll(io.LimitReader(f, 50<<20))
		if err != nil || len(fileBytes) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "读取上传文件失败或文件为空"})
			return
		}
		filename = fh.Filename
	} else {
		var req struct {
			FilePath string `json:"file_path"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.FilePath == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请上传 file 文件,或提供 file_path"})
			return
		}
		filePath = req.FilePath
		filename = filePath
	}

	// 与解析服务白名单对齐,不支持的格式即时 400(否则要等异步任务失败)
	lowerName := strings.ToLower(filename)
	allowed := false
	for _, e := range []string{".docx", ".xlsx", ".xlsm", ".pptx", ".pdf", ".md", ".txt", ".csv", ".html"} {
		if strings.HasSuffix(lowerName, e) {
			allowed = true
			break
		}
	}
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件格式:仅支持 .docx/.xlsx/.pptx/.pdf/.md/.txt/.csv/.html"})
		return
	}

	taskID := uuid.New().String()
	h.taskQueue.CreateTask(taskID, sessionID, "parse", "")
	go func() {
		h.taskQueue.UpdateStatus(taskID, "running")
		client := service.NewParseClient(h.config.PythonService)
		var (
			output *model.BidParsingOutput
			err    error
		)
		if fileBytes != nil {
			output, err = client.ParseReader(filename, bytes.NewReader(fileBytes))
		} else {
			output, err = client.ParseFile(filePath)
		}
		if err != nil {
			h.taskQueue.UpdateError(taskID, err.Error())
			return
		}
		h.sessionStore.SetParsingOutput(sessionID, output)
		h.taskQueue.UpdateResult(taskID, map[string]any{"output": output})
	}()
	c.JSON(http.StatusOK, gin.H{"task_id": taskID, "status": "started"})
}

func (h *Handler) RunOutlineSkill(c *gin.Context) {
	sessionID := c.Param("id")
	session, ok := h.sessionStore.GetSession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	var req struct {
		Mode      string `json:"mode"`
		UserInput string `json:"user_input,omitempty"`
		Provider  string `json:"provider,omitempty"`
		Model     string `json:"model,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	outlineID := uuid.New().String()
	mode := req.Mode
	if mode == "" {
		mode = "auto"
	}
	outline := &model.BidOutline{OutlineID: outlineID, SourceMode: mode, Locked: false}
	if session.ParsingOutput != nil && session.ParsingOutput.Template != nil {
		for _, sec := range session.ParsingOutput.Template.PresetSections {
			outline.Sections = append(outline.Sections, &model.Section{
				SectionID: sec.SectionID, Title: sec.Title, Level: sec.Level,
				Order: sec.Order, Required: sec.Required, ParentID: sec.ParentID,
				Source: model.SourceAuto, Status: model.ChapterStatusPending,
			})
		}
		service.BuildChildrenTree(outline.Sections)
	}
	// 通道B/hybrid 需要走模型
	if mode == "custom" || mode == "hybrid" {
		if strings.TrimSpace(req.UserInput) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "custom/hybrid 模式需要 user_input"})
			return
		}
		hint := ""
		if session.ParsingOutput != nil && session.ParsingOutput.SourceMeta != nil {
			hint = session.ParsingOutput.SourceMeta.Title + " " + session.ParsingOutput.SourceMeta.BidNo
		}
		var (
			sections []*model.Section
			err      error
		)
		if mode == "custom" {
			sections, err = service.BuildCustomOutline(c.Request.Context(), h.gateway, req.Provider, req.Model, req.UserInput, hint)
		} else {
			sections, err = service.BuildHybridOutline(c.Request.Context(), h.gateway, req.Provider, req.Model, outline.Sections, req.UserInput)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "大纲生成失败: " + err.Error()})
			return
		}
		outline.Sections = sections
	}
	// 挂载解析层要求,持久化到会话
	if session.ParsingOutput != nil {
		service.MountRequirements(session.ParsingOutput, outline.Sections)
	}
	h.sessionStore.SetOutline(sessionID, outline)
	c.JSON(http.StatusOK, outline)
}

func (h *Handler) UpdateOutline(c *gin.Context) {
	sessionID := c.Param("id")
	session, ok := h.sessionStore.GetSession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	var outline model.BidOutline
	if err := c.ShouldBindJSON(&outline); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 按纲约束:锁定后禁删章节,标题非空、id 唯一,新增标 edited
	if err := service.ValidateOutlineUpdate(session.Outline, &outline); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.sessionStore.SetOutline(sessionID, &outline)
	c.JSON(http.StatusOK, gin.H{"message": "outline updated"})
}

// LockOutline 锁定大纲:锁定后生成阶段不可改结构,仅可增补子章节
func (h *Handler) LockOutline(c *gin.Context) {
	sessionID := c.Param("id")
	session, ok := h.sessionStore.GetSession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if session.Outline == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先生成大纲再锁定"})
		return
	}
	session.Outline.Locked = true
	h.sessionStore.UpdateSession(session)
	c.JSON(http.StatusOK, gin.H{"message": "outline locked", "outline_id": session.Outline.OutlineID})
}

func (h *Handler) RunGenerateSection(c *gin.Context) {
	sessionID := c.Param("id")
	if _, ok := h.sessionStore.GetSession(sessionID); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	var req struct {
		SectionID string `json:"section_id"`
		Provider  string `json:"provider"`
		Model     string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	taskID := uuid.New().String()
	h.taskQueue.CreateTask(taskID, sessionID, "generate_section", req.SectionID)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		h.taskQueue.UpdateStatus(taskID, "running")
		h.taskQueue.UpdateProgress(taskID, 10)
		messages := []map[string]string{
			{"role": "system", "content": "你是一个专业的标书编写助手。请根据大纲章节要求，撰写专业、完整的标书内容。"},
			{"role": "user", "content": fmt.Sprintf("请撰写章节：%s", req.SectionID)},
		}
		content, err := h.gateway.Chat(ctx, req.Provider, messages, req.Model)
		if err != nil {
			h.taskQueue.UpdateError(taskID, err.Error())
			return
		}
		h.taskQueue.UpdateProgress(taskID, 90)
		h.sessionStore.UpdateSectionContent(sessionID, req.SectionID, content)
		h.taskQueue.UpdateProgress(taskID, 100)
		h.taskQueue.UpdateResult(taskID, map[string]any{"content": content})
	}()
	c.JSON(http.StatusOK, gin.H{"task_id": taskID, "status": "started"})
}

func (h *Handler) RunGenerateFull(c *gin.Context) {
	sessionID := c.Param("id")
	session, ok := h.sessionStore.GetSession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	var req struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if session.Outline == nil || len(session.Outline.Sections) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先在大纲中生成章节，再执行全文生成"})
		return
	}
	taskID := uuid.New().String()
	h.taskQueue.CreateTask(taskID, sessionID, "generate_full", "")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		h.taskQueue.UpdateStatus(taskID, "running")
		sections := flattenSections(session.Outline.Sections)
		total := len(sections)
		if total == 0 {
			h.taskQueue.UpdateError(taskID, "no sections to generate")
			return
		}
		completed := 0
		var errMsg string
		for _, sec := range sections {
			if sec.Content != "" {
				completed++
				h.taskQueue.UpdateProgress(taskID, int(float64(completed)/float64(total)*100))
				continue
			}
			prompt := buildSectionPrompt(session, sec)
			messages := []map[string]string{
				{"role": "system", "content": "你是一名资深的投标文件撰写专家。请根据大纲章节要求，撰写专业、完整、可直接使用的标书内容。输出纯文本，不要加额外的Markdown标题。"},
				{"role": "user", "content": prompt},
			}
			content, err := h.gateway.Chat(ctx, req.Provider, messages, req.Model)
			if err != nil {
				h.sessionStore.MarkSectionFailed(sessionID, sec.SectionID, err.Error())
				errMsg = fmt.Sprintf("section %s: %s", sec.Title, err.Error())
				continue
			}
			now := time.Now().UTC()
			h.sessionStore.SetSectionContent(sessionID, sec.SectionID, sec.Title, sec.Level, content, &now)
			completed++
			progress := int(float64(completed) / float64(total) * 100)
			h.taskQueue.UpdateProgress(taskID, progress)
			// 注:实时推送走 GenerateFullStream;此处直接写 c.Writer 已无连接,不再推送
		}
		result := map[string]any{"completed": completed, "total": total}
		if errMsg != "" {
			result["error"] = errMsg
		}
		h.taskQueue.UpdateResult(taskID, result)
	}()
	c.JSON(http.StatusOK, gin.H{"task_id": taskID, "status": "started"})
}

func flattenSections(sections []*model.Section) []*model.Section {
	byID := make(map[string]*model.Section, len(sections))
	for _, s := range sections {
		byID[s.SectionID] = s
	}
	var result []*model.Section
	visited := make(map[string]bool)
	var visit func(s *model.Section)
	visit = func(s *model.Section) {
		if s == nil || visited[s.SectionID] {
			return
		}
		visited[s.SectionID] = true
		result = append(result, s)
		for _, cid := range s.Children {
			visit(byID[cid])
		}
	}
	for _, s := range sections {
		visit(s)
	}
	return result
}

func buildSectionPrompt(session *model.Session, sec *model.Section) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("请撰写标书章节「%s」的详细内容。\n\n", sec.Title))
	sb.WriteString(fmt.Sprintf("标书标题：%s\n", session.Title))
	if session.BidNo != "" {
		sb.WriteString(fmt.Sprintf("招标编号：%s\n", session.BidNo))
	}
	if session.ParsingOutput != nil {
		if len(session.ParsingOutput.TechnicalIndicators) > 0 {
			sb.WriteString("\n技术指标要求：\n")
			for _, ti := range session.ParsingOutput.TechnicalIndicators {
				sb.WriteString(fmt.Sprintf("- %s: %s (必填: %v)\n", ti.Name, ti.Requirement, ti.IsMandatory))
			}
		}
		if len(session.ParsingOutput.CommercialReqs) > 0 {
			sb.WriteString("\n商务要求：\n")
			for _, cr := range session.ParsingOutput.CommercialReqs {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", cr.Category, cr.Description))
			}
		}
		if len(session.ParsingOutput.ScoringPoints) > 0 {
			sb.WriteString("\n评分要点：\n")
			for _, sp := range session.ParsingOutput.ScoringPoints {
				sb.WriteString(fmt.Sprintf("- %s (%g分): %s\n", sp.ScoringItem, sp.FullScore, sp.Description))
			}
		}
	}
	sb.WriteString("\n请输出专业的标书内容，不要输出章节标题本身。")
	return sb.String()
}

func (h *Handler) GenerateSectionStream(c *gin.Context) {
	sessionID := c.Param("id")
	sectionID := c.Query("section_id")
	provider := c.Query("provider")
	modelName := c.Query("model")
	if sectionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "section_id is required"})
		return
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	fmt.Fprintf(c.Writer, "event: start\ndata: %s\n\n", jsonStr(gin.H{"section_id": sectionID}))
	// 带上章节标题与解析上下文,避免只传 ID 导致内容泛化
	userPrompt := fmt.Sprintf("请撰写章节：%s", sectionID)
	if session, ok := h.sessionStore.GetSession(sessionID); ok && session.Outline != nil {
		for _, s := range flattenSections(session.Outline.Sections) {
			if s.SectionID == sectionID {
				userPrompt = buildSectionPrompt(session, s)
				break
			}
		}
	}
	messages := []map[string]string{
		{"role": "system", "content": "你是一个专业的标书编写助手。请根据大纲章节要求，撰写专业、完整的标书内容。"},
		{"role": "user", "content": userPrompt},
	}
	err := h.gateway.ChatStream(c.Request.Context(), provider, messages, modelName, func(chunk string) {
		fmt.Fprintf(c.Writer, "event: chunk\ndata: %s\n\n", jsonStr(gin.H{"chunk": chunk, "section_id": sectionID}))
		c.Writer.Flush()
	})
	if err != nil {
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", jsonStr(gin.H{"error": err.Error()}))
		return
	}
	fmt.Fprintf(c.Writer, "event: done\ndata: %s\n\n", jsonStr(gin.H{"section_id": sectionID}))
}

// GenerateFullStream SSE 全文生成流 — 前端通过 GET /sessions/:id/skills/generate/full/stream 连接
func (h *Handler) GenerateFullStream(c *gin.Context) {
	sessionID := c.Param("id")
	session, ok := h.sessionStore.GetSession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	provider := c.Query("provider")
	modelName := c.Query("model")
	if session.Outline == nil || len(session.Outline.Sections) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先在大纲中生成章节，再执行全文生成"})
		return
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	sections := flattenSections(session.Outline.Sections)
	total := len(sections)
	if total == 0 {
		fmt.Fprintf(c.Writer, "event: done\ndata: %s\n\n", jsonStr(gin.H{}))
		return
	}
	completed := 0
	for _, sec := range sections {
		if sec.Content != "" {
			completed++
			continue
		}
		prompt := buildSectionPrompt(session, sec)
		messages := []map[string]string{
			{"role": "system", "content": "你是一名资深的投标文件撰写专家。请根据大纲章节要求，撰写专业、完整、可直接使用的标书内容。输出纯文本，不要加额外的Markdown标题。"},
			{"role": "user", "content": prompt},
		}
		var content string
		err := h.gateway.ChatStream(c.Request.Context(), provider, messages, modelName, func(chunk string) {
			content += chunk
			fmt.Fprintf(c.Writer, "event: chunk\ndata: %s\n\n", jsonStr(gin.H{"section_id": sec.SectionID, "chunk": chunk, "title": sec.Title, "progress": int(float64(completed+1)/float64(total)*100)}))
			c.Writer.Flush()
		})
		if err != nil {
			h.sessionStore.MarkSectionFailed(sessionID, sec.SectionID, err.Error())
			fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", jsonStr(gin.H{"section_id": sec.SectionID, "error": err.Error()}))
			c.Writer.Flush()
			continue
		}
		now := time.Now().UTC()
		h.sessionStore.SetSectionContent(sessionID, sec.SectionID, sec.Title, sec.Level, content, &now)
		completed++
		progress := int(float64(completed) / float64(total) * 100)
		fmt.Fprintf(c.Writer, "event: progress\ndata: %s\n\n", jsonStr(gin.H{"progress": progress, "completed": completed, "total": total}))
		c.Writer.Flush()
	}
	fmt.Fprintf(c.Writer, "event: done\ndata: %s\n\n", jsonStr(gin.H{"completed": completed, "total": total}))
}

func (h *Handler) ExportDocument(c *gin.Context) {
	sessionID := c.Param("id")
	session, ok := h.sessionStore.GetSession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 内容层:优先用大纲章节(含生成正文),否则从 Markdown 按标题切分
	sections := outlineExportSections(session)
	if len(sections) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有可导出的章节内容(请先生成大纲/正文)"})
		return
	}
	payload := map[string]any{
		"title":    session.Title,
		"bid_no":   session.BidNo,
		"sections": sections,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp, err := http.Post(h.config.PythonService+"/export", "application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "导出服务不可用: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("导出失败: %s", string(raw))})
		return
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.docx", sessionID))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", data)
}

// ExportPDF 导出 PDF:与 Word 同内容源,选项由前端 query 传入(导出设置页)
func (h *Handler) ExportPDF(c *gin.Context) {
	sessionID := c.Param("id")
	session, ok := h.sessionStore.GetSession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	sections := outlineExportSections(session)
	if len(sections) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有可导出的章节内容(请先生成大纲/正文)"})
		return
	}
	payload := map[string]any{
		"title":    session.Title,
		"bid_no":   session.BidNo,
		"sections": sections,
		"options": map[string]any{
			"vectorHD":      c.DefaultQuery("high_res", "1") == "1",
			"bookmarks":     c.DefaultQuery("bookmarks", "1") == "1",
			"watermark":     c.DefaultQuery("watermark", "0") == "1",
			"watermarkText": c.DefaultQuery("watermark_text", "内部稿"),
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp, err := http.Post(h.config.PythonService+"/export-pdf", "application/json", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "导出服务不可用: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("导出失败: %s", string(raw))})
		return
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 50<<20))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", sessionID))
	c.Data(http.StatusOK, "application/pdf", data)
}

// outlineExportSections 把会话大纲展平为导出行;正文缺失时回退解析 Markdown
func outlineExportSections(session *model.Session) []map[string]any {
	var sections []map[string]any
	if session.Outline != nil {
		for _, s := range session.Outline.Sections {
			sections = append(sections, map[string]any{
				"title":   s.Title,
				"content": s.Content,
				"level":   s.Level,
			})
		}
	}
	hasContent := false
	for _, s := range sections {
		if strings.TrimSpace(s["content"].(string)) != "" {
			hasContent = true
			break
		}
	}
	if hasContent {
		return sections
	}
	return splitMarkdownSections(session.Markdown)
}

// splitMarkdownSections 按 # 标题把 Markdown 切为章节
func splitMarkdownSections(md string) []map[string]any {
	var sections []map[string]any
	var cur map[string]any
	var buf strings.Builder
	flush := func() {
		if cur != nil {
			cur["content"] = strings.TrimSpace(buf.String())
			sections = append(sections, cur)
		}
		buf.Reset()
	}
	for _, line := range strings.Split(md, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for level < len(trimmed) && trimmed[level] == '#' {
				level++
			}
			if level > 6 {
				level = 6
			}
			title := strings.TrimSpace(trimmed[level:])
			if title != "" {
				flush()
				cur = map[string]any{"title": title, "content": "", "level": level}
				continue
			}
		}
		if cur != nil {
			buf.WriteString(line + "\n")
		}
	}
	flush()
	return sections
}

func (h *Handler) GetMarkdown(c *gin.Context) {
	sessionID := c.Param("id")
	session, ok := h.sessionStore.GetSession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"markdown": session.Markdown})
}

func (h *Handler) ListSkills(c *gin.Context) {
	skills := []map[string]string{
		{"name": "parse", "description": "解析招标文件，提取模板、技术指标、商务要求、评分点、废标条款"},
		{"name": "outline", "description": "根据招标文件或用户输入生成标书大纲"},
		{"name": "generate", "description": "按大纲章节生成标书内容"},
		{"name": "export", "description": "将Markdown内容导出为Word文档"},
		{"name": "rag", "description": "从企业知识库检索相关素材"},
	}
	c.JSON(http.StatusOK, gin.H{"skills": skills})
}

func (h *Handler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("taskId")
	task, ok := h.taskQueue.GetTask(taskID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *Handler) ListModels(c *gin.Context) {
	models := h.gateway.ListModels()
	c.JSON(http.StatusOK, gin.H{"models": models})
}

func (h *Handler) TestApiKey(c *gin.Context) {
	var req struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
		Key      string `json:"key"`
		Endpoint string `json:"endpoint"`
		Format   string `json:"format"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	available, errMsg := h.gateway.TestProvider(req.Provider, req.Key, req.Model, req.Endpoint, req.Format)
	c.JSON(http.StatusOK, gin.H{"available": available, "error": errMsg})
}

func (h *Handler) Chat(c *gin.Context) {
	var req struct {
		DocumentID string            `json:"document_id"`
		Message    string            `json:"message"`
		Mode       string            `json:"mode"`
		SectionID  string            `json:"section_id"`
		History    []map[string]string `json:"history"`
		Provider   string            `json:"provider"`
		Model      string            `json:"model"`
		Endpoint   string            `json:"endpoint"`
		Format     string            `json:"format"`
		APIKey     string            `json:"apiKey"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	// Build context from session if available
	var sessionContext string
	if req.DocumentID != "" {
		if session, ok := h.sessionStore.GetSession(req.DocumentID); ok {
			var sb strings.Builder
			if session.Outline != nil {
				sb.WriteString("当前标书大纲：\n")
				for _, s := range session.Outline.Sections {
					indent := strings.Repeat("  ", s.Level-1)
					sb.WriteString(fmt.Sprintf("%s%s\n", indent, s.Title))
				}
			}
			if session.Markdown != "" {
				sb.WriteString("\n当前标书内容（摘要）：\n")
				md := session.Markdown
				if len(md) > 3000 {
					md = md[:3000] + "..."
				}
				sb.WriteString(md)
			}
			if session.ParsingOutput != nil && session.ParsingOutput.SourceMeta != nil && session.ParsingOutput.SourceMeta.Title != "" {
				sb.WriteString("\n招标文件信息：" + session.ParsingOutput.SourceMeta.Title + "\n")
			}
			sessionContext = sb.String()
		}
	}

	var messages []map[string]string
	systemPrompt := "你是一个专业的标书写作助手，帮助用户撰写和优化投标文件内容。请用中文回答。"
	if sessionContext != "" {
		systemPrompt += "\n\n" + sessionContext
	}
	messages = append(messages, map[string]string{"role": "system", "content": systemPrompt})
	for _, h := range req.History {
		if h["role"] == "ai" {
			h["role"] = "assistant"
		}
		messages = append(messages, h)
	}
	messages = append(messages, map[string]string{"role": "user", "content": req.Message})

	content, err := h.gateway.Chat(c.Request.Context(), req.Provider, messages, req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 服务不可用: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reply": content})
}

func (h *Handler) ListLocalSkills(c *gin.Context) {
	skills := service.ScanLocalSkills()
	c.JSON(http.StatusOK, gin.H{"skills": skills})
}

func buildMarkdownFromOutline(outline *model.BidOutline) string {
	var md strings.Builder
	for _, section := range outline.Sections {
		indent := strings.Repeat("#", section.Level)
		md.WriteString(indent + " " + section.Title + "\n\n")
	}
	return md.String()
}

func strPtr(s string) *string { return &s }
func floatPtr(f float64) *float64 { return &f }
func jsonStr(v any) string { b, _ := json.Marshal(v); return string(b) }

func toServiceProviders(cfgs []config.LLMProviderConfig) []service.LLMProviderConfig {
	result := make([]service.LLMProviderConfig, len(cfgs))
	for i, c := range cfgs {
		result[i] = service.LLMProviderConfig{
			Name: c.Name, APIKey: c.APIKey, BaseURL: c.BaseURL,
			ModelName: c.ModelName, Type: c.Type,
		}
	}
	return result
}
