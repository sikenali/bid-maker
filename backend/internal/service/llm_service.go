package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/example/bid-maker-backend/internal/config"
	"github.com/sashabaranov/go-openai"
)

// Message represents a chat message in the conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMClient is the interface that all LLM providers must implement.
type LLMClient interface {
	Chat(ctx context.Context, messages []Message, model string) (string, error)
	Close() error
}

// LLMRegistry manages LLM provider instances keyed by name.
type LLMRegistry struct {
	clients map[string]LLMClient
	models  []string
}

// NewLLMRegistry creates a registry from the given provider configurations.
func NewLLMRegistry(providers []config.LLMProviderConfig) *LLMRegistry {
	reg := &LLMRegistry{
		clients: make(map[string]LLMClient),
		models:  make([]string, 0, len(providers)),
	}
	for _, p := range providers {
		reg.clients[p.Name] = NewOpenAIProvider(p)
		reg.models = append(reg.models, p.ModelName)
	}
	return reg
}

// GetProvider returns an LLMClient by name.
func (r *LLMRegistry) GetProvider(name string) (LLMClient, error) {
	client, ok := r.clients[name]
	if !ok {
		// Return the first available provider if name is empty
		if name == "" {
			for _, c := range r.clients {
				return c, nil
			}
			return nil, errors.New("no providers configured")
		}
		return nil, fmt.Errorf("provider %q not found", name)
	}
	return client, nil
}

// ListModels returns all available model names.
func (r *LLMRegistry) ListModels() []string {
	return r.models
}

// OpenAIProvider implements LLMClient for any OpenAI-compatible API.
type OpenAIProvider struct {
	client  *openai.Client
	baseURL string
	apiKey  string
	model   string
}

// NewOpenAIProvider creates a provider from config.
func NewOpenAIProvider(cfg config.LLMProviderConfig) *OpenAIProvider {
	return &OpenAIProvider{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		model:   cfg.ModelName,
	}
}

// Chat sends a chat completion request and returns the response content.
func (p *OpenAIProvider) Chat(ctx context.Context, messages []Message, model string) (string, error) {
	if p.apiKey == "" {
		return "", errors.New("API key not configured")
	}

	if model == "" {
		model = p.model
	}

	openAIMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		openAIMessages[i] = openai.ChatCompletionMessage{
			Role:      m.Role,
			Content:   m.Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: openAIMessages,
	}

	client, err := p.getClient()
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no choices in response")
	}

	return resp.Choices[0].Message.Content, nil
}

// Close releases resources held by the provider.
func (p *OpenAIProvider) Close() error {
	return nil
}

// getClient returns a configured OpenAI client.
func (p *OpenAIProvider) getClient() (*openai.Client, error) {
	if _, err := url.Parse(p.baseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL %q: %w", p.baseURL, err)
	}

	cfg := openai.DefaultConfig(p.apiKey)
	cfg.BaseURL = p.baseURL
	return openai.NewClientWithConfig(cfg), nil
}
