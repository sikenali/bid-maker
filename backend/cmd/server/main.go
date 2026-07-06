package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/example/bid-maker-backend/internal/config"
	"github.com/example/bid-maker-backend/internal/handler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	h := handler.New()
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
