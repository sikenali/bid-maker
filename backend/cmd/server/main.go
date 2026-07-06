package main

import (
	"fmt"
	"log"

	"github.com/example/bid-maker-backend/internal/config"
	"github.com/example/bid-maker-backend/internal/handler"
	"github.com/example/bid-maker-backend/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	h := handler.New()

	if len(cfg.LLMProviders) > 0 {
		registry := service.NewLLMRegistry(cfg.LLMProviders)
		h.WithLLMRegistry(registry)
	}

	r := gin.Default()
	h.RegisterRoutes(r)

	port := cfg.Port
	if port == 0 {
		port = 8080
	}

	log.Printf("Starting server on port %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
