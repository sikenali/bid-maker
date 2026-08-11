package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/bid-maker-backend/internal/model"
	"github.com/example/bid-maker-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSaveMarkdownUpdatesTitleAndOutline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	doc := &model.Document{
		ID:       "doc-test-save",
		Title:    "Old Title",
		Markdown: "# 旧标题\n\n旧正文。",
	}
	service.StoreDocument(doc)

	h := &Handler{}
	router := gin.New()
	router.PUT("/document/:id/markdown", h.SaveMarkdown)

	req := httptest.NewRequest(http.MethodPut, "/document/doc-test-save/markdown",
		bytes.NewBufferString(`{"markdown":"## 全新标题\n\n更新后的正文。"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	stored, ok := service.GetDocument(doc.ID)
	assert.True(t, ok)
	assert.Equal(t, "全新标题", stored.Title)
	if assert.Len(t, stored.Outline, 1) {
		assert.Equal(t, 2, stored.Outline[0].Level)
		assert.Equal(t, "更新后的正文。", stored.Outline[0].Content)
	}

	req2 := httptest.NewRequest(http.MethodPut, "/document/doc-test-save/markdown",
		bytes.NewBufferString(`{"markdown":""}`))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	stored2, _ := service.GetDocument(doc.ID)
	assert.Len(t, stored2.Outline, 0)
	assert.Equal(t, "全新标题", stored2.Title)
}