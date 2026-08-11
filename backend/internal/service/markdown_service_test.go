package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSectionsWithContent_HierarchyAndContent(t *testing.T) {
	md := `# 第一章 项目概况

本项目概况导语。

## 1.1 项目背景

背景内容第一段。

背景内容第二段。

## 1.2 建设目标

目标内容。

# 第二章 技术方案

方案内容。`
	sections, err := ParseSectionsWithContent(md)
	assert.NoError(t, err)
	assert.Len(t, sections, 2)
	assert.Equal(t, "第一章 项目概况", sections[0].Title)
	assert.Equal(t, 1, sections[0].Level)
	assert.Len(t, sections[0].Children, 2)
	// 正文归属到无子章节的叶子
	leaf := sections[0].Children[1]
	assert.Equal(t, "1.2 建设目标", leaf.Title)
	assert.Contains(t, leaf.Content, "目标内容。")
	// 有子章节的父章节不应吞掉子标题后的正文（该正文并入其叶子）
	assert.Contains(t, sections[0].Content, "本项目概况导语。")
}

func TestParseSectionsWithContent_Empty(t *testing.T) {
	sections, err := ParseSectionsWithContent("")
	assert.NoError(t, err)
	assert.Len(t, sections, 0)
}
