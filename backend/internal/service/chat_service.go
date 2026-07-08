package service

import (
	"context"
	"strings"

	"github.com/example/bid-maker-backend/internal/model"
)

type ChatService struct {
	registry *LLMRegistry
}

func NewChatService(registry *LLMRegistry) *ChatService {
	return &ChatService{registry: registry}
}

type ChatRequest struct {
	Message   string     `json:"message"`
	Mode      string     `json:"mode"`
	SectionID string     `json:"section_id,omitempty"`
	History   []Message  `json:"history"`
	Provider  string     `json:"provider,omitempty"`
	Model     string     `json:"model,omitempty"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
	Model string `json:"model"`
}

func (s *ChatService) Chat(ctx context.Context, req ChatRequest, doc *model.Document) (*ChatResponse, error) {
	client, err := s.registry.GetProvider(req.Provider)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var messages []Message

	if req.Mode == "context" && req.SectionID != "" && doc != nil {
		section := findSectionInOutline(doc.Outline, req.SectionID)
		if section != nil {
			contextPrompt := "You are helping write a bid document. Current section: " + section.Title + "\nContent so far: " + section.Content + "\nOutline: " + buildOutlineString(doc.Outline, 0) + "\n\n"
			messages = append(messages, Message{Role: "system", Content: contextPrompt})
		}
	} else {
		messages = append(messages, Message{Role: "system", Content: "You are a helpful assistant for bid document creation."})
	}

	for _, m := range req.History {
		messages = append(messages, m)
	}

	messages = append(messages, Message{Role: "user", Content: req.Message})

	model := req.Model
	if model == "" {
		if p, ok := client.(*OpenAIProvider); ok {
			model = p.model
		}
	}

	reply, err := client.Chat(ctx, messages, model)
	if err != nil {
		return nil, err
	}

	return &ChatResponse{Reply: reply, Model: model}, nil
}

func buildOutlineString(sections []model.Section, depth int) string {
	var sb strings.Builder
	for _, s := range sections {
		sb.WriteString(strings.Repeat("  ", depth) + s.Title + "\n")
		sb.WriteString(buildOutlineString(s.Children, depth+1))
	}
	return sb.String()
}

func findSectionInOutline(sections []model.Section, id string) *model.Section {
	for i := range sections {
		if sections[i].ID == id {
			return &sections[i]
		}
		if child := findSectionInOutline(sections[i].Children, id); child != nil {
			return child
		}
	}
	return nil
}
