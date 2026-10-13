package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port           int                `json:"port"`
	LLMProviders   []LLMProviderConfig
	PythonService  string             `json:"python_service"`
	RAGEnabled     bool               `json:"rag_enabled"`
}

type LLMProviderConfig struct {
	Name      string `json:"name"`
	APIKey    string `json:"api_key"`
	BaseURL   string `json:"base_url"`
	ModelName string `json:"model_name"`
	Type      string `json:"type"`
}

func Load() (*Config, error) {
	cfg := &Config{Port: 8080, RAGEnabled: true}

	portStr := os.Getenv("PORT")
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		cfg.Port = p
	}

	cfg.PythonService = os.Getenv("PYTHON_SERVICE")
	if cfg.PythonService == "" {
		cfg.PythonService = "http://localhost:8000"
	}

	rag := os.Getenv("RAG_ENABLED")
	if rag == "false" {
		cfg.RAGEnabled = false
	}

	providersJSON := os.Getenv("LLM_PROVIDERS")
	if providersJSON != "" {
		var providers []LLMProviderConfig
		providersJSON = expandEnvPlaceholders(providersJSON)
		if err := json.Unmarshal([]byte(providersJSON), &providers); err != nil {
			return nil, fmt.Errorf("parse LLM_PROVIDERS: %w", err)
		}
		cfg.LLMProviders = providers
	}

	return cfg, nil
}

func expandEnvPlaceholders(s string) string {
	result := s
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		varName := result[start+2 : start+end]
		varValue := os.Getenv(varName)
		result = result[:start] + varValue + result[start+end+1:]
	}
	return result
}
