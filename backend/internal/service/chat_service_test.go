package service

import (
	"testing"

	"github.com/example/bid-maker-backend/internal/model"
)

func TestBuildOutlineString_WithSections(t *testing.T) {
	sections := []model.Section{
		{Title: "Chapter 1", Level: 1},
		{Title: "Section 1.1", Level: 2},
		{Title: "Chapter 2", Level: 1},
	}
	result := buildOutlineString(sections)
	if result == "" {
		t.Fatal("expected non-empty outline string")
	}
	if len(result) < 20 {
		t.Fatal("expected substantial outline string")
	}
}

func TestBuildOutlineString_Empty(t *testing.T) {
	result := buildOutlineString(nil)
	if result != "" {
		t.Fatal("expected empty outline string for nil input")
	}
}

func TestFindSectionInOutline_Found(t *testing.T) {
	sections := []model.Section{
		{ID: "1", Title: "Chapter 1", Level: 1},
		{ID: "2", Title: "Section 1.1", Level: 2, ParentID: "1"},
	}
	sec := findSectionInOutline(sections, "2")
	if sec == nil {
		t.Fatal("expected to find section")
	}
	if sec.Title != "Section 1.1" {
		t.Fatalf("expected title 'Section 1.1', got %q", sec.Title)
	}
}

func TestFindSectionInOutline_NotFound(t *testing.T) {
	sections := []model.Section{
		{ID: "1", Title: "Chapter 1", Level: 1},
	}
	sec := findSectionInOutline(sections, "nonexistent")
	if sec != nil {
		t.Fatal("expected nil for nonexistent section")
	}
}

func TestChatService_New(t *testing.T) {
	reg := NewLLMRegistry(nil)
	svc := NewChatService(reg)
	if svc == nil {
		t.Fatal("expected non-nil ChatService")
	}
}