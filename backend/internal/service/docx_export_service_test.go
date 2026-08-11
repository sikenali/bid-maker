package service

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDocxFromMarkdown_CreatesValidZip(t *testing.T) {
	md := "# 第一章 项目概况\n\n导语。\n\n## 1.1 背景\n\n背景内容。"
	data, err := BuildDocxFromMarkdown(md)
	assert.NoError(t, err)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	assert.NoError(t, err)
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	assert.True(t, names["[Content_Types].xml"])
	assert.True(t, names["_rels/.rels"])
	assert.True(t, names["word/document.xml"])

	var docXML []byte
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, _ := f.Open()
			buf := new(bytes.Buffer)
			buf.ReadFrom(rc)
			rc.Close()
			docXML = buf.Bytes()
		}
	}
	body := string(docXML)
	assert.Contains(t, body, "第一章 项目概况")
	assert.Contains(t, body, `w:pStyle w:val="Heading1"`)
	assert.Contains(t, body, "背景内容。")
}

func TestBuildDocxFromMarkdown_Empty(t *testing.T) {
	data, err := BuildDocxFromMarkdown("")
	assert.NoError(t, err)
	assert.NotNil(t, data)
}
