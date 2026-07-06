package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port         int              `json:"port" env:"PORT"`
	LLMProviders []LLMProviderConfig
}

type LLMProviderConfig struct {
	Name      string `json:"name" env:"NAME"`
	APIKey    string `json:"api_key" env:"API_KEY"`
	BaseURL   string `json:"base_url" env:"BASE_URL"`
	ModelName string `json:"model_name" env:"MODEL_NAME"`
	Type      string `json:"type" env:"TYPE"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	portStr := os.Getenv("PORT")
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT value: %w", err)
		}
		cfg.Port = p
	}

	providersJSON := os.Getenv("LLM_PROVIDERS")
	if providersJSON != "" {
		var providers []LLMProviderConfig
		// Replace ${VAR} placeholders with env values
		providersJSON = expandEnvPlaceholders(providersJSON)
		err := json.Unmarshal([]byte(providersJSON), &providers)
		if err != nil {
			return nil, fmt.Errorf("failed to parse LLM_PROVIDERS JSON: %w", err)
		}
		cfg.LLMProviders = providers
	}

	return cfg, nil
}

func expandEnvPlaceholders(s string) string {
	result := s
	start := strings.Index(result, "${")
	for start != -1 {
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		varName := result[start+2 : start+end]
		varValue := os.Getenv(varName)
		result = result[:start] + varValue + result[start+end+1:]
		start = strings.Index(result, "${")
	}
	return result
}
