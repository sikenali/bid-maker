package service

import (
	"regexp"
	"strings"

	"github.com/example/bid-maker-backend/internal/model"
	"github.com/google/uuid"
)

var mdHeadingRe = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

// ParseSectionsWithContent 从 Markdown 构建章节树，并把每级标题下、
// 属于其叶子章节的正文写入该叶子的 Content（含子标题间的正文）。
func ParseSectionsWithContent(md string) ([]model.Section, error) {
	lines := strings.Split(md, "\n")
	var root []model.Section
	type nodeLink struct {
		sec *model.Section
	}
	var stack []nodeLink

	appendToLeaf := func(text string) {
		if strings.TrimSpace(text) == "" {
			return
		}
		if len(stack) == 0 {
			return
		}
		leaf := stack[len(stack)-1].sec
		if leaf.Content != "" {
			leaf.Content += "\n"
		}
		leaf.Content += text
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		matches := mdHeadingRe.FindStringSubmatch(trimmed)
		if matches != nil {
			level := len(matches[1])
			title := strings.TrimSpace(matches[2])
			sec := &model.Section{
				ID:    uuid.NewString(),
				Title: title,
				Level: level,
			}
			// 弹出层级 >= 当前标题的栈
			for len(stack) > 0 && stack[len(stack)-1].sec.Level >= level {
				stack = stack[:len(stack)-1]
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1].sec
				parent.Children = append(parent.Children, *sec)
				// 重新取指针（Children append 可能 realloc）
				ptr := &parent.Children[len(parent.Children)-1]
				stack = append(stack, nodeLink{sec: ptr})
			} else {
				root = append(root, *sec)
				stack = append(stack, nodeLink{sec: &root[len(root)-1]})
			}
			continue
		}
		if trimmed == "" {
			continue
		}
		appendToLeaf(trimmed)
	}
	return root, nil
}
