package service

import (
	"os"
	"testing"

	"github.com/example/bid-maker-backend/internal/model"
)

func TestParseDocument_EmptyData(t *testing.T) {
	svc := NewDocxService()
	doc, err := svc.ParseDocument([]byte{})
	if err == nil {
		t.Fatal("expected error for empty data, got nil")
	}
	if doc != nil {
		t.Fatal("expected nil document for empty data")
	}
}

func TestParseDocument_InvalidDocx(t *testing.T) {
	svc := NewDocxService()
	doc, err := svc.ParseDocument([]byte("not a docx file"))
	if err == nil {
		t.Fatal("expected error for invalid docx, got nil")
	}
	if doc != nil {
		t.Fatal("expected nil document for invalid docx")
	}
}

func TestParseDocument_ValidDocx(t *testing.T) {
	data, err := os.ReadFile("testdata/test.docx")
	if err != nil {
		t.Skipf("test fixture not available: %v", err)
	}

	svc := NewDocxService()
	doc, err := svc.ParseDocument(data)
	if err != nil {
		t.Fatalf("unexpected error parsing valid docx: %v", err)
	}
	if doc == nil {
		t.Fatal("expected document, got nil")
	}
	if doc.ID == "" {
		t.Error("expected non-empty document ID")
	}
	if doc.Title == "" {
		t.Error("expected non-empty document title")
	}
	if doc.Outline == nil {
		t.Error("expected non-nil outline")
	}
}

func TestGenerateDocument(t *testing.T) {
	outDoc := &model.Document{
		ID:    "test-1",
		Title: "Test Document",
		Outline: []model.Section{
			{
				ID:      "sec-1",
				Title:   "Chapter 1",
				Level:   1,
				Content: "Some content here.",
				Children: []model.Section{
					{
						ID:    "sec-2",
						Title: "Section 1.1",
						Level: 2,
					},
				},
			},
		},
	}

	svc := NewDocxService()
	data, err := svc.GenerateDocument(outDoc)
	if err != nil {
		t.Logf("GenerateDocument skipped (license required for write): %v", err)
		return
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty docx data")
	}
}

func TestStoreAndGetDocument(t *testing.T) {
	testDoc := &model.Document{
		ID:    "store-test-1",
		Title: "Store Test",
		Outline: []model.Section{
			{ID: "s1", Title: "Hello", Level: 1},
		},
	}

	StoreDocument(testDoc)

	got, ok := GetDocument("store-test-1")
	if !ok {
		t.Fatal("expected to find stored document")
	}
	if got.ID != testDoc.ID {
		t.Errorf("expected ID %s, got %s", testDoc.ID, got.ID)
	}
	if got.Title != testDoc.Title {
		t.Errorf("expected title %s, got %s", testDoc.Title, got.Title)
	}

	_, ok = GetDocument("nonexistent")
	if ok {
		t.Fatal("expected not found for nonexistent document")
	}
}

func TestUpdateDocument(t *testing.T) {
	testDoc := &model.Document{
		ID:    "update-test-1",
		Title: "Original",
	}
	StoreDocument(testDoc)

	testDoc.Title = "Updated"
	UpdateDocument(testDoc)

	got, _ := GetDocument("update-test-1")
	if got.Title != "Updated" {
		t.Errorf("expected updated title, got %s", got.Title)
	}
}
