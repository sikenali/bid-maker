package service

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/example/bid-maker-backend/internal/model"
	"github.com/unidoc/unioffice/document"
)

type DocxService struct{}

func NewDocxService() *DocxService {
	return &DocxService{}
}

func (s *DocxService) ParseDocument(data []byte) (*model.Document, error) {
	doc, err := document.Read(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open docx: %w", err)
	}
	defer doc.Close()

	id := fmt.Sprintf("doc-%d", time.Now().Unix())
	sections := s.extractSections(doc)

	title := "Untitled Document"
	if len(sections) > 0 {
		title = sections[0].Title
	}

	return &model.Document{
		ID:        id,
		Title:     title,
		Outline:   sections,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func (s *DocxService) extractSections(doc *document.Document) []model.Section {
	var sections []model.Section
	var parentStack []*model.Section

	for _, para := range doc.Paragraphs() {
		text := ""
		for _, run := range para.Runs() {
			text += run.Text()
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		level := s.detectHeadingLevel(para)

		if level > 0 {
			section := model.Section{
				ID:    fmt.Sprintf("sec-%d", len(sections)+1),
				Title: text,
				Level: level,
			}

			for len(parentStack) > 0 && parentStack[len(parentStack)-1].Level >= level {
				parentStack = parentStack[:len(parentStack)-1]
			}

			var parent *model.Section
			if len(parentStack) > 0 {
				parent = parentStack[len(parentStack)-1]
			}

			if parent != nil {
				parent.Children = append(parent.Children, section)
			} else {
				sections = append(sections, section)
			}

			parentStack = append(parentStack, &sections[len(sections)-1])
		} else if len(sections) > 0 {
			lastSection := s.findLeafSection(&sections[len(sections)-1])
			lastSection.Content += text + "\n"
		}
	}

	return sections
}

func (s *DocxService) findLeafSection(sec *model.Section) *model.Section {
	if len(sec.Children) == 0 || sec.Content != "" {
		return sec
	}
	return s.findLeafSection(&sec.Children[len(sec.Children)-1])
}

func (s *DocxService) detectHeadingLevel(para document.Paragraph) int {
	props := para.Properties()
	style := props.RStyle()

	if style == "" {
		return 0
	}

	switch {
	case strings.HasPrefix(style, "Heading 1"):
		return 1
	case strings.HasPrefix(style, "Heading 2"):
		return 2
	case strings.HasPrefix(style, "Heading 3"):
		return 3
	case strings.HasPrefix(style, "Heading 4"):
		return 4
	case strings.HasPrefix(style, "Heading 5"):
		return 5
	case strings.HasPrefix(style, "Heading 6"):
		return 6
	case strings.HasPrefix(style, "Heading 7"):
		return 7
	case strings.HasPrefix(style, "Heading 8"):
		return 8
	case strings.HasPrefix(style, "Heading 9"):
		return 9
	}

	return 0
}

func (s *DocxService) GenerateDocument(doc *model.Document) ([]byte, error) {
	out := document.New()

	for _, sec := range doc.Outline {
		s.addSectionToDoc(out, &sec)
	}

	buf := new(bytes.Buffer)
	if err := out.Save(buf); err != nil {
		out.Close()
		return nil, fmt.Errorf("failed to write docx: %w", err)
	}
	out.Close()
	return buf.Bytes(), nil
}

func (s *DocxService) addSectionToDoc(out *document.Document, sec *model.Section) {
	p := out.AddParagraph()
	p.Properties().SetStyle(fmt.Sprintf("Heading %d", sec.Level))
	run := p.AddRun()
	run.AddText(sec.Title)

	if sec.Content != "" {
		contentLines := strings.Split(strings.TrimSpace(sec.Content), "\n")
		for _, line := range contentLines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			p2 := out.AddParagraph()
			p2.AddRun().AddText(line)
		}
	}

	for _, child := range sec.Children {
		s.addSectionToDoc(out, &child)
	}
}

func NowUTC() time.Time {
	return time.Now().UTC()
}
