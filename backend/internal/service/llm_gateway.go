package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// LLMGateway 模型网关 - 统一OpenAI兼容接口
type LLMGateway struct {
	providers map[string]*LLMProvider
	defaultModel string
}

type LLMProvider struct {
	Name      string
	APIKey    string
	BaseURL   string
	ModelName string
	Client    *openai.Client
}

func NewLLMGateway(providers []LLMProviderConfig, defaultModel string) *LLMGateway {
	gw := &LLMGateway{
		providers: make(map[string]*LLMProvider),
		defaultModel: defaultModel,
	}
	
	for _, p := range providers {
		cfg := openai.DefaultConfig(p.APIKey)
		cfg.BaseURL = p.BaseURL
		client := openai.NewClientWithConfig(cfg)
		
		gw.providers[p.Name] = &LLMProvider{
			Name:      p.Name,
			APIKey:    p.APIKey,
			BaseURL:   p.BaseURL,
			ModelName: p.ModelName,
			Client:    client,
		}
	}
	
	return gw
}

func (gw *LLMGateway) GetProvider(name string) (*LLMProvider, error) {
	if name == "" {
		for _, p := range gw.providers {
			return p, nil
		}
		return nil, fmt.Errorf("no LLM providers configured")
	}
	
	p, ok := gw.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %q not found", name)
	}
	return p, nil
}

func (gw *LLMGateway) ListModels() []string {
	models := make([]string, 0, len(gw.providers))
	for _, p := range gw.providers {
		models = append(models, p.ModelName)
	}
	return models
}

// Chat 发送聊天请求
func (gw *LLMGateway) Chat(ctx context.Context, providerName string, messages []map[string]string, model string) (string, error) {
	provider, err := gw.GetProvider(providerName)
	if err != nil {
		return "", err
	}
	
	if model == "" {
		model = provider.ModelName
	}
	
	openAIMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		role, _ := m["role"]
		content, _ := m["content"]
		openAIMessages[i] = openai.ChatCompletionMessage{
			Role:      role,
			Content:   content,
		}
	}
	
	resp, err := provider.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    model,
		Messages: openAIMessages,
	})
	if err != nil {
		return "", fmt.Errorf("chat completion failed: %w", err)
	}
	
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	
	return resp.Choices[0].Message.Content, nil
}

// ChatStream 流式聊天
func (gw *LLMGateway) ChatStream(ctx context.Context, providerName string, messages []map[string]string, model string, handler func(chunk string)) error {
	provider, err := gw.GetProvider(providerName)
	if err != nil {
		return err
	}
	
	if model == "" {
		model = provider.ModelName
	}
	
	openAIMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		role, _ := m["role"]
		content, _ := m["content"]
		openAIMessages[i] = openai.ChatCompletionMessage{
			Role:      role,
			Content:   content,
		}
	}
	
	stream, err := provider.Client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    model,
		Messages: openAIMessages,
		Stream:   true,
	})
	if err != nil {
		return fmt.Errorf("stream creation failed: %w", err)
	}
	defer stream.Close()
	
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("stream receive error: %w", err)
		}
		
		if len(resp.Choices) > 0 {
			handler(resp.Choices[0].Delta.Content)
		}
	}
}

// TestProvider 测试提供商连接
func (gw *LLMGateway) TestProvider(name, apiKey, model, endpoint, format string) (bool, string) {
	if endpoint != "" {
		return testCustomEndpoint(endpoint, apiKey, model)
	}
	
	provider, err := gw.GetProvider(name)
	if err != nil {
		return false, err.Error()
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	
	_, err = provider.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{Role: "user", Content: "Hi"},
		},
		MaxTokens: 1,
	})
	
	if err != nil {
		return false, err.Error()
	}
	return true, ""
}

func testCustomEndpoint(endpoint, apiKey, model string) (bool, string) {
	modelsURL := strings.TrimRight(endpoint, "/") + "/v1/models"
	
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return false, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()
	
	switch resp.StatusCode {
	case http.StatusOK:
		return true, ""
	case http.StatusUnauthorized, http.StatusForbidden:
		return false, "API key invalid"
	case http.StatusNotFound:
		// Try chat endpoint
		return testChatEndpoint(endpoint, apiKey, model)
	default:
		return false, fmt.Sprintf("unexpected status %d", resp.StatusCode)
	}
}

func testChatEndpoint(endpoint, apiKey, model string) (bool, string) {
	url := strings.TrimRight(endpoint, "/") + "/v1/chat/completions"
	
	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": "Hi"},
		},
		"max_tokens": 1,
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return false, err.Error()
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return false, err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return false, "API key invalid"
	}
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return false, fmt.Sprintf("client error %d", resp.StatusCode)
	}
	if resp.StatusCode >= 500 {
		return false, fmt.Sprintf("server error %d", resp.StatusCode)
	}
	return true, ""
}
