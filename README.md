# Bid-Maker

Bid document generator with AI assistance. Upload a .docx, extract the outline, edit content with AI help, and export.

## Tech Stack

- **Frontend:** Vue 3 + Vite + TypeScript + tiptap + Pinia
- **Backend:** Go + Gin + unioffice + multi-provider LLM

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+ and pnpm

### Backend

```bash
cd backend
go mod tidy
go run ./cmd/server/
```

The server starts on port 8080 by default.

Configure LLM providers via the `LLM_PROVIDERS` environment variable (JSON array):

```bash
export LLM_PROVIDERS='[{"name":"openai","api_key":"sk-xxx","base_url":"https://api.openai.com/v1","model_name":"gpt-4","type":"openai"}]'
```

### Frontend

```bash
cd frontend
pnpm install
pnpm dev
```

Open http://localhost:3000 in your browser. The dev server proxies `/api` requests to the backend.

## Features

- **Upload:** Drag-and-drop or click to upload a .docx file
- **Outline:** Auto-extracted heading structure, editable (add/remove/reorder sections)
- **Editor:** Rich text editing with tiptap, auto-saves to backend
- **AI Chat:** Context-aware or free-form AI assistance with swappable LLM providers
- **Export:** Generate and download a complete .docx bid document

## Project Structure

```
bid-maker/
├── frontend/            # Vue 3 SPA
│   ├── src/
│   │   ├── api/         # Axios client + endpoints
│   │   ├── components/  # OutlineTree, ContentEditor, AIChat
│   │   ├── router/      # Vue Router
│   │   ├── stores/      # Pinia stores
│   │   └── views/       # UploadView, EditorView
│   └── package.json
├── backend/             # Go service
│   ├── cmd/server/      # Main entry
│   └── internal/
│       ├── config/      # Env-based configuration
│       ├── handler/     # HTTP handlers
│       ├── model/       # Data models
│       └── service/     # Business logic (docx, llm, chat)
└── docs/superpowers/    # Design specs and plans
```

## License

Private
