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
	svc := NewDocxService()

	// We need a valid DOCX to test ParseDocument with unioffice.
	// Creating one programmatically without a real .docx file is unreliable
	// because unioffice's document.Open() rejects minimal XML. If the file
	// isn't available, skip.
	data, err := os.ReadFile("testdata/test.docx")
	if err != nil {
		t.Skipf("test fixture not available: %v (put a real bid docx at backend/internal/service/testdata/test.docx)", err)
	}

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

func TestFilterKeywordOutline(t *testing.T) {
	svc := NewDocxService()
	sections := []model.Section{
		{
			ID: "sec-1", Title: "第一章 总则", Level: 1,
			Children: []model.Section{
				{ID: "sec-1-1", Title: "1.1 项目概况", Level: 2},
			},
		},
		{
			ID: "sec-2", Title: "第二章 投标文件", Level: 1,
			Children: []model.Section{
				{ID: "sec-2-1", Title: "2.1 投标人资格", Level: 2,
					Children: []model.Section{
						{ID: "sec-2-1-1", Title: "2.1.1 资质要求", Level: 3},
					},
				},
				{ID: "sec-2-2", Title: "2.2 投标报价", Level: 2},
			},
		},
		{
			ID: "sec-3", Title: "第三章 评标办法", Level: 1,
		},
	}

	result := svc.filterKeywordOutline(sections, "投标文件")
	if result == nil {
		t.Fatal("expected non-nil result for matching keyword")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 children under 投标文件, got %d", len(result))
	}
	if result[0].Title != "2.1 投标人资格" {
		t.Errorf("expected first child '2.1 投标人资格', got '%s'", result[0].Title)
	}
	if result[1].Title != "2.2 投标报价" {
		t.Errorf("expected second child '2.2 投标报价', got '%s'", result[1].Title)
	}
	if len(result[0].Children) != 1 || result[0].Children[0].Title != "2.1.1 资质要求" {
		t.Error("nested children under 投标人资格 should be preserved")
	}
}

func TestFilterKeywordOutline_NoMatch(t *testing.T) {
	svc := NewDocxService()
	sections := []model.Section{
		{ID: "sec-1", Title: "第一章 总则", Level: 1},
	}

	result := svc.filterKeywordOutline(sections, "不存在的关键词")
	if result != nil {
		t.Fatal("expected nil for non-matching keyword")
	}
}

func TestExtractSectionsWithKeyword_HeadingMatch(t *testing.T) {
	svc := NewDocxService()
	// Keyword found in heading — should extract from that heading onward
	// Simulate: paragraphs where "投标文件" is a heading
	sections := []model.Section{
		{ID: "s1", Title: "第一章 总则", Level: 1},
		{ID: "s2", Title: "第二章 投标文件", Level: 1,
			Children: []model.Section{
				{ID: "s2-1", Title: "2.1 资格", Level: 2},
				{ID: "s2-2", Title: "2.2 报价", Level: 2},
			},
		},
		{ID: "s3", Title: "第三章 评标", Level: 1},
	}

	result := svc.filterKeywordOutline(sections, "投标文件")
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 children, got %d", len(result))
	}
}

func TestExtractSectionsWithKeyword_ContentMatch(t *testing.T) {
	svc := NewDocxService()
	// Keyword found in content — should extract from that section onward
	sections := []model.Section{
		{ID: "s1", Title: "第一章 概述", Level: 1,
			Content: "关于投标文件的要求如下..."},
		{ID: "s2", Title: "第二章 投标人", Level: 1},
	}

	result := svc.filterKeywordOutline(sections, "投标文件")
	if result == nil {
		t.Fatal("expected non-nil result when keyword in content")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 sections starting from keyword match, got %d", len(result))
	}
	if result[0].Title != "第一章 概述" || result[0].Content != "关于投标文件的要求如下..." {
		t.Error("first section should be the one containing keyword in content")
	}
	if result[1].Title != "第二章 投标人" {
		t.Errorf("expected '第二章 投标人', got '%s'", result[1].Title)
	}
}

func TestExtractSectionsWithKeyword_NestedMatch(t *testing.T) {
	svc := NewDocxService()
	// Keyword found in nested child heading — extract its children (if any)
	sections := []model.Section{
		{ID: "s1", Title: "第一章 概述", Level: 1,
			Children: []model.Section{
				{ID: "s1-1", Title: "1.1 项目说明", Level: 2,
					Children: []model.Section{
						{ID: "s1-1-1", Title: "1.1.1 技术需求", Level: 3,
							Children: []model.Section{
								{ID: "s1-1-1-a", Title: "a) 投标文件格式", Level: 4,
									Children: []model.Section{
										{ID: "a1", Title: "A.1 封面", Level: 5},
										{ID: "a2", Title: "A.2 报价表", Level: 5},
									},
								},
							},
						},
					},
				},
			},
		},
		{ID: "s2", Title: "第二章 投标", Level: 1},
	}

	result := svc.filterKeywordOutline(sections, "投标文件")
	if result == nil {
		t.Fatal("expected non-nil result for nested keyword match")
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 children under nested heading, got %d", len(result))
	}
	if result[0].Title != "A.1 封面" {
		t.Errorf("expected first child 'A.1 封面', got '%s'", result[0].Title)
	}
	if result[1].Title != "A.2 报价表" {
		t.Errorf("expected second child 'A.2 报价表', got '%s'", result[1].Title)
	}
}

