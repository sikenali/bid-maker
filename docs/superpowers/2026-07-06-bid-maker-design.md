# Bid-Maker Design Spec

**Date:** 2026-07-06
**Status:** Approved
**Stack:** Vue 3 + Go + unioffice + Multi-provider LLM

## Architecture

Monorepo with pnpm workspaces. Frontend (Vue 3 SPA) and backend (Go service) run as separate processes during local development.

```
bid-maker/
├── pnpm-workspace.yaml
├── frontend/              # Vue 3 + Vite + TypeScript
│   ├── src/
│   │   ├── views/         # UploadView, EditorView
│   │   ├── components/    # OutlineTree, ContentEditor, AIChat
│   │   ├── stores/        # Pinia stores
│   │   ├── api/           # Axios + endpoints
│   │   └── router/
│   └── package.json
└── backend/               # Go 1.21+ + Gin
    ├── cmd/server/main.go
    ├── internal/
    │   ├── handler/       # HTTP handlers
    │   ├── service/       # Business logic
    │   ├── model/         # Data models
    │   └── config/        # Config management
    └── go.mod
```

## Frontend

### Views
- **UploadView** — Drag & drop + button upload, transitions to EditorView after upload
- **EditorView** — Three-panel layout (outline | editor | AI chat)

### Components
- **OutlineTree** — Left sidebar, editable tree nodes (expand/collapse/reorder), syncs to backend
- **ContentEditor** — Uses docx-editor (rich text editor for Vue), supports bold/italic/lists/tables, synced to backend on change
- **AIChat** — Right panel, two modes:
  - Free-form mode: general AI conversation
  - Context mode: AI knows current outline section + content being edited
- **ModelSelector** — Dropdown to switch LLM provider/model

### State Management (Pinia)
- **documentStore** — Current document state (outline, content sections, file ID)
- **chatStore** — Message history, current mode, active section context
- **uiStore** — Loading states, active panel, selected section

## Backend

### Routes
- `POST /api/upload` — Upload .docx file, returns document ID
- `GET /api/document/:id/outline` — Get outline from document
- `PUT /api/document/:id/outline` — Update outline (reorder/add/remove sections)
- `GET /api/document/:id/section/:sectionId` — Get content of a section
- `PUT /api/document/:id/section/:sectionId` — Save section content
- `POST /api/document/:id/export` — Generate and download .docx
- `POST /api/chat` — AI chat endpoint (supports context/free-form modes)
- `GET /api/config/models` — List available LLM models

### Services
- **DocxService** — Uses unioffice to parse .docx (extract headings/content) and generate final document
- **OutlineService** — Extract headings from parsed docx, manage outline structure
- **LLMService** — Abstract LLM provider interface, supports multiple backends (OpenAI, 通义千问, 智谱, etc.) via config
- **ChatService** — Manages conversation context, routes to LLM

### Config
Environment variables or `.env` file for LLM API keys, base URLs, model names. Pluggable provider registry.

## Data Flow

### Upload Flow
1. User uploads .docx → Frontend calls `POST /api/upload`
2. Backend parses docx with unioffice, extracts outline + content
3. Returns document ID + initial outline structure
4. Frontend navigates to EditorView, loads outline into OutlineTree

### Editing Flow
1. User clicks an outline node → Frontend fetches that section's content
2. Content loads into ContentEditor (docx-editor)
3. As user types, debounced save to backend (`PUT /api/document/:id/section/:sectionId`)
4. Backend persists changes

### AI Chat Flow
- **Free-form:** User types prompt → `POST /api/chat` with `mode: "free"` → LLM responds
- **Context:** User types prompt → Frontend sends current section content + outline as context → `POST /api/chat` with `mode: "context"` → LLM generates content-aware response → User can "insert" suggested content into editor

### Export Flow
1. User clicks "标书生成" → Frontend calls `POST /api/document/:id/export`
2. Backend collects all sections in outline order, builds complete .docx with unioffice
3. Returns file download stream

## Decisions Summary

| Decision | Choice |
|----------|--------|
| Project structure | Monorepo with pnpm workspaces |
| Backend language | Go |
| LLM provider | Multi-provider, configurable |
| Document format | .docx only |
| Backend structure | Simple flat (routes + handlers together) |
| AI chat modes | Hybrid (context + free-form) |
| Outline editing | User-editable after extraction |
| Content editor | docx-editor (rich text) |
| Auth | Not yet, leave room for it |
| Deployment | Local dev only for now |
| Export | Single .docx download |
