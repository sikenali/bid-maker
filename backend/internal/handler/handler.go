package handler

import (
	"bytes"
	"fmt"
	"log"
	"net/http"

	"github.com/example/bid-maker-backend/internal/model"
	"github.com/example/bid-maker-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	docxService *service.DocxService
	llmRegistry *service.LLMRegistry
}

func New() *Handler {
	return &Handler{
		docxService: service.NewDocxService(),
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
		config := api.Group("/config")
		{
			config.GET("/models", h.ListModels)
		}
		api.POST("/chat", h.Chat)
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
	if h.llmRegistry == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "LLM not configured"})
		return
	}

	var req struct {
		DocumentID string              `json:"document_id"`
		SectionID  string              `json:"section_id"`
		Message    string              `json:"message"`
		Mode       string              `json:"mode"`
		History    []service.Message   `json:"history"`
		Provider   string              `json:"provider"`
		Model      string              `json:"model"`
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
