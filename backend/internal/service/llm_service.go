package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

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

// ApiKeyTestResult holds the result of an API key verification.
type ApiKeyTestResult struct {
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

// detectProviderType determines whether the provider uses Anthropic or OpenAI format.
func detectProviderType(provider string, model string) string {
	lower := strings.ToLower(provider + " " + model)
	if strings.Contains(lower, "anthropic") ||
		strings.Contains(lower, "claude") {
		return "anthropic"
	}
	return "openai"
}

// getDefaultBaseURL returns the default endpoint for the given provider.
func getDefaultBaseURL(provider string) string {
	switch provider {
	case "阿里云", "Alibaba":
		return "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	case "百度", "Baidu":
		return "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions"
	case "智谱", "Zhipu":
		return "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	case "OpenAI":
		return "https://api.openai.com/v1/chat/completions"
	case "Anthropic":
		return "https://api.anthropic.com/v1/messages"
	default:
		return "https://api.openai.com/v1/chat/completions"
	}
}

// TestApiKey sends a lightweight request to verify the API key is valid.
// Uses GET /v1/models first (like `curl -H "Authorization: Bearer KEY" URL/v1/models`),
// falls back to minimal POST /chat/completions if models endpoint not available.
func TestApiKey(provider string, apiKey string, model string, customEndpoint string, customFormat string) *ApiKeyTestResult {
	if customEndpoint != "" {
		return testCustomEndpoint(customEndpoint, apiKey, model, customFormat)
	}

	ptype := detectProviderType(provider, model)

	switch ptype {
	case "anthropic":
		return testAnthropicMinimal(apiKey, model)
	default:
		baseURL := normalizeEndpoint(getDefaultBaseURL(provider))
		return testModelsEndpoint(baseURL, apiKey)
	}
}

// normalizeEndpoint removes trailing /chat/completions or /messages segments.
func normalizeEndpoint(ep string) string {
	ep = strings.TrimRight(ep, "/")
	if idx := strings.Index(ep, "/chat/completions"); idx >= 0 {
		ep = ep[:idx]
	} else if idx := strings.Index(ep, "/messages"); idx >= 0 {
		ep = ep[:idx]
	}
	return ep
}

// testCustomEndpoint follows curl-style connectivity testing:
// 1. GET /v1/models — no body, instant, no token usage
// 2. If 404, fall back to minimal POST with single message
func testCustomEndpoint(endpoint string, apiKey string, model string, format string) *ApiKeyTestResult {
	normalized := normalizeEndpoint(endpoint)

	// Step 1: GET /v1/models (standard curl approach)
	modelsURL := normalized + "/models"
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("request error: %v", err)}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	httpClient := &http.Client{Timeout: 8 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("connection failed: %v", err)}
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
		return &ApiKeyTestResult{Available: true}
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return &ApiKeyTestResult{Available: false, Error: "API key invalid"}
	case resp.StatusCode == http.StatusNotFound:
		// Models endpoint not supported, try minimal message request (step 2)
		return testCustomMessage(normalized, apiKey, model, format)
	case resp.StatusCode >= 500:
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("server error %d", resp.StatusCode)}
	default:
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("unexpected status %d", resp.StatusCode)}
	}
}

// testCustomMessage sends a minimal chat request for APIs without /models endpoint.
// Equivalent to: curl -X POST $URL/chat/completions -H "Authorization: Bearer KEY" -d '{"model":"...","messages":[{"role":"user","content":"Hi"}]}'
func testCustomMessage(baseURL string, apiKey string, model string, format string) *ApiKeyTestResult {
	var messagesURL string
	isAnthropic := strings.EqualFold(format, "anthropic")

	if isAnthropic {
		messagesURL = baseURL + "/messages"
	} else {
		messagesURL = baseURL + "/chat/completions"
	}

	payload := map[string]any{
		"model":    model,
		"messages": []map[string]string{{"role": "user", "content": "Hi"}},
	}
	if isAnthropic {
		payload["max_tokens"] = 1
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("marshal error: %v", err)}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, messagesURL, bytes.NewReader(body))
	if err != nil {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("request error: %v", err)}
	}

	req.Header.Set("Content-Type", "application/json")
	if isAnthropic {
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	} else {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	httpClient := &http.Client{Timeout: 8 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("network error: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return &ApiKeyTestResult{Available: false, Error: "API key invalid"}
	}

	// Accept success or model-not-found (key is valid, just invalid model name)
	if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooEarly {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("client error %d", resp.StatusCode)}
	}

	if resp.StatusCode >= 500 {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("server error %d", resp.StatusCode)}
	}

	return &ApiKeyTestResult{Available: true}
}

// testModelsEndpoint verifies connectivity and auth via GET /v1/models.
// Same as: curl -H "Authorization: Bearer $KEY" $BASE/v1/models
func testModelsEndpoint(baseURL string, apiKey string) *ApiKeyTestResult {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if _, err := url.Parse(baseURL); err != nil {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("invalid base URL %q", baseURL)}
	}

	modelsURL := strings.TrimRight(baseURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("request error: %v", err)}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	httpClient := &http.Client{Timeout: 8 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("connection failed: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return &ApiKeyTestResult{Available: false, Error: "API key invalid"}
	}

	if resp.StatusCode == http.StatusOK {
		return &ApiKeyTestResult{Available: true}
	}

	if resp.StatusCode >= 500 {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("server error %d", resp.StatusCode)}
	}

	return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("unexpected status %d", resp.StatusCode)}
}

// testAnthropicMinimal verifies Anthropic API key with a bare-minimum request.
func testAnthropicMinimal(apiKey string, model string) *ApiKeyTestResult {
	endpoint := "https://api.anthropic.com/v1/messages"

	payload := map[string]any{
		"model":      model,
		"max_tokens": 1,
		"messages": []map[string]string{
			{"role": "user", "content": "Hi"},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("marshal error: %v", err)}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("request error: %v", err)}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	httpClient := &http.Client{Timeout: 8 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("network error: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return &ApiKeyTestResult{Available: false, Error: "API key invalid"}
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return &ApiKeyTestResult{Available: false, Error: fmt.Sprintf("client error %d", resp.StatusCode)}
	}

	return &ApiKeyTestResult{Available: true}
}
