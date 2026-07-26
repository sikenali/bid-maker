package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOutlineFromMD(t *testing.T) {
	md := `# 第一章 项目概况
## 1.1 项目背景
## 1.2 建设目标
### 1.2.1 总体目标
# 第二章 技术方案`
	sections := ParseOutlineFromMD(md)
	assert.Len(t, sections, 2)
	assert.Equal(t, "第一章 项目概况", sections[0].Title)
	assert.Equal(t, 1, sections[0].Level)
	assert.Len(t, sections[0].Children, 2)
	assert.Equal(t, "1.1 项目背景", sections[0].Children[0].Title)
	assert.Equal(t, "第二章 技术方案", sections[1].Title)
	assert.Equal(t, "1.2 建设目标", sections[0].Children[1].Title)
	assert.Len(t, sections[0].Children[1].Children, 1)
	assert.Equal(t, "1.2.1 总体目标", sections[0].Children[1].Children[0].Title)
}

func TestParseOutlineFromMD_Empty(t *testing.T) {
	sections := ParseOutlineFromMD("")
	assert.Len(t, sections, 0)
}

func TestParseOutlineFromMD_NoHeadings(t *testing.T) {
	sections := ParseOutlineFromMD("纯文本内容\n没有标题\n")
	assert.Len(t, sections, 0)
}
