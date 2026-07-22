package handler

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/example/bid-maker-backend/internal/model"
	"github.com/example/bid-maker-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	docxService  *service.DocxService
	llmRegistry  *service.LLMRegistry
	skillService *service.SkillService
}

func New() *Handler {
	skillsDir := os.Getenv("AGENTS_SKILLS_DIR")
	if skillsDir == "" {
		skillsDir = "~/.agents/skills"
	}
	if strings.HasPrefix(skillsDir, "~/") {
		skillsDir = filepath.Join(os.Getenv("HOME"), skillsDir[2:])
	}
	return &Handler{
		docxService:  service.NewDocxService(),
		skillService: service.NewSkillService(skillsDir),
	}
}

func (h *Handler) WithLLMRegistry(reg *service.LLMRegistry) *Handler {
	h.llmRegistry = reg
	return h
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		upload := api.Group("/upload")
		{
			upload.POST("", h.UploadDocument)
		}
		doc := api.Group("/document")
		{
			doc.GET("/:id/outline", h.GetOutline)
			doc.PUT("/:id/outline", h.UpdateOutline)
			doc.GET("/:id/section/:sectionId", h.GetSection)
			doc.PUT("/:id/section/:sectionId", h.SaveSection)
			doc.POST("/:id/export", h.ExportDocument)
		}
		tpl := api.Group("/templates")
		{
			tpl.GET("", h.ListTemplates)
			tpl.GET("/:id", h.GetTemplate)
			tpl.POST("", h.CreateTemplate)
		}
		config := api.Group("/config")
		{
			config.GET("/models", h.ListModels)
			config.POST("/test-key", h.TestApiKey)
		}
		api.POST("/chat", h.Chat)
		api.GET("/local-skills", h.ScanLocalSkills)
		api.GET("/skills/content", h.GetSkillContent)
	}
}

func (h *Handler) UploadDocument(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(src)

	doc, err := h.docxService.ParseDocument(buf.Bytes())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	service.StoreDocument(doc)

	c.JSON(http.StatusOK, doc)
}

func (h *Handler) GetOutline(c *gin.Context) {
	id := c.Param("id")
	doc, ok := service.GetDocument(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"outline": doc.Outline})
}

func (h *Handler) UpdateOutline(c *gin.Context) {
	id := c.Param("id")
	doc, ok := service.GetDocument(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	var outline []model.Section
	if err := c.ShouldBindJSON(&outline); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc.Outline = outline
	service.UpdateDocument(doc)

	c.JSON(http.StatusOK, gin.H{"message": "outline updated"})
}

func (h *Handler) GetSection(c *gin.Context) {
	id := c.Param("id")
	sectionID := c.Param("sectionId")
	doc, ok := service.GetDocument(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	section := h.findSection(&doc.Outline, sectionID)
	if section == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "section not found"})
		return
	}

	c.JSON(http.StatusOK, section)
}

func (h *Handler) SaveSection(c *gin.Context) {
	id := c.Param("id")
	sectionID := c.Param("sectionId")
	doc, ok := service.GetDocument(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	section := h.findSection(&doc.Outline, sectionID)
	if section == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "section not found"})
		return
	}

	var updated model.Section
	if err := c.ShouldBindJSON(&updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	section.Title = updated.Title
	section.Content = updated.Content
	section.Level = updated.Level

	doc.UpdatedAt = service.NowUTC()
	service.UpdateDocument(doc)

	c.JSON(http.StatusOK, gin.H{"message": "section saved"})
}

func (h *Handler) ExportDocument(c *gin.Context) {
	id := c.Param("id")
	doc, ok := service.GetDocument(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	var req struct {
		Format string `json:"format"`
	}
	c.ShouldBindJSON(&req)

	if req.Format == "md" {
		data := h.docxService.GenerateMarkdown(doc)
		c.Header("Content-Type", "text/markdown; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.md", doc.Title))
		c.Data(http.StatusOK, "text/markdown; charset=utf-8", data)
		return
	}

	data, err := h.docxService.GenerateDocument(doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to generate docx: %v", err)})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.docx", doc.Title))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", data)
}

func (h *Handler) Chat(c *gin.Context) {
	var req struct {
		DocumentID string              `json:"document_id"`
		SectionID  string              `json:"section_id"`
		Message    string              `json:"message"`
		Mode       string              `json:"mode"`
		History    []service.Message   `json:"history"`
		Provider   string              `json:"provider"`
		Model      string              `json:"model"`
		Endpoint   string              `json:"endpoint"`
		Format     string              `json:"format"`
		APIKey     string              `json:"apiKey"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var doc *model.Document
	if req.DocumentID != "" {
		d, ok := service.GetDocument(req.DocumentID)
		if ok {
			doc = d
		}
	}

	chatSvc := service.NewChatService(h.llmRegistry)
	chatReq := service.ChatRequest{
		Message:   req.Message,
		Mode:      req.Mode,
		SectionID: req.SectionID,
		History:   req.History,
		Provider:  req.Provider,
		Model:     req.Model,
		Endpoint:  req.Endpoint,
		Format:    req.Format,
		APIKey:    req.APIKey,
	}
	resp, err := chatSvc.Chat(c.Request.Context(), chatReq, doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListModels(c *gin.Context) {
	if h.llmRegistry == nil {
		c.JSON(http.StatusOK, gin.H{"models": []string{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"models": h.llmRegistry.ListModels()})
}

func (h *Handler) ListTemplates(c *gin.Context) {
	templates := service.GetTemplateStore().List()
	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *Handler) GetTemplate(c *gin.Context) {
	id := c.Param("id")
	t, ok := service.GetTemplateStore().Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *Handler) TestApiKey(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required"`
		Model    string `json:"model" binding:"required"`
		Key      string `json:"key" binding:"required"`
		Endpoint string `json:"endpoint"`
		Format   string `json:"format"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := service.TestApiKey(req.Provider, req.Key, req.Model, req.Endpoint, req.Format)
	c.JSON(http.StatusOK, gin.H{"available": result.Available, "error": result.Error})
}

func (h *Handler) findSection(sections *[]model.Section, id string) *model.Section {
	for i := range *sections {
		if (*sections)[i].ID == id {
			return &(*sections)[i]
		}
		if (*sections)[i].Children != nil {
			if found := h.findSection(&(*sections)[i].Children, id); found != nil {
				return found
			}
		}
	}
	return nil
}

func init() {
	log.Println("bid-maker handler initialized")
}

func (h *Handler) ScanLocalSkills(c *gin.Context) {
	skills, err := h.skillService.ScanLocalSkills()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"skills": skills})
}

func (h *Handler) GetSkillContent(c *gin.Context) {
	skillName := c.Query("path")
	if skillName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path parameter is required"})
		return
	}

	content, err := h.skillService.GetSkillContent(skillName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, content)
}

func (h *Handler) CreateTemplate(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := form.Value["name"][0]
	fileHeader := form.File["file"][0]
	if fileHeader == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	ext := filepath.Ext(fileHeader.Filename)
	if ext != ".docx" && ext != ".doc" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .docx files are supported"})
		return
	}

	src, _ := fileHeader.Open()
	defer src.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(src)

	doc, err := h.docxService.ParseDocument(buf.Bytes())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	template := model.Template{
		ID:          fmt.Sprintf("tpl_%d", time.Now().Unix()),
		Name:        name,
		Description: "User uploaded template",
		Category:    "custom",
		Icon:        "RiFileTextLine",
		Outline:     doc.Outline,
	}

	service.GetTemplateStore().Save(template)
	c.JSON(http.StatusOK, template)
}
