package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultPort(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("LLM_PROVIDERS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 0 {
		t.Errorf("expected default port 0, got %d", cfg.Port)
	}
}

func TestLoad_CustomPort(t *testing.T) {
	os.Setenv("PORT", "3000")
	defer os.Unsetenv("PORT")
	os.Unsetenv("LLM_PROVIDERS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != 3000 {
		t.Errorf("expected port 3000, got %d", cfg.Port)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	os.Setenv("PORT", "not-a-number")
	defer os.Unsetenv("PORT")
	os.Unsetenv("LLM_PROVIDERS")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestLoad_LLMProviders(t *testing.T) {
	os.Unsetenv("PORT")
	os.Setenv("LLM_PROVIDERS", `[{"name":"test-provider","api_key":"key123","base_url":"http://localhost/v1","model_name":"test-model","type":"openai"}]`)
	defer os.Unsetenv("LLM_PROVIDERS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.LLMProviders) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(cfg.LLMProviders))
	}
	p := cfg.LLMProviders[0]
	if p.Name != "test-provider" {
		t.Errorf("expected name 'test-provider', got '%s'", p.Name)
	}
	if p.APIKey != "key123" {
		t.Errorf("expected api_key 'key123', got '%s'", p.APIKey)
	}
	if p.BaseURL != "http://localhost/v1" {
		t.Errorf("expected base_url 'http://localhost/v1', got '%s'", p.BaseURL)
	}
	if p.ModelName != "test-model" {
		t.Errorf("expected model_name 'test-model', got '%s'", p.ModelName)
	}
	if p.Type != "openai" {
		t.Errorf("expected type 'openai', got '%s'", p.Type)
	}
}

func TestLoad_LLMProviders_EnvExpansion(t *testing.T) {
	os.Unsetenv("PORT")
	os.Setenv("MY_API_KEY", "secret-from-env")
	os.Setenv("LLM_PROVIDERS", `[{"name":"expanded","api_key":"${MY_API_KEY}","base_url":"http://example.com","model_name":"gpt-4","type":"openai"}]`)
	defer func() {
		os.Unsetenv("MY_API_KEY")
		os.Unsetenv("LLM_PROVIDERS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.LLMProviders) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(cfg.LLMProviders))
	}
	if cfg.LLMProviders[0].APIKey != "secret-from-env" {
		t.Errorf("expected expanded api_key 'secret-from-env', got '%s'", cfg.LLMProviders[0].APIKey)
	}
}

func TestLoad_LLMProviders_InvalidJSON(t *testing.T) {
	os.Unsetenv("PORT")
	os.Setenv("LLM_PROVIDERS", "not-valid-json")
	defer os.Unsetenv("LLM_PROVIDERS")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
