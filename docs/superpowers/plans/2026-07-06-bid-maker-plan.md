# Bid-Maker Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a bid document generator with Vue 3 frontend and Go backend, supporting file upload, outline editing, AI-assisted content editing, and .docx export.

**Architecture:** Monorepo with pnpm workspaces. Frontend (Vue 3 SPA) runs on Vite dev server, backend (Go + Gin) runs as separate process. Frontend proxies API calls to backend via Vite dev proxy.

**Tech Stack:**
- Frontend: Vue 3, Vite, TypeScript, Pinia, Vue Router, docx-editor, Axios
- Backend: Go 1.21+, Gin, unioffice, multi-provider LLM SDK
- Dev: pnpm workspaces, nodemon (Go hot reload), Vite proxy

## Global Constraints

- **Document format:** .docx only (Word format)
- **Backend structure:** Simple flat structure (routes + handlers together)
- **LLM providers:** Multi-provider, configurable via environment variables
- **Outline editing:** User-editable after extraction (expand/collapse/reorder/add/remove)
- **Content editor:** docx-editor (rich text editor for Vue)
- **AI chat modes:** Hybrid — context mode + free-form mode
- **Auth:** Not yet, leave room for it
- **Deployment:** Local development only for now
- **Export:** Single .docx download triggered by button click
- **State sync:** Backend-synced — each panel syncs to backend in real-time

---

### Task 1: Initialize Backend Go Module and Project Structure

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/internal/config/config.go`
- Create: `backend/internal/handler/handler.go`
- Create: `backend/internal/service/docx_service.go`
- Create: `backend/internal/service/llm_service.go`
- Create: `backend/internal/service/chat_service.go`
- Create: `backend/internal/model/document.go`
- Create: `backend/.env.example`

**Interfaces:**
- Consumes: Nothing (setup task)
- Produces: Go module with Gin + unioffice dependencies, basic config struct, empty handler template

- [ ] **Step 1: Initialize Go module**

Run:
```bash
cd backend
go mod init github.com/example/bid-maker-backend
```

Then install dependencies:
```bash
go get github.com/gin-gonic/gin
go get github.com/unidoc/unidoc-office/v4
```

- [ ] **Step 2: Create config package**

Create `backend/internal/config/config.go`:
```go
package config

type Config struct {
	Port     int    `env:"PORT"`
	LLMProviders []LLMProviderConfig
}

type LLMProviderConfig struct {
	Name      string `env:"NAME"`
	APIKey    string `env:"API_KEY"`
	BaseURL   string `env:"BASE_URL"`
	ModelName string `env:"MODEL_NAME"`
	Type      string `env:"TYPE"` // "openai", "dashscope", "zhipu", etc.
}

func Load() (*Config, error)
```

Use `github.com/caarlos0/env/v10` for env loading. Install it:
```bash
go get github.com/caarlos0/env/v10
```

- [ ] **Step 3: Create model package**

Create `backend/internal/model/document.go`:
```go
package model

type Section struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Level    int    `json:"level"`    // heading level (1, 2, 3...)
	ParentID string `json:"parent_id"`
	Content  string `json:"content"`  // HTML content for docx-editor
	Children []Section `json:"children"`
}

type Document struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Outline   []Section `json:"outline"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}
```

- [ ] **Step 4: Create handler template**

Create `backend/internal/handler/handler.go`:
```go
package handler

import "github.com/gin-gonic/gin"

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {}
```

- [ ] **Step 5: Create main.go entry point**

Create `backend/cmd/server/main.go`:
```go
package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"github.com/example/bid-maker-backend/internal/config"
	"github.com/example/bid-maker-backend/internal/handler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	h := handler.New()
	r := gin.Default()
	h.RegisterRoutes(r)

	port := cfg.Port
	if port == 0 {
		port = 8080
	}
	r.Run(":" + string(rune(port+'0')))
}
```

Fix: Use `fmt.Sprintf(":%d", port)` instead of rune math.

- [ ] **Step 6: Create .env.example**

Create `backend/.env.example`:
```env
PORT=8080

# LLM Providers
LLM_PROVIDERS=[
  {
    "name": "openai",
    "api_key": "${OPENAI_API_KEY}",
    "base_url": "https://api.openai.com/v1",
    "model_name": "gpt-4",
    "type": "openai"
  },
  {
    "name": "dashscope",
    "api_key": "${DASHSCOPE_API_KEY}",
    "base_url": "https://dashscope.aliyuncs.com/compatible-mode/v1",
    "model_name": "qwen-max",
    "type": "dashscope"
  }
]
```

- [ ] **Step 7: Verify backend builds**

Run:
```bash
cd backend
go build ./cmd/server/
./server
```

Expected: Server starts on port 8080 (or configured port).

- [ ] **Step 8: Commit**

```bash
git add backend/
git commit -m "feat: initialize Go backend with project structure"
```

---

### Task 2: Implement LLM Provider Abstraction

**Files:**
- Modify: `backend/internal/service/llm_service.go`
- Modify: `backend/internal/config/config.go`

**Interfaces:**
- Consumes: `config.Config` with LLMProviders array
- Produces: `LLMClient` interface with `Chat(ctx, messages, model) (string, error)` method

- [ ] **Step 1: Define LLM interface and provider registry**

Create `backend/internal/service/llm_service.go`:
```go
package service

import (
	"context"
	"github.com/example/bid-maker-backend/internal/config"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMClient interface {
	Chat(ctx context.Context, messages []Message, model string) (string, error)
	Close() error
}

type LLMRegistry struct {
	clients map[string]LLMClient
}

func NewLLMRegistry(providers []config.LLMProviderConfig) *LLMRegistry
func (r *LLMRegistry) GetProvider(name string) (LLMClient, error)
func (r *LLMRegistry) ListModels() []string
```

- [ ] **Step 2: Implement OpenAI-compatible provider**

Still in `llm_service.go`, add:
```go
type OpenAIProvider struct {
	client  *http.Client
	baseURL string
	apiKey  string
	model   string
}

func NewOpenAIProvider(cfg config.LLMProviderConfig) *OpenAIProvider

func (p *OpenAIProvider) Chat(ctx context.Context, messages []Message, model string) (string, error) {
	// Build OpenAI-compatible chat completion request
	// POST {baseURL}/chat/completions
	// Use standard OpenAI message format
	// Parse response and return content
}
```

Install OpenAI SDK:
```bash
go get github.com/sashabaranov/go-openai
```

Use `github.com/sashabaranov/go-openai` for all OpenAI-compatible providers (OpenAI, DashScope, Zhipu all share the same API format).

- [ ] **Step 3: Register providers in registry**

In `NewLLMRegistry`, register each provider from config:
```go
func NewLLMRegistry(providers []config.LLMProviderConfig) *LLMRegistry {
	reg := &LLMRegistry{clients: make(map[string]LLMClient)}
	for _, p := range providers {
		switch p.Type {
		case "openai":
			reg.clients[p.Name] = NewOpenAIProvider(p)
		default:
			// Support any OpenAI-compatible provider
			reg.clients[p.Name] = NewOpenAIProvider(p)
		}
	}
	return reg
}
```

- [ ] **Step 4: Write and run tests**

Create `backend/internal/service/llm_service_test.go`:
```go
package service

import (
	"context"
	"testing"
)

func TestLLMRegistry_GetProvider(t *testing.T) {
	providers := []config.LLMProviderConfig{
		{Name: "test", APIKey: "key", BaseURL: "https://test.com/v1", ModelName: "gpt-3.5", Type: "openai"},
	}
	reg := NewLLMRegistry(providers)
	client, err := reg.GetProvider("test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
}

func TestLLMRegistry_ListModels(t *testing.T) {
	providers := []config.LLMProviderConfig{
		{Name: "openai", ModelName: "gpt-4", Type: "openai"},
		{Name: "dashscope", ModelName: "qwen-max", Type: "openai"},
	}
	reg := NewLLMRegistry(providers)
	models := reg.ListModels()
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}
}
```

Run:
```bash
cd backend
go test ./internal/service/ -v -run TestLLM
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/
git commit -m "feat: implement LLM provider abstraction with OpenAI-compatible client"
```

---

### Task 3: Implement .docx Upload and Outline Extraction

**Files:**
- Modify: `backend/internal/service/docx_service.go`
- Modify: `backend/internal/handler/handler.go`
- Create: `backend/internal/service/docx_service_test.go`

**Interfaces:**
- Consumes: Uploaded .docx file via multipart form
- Produces: `Document` with extracted outline (sections tree) and content

- [ ] **Step 1: Implement docx parser service**

Create `backend/internal/service/docx_service.go`:
```go
package service

import (
	"bytes"
	"github.com/example/bid-maker-backend/internal/model"
	"github.com/unidoc/unipdf/v4/common/log"
	unidoc "github.com/unidoc/unidoc-office/v4/pkg/document"
)

type DocxService struct{}

func NewDocxService() *DocxService {
	return &DocxService{}
}

// ParseDocument extracts outline and content from a .docx file
func (s *DocxService) ParseDocument(data []byte) (*model.Document, error)
// Uses unidoc to open the docx
// Walks the document tree to find heading paragraphs
// Builds a Section tree based on heading levels (1, 2, 3...)
// Extracts body text for each section
// Returns Document with ID, Title, Outline, timestamps

// extractOutline returns []model.Section from docx paragraphs
func (s *DocxService) extractOutline(paragraphs []unidoc.Paragraph) []model.Section
```

Install unidoc:
```bash
go get github.com/unidoc/unidoc-office/v4
```

Note: unidoc requires a license for production. For local dev, use the trial or evaluate mode. The basic document structure reading (paragraphs, styles) works without a full license.

- [ ] **Step 2: Implement upload handler**

Modify `backend/internal/handler/handler.go`:
```go
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
		api.POST("/chat", h.Chat)
	}
}

func (h *Handler) UploadDocument(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(src)

	svc := service.NewDocxService()
	doc, err := svc.ParseDocument(buf.Bytes())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Store document (in-memory for now, persist later)
	store.StoreDocument(doc)

	c.JSON(200, doc)
}
```

- [ ] **Step 3: Add simple in-memory document store**

Create `backend/internal/service/store.go`:
```go
package service

import (
	"github.com/example/bid-maker-backend/internal/model"
	"sync"
)

var documentStore struct {
	sync.RWMutex
	documents map[string]*model.Document
}

func init() {
	documentStore.documents = make(map[string]*model.Document)
}

func StoreDocument(doc *model.Document) {
	documentStore.Lock()
	defer documentStore.Unlock()
	documentStore.documents[doc.ID] = doc
}

func GetDocument(id string) (*model.Document, bool) {
	documentStore.RLock()
	defer documentStore.RUnlock()
	doc, ok := documentStore.documents[id]
	return doc, ok
}

func UpdateDocument(doc *model.Document) {
	documentStore.Lock()
	defer documentStore.Unlock()
	documentStore.documents[doc.ID] = doc
}
```

- [ ] **Step 4: Write tests for docx parsing**

Create `backend/internal/service/docx_service_test.go`:
```go
package service

import (
	"os"
	"testing"
)

func TestDocxService_ParseDocument(t *testing.T) {
	// Create a minimal test .docx file
	// Since creating a real .docx programmatically is complex,
	// use a zip archive with minimal docx structure
	// OR use an existing test fixture

	svc := NewDocxService()
	doc, err := svc.ParseDocument([]byte{})
	if err != nil {
		t.Logf("Parse error (expected for empty file): %v", err)
	}
	// Test with actual fixture when available
}
```

For a real test, create a minimal .docx fixture in `backend/testdata/test.docx` with at least one Heading 1 and some body text.

- [ ] **Step 5: Wire up main.go to use handler.RegisterRoutes**

Update `backend/cmd/server/main.go`:
```go
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	h := handler.New()
	r := gin.Default()
	h.RegisterRoutes(r)

	port := cfg.Port
	if port == 0 {
		port = 8080
	}
	log.Printf("Server starting on port %d", port)
	r.Run(fmt.Sprintf(":%d", port))
}
```

Install `fmt`:
```bash
go get fmt
```
(Note: fmt is stdlib, just import it)

- [ ] **Step 6: Verify backend builds**

```bash
cd backend
go build ./cmd/server/
```

- [ ] **Step 7: Commit**

```bash
git add backend/internal/
git commit -m "feat: implement .docx upload and outline extraction"
```

---

### Task 4: Implement Document Editing Endpoints

**Files:**
- Modify: `backend/internal/handler/handler.go`
- Modify: `backend/internal/service/store.go`

**Interfaces:**
- Consumes: `Document` from store, section ID, content (HTML string)
- Produces: Updated document with modified section content

- [ ] **Step 1: Implement GetOutline handler**

In `handler.go`:
```go
func (h *Handler) GetOutline(c *gin.Context) {
	id := c.Param("id")
	doc, ok := service.GetDocument(id)
	if !ok {
		c.JSON(404, gin.H{"error": "document not found"})
		return
	}
	c.JSON(200, gin.H{"outline": doc.Outline})
}
```

- [ ] **Step 2: Implement UpdateOutline handler**

```go
func (h *Handler) UpdateOutline(c *gin.Context) {
	id := c.Param("id")
	doc, ok := service.GetDocument(id)
	if !ok {
		c.JSON(404, gin.H{"error": "document not found"})
		return
	}

	var outline []model.Section
	if err := c.ShouldBindJSON(&outline); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	doc.Outline = outline
	service.UpdateDocument(doc)
	c.JSON(200, gin.H{"ok": true})
}
```

- [ ] **Step 3: Implement GetSection handler**

```go
func (h *Handler) GetSection(c *gin.Context) {
	id := c.Param("id")
	sectionID := c.Param("sectionId")
	doc, ok := service.GetDocument(id)
	if !ok {
		c.JSON(404, gin.H{"error": "document not found"})
		return
	}

	section := findSection(doc.Outline, sectionID)
	if section == nil {
		c.JSON(404, gin.H{"error": "section not found"})
		return
	}
	c.JSON(200, section)
}

func findSection(sections []model.Section, id string) *model.Section {
	for i := range sections {
		if sections[i].ID == id {
			return &sections[i]
		}
		if child := findSection(sections[i].Children, id); child != nil {
			return child
		}
	}
	return nil
}
```

- [ ] **Step 4: Implement SaveSection handler**

```go
func (h *Handler) SaveSection(c *gin.Context) {
	id := c.Param("id")
	sectionID := c.Param("sectionId")
	doc, ok := service.GetDocument(id)
	if !ok {
		c.JSON(404, gin.H{"error": "document not found"})
		return
	}

	section := findSection(doc.Outline, sectionID)
	if section == nil {
		c.JSON(404, gin.H{"error": "section not found"})
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	section.Content = req.Content
	service.UpdateDocument(doc)
	c.JSON(200, gin.H{"ok": true})
}
```

- [ ] **Step 5: Write handler tests**

Create `backend/internal/handler/handler_test.go`:
```go
package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/example/bid-maker-backend/internal/model"
	"github.com/example/bid-maker-backend/internal/service"
)

func TestGetOutline_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := New()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/document/nonexistent/outline", nil)
	c.Params = gin.Param{Key: "id", Value: "nonexistent"}
	h.GetOutline(c)
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
```

Run:
```bash
cd backend
go test ./internal/handler/ -v
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/handler/
git commit -m "feat: implement document outline and section editing endpoints"
```

---

### Task 5: Implement AI Chat Endpoint

**Files:**
- Modify: `backend/internal/service/chat_service.go`
- Modify: `backend/internal/handler/handler.go`

**Interfaces:**
- Consumes: LLMRegistry, document context
- Produces: AI response based on chat mode (free-form or context-aware)

- [ ] **Step 1: Implement chat service**

Create `backend/internal/service/chat_service.go`:
```go
package service

import (
	"context"
	"github.com/example/bid-maker-backend/internal/model"
	"strings"
)

type ChatService struct {
	registry *LLMRegistry
}

func NewChatService(registry *LLMRegistry) *ChatService {
	return &ChatService{registry: registry}
}

type ChatRequest struct {
	Message    string         `json:"message"`
	Mode       string         `json:"mode"` // "free" or "context"
	SectionID  string         `json:"section_id,omitempty"`
	History    []Message      `json:"history"`
	Model      string         `json:"model,omitempty"`
}

type ChatResponse struct {
	Reply   string `json:"reply"`
	Model   string `json:"model"`
}

func (s *ChatService) Chat(ctx context.Context, req ChatRequest, doc *model.Document) (*ChatResponse, error) {
	client, err := s.registry.GetProvider("")
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var messages []Message

	if req.Mode == "context" && req.SectionID != "" {
		// Build context-aware prompt
		section := findSection(doc.Outline, req.SectionID)
		if section != nil {
			contextPrompt := "You are helping write a bid document. Current section: " + section.Title + "\nContent so far: " + section.Content + "\nOutline: " + buildOutlineString(doc.Outline) + "\n\n"
			messages = append(messages, Message{Role: "system", Content: contextPrompt})
		}
	} else {
		messages = append(messages, Message{Role: "system", Content: "You are a helpful assistant for bid document creation."})
	}

	// Add conversation history
	for _, m := range req.History {
		messages = append(messages, m)
	}

	// Add current message
	messages = append(messages, Message{Role: "user", Content: req.Message})

	reply, err := client.Chat(ctx, messages, req.Model)
	if err != nil {
		return nil, err
	}

	return &ChatResponse{Reply: reply, Model: req.Model}, nil
}

func buildOutlineString(sections []model.Section) string {
	var sb strings.Builder
	for _, s := range sections {
		sb.WriteString(strings.Repeat("  ", s.Level-1) + s.Title + "\n")
		for _, c := range s.Children {
			sb.WriteString(strings.Repeat("  ", c.Level-1) + c.Title + "\n")
		}
	}
	return sb.String()
}
```

- [ ] **Step 2: Implement Chat handler**

In `handler.go`:
```go
func (h *Handler) Chat(c *gin.Context) {
	var req service.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Document ID should come from URL or context
	docID := c.GetString("document_id") // Set via middleware or from request
	doc, ok := service.GetDocument(docID)
	if !ok {
		c.JSON(404, gin.H{"error": "document not found"})
		return
	}

	registry := h.getLLMRegistry() // From config
	chatSvc := service.NewChatService(registry)
	resp, err := chatSvc.Chat(c.Request.Context(), req, doc)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, resp)
}
```

- [ ] **Step 3: Write chat service test**

Create `backend/internal/service/chat_service_test.go`:
```go
package service

import (
	"testing"
)

func TestBuildOutlineString(t *testing.T) {
	sections := []model.Section{
		{Title: "Chapter 1", Level: 1},
		{Title: "Section 1.1", Level: 2, ParentID: "1"},
		{Title: "Chapter 2", Level: 1},
	}
	result := buildOutlineString(sections)
	if result == "" {
		t.Fatal("expected non-empty outline string")
	}
}
```

Run:
```bash
cd backend
go test ./internal/service/ -v -run TestBuildOutline
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/service/chat_service.go backend/internal/handler/handler.go
git commit -m "feat: implement AI chat endpoint with context-aware mode"
```

---

### Task 6: Implement .docx Export

**Files:**
- Modify: `backend/internal/service/docx_service.go`
- Modify: `backend/internal/handler/handler.go`

**Interfaces:**
- Consumes: `Document` with outline and sections
- Produces: .docx file download stream

- [ ] **Step 1: Implement export method in DocxService**

In `docx_service.go`:
```go
import (
	"bytes"
	unidoc "github.com/unidoc/unidoc-office/v4/pkg/document"
)

func (s *DocxService) ExportDocument(doc *model.Document) ([]byte, error) {
	writer := unidoc.NewDocumentWriter()
	section := writer.NewSection()

	for _, s := range doc.Outline {
		// Write heading
		p := section.AddParagraph()
		run := p.AddRun(s.Title)
		run.Bold = true
		run.FontSize = float64(14 - s.Level) // Smaller font for deeper headings

		// Write content if exists
		if s.Content != "" {
			contentP := section.AddParagraph()
			contentP.InsertXMLFromString(s.Content) // docx-editor outputs HTML, convert or use raw text
		}
	}

	buf := new(bytes.Buffer)
	if err := writer.Write(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
```

Note: docx-editor outputs HTML. For the initial implementation, strip HTML tags and write plain text. A full HTML-to-docx conversion can be added later.

Helper to strip HTML:
```go
import "golang.org/x/net/html"

func stripHTML(htmlStr string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlStr))
	var result strings.Builder
	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}
		if tt == html.TextToken {
			result.WriteString(tokenizer.Token().Data)
		}
	}
	return result.String()
}
```

- [ ] **Step 2: Implement Export handler**

In `handler.go`:
```go
func (h *Handler) ExportDocument(c *gin.Context) {
	id := c.Param("id")
	doc, ok := service.GetDocument(id)
	if !ok {
		c.JSON(404, gin.H{"error": "document not found"})
		return
	}

	docxSvc := service.NewDocxService()
	data, err := docxSvc.ExportDocument(doc)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	c.Header("Content-Disposition", "attachment; filename=bid-document.docx")
	c.Data(200, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", data)
}
```

- [ ] **Step 3: Write export test**

Create test in `docx_service_test.go`:
```go
func TestDocxService_ExportDocument(t *testing.T) {
	doc := &model.Document{
		ID:    "test",
		Title: "Test Bid",
		Outline: []model.Section{
			{ID: "1", Title: "Chapter 1", Level: 1, Content: "<p>Hello</p>"},
		},
	}
	svc := NewDocxService()
	data, err := svc.ExportDocument(doc)
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty export data")
	}
}
```

Run:
```bash
cd backend
go test ./internal/service/ -v -run TestExport
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/service/docx_service.go backend/internal/handler/handler.go
git commit -m "feat: implement .docx export with unidoc"
```

---

### Task 7: Initialize Frontend Vue Project

**Files:**
- Create: `frontend/package.json`
- Create: `frontend/tsconfig.json`
- Create: `frontend/vite.config.ts`
- Create: `frontend/index.html`
- Create: `frontend/src/main.ts`
- Create: `frontend/src/App.vue`
- Create: `frontend/src/router/index.ts`
- Create: `frontend/src/api/client.ts`
- Create: `pnpm-workspace.yaml`
- Create: `pnpm-lock.yaml` (generated)

**Interfaces:**
- Consumes: Nothing (setup task)
- Produces: Working Vue 3 + Vite + TypeScript project with routing and API client

- [ ] **Step 1: Create pnpm workspace config**

Create `pnpm-workspace.yaml` at repo root:
```yaml
packages:
  - 'frontend'
  - 'backend'
```

- [ ] **Step 2: Scaffold frontend project**

Run:
```bash
cd frontend
npm create vite@latest . -- --template vue-ts
```

Or manually create:

`frontend/package.json`:
```json
{
  "name": "bid-maker-frontend",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "vue": "^3.4.0",
    "vue-router": "^4.3.0",
    "pinia": "^2.1.0",
    "axios": "^1.6.0",
    "docx-editor": "^0.2.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0",
    "vue-tsc": "^1.8.0"
  }
}
```

Check if `docx-editor` is available on npm. If not, use `wangeditor` or `tiptap` as alternative:
```bash
npm install @tiptap/vue-3 @tiptap/starter-kit @tiptap/pm
```

Update dependencies to use tiptap instead:
```json
"dependencies": {
  "vue": "^3.4.0",
  "vue-router": "^4.3.0",
  "pinia": "^2.1.0",
  "axios": "^1.6.0",
  "@tiptap/vue-3": "^2.4.0",
  "@tiptap/starter-kit": "^2.4.0",
  "@tiptap/pm": "^2.4.0"
}
```

- [ ] **Step 3: Create vite.config.ts**

```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 4: Create router**

`frontend/src/router/index.ts`:
```typescript
import { createRouter, createWebHistory } from 'vue-router'
import UploadView from '../views/UploadView.vue'
import EditorView from '../views/EditorView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: UploadView },
    { path: '/editor/:id', component: EditorView, props: true },
  ],
})

export default router
```

- [ ] **Step 5: Create API client**

`frontend/src/api/client.ts`:
```typescript
import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
})

export const uploadDocument = (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export const getOutline = (docId: string) => api.get(`/document/${docId}/outline`)
export const updateOutline = (docId: string, outline: any) => api.put(`/document/${docId}/outline`, outline)
export const getSection = (docId: string, sectionId: string) => api.get(`/document/${docId}/section/${sectionId}`)
export const saveSection = (docId: string, sectionId: string, content: string) => api.put(`/document/${docId}/section/${sectionId}`, { content })
export const exportDocument = (docId: string) => api.post(`/document/${docId}/export`, null, {
  responseType: 'blob',
})
export const sendChat = (data: any) => api.post('/chat', data)

export default api
```

- [ ] **Step 6: Create main.ts and App.vue**

`frontend/src/main.ts`:
```typescript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
```

`frontend/src/App.vue`:
```vue
<template>
  <div id="app">
    <router-view />
  </div>
</template>
```

- [ ] **Step 7: Install dependencies and verify**

```bash
cd frontend
pnpm install
pnpm dev
```

Expected: Dev server starts on port 3000, shows UploadView at `/`.

- [ ] **Step 8: Commit**

```bash
git add frontend/ pnpm-workspace.yaml
git commit -m "feat: initialize Vue 3 frontend with Vite, TypeScript, router, and API client"
```

---

### Task 8: Implement Upload View

**Files:**
- Create: `frontend/src/views/UploadView.vue`
- Create: `frontend/src/stores/uiStore.ts`

**Interfaces:**
- Consumes: `uploadDocument()` from API client
- Produces: Upload UI that navigates to EditorView after successful upload

- [ ] **Step 1: Create UploadView component**

`frontend/src/views/UploadView.vue`:
```vue
<template>
  <div class="upload-view">
    <div class="upload-area" @drop.prevent="onDrop" @dragover.prevent>
      <div class="upload-icon">
        <svg><!-- upload icon --></svg>
      </div>
      <p class="upload-text">Drop your .docx file here, or</p>
      <button class="upload-button" @click="triggerUpload">Upload</button>
      <input
        ref="fileInput"
        type="file"
        accept=".docx"
        hidden
        @change="onFileSelected"
      />
    </div>
    <div v-if="loading" class="loading">Processing document...</div>
    <div class="bottom-actions" v-if="documentId">
      <!-- Will be replaced by router transition to EditorView -->
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { uploadDocument } from '../api/client'

const router = useRouter()
const fileInput = ref<HTMLInputElement>()
const loading = ref(false)
const documentId = ref('')

const triggerUpload = () => fileInput.value?.click()

const onFileSelected = async (e: Event) => {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  await handleFile(file)
}

const onDrop = async (e: DragEvent) => {
  const file = e.dataTransfer?.files?.[0]
  if (!file) return
  await handleFile(file)
}

const handleFile = async (file: File) => {
  loading.value = true
  try {
    const res = await uploadDocument(file)
    documentId.value = res.data.id
    router.push(`/editor/${res.data.id}`)
  } catch (err) {
    console.error('Upload failed:', err)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.upload-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 40px;
}
.upload-area {
  border: 2px dashed #ccc;
  border-radius: 12px;
  padding: 60px;
  text-align: center;
  cursor: pointer;
  width: 600px;
}
.upload-area:hover {
  border-color: #666;
}
.upload-button {
  margin-top: 16px;
  padding: 12px 32px;
  background: #1677ff;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
}
.loading {
  margin-top: 16px;
  color: #666;
}
</style>
```

- [ ] **Step 2: Create uiStore**

`frontend/src/stores/uiStore.ts`:
```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUiStore = defineStore('ui', () => {
  const loading = ref(false)
  const setActivePanel = (panel: string) => {}
  return { loading, setActivePanel }
})
```

- [ ] **Step 3: Test upload view manually**

Start both frontend and backend:
```bash
# Terminal 1
cd backend && go run ./cmd/server/

# Terminal 2
cd frontend && pnpm dev
```

Navigate to `http://localhost:3000`, drag a .docx file, verify it uploads and navigates to `/editor/:id`.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/UploadView.vue frontend/src/stores/uiStore.ts
git commit -m "feat: implement upload view with drag-and-drop"
```

---

### Task 9: Implement Editor View Layout

**Files:**
- Create: `frontend/src/views/EditorView.vue`
- Create: `frontend/src/components/OutlineTree.vue`
- Create: `frontend/src/components/AIChat.vue`
- Create: `frontend/src/stores/documentStore.ts`
- Create: `frontend/src/stores/chatStore.ts`

**Interfaces:**
- Consumes: `documentId` from route params
- Produces: Three-panel layout (outline | editor | AI chat)

- [ ] **Step 1: Create EditorView with three-panel layout**

`frontend/src/views/EditorView.vue`:
```vue
<template>
  <div class="editor-view">
    <header class="top-bar">
      <div class="logo">Bid-Maker</div>
      <div class="actions">
        <button title="Help">?</button>
        <button title="Settings">⚙</button>
      </div>
    </header>
    <main class="editor-body">
      <aside class="left-panel">
        <OutlineTree />
      </aside>
      <section class="center-panel">
        <ContentEditor />
      </section>
      <aside class="right-panel">
        <AIChat />
      </aside>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import OutlineTree from '../components/OutlineTree.vue'
import ContentEditor from '../components/ContentEditor.vue'
import AIChat from '../components/AIChat.vue'

const route = useRoute()
const docId = route.params.id as string
</script>

<style scoped>
.editor-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
}
.top-bar {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  border-bottom: 1px solid #eee;
}
.logo {
  font-weight: bold;
  font-size: 18px;
}
.actions {
  display: flex;
  gap: 8px;
}
.editor-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}
.left-panel {
  width: 260px;
  border-right: 1px solid #eee;
  overflow-y: auto;
}
.center-panel {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}
.right-panel {
  width: 300px;
  border-left: 1px solid #eee;
  display: flex;
  flex-direction: column;
}
</style>
```

- [ ] **Step 2: Create documentStore**

`frontend/src/stores/documentStore.ts`:
```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getOutline, getSection, saveSection, updateOutline } from '../api/client'

export interface Section {
  id: string
  title: string
  level: number
  parent_id: string
  content: string
  children: Section[]
}

export const useDocumentStore = defineStore('document', () => {
  const outline = ref<Section[]>([])
  const sections = ref<Map<string, Section>>(new Map())
  const activeSectionId = ref('')

  const loadOutline = async (docId: string) => {
    const res = await getOutline(docId)
    outline.value = res.data.outline
    res.data.outline.forEach((s: Section) => sections.value.set(s.id, s))
  }

  const loadSection = async (docId: string, sectionId: string) => {
    const res = await getSection(docId, sectionId)
    sections.value.set(sectionId, res.data)
    activeSectionId.value = sectionId
  }

  const saveSectionContent = async (docId: string, sectionId: string, content: string) => {
    await saveSection(docId, sectionId, content)
    const section = sections.value.get(sectionId)
    if (section) section.content = content
  }

  const updateOutlineTree = async (docId: string, newOutline: Section[]) => {
    await updateOutline(docId, newOutline)
    outline.value = newOutline
  }

  return { outline, sections, activeSectionId, loadOutline, loadSection, saveSectionContent, updateOutlineTree }
})
```

- [ ] **Step 3: Create placeholder OutlineTree component**

`frontend/src/components/OutlineTree.vue`:
```vue
<template>
  <div class="outline-tree">
    <div class="tree-header">
      <span>Outline</span>
      <button title="Add section">+</button>
    </div>
    <div class="tree-list">
      <div
        v-for="section in outline"
        :key="section.id"
        class="tree-item"
        :class="{ active: section.id === activeSectionId }"
        @click="$emit('select', section.id)"
      >
        <span class="icon">📄</span>
        <span class="title">{{ section.title }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useDocumentStore } from '../stores/documentStore'

const docStore = useDocumentStore()
const outline = computed(() => docStore.outline)
const activeSectionId = computed(() => docStore.activeSectionId)
defineEmits<{ select: [id: string] }>()
</script>

<style scoped>
.outline-tree {
  padding: 16px;
}
.tree-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-weight: bold;
}
.tree-item {
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
}
.tree-item.active {
  background: #e6f0ff;
}
</style>
```

- [ ] **Step 4: Create placeholder ContentEditor component**

`frontend/src/components/ContentEditor.vue`:
```vue
<template>
  <div class="content-editor">
    <div class="editor-header">
      <span class="current-section">{{ currentSectionTitle }}</span>
      <div class="toolbar">
        <button title="AI Assist">✨</button>
        <button title="Save">💾</button>
      </div>
    </div>
    <div class="editor-body">
      <!-- docx-editor or tiptap will mount here -->
      <div ref="editorRef" class="editor-container"></div>
    </div>
    <div class="editor-footer">
      <button class="outline-btn">Extract Outline</button>
      <button class="export-btn">Generate Bid</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useDocumentStore } from '../stores/documentStore'

const docStore = useDocumentStore()
const activeSectionId = computed(() => docStore.activeSectionId)
const currentSectionTitle = computed(() => {
  // Find active section title
  return 'Editing...'
})
const editorRef = ref<HTMLElement>()
</script>

<style scoped>
.content-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #eee;
}
.toolbar {
  display: flex;
  gap: 8px;
}
.editor-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 0;
}
.editor-footer {
  display: flex;
  justify-content: center;
  gap: 16px;
  padding: 16px 0;
}
.outline-btn, .export-btn {
  padding: 12px 24px;
  border-radius: 6px;
  border: 1px solid #ddd;
  background: white;
  cursor: pointer;
}
.export-btn {
  background: #1677ff;
  color: white;
  border: none;
}
</style>
```

- [ ] **Step 5: Create placeholder AIChat component**

`frontend/src/components/AIChat.vue`:
```vue
<template>
  <div class="ai-chat">
    <div class="chat-header">
      <span>AI</span>
      <select v-model="model" class="model-select">
        <option value="">Default</option>
      </select>
    </div>
    <div class="chat-messages" ref="messagesRef">
      <div v-for="(msg, i) in messages" :key="i" class="message" :class="msg.role">
        <div v-if="msg.role === 'ai'" class="avatar">🤖</div>
        <div class="bubble">{{ msg.content }}</div>
      </div>
    </div>
    <div class="chat-input">
      <input v-model="inputText" @keyup.enter="sendMessage" placeholder="Ask AI..." />
      <button @click="sendMessage">Send</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'

const messages = reactive<Array<{ role: string; content: string }>>([])
const inputText = ref('')
const model = ref('')
const messagesRef = ref<HTMLElement>()

const sendMessage = () => {
  if (!inputText.value.trim()) return
  messages.push({ role: 'user', content: inputText.value })
  inputText.value = ''
  // TODO: Call API
}
</script>

<style scoped>
.ai-chat {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
  font-weight: bold;
}
.model-select {
  font-size: 12px;
  padding: 2px 4px;
}
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}
.message {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.message.user {
  flex-direction: row-reverse;
}
.bubble {
  max-width: 80%;
  padding: 8px 12px;
  border-radius: 8px;
  background: #f0f0f0;
  font-size: 14px;
}
.message.user .bubble {
  background: #1677ff;
  color: white;
}
.chat-input {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid #eee;
}
.chat-input input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
}
.chat-input button {
  padding: 8px 16px;
  background: #1677ff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}
</style>
```

- [ ] **Step 6: Wire up EditorView to load outline on mount**

Update `EditorView.vue`:
```vue
<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import OutlineTree from '../components/OutlineTree.vue'
import ContentEditor from '../components/ContentEditor.vue'
import AIChat from '../components/AIChat.vue'

const route = useRoute()
const docId = route.params.id as string
const docStore = useDocumentStore()

onMounted(() => {
  docStore.loadOutline(docId)
})
</script>
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/views/EditorView.vue frontend/src/components/ frontend/src/stores/documentStore.ts frontend/src/stores/chatStore.ts
git commit -m "feat: implement editor view with three-panel layout"
```

---

### Task 10: Integrate docx-editor / Rich Text Editor

**Files:**
- Modify: `frontend/src/components/ContentEditor.vue`
- Modify: `frontend/package.json`

**Interfaces:**
- Consumes: Section content (HTML string), section ID
- Produces: Updated HTML content on change

- [ ] **Step 1: Install tiptap**

```bash
cd frontend
pnpm add @tiptap/vue-3 @tiptap/starter-kit @tiptap/pm
```

- [ ] **Step 2: Replace placeholder with tiptap editor**

Update `frontend/src/components/ContentEditor.vue`:
```vue
<template>
  <div class="content-editor">
    <div class="editor-header">
      <span class="current-section">{{ currentSectionTitle }}</span>
      <div class="toolbar">
        <button title="AI Assist" @click="toggleAI">✨</button>
        <button title="Save" @click="save">💾</button>
      </div>
    </div>
    <div class="editor-body">
      <Editor v-if="editor" :editor-content="htmlContent" @update:content="onContentChange" />
    </div>
    <div class="editor-footer">
      <button class="outline-btn" @click="extractOutline">Extract Outline</button>
      <button class="export-btn" @click="generateBid">Generate Bid</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useDocumentStore, Section } from '../stores/documentStore'
import { Editor } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'

const docStore = useDocumentStore()
const activeSectionId = computed(() => docStore.activeSectionId)
const currentSection = computed(() => {
  if (!activeSectionId.value) return null
  return docStore.sections.get(activeSectionId.value)
})
const currentSectionTitle = computed(() => currentSection.value?.title || 'No section selected')
const htmlContent = ref<string | null>(null)
const saving = ref(false)

const onContentChange = (newHtml: string) => {
  if (!activeSectionId.value) return
  debouncedSave(activeSectionId.value, newHtml)
}

let debounceTimer: ReturnType<typeof setTimeout>
const debouncedSave = (sectionId: string, content: string) => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    docStore.saveSectionContent(docId, sectionId, content)
  }, 1000)
}

const save = async () => {
  if (!activeSectionId.value || !htmlContent.value) return
  saving.value = true
  await docStore.saveSectionContent(docId, activeSectionId.value, htmlContent.value)
  saving.value = false
}

const docId = computed(() => route.params.id as string)
import { useRoute } from 'vue-router'
const route = useRoute()

// Watch for section changes and load content
watch(activeSectionId, async (newId) => {
  if (newId) {
    await docStore.loadSection(docId.value, newId)
    const section = docStore.sections.get(newId)
    htmlContent.value = section?.content || '<p></p>'
  }
})

onMounted(async () => {
  if (activeSectionId.value) {
    await docStore.loadSection(docId.value, activeSectionId.value)
    const section = docStore.sections.get(activeSectionId.value)
    htmlContent.value = section?.content || '<p></p>'
  }
})

const toggleAI = () => {}
const extractOutline = () => {}
const generateBid = () => {}
</script>
```

Note: The above imports `route` and `docId` — refactor to proper setup order:

```vue
<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useDocumentStore } from '../stores/documentStore'
import { Editor } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'

const route = useRoute()
const docStore = useDocumentStore()
const docId = computed(() => route.params.id as string)

const activeSectionId = computed(() => docStore.activeSectionId)
const currentSection = computed(() => {
  if (!activeSectionId.value) return null
  return docStore.sections.get(activeSectionId.value)
})
const currentSectionTitle = computed(() => currentSection.value?.title || 'No section selected')
const htmlContent = ref<string | null>(null)
const saving = ref(false)

let debounceTimer: ReturnType<typeof setTimeout>
const debouncedSave = (sectionId: string, content: string) => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    docStore.saveSectionContent(docId.value, sectionId, content)
  }, 1000)
}

const onContentChange = (newHtml: string) => {
  if (!activeSectionId.value) return
  debouncedSave(activeSectionId.value, newHtml)
}

const save = async () => {
  if (!activeSectionId.value || !htmlContent.value) return
  saving.value = true
  await docStore.saveSectionContent(docId.value, activeSectionId.value, htmlContent.value)
  saving.value = false
}

watch(activeSectionId, async (newId) => {
  if (newId) {
    await docStore.loadSection(docId.value, newId)
    const section = docStore.sections.get(newId)
    htmlContent.value = section?.content || '<p></p>'
  }
})

onMounted(async () => {
  // Content will be loaded when a section is selected from the tree
})

const toggleAI = () => {}
const extractOutline = () => {}
const generateBid = () => {}
</script>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/ContentEditor.vue frontend/package.json
git commit -m "feat: integrate tiptap rich text editor for content editing"
```

---

### Task 11: Implement Outline Tree Interactivity

**Files:**
- Modify: `frontend/src/components/OutlineTree.vue`

**Interfaces:**
- Consumes: `documentStore.outline`, `documentStore.activeSectionId`
- Produces: Select event, reorder callback

- [ ] **Step 1: Make outline tree interactive**

Update `frontend/src/components/OutlineTree.vue`:
```vue
<template>
  <div class="outline-tree">
    <div class="tree-header">
      <span>Outline</span>
      <button @click="addSection" title="Add section">+</button>
    </div>
    <div class="tree-list">
      <OutlineNode
        v-for="section in outline"
        :key="section.id"
        :section="section"
        :depth="0"
        :active-id="activeSectionId"
        @select="$emit('select', $event)"
        @add-child="$emit('add-child', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useDocumentStore, Section } from '../stores/documentStore'
import OutlineNode from './OutlineNode.vue'

const docStore = useDocumentStore()
const outline = computed(() => docStore.outline)
const activeSectionId = computed(() => docStore.activeSectionId)
defineEmits<{ select: [id: string]; addChild: [parentId: string] }>()

const addSection = () => {
  // Add new top-level section
  const newSection: Section = {
    id: Date.now().toString(),
    title: 'New Section',
    level: 1,
    parent_id: '',
    content: '',
    children: [],
  }
  docStore.outline.push(newSection)
}
</script>
```

- [ ] **Step 2: Create recursive OutlineNode component**

`frontend/src/components/OutlineNode.vue`:
```vue
<template>
  <div class="outline-node" :style="{ paddingLeft: depth * 16 + 'px' }">
    <div
      class="node-row"
      :class="{ active: section.id === activeId }"
      @click="$emit('select', section.id)"
    >
      <button
        v-if="section.children.length > 0"
        class="toggle-btn"
        @click.stop="toggleExpand"
      >
        {{ expanded ? '▼' : '▶' }}
      </button>
      <span v-else class="spacer"></span>
      <span class="icon">📄</span>
      <input
        v-if="editing"
        v-model="editTitle"
        class="edit-input"
        @blur="finishEdit"
        @keyup.enter="finishEdit"
        @keyup.escape="cancelEdit"
      />
      <span v-else class="title">{{ section.title }}</span>
      <button class="delete-btn" @click.stop="remove" title="Delete">×</button>
    </div>
    <div v-if="expanded && section.children.length" class="children">
      <OutlineNode
        v-for="child in section.children"
        :key="child.id"
        :section="child"
        :depth="depth + 1"
        :active-id="activeId"
        @select="$emit('select', $event)"
        @add-child="$emit('add-child', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Section } from '../../stores/documentStore'

const props = defineProps<{
  section: Section
  depth: number
  activeId: string
}>()

defineEmits<{ select: [id: string]; addChild: [parentId: string] }>()

const expanded = ref(true)
const editing = ref(false)
const editTitle = ref(props.section.title)

watch(() => props.section.title, (val) => {
  editTitle.value = val
})

const toggleExpand = () => {
  expanded.value = !expanded.value
}

const finishEdit = () => {
  props.section.title = editTitle.value
  editing.value = false
}

const cancelEdit = () => {
  editTitle.value = props.section.title
  editing.value = false
}

const remove = () => {
  // Remove from store
  emit('select', props.section.id) // Select sibling or parent
}
</script>

<style scoped>
.outline-node {
  margin-bottom: 2px;
}
.node-row {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}
.node-row.active {
  background: #e6f0ff;
}
.toggle-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  font-size: 10px;
  width: 16px;
}
.spacer {
  width: 16px;
}
.delete-btn {
  background: none;
  border: none;
  cursor: pointer;
  opacity: 0;
  color: #999;
}
.node-row:hover .delete-btn {
  opacity: 1;
}
.edit-input {
  flex: 1;
  border: 1px solid #1677ff;
  border-radius: 2px;
  padding: 2px 4px;
  font-size: 14px;
}
</style>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/OutlineTree.vue frontend/src/components/OutlineNode.vue
git commit -m "feat: implement interactive outline tree with expand/collapse/edit"
```

---

### Task 12: Integrate AI Chat with Backend

**Files:**
- Modify: `frontend/src/components/AIChat.vue`
- Modify: `frontend/src/stores/chatStore.ts`
- Modify: `frontend/src/api/client.ts`

**Interfaces:**
- Consumes: `sendChat()` API call, document context
- Produces: Chat messages displayed in AI chat panel

- [ ] **Step 1: Enhance chatStore**

`frontend/src/stores/chatStore.ts`:
```typescript
import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import { sendChat } from '../api/client'

export interface ChatMessage {
  role: 'user' | 'ai'
  content: string
}

export const useChatStore = defineStore('chat', () => {
  const messages = reactive<ChatMessage[]>([])
  const mode = ref<'free' | 'context'>('free')
  const model = ref('')
  const isSending = ref(false)

  const sendMessage = async (text: string, sectionId?: string) => {
    if (isSending.value) return
    isSending.value = true

    messages.push({ role: 'user', content: text })

    try {
      const res = await sendChat({
        message: text,
        mode: mode.value,
        section_id: sectionId || null,
        history: messages.map(m => ({ role: m.role, content: m.content })),
        model: model.value,
      })

      messages.push({ role: 'ai', content: res.data.reply })
    } catch (err) {
      console.error('Chat failed:', err)
      messages.push({ role: 'ai', content: 'Sorry, something went wrong.' })
    } finally {
      isSending.value = false
    }
  }

  const setMode = (m: 'free' | 'context') => { mode.value = m }
  const setModel = (m: string) => { model.value = m }

  return { messages, mode, model, isSending, sendMessage, setMode, setModel }
})
```

- [ ] **Step 2: Wire up AIChat component**

Update `frontend/src/components/AIChat.vue`:
```vue
<template>
  <div class="ai-chat">
    <div class="chat-header">
      <span>AI Assistant</span>
      <div class="header-controls">
        <select v-model="chatStore.model" class="model-select">
          <option value="">Select Model</option>
        </select>
        <button
          :class="{ active: chatStore.mode === 'context' }"
          @click="chatStore.setMode(chatStore.mode === 'context' ? 'free' : 'context')"
          title="Toggle context mode"
        >
          {{ chatStore.mode === 'context' ? '📝 Context' : '💬 Free' }}
        </button>
      </div>
    </div>
    <div class="chat-messages" ref="messagesRef">
      <div
        v-for="(msg, i) in chatStore.messages"
        :key="i"
        class="message"
        :class="msg.role"
      >
        <div v-if="msg.role === 'ai'" class="avatar">🤖</div>
        <div class="bubble">{{ msg.content }}</div>
      </div>
      <div v-if="chatStore.isSending" class="message ai">
        <div class="avatar">🤖</div>
        <div class="bubble">Thinking...</div>
      </div>
    </div>
    <div class="chat-input">
      <input
        v-model="inputText"
        @keyup.enter="handleSend"
        placeholder="Type a message..."
      />
      <button @click="handleSend" :disabled="chatStore.isSending">
        {{ chatStore.isSending ? '...' : 'Send' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from 'vue'
import { useChatStore } from '../stores/chatStore'
import { useDocumentStore } from '../stores/documentStore'

const chatStore = useChatStore()
const docStore = useDocumentStore()
const inputText = ref('')
const messagesRef = ref<HTMLElement>()

const handleSend = () => {
  if (!inputText.value.trim()) return
  const text = inputText.value
  inputText.value = ''
  const sectionId = chatStore.mode === 'context' ? docStore.activeSectionId : undefined
  chatStore.sendMessage(text, sectionId)
  scrollToBottom()
}

const scrollToBottom = async () => {
  await nextTick()
  if (messagesRef.value) {
    messagesRef.value.scrollTop = messagesRef.value.scrollHeight
  }
}
</script>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/stores/chatStore.ts frontend/src/components/AIChat.vue frontend/src/api/client.ts
git commit -m "feat: integrate AI chat with backend and context mode"
```

---

### Task 13: Implement Export and Bottom Actions

**Files:**
- Modify: `frontend/src/components/ContentEditor.vue`
- Modify: `backend/internal/handler/handler.go` (ensure export route is wired)

**Interfaces:**
- Consumes: Document ID
- Produces: .docx file download

- [ ] **Step 1: Wire up export button in ContentEditor**

Update `frontend/src/components/ContentEditor.vue`, add to footer buttons:
```vue
<script setup>
// ... existing code ...

import { exportDocument } from '../api/client'
import { useRoute } from 'vue-router'

const route = useRoute()
const docId = computed(() => route.params.id as string)

const generateBid = async () => {
  try {
    const res = await exportDocument(docId.value)
    // Trigger download
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', 'bid-document.docx')
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  } catch (err) {
    console.error('Export failed:', err)
    alert('Export failed. Please try again.')
  }
}
</script>
```

- [ ] **Step 2: Wire up extract outline button**

The "Extract Outline" button in the editor footer should be a no-op after initial upload (outline is already extracted). Could be used to re-extract:
```vue
const extractOutline = () => {
  // Re-extract outline from current document
  // For now, just a placeholder
  console.log('Re-extracting outline...')
}
```

- [ ] **Step 3: Verify backend export route**

Ensure `handler.go` has the export route registered in `RegisterRoutes`:
```go
doc.POST("/:id/export", h.ExportDocument)
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ContentEditor.vue backend/internal/handler/handler.go
git commit -m "feat: implement document export and bottom action buttons"
```

---

### Task 14: Polish, Styling, and Dev Experience

**Files:**
- Modify: `frontend/src/App.vue`
- Modify: `frontend/src/views/UploadView.vue` (match design)
- Modify: `backend/cmd/server/main.go` (add logging, graceful shutdown)

**Interfaces:**
- Consumes: None
- Produces: Polished UI matching Calicat design, better dev experience

- [ ] **Step 1: Add top navigation bar to EditorView matching design**

Update `frontend/src/views/EditorView.vue` header:
```vue
<header class="top-bar">
  <div class="logo-area">
    <div class="logo-icon">📋</div>
    <span class="brand-name">Bid-Maker</span>
  </div>
  <div class="right-actions">
    <button class="icon-btn" title="Help">?</button>
    <button class="icon-btn" title="Settings">⚙</button>
  </div>
</header>
```

- [ ] **Step 2: Match UploadView to design**

Ensure UploadView has:
- Upload icon (centered, large)
- "Drop your .docx file here" text
- Upload button
- Bottom action buttons: "Extract Outline" and "Generate Bid" (shown after upload)

- [ ] **Step 3: Add graceful shutdown to Go backend**

Update `backend/cmd/server/main.go`:
```go
import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

// In main():
srv := &http.Server{
    Addr:    ":" + fmt.Sprintf("%d", port),
    Handler: r,
}

go func() {
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Server failed: %v", err)
    }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit
log.Println("Shutting down server...")

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
srv.Shutdown(ctx)
log.Println("Server exited")
```

- [ ] **Step 4: Add Go dev hot reload**

Install `air` for Go hot reload:
```bash
cd backend
go install github.com/cosmtrek/air@latest
```

Create `backend/.air.toml`:
```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build ./cmd/server/"
bin = "tmp/server"
full_bin = "tmp/server"
include_ext = ["go"]
exclude_dir = ["tmp"]
```

- [ ] **Step 5: Final commit**

```bash
git add frontend/src/ backend/cmd/server/main.go
git commit -m "chore: polish UI, add graceful shutdown and dev hot reload"
```

---

## Execution Order

Tasks must be completed in order (1 → 14):

1. Backend init & structure
2. LLM provider abstraction
3. .docx upload & outline extraction
4. Document editing endpoints
5. AI chat endpoint
6. .docx export
7. Frontend Vue init
8. Upload view
9. Editor view layout
10. Rich text editor integration
11. Outline tree interactivity
12. AI chat integration
13. Export & bottom actions
14. Polish & dev experience

## Testing Strategy

- **Backend:** Table-driven tests for each handler, unit tests for services
- **Frontend:** Manual testing via browser during dev, add Vitest tests later
- **Integration:** Test full flow: upload → edit → chat → export

## Notes

- unidoc may require a license key for production use. For local development, the trial mode is sufficient.
- docx-editor outputs HTML. The current implementation strips HTML for plain text export. Full HTML-to-docx conversion with formatting can be added later.
- The in-memory document store is a placeholder. Add file-based or database persistence in a future iteration.
- Auth is intentionally omitted but the handler structure leaves room for middleware-based auth.
