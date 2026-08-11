package service

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildDocxFromMarkdown_CreatesValidZip(t *testing.T) {
	md := `# 第一章 项目概况

导语。

## 1.1 背景

背景内容。

### 1.1.1 细分

价格 < 100 > 50 & 说明 "abc"`
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
	assert.True(t, names["word/styles.xml"])
	assert.True(t, names["word/_rels/document.xml.rels"])

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
	assert.Contains(t, body, "背景内容。")
	assert.Contains(t, body, "细分")
	assert.Contains(t, body, `w:pStyle w:val="Heading1"`)
	assert.Contains(t, body, `w:pStyle w:val="Heading2"`)
	assert.Contains(t, body, `w:pStyle w:val="Heading3"`)
	assert.Contains(t, body, `价格 &lt; 100 &gt; 50 &amp; 说明 &quot;abc&quot;`)
	assert.NotContains(t, body, `价格 < 100 > 50 & 说明 "abc"`)
}

func TestBuildDocxFromMarkdown_Empty(t *testing.T) {
	data, err := BuildDocxFromMarkdown("")
	assert.NoError(t, err)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	assert.NoError(t, err)
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	assert.True(t, names["word/document.xml"])
}
