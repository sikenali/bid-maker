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

func TestParseSectionsWithContent_LevelSkip(t *testing.T) {
	md := `# 第一章
### 1.1 跳过二级
跳跃正文内容。`
	sections, err := ParseSectionsWithContent(md)
	assert.NoError(t, err)
	assert.Len(t, sections, 1)
	assert.Equal(t, "第一章", sections[0].Title)
	assert.Len(t, sections[0].Children, 1)
	child := sections[0].Children[0]
	assert.Equal(t, "1.1 跳过二级", child.Title)
	assert.Equal(t, 3, child.Level)
	assert.Contains(t, child.Content, "跳跃正文内容。")
}

func TestParseSectionsWithContent_SiblingIsolation(t *testing.T) {
	md := `# 第一章
第一章专属内容。
## 1.1 小节
第一小节内容。
# 第二章
第二章专属内容。
## 2.1 小节
第二小节内容。`
	sections, err := ParseSectionsWithContent(md)
	assert.NoError(t, err)
	assert.Len(t, sections, 2)
	assert.Contains(t, sections[0].Content, "第一章专属内容。")
	assert.NotContains(t, sections[0].Content, "第二章专属内容。")
	assert.Contains(t, sections[0].Children[0].Content, "第一小节内容。")
	assert.NotContains(t, sections[0].Children[0].Content, "第二小节内容。")
	assert.Contains(t, sections[1].Content, "第二章专属内容。")
	assert.NotContains(t, sections[1].Content, "第一章专属内容。")
	assert.Contains(t, sections[1].Children[0].Content, "第二小节内容。")
	assert.NotContains(t, sections[1].Children[0].Content, "第一小节内容。")
}
