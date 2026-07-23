package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	Endpoint  string     `json:"endpoint,omitempty"`
	Format    string     `json:"format,omitempty"`
	APIKey    string     `json:"apiKey,omitempty"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
	Model string `json:"model"`
}

func (s *ChatService) Chat(ctx context.Context, req ChatRequest, doc *model.Document) (*ChatResponse, error) {
	if req.Endpoint != "" {
		return s.chatWithCustomEndpoint(ctx, req, doc)
	}

	if s.registry == nil {
		return nil, fmt.Errorf("no LLM providers configured and no custom endpoint provided")
	}

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

	modelName := req.Model
	if modelName == "" {
		if p, ok := client.(*OpenAIProvider); ok {
			modelName = p.model
		}
	}

	reply, err := client.Chat(ctx, messages, modelName)
	if err != nil {
		return nil, err
	}

	return &ChatResponse{Reply: reply, Model: modelName}, nil
}

func (s *ChatService) chatWithCustomEndpoint(ctx context.Context, req ChatRequest, doc *model.Document) (*ChatResponse, error) {
	if req.APIKey == "" {
		return nil, fmt.Errorf("API key is required for custom endpoint")
	}

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

	baseURL := normalizeEndpointForChat(req.Endpoint)

	if strings.EqualFold(req.Format, "anthropic") {
		return s.chatAnthropic(ctx, baseURL, messages, req.Model, req.APIKey)
	}

	return s.chatOpenAI(ctx, baseURL, messages, req.Model, req.APIKey)
}

func (s *ChatService) chatOpenAI(ctx context.Context, baseURL string, messages []Message, model string, apiKey string) (*ChatResponse, error) {
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/chat/completions"

	openAIMessages := make([]map[string]string, len(messages))
	for i, m := range messages {
		openAIMessages[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	payload := map[string]any{
		"model":    model,
		"messages": openAIMessages,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	httpClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("status %d", resp.StatusCode)
		var rawResp map[string]any
		json.NewDecoder(resp.Body).Decode(&rawResp)
		if errResp, ok := rawResp["error"].(map[string]any); ok {
			if msg, ok := errResp["message"].(string); ok {
				errMsg = msg
			}
		}
		return nil, fmt.Errorf("%s", errMsg)
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response error: %v", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &ChatResponse{Reply: apiResp.Choices[0].Message.Content, Model: model}, nil
}

func (s *ChatService) chatAnthropic(ctx context.Context, baseURL string, messages []Message, model string, apiKey string) (*ChatResponse, error) {
	if model == "" {
		model = "claude-3-haiku-20240307"
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/messages"

	openAIMessages := make([]map[string]any, len(messages))
	for i, m := range messages {
		openAIMessages[i] = map[string]any{"role": m.Role, "content": m.Content}
	}

	var systemMsg string
	if len(openAIMessages) > 0 && openAIMessages[0]["role"] == "system" {
		if content, ok := openAIMessages[0]["content"].(string); ok {
			systemMsg = content
		}
		openAIMessages = openAIMessages[1:]
	}

	payload := map[string]any{
		"model":      model,
		"max_tokens": 4096,
		"messages":   openAIMessages,
	}
	if systemMsg != "" {
		payload["system"] = systemMsg
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	httpClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("status %d", resp.StatusCode)
		var rawResp map[string]any
		json.NewDecoder(resp.Body).Decode(&rawResp)
		if errObj, ok := rawResp["error"].(map[string]any); ok {
			if msg, ok := errObj["message"].(string); ok {
				errMsg = msg
			}
		}
		return nil, fmt.Errorf("%s", errMsg)
	}

	var apiResp struct {
		Content []struct {
			Type  string `json:"type"`
			Text  string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response error: %v", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	return &ChatResponse{Reply: apiResp.Content[0].Text, Model: model}, nil
}

// normalizeEndpointForChat strips the final path segment (/chat/completions or /messages)
// to get the base URL for chat requests.
func normalizeEndpointForChat(ep string) string {
	ep = strings.TrimRight(ep, "/")
	if idx := strings.Index(ep, "/chat/completions"); idx >= 0 {
		return ep[:idx]
	}
	if idx := strings.Index(ep, "/messages"); idx >= 0 {
		return ep[:idx]
	}
	return ep
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
