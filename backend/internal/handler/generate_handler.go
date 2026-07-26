package handler

import (
	"net/http"

	"github.com/example/bid-maker-backend/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GenerateOutline(c *gin.Context) {
	if h.generateSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generate service not initialized"})
		return
	}
	var req service.GenerateOutlineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	outline, err := h.generateSvc.GenerateOutline(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"outline": outline})
}

func (h *Handler) GenerateSection(c *gin.Context) {
	if h.generateSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "generate service not initialized"})
		return
	}
	var req service.GenerateSectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.generateSvc.GenerateSectionStream(c.Request.Context(), req, c.Writer)
}
