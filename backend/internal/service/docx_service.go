package service

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/example/bid-maker-backend/internal/model"
	"github.com/unidoc/unioffice/v2/document"
)

type DocxService struct {
	Keyword string
}

func NewDocxService() *DocxService {
	return &DocxService{Keyword: "投标文件"}
}

func paragraphText(para document.Paragraph) string {
	var sb strings.Builder
	for _, r := range para.Runs() {
		sb.WriteString(r.Text())
	}
	return sb.String()
}

func isHeading(para document.Paragraph) (bool, int) {
	props := para.X().PPr
	if props != nil && props.PStyle != nil {
		styleVal := props.PStyle.ValAttr
		for i := 1; i <= 9; i++ {
			patterns := []string{
				fmt.Sprintf("Heading%d", i),
				fmt.Sprintf("Heading %d", i),
				fmt.Sprintf("heading%d", i),
				fmt.Sprintf("heading %d", i),
				fmt.Sprintf("标题%d", i),
				fmt.Sprintf("标题 %d", i),
			}
			for _, p := range patterns {
				if styleVal == p {
					return true, i
				}
			}
		}
	}
	return false, 0
}

func (s *DocxService) ParseDocument(data []byte) (*model.Document, error) {
	tmpFile, err := os.CreateTemp("", "bid-*.docx")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	doc, err := document.Open(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open docx with unioffice: %w", err)
	}
	defer doc.Close()

	paras := doc.Paragraphs()
	sections := s.extractSections(paras)
	filtered := s.filterKeywordOutline(sections, s.Keyword)

	id := fmt.Sprintf("doc-%d", time.Now().Unix())
	title := "Untitled Document"
	if len(filtered) > 0 {
		title = filtered[0].Title
	}

	return &model.Document{
		ID:        id,
		Title:     title,
		Outline:   filtered,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func (s *DocxService) extractSections(paras []document.Paragraph) []model.Section {
	var sections []model.Section
	var parentStack []*model.Section

	for _, p := range paras {
		text := strings.TrimSpace(paragraphText(p))
		if text == "" {
			continue
		}

		isH, level := isHeading(p)
		if isH {
			section := model.Section{
				ID:    fmt.Sprintf("sec-%d", len(sections)+1),
				Title: text,
				Level: level,
			}

			for len(parentStack) > 0 && parentStack[len(parentStack)-1].Level >= level {
				parentStack = parentStack[:len(parentStack)-1]
			}

			if len(parentStack) > 0 {
				parent := parentStack[len(parentStack)-1]
				parent.Children = append(parent.Children, section)
				parentStack = append(parentStack, &parent.Children[len(parent.Children)-1])
			} else {
				sections = append(sections, section)
				parentStack = append(parentStack, &sections[len(sections)-1])
			}
		} else if len(sections) > 0 {
			last := findLeaf(&sections[len(sections)-1])
			if last.Content != "" {
				last.Content += "\n"
			}
			last.Content += text
		}
	}

	return sections
}

func findLeaf(sec *model.Section) *model.Section {
	if len(sec.Children) == 0 || sec.Content != "" {
		return sec
	}
	return findLeaf(&sec.Children[len(sec.Children)-1])
}

func (s *DocxService) filterKeywordOutline(sections []model.Section, keyword string) []model.Section {
	for _, sec := range sections {
		if strings.Contains(sec.Title, keyword) {
			return sec.Children
		}
		if found := s.filterKeywordOutline(sec.Children, keyword); found != nil {
			return found
		}
	}
	return nil
}

func (s *DocxService) GenerateDocument(doc *model.Document) ([]byte, error) {
	return s.generateDocumentXML(doc)
}

func (s *DocxService) generateDocumentXML(doc *model.Document) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := NewZipWriter(buf)

	docContent := s.buildDocumentXML(doc.Outline)
	contentTypes := s.buildContentTypes()
	rels := s.buildRels()

	for name, data := range map[string]string{
		"[Content_Types].xml":               contentTypes,
		"word/document.xml":                 docContent,
		"word/_rels/document.xml.rels":      rels,
	} {
		if err := w.AddFile(name, []byte(data)); err != nil {
			return nil, fmt.Errorf("failed to create %s in zip: %w", name, err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize docx: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *DocxService) buildDocumentXML(sections []model.Section) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	b.WriteString(`<w:body>`)
	for _, sec := range sections {
		s.writeSectionXML(&b, &sec)
	}
	b.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/></w:sectPr>`)
	b.WriteString(`</w:body></w:document>`)
	return b.String()
}

func (s *DocxService) writeSectionXML(b *strings.Builder, sec *model.Section) {
	writeParagraphXML(b, sec.Title, fmt.Sprintf("Heading %d", sec.Level))

	if sec.Content != "" {
		for _, line := range strings.Split(strings.TrimSpace(sec.Content), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			writeParagraphXML(b, line, "")
		}
	}

	for _, child := range sec.Children {
		s.writeSectionXML(b, &child)
	}
}

func writeParagraphXML(b *strings.Builder, text, style string) {
	b.WriteString(`<w:p>`)
	if style != "" {
		b.WriteString(`<w:pPr><w:pStyle w:val="` + style + `"/></w:pPr>`)
	}
	b.WriteString(`<w:r><w:rPr><w:sz w:val="24"/></w:rPr><w:t xml:space="preserve">` + xmlEscape(text) + `</w:t></w:r>`)
	b.WriteString(`</w:p>`)
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func (s *DocxService) buildContentTypes() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
		`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
		`<Default Extension="xml" ContentType="application/xml"/>` +
		`<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>` +
		`</Types>`
}

func (s *DocxService) buildRels() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="document.xml"/>` +
		`</Relationships>`
}

func (s *DocxService) GenerateMarkdown(doc *model.Document) []byte {
	var b strings.Builder
	b.WriteString("# " + doc.Title + "\n\n")
	for _, sec := range doc.Outline {
		s.writeSectionMarkdown(&b, &sec, sec.Level)
	}
	return []byte(b.String())
}

func (s *DocxService) writeSectionMarkdown(b *strings.Builder, sec *model.Section, level int) {
	prefix := strings.Repeat("#", level)
	b.WriteString(prefix + " " + sec.Title + "\n\n")
	if sec.Content != "" {
		b.WriteString(sec.Content + "\n\n")
	}
	for i := range sec.Children {
		s.writeSectionMarkdown(b, &sec.Children[i], level+1)
	}
}

func NowUTC() time.Time {
	return time.Now().UTC()
}