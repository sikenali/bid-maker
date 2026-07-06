package service

import (
	"context"
	"testing"

	"github.com/example/bid-maker-backend/internal/config"
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

func TestLLMRegistry_GetProvider_NotFound(t *testing.T) {
	providers := []config.LLMProviderConfig{}
	reg := NewLLMRegistry(providers)
	_, err := reg.GetProvider("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
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

func TestLLMRegistry_ListModels_Empty(t *testing.T) {
	reg := NewLLMRegistry(nil)
	models := reg.ListModels()
	if len(models) != 0 {
		t.Fatalf("expected 0 models, got %d", len(models))
	}
}

func TestOpenAIProvider_New(t *testing.T) {
	cfg := config.LLMProviderConfig{
		Name:      "test",
		APIKey:    "sk-test",
		BaseURL:   "https://api.openai.com/v1",
		ModelName: "gpt-3.5-turbo",
		Type:      "openai",
	}
	p := NewOpenAIProvider(cfg)
	if p == nil {
		t.Fatal("expected provider to be non-nil")
	}
}

func TestOpenAIProvider_Close(t *testing.T) {
	cfg := config.LLMProviderConfig{
		APIKey:    "sk-test",
		BaseURL:   "https://api.openai.com/v1",
		ModelName: "gpt-3.5-turbo",
		Type:      "openai",
	}
	p := NewOpenAIProvider(cfg)
	err := p.Close()
	if err != nil {
		t.Fatalf("expected no error on close, got %v", err)
	}
}

func TestOpenAIProvider_Chat_NoApiKey(t *testing.T) {
	cfg := config.LLMProviderConfig{
		Name:      "test",
		BaseURL:   "https://api.openai.com/v1",
		ModelName: "gpt-3.5-turbo",
		Type:      "openai",
	}
	p := NewOpenAIProvider(cfg)
	defer p.Close()

	_, err := p.Chat(context.Background(), []Message{}, "gpt-3.5-turbo")
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestOpenAIProvider_Chat_InvalidBaseURL(t *testing.T) {
	cfg := config.LLMProviderConfig{
		Name:      "test",
		APIKey:    "sk-test",
		BaseURL:   "://invalid",
		ModelName: "gpt-3.5-turbo",
		Type:      "openai",
	}
	p := NewOpenAIProvider(cfg)
	defer p.Close()

	_, err := p.Chat(context.Background(), []Message{{Role: "user", Content: "hello"}}, "gpt-3.5-turbo")
	if err == nil {
		t.Fatal("expected error for invalid base URL")
	}
}
