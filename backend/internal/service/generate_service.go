package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/example/bid-maker-backend/internal/model"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

var reHeading = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

var DefaultSkillPrompt = `You are a professional bid document writer with deep expertise in creating comprehensive, well-structured, and persuasive bid proposals. Follow the user's requirements carefully and produce high-quality content.`

type GenerateOutlineRequest struct {
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	Endpoint   string `json:"endpoint"`
	APIKey     string `json:"apiKey"`
	SkillID    string `json:"skill_id"`
	UserPrompt string `json:"user_prompt"`
}

type GenerateSectionRequest struct {
	Provider   string          `json:"provider"`
	Model      string          `json:"model"`
	Endpoint   string          `json:"endpoint"`
	APIKey     string          `json:"apiKey"`
	SkillID    string          `json:"skill_id"`
	SectionID  string          `json:"section_id"`
	Section    model.Section   `json:"section"`
	Outline    []model.Section `json:"outline"`
	UserPrompt string          `json:"user_prompt"`
}

type SectionChunk struct {
	SectionID string `json:"section_id"`
	Chunk     string `json:"chunk,omitempty"`
	Done      bool   `json:"done,omitempty"`
	Error     string `json:"error,omitempty"`
}

type GenerateService struct {
	llm *LLMRegistry
}

func NewGenerateService(llm *LLMRegistry) *GenerateService {
	return &GenerateService{llm: llm}
}

func ParseOutlineFromMD(md string) []model.Section {
	lines := strings.Split(md, "\n")
	var root []model.Section
	var stack []*model.Section

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		matches := reHeading.FindStringSubmatch(trimmed)
		if matches == nil {
			continue
		}
		level := len(matches[1])
		title := strings.TrimSpace(matches[2])

		for len(stack) > 0 && stack[len(stack)-1].Level >= level {
			stack = stack[:len(stack)-1]
		}

		sec := model.Section{
			ID:    uuid.NewString(),
			Title: title,
			Level: level,
		}

		if len(stack) > 0 {
			parent := stack[len(stack)-1]
			sec.ParentID = parent.ID
			parent.Children = append(parent.Children, sec)
			stack = append(stack, &parent.Children[len(parent.Children)-1])
		} else {
			root = append(root, sec)
			stack = append(stack, &root[len(root)-1])
		}
	}
	return root
}

func writeSSEChunk(w http.ResponseWriter, sectionID, chunk string) {
	data, _ := json.Marshal(SectionChunk{
		SectionID: sectionID,
		Chunk:     chunk,
	})
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func writeSSEError(w http.ResponseWriter, sectionID, errMsg string) {
	data, _ := json.Marshal(SectionChunk{
		SectionID: sectionID,
		Error:     errMsg,
	})
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func writeSSEDone(w http.ResponseWriter, sectionID string) {
	data, _ := json.Marshal(SectionChunk{
		SectionID: sectionID,
		Done:      true,
	})
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func (s *GenerateService) generateOutlineSystemPrompt(req GenerateOutlineRequest) string {
	return DefaultSkillPrompt + `

Please generate a detailed outline for a bid document based on the user's requirements below. Return ONLY a markdown outline using headings (# for top-level, ## for subsections, etc.). Do not include any explanatory text outside the markdown.

User requirements: ` + req.UserPrompt + `

Generate a comprehensive markdown outline with at least 2-3 levels of hierarchy where appropriate.`
}

func (s *GenerateService) GenerateOutline(ctx context.Context, req GenerateOutlineRequest) ([]model.Section, error) {
	systemPrompt := s.generateOutlineSystemPrompt(req)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: req.UserPrompt},
	}

	modelName := req.Model
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}

	var reply string

	if req.Endpoint != "" && req.APIKey != "" {
		cfg := openai.DefaultConfig(req.APIKey)
		cfg.BaseURL = req.Endpoint
		client := openai.NewClientWithConfig(cfg)

		openAIMessages := make([]openai.ChatCompletionMessage, len(messages))
		for i, m := range messages {
			openAIMessages[i] = openai.ChatCompletionMessage{
				Role:    m.Role,
				Content: m.Content,
			}
		}

		resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    modelName,
			Messages: openAIMessages,
		})
		if err != nil {
			return nil, fmt.Errorf("custom endpoint chat completion failed: %w", err)
		}
		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("no choices in response")
		}
		reply = resp.Choices[0].Message.Content
	} else {
		client, err := s.llm.GetProvider(req.Provider)
		if err != nil {
			return nil, err
		}

		clientModel := req.Model
		if clientModel == "" {
			if p, ok := client.(*OpenAIProvider); ok {
				clientModel = p.model
			}
		}

		reply, err = client.Chat(ctx, messages, clientModel)
		if err != nil {
			return nil, err
		}
	}

	return ParseOutlineFromMD(reply), nil
}

func (s *GenerateService) GenerateSectionStream(ctx context.Context, req GenerateSectionRequest, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	outlineStr := buildOutlineString(req.Outline, 0)

	systemPrompt := DefaultSkillPrompt + fmt.Sprintf(`

You are now writing a specific section of a bid document.

Section to write: %s

Overall document outline for context:
%s

Write detailed, professional, and comprehensive content for this section. Format the content with appropriate markdown.`,
		req.Section.Title, outlineStr)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: req.UserPrompt},
	}

	modelName := req.Model
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}

	if req.Endpoint != "" && req.APIKey != "" {
		cfg := openai.DefaultConfig(req.APIKey)
		cfg.BaseURL = req.Endpoint
		client := openai.NewClientWithConfig(cfg)

		openAIMessages := make([]openai.ChatCompletionMessage, len(messages))
		for i, m := range messages {
			openAIMessages[i] = openai.ChatCompletionMessage{
				Role:    m.Role,
				Content: m.Content,
			}
		}

		stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model:    modelName,
			Messages: openAIMessages,
		})
		if err != nil {
			writeSSEError(w, req.SectionID, fmt.Sprintf("stream creation failed: %v", err))
			return
		}
		defer stream.Close()

		for {
			resp, err := stream.Recv()
			if err != nil {
				if strings.Contains(err.Error(), "EOF") {
					writeSSEDone(w, req.SectionID)
					return
				}
				writeSSEError(w, req.SectionID, fmt.Sprintf("stream receive error: %v", err))
				return
			}
			if len(resp.Choices) > 0 {
				writeSSEChunk(w, req.SectionID, resp.Choices[0].Delta.Content)
			}
		}
	}

	client, err := s.llm.GetProvider(req.Provider)
	if err != nil {
		writeSSEError(w, req.SectionID, fmt.Sprintf("provider error: %v", err))
		return
	}

	clientModel := req.Model
	if clientModel == "" {
		if p, ok := client.(*OpenAIProvider); ok {
			clientModel = p.model
		}
	}

	stream, err := client.CreateChatCompletionStream(ctx, messages, clientModel)
	if err != nil {
		writeSSEError(w, req.SectionID, fmt.Sprintf("stream creation failed: %v", err))
		return
	}
	defer stream.Close()

	for {
		resp, err := stream.Recv()
		if err != nil {
			if strings.Contains(err.Error(), "EOF") {
				writeSSEDone(w, req.SectionID)
				return
			}
			writeSSEError(w, req.SectionID, fmt.Sprintf("stream receive error: %v", err))
			return
		}
		if len(resp.Choices) > 0 {
			writeSSEChunk(w, req.SectionID, resp.Choices[0].Delta.Content)
		}
	}
}
