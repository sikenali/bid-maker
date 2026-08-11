package service

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertFileToMarkdown_Integration(t *testing.T) {
	bin := os.Getenv("ANYDOC_BIN")
	if bin == "" {
		t.Skip("ANYDOC_BIN not set; skipping integration test")
	}
	mdIn := "# 第一章 项目概况\n\n导语。\n\n## 1.1 背景\n\n背景内容。"
	data, err := BuildDocxFromMarkdown(mdIn)
	assert.NoError(t, err)
	dir := t.TempDir()
	p := dir + "/sample.docx"
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatal(err)
	}
	svc := NewAnyDocService()
	mdOut, err := svc.ConvertFileToMarkdown(context.Background(), p)
	assert.NoError(t, err)
	sections, err := ParseSectionsWithContent(mdOut)
	assert.NoError(t, err)
	assert.Len(t, sections, 1)
	assert.Equal(t, "第一章 项目概况", sections[0].Title)
}
