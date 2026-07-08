package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/example/bid-maker-backend/internal/model"
)

type DocxService struct{}

func NewDocxService() *DocxService {
	return &DocxService{}
}

const nsW = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

type wDoc struct {
	XMLName xml.Name `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main document"`
	Body    wBody    `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main body"`
}

type wBody struct {
	Paragraphs []wPara `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main p"`
}

type wPara struct {
	Props *wParaProps `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pPr"`
	Runs  []wRun      `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main r"`
}

type wParaProps struct {
	Style *wStyle `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main pStyle"`
}

type wStyle struct {
	Val string `xml:"val,attr"`
}

type wRun struct {
	Texts []wText `xml:"http://schemas.openxmlformats.org/wordprocessingml/2006/main t"`
}

type wText struct {
	Text string `xml:",chardata"`
}

func extractText(runs []wRun) string {
	var b strings.Builder
	for _, r := range runs {
		for _, t := range r.Texts {
			b.WriteString(t.Text)
		}
	}
	return b.String()
}

func headingLevel(p wPara) int {
	if p.Props == nil || p.Props.Style == nil {
		return 0
	}
	style := p.Props.Style.Val
	var level int
	if n, _ := fmt.Sscanf(style, "Heading %d", &level); n == 1 && level >= 1 && level <= 9 {
		return level
	}
	return 0
}

func (s *DocxService) ParseDocument(data []byte) (*model.Document, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open docx: not a valid zip file: %w", err)
	}

	var docXML []byte
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to read word/document.xml: %w", err)
			}
			docXML, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read word/document.xml: %w", err)
			}
			break
		}
	}
	if docXML == nil {
		return nil, fmt.Errorf("invalid docx: word/document.xml not found")
	}

	var doc wDoc
	if err := xml.Unmarshal(docXML, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse document xml: %w", err)
	}

	id := fmt.Sprintf("doc-%d", time.Now().Unix())
	sections := s.extractSections(doc.Body.Paragraphs)

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

func (s *DocxService) extractSections(paras []wPara) []model.Section {
	var sections []model.Section
	var parentStack []int

	for _, p := range paras {
		text := strings.TrimSpace(extractText(p.Runs))
		if text == "" {
			continue
		}

		level := headingLevel(p)

		if level > 0 {
			section := model.Section{
				ID:    fmt.Sprintf("sec-%d", len(sections)+1),
				Title: text,
				Level: level,
			}

			for len(parentStack) > 0 {
				topIdx := parentStack[len(parentStack)-1]
				if sections[topIdx].Level >= level {
					parentStack = parentStack[:len(parentStack)-1]
				} else {
					break
				}
			}

			if len(parentStack) > 0 {
				parentIdx := parentStack[len(parentStack)-1]
				sections[parentIdx].Children = append(sections[parentIdx].Children, section)
			} else {
				sections = append(sections, section)
			}

			parentStack = append(parentStack, len(sections)-1)
		} else if len(sections) > 0 {
			last := findLeaf(&sections[len(sections)-1])
			last.Content += text + "\n"
		}
	}

	if len(sections) == 0 {
		var allText strings.Builder
		for _, p := range paras {
			text := strings.TrimSpace(extractText(p.Runs))
			if text == "" {
				continue
			}
			if allText.Len() > 0 {
				allText.WriteRune('\n')
			}
			allText.WriteString(text)
		}
		if allText.Len() > 0 {
			sections = append(sections, model.Section{
				ID:      "sec-1",
				Title:   "全文",
				Level:   1,
				Content: allText.String(),
			})
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

func (s *DocxService) GenerateDocument(doc *model.Document) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	docContent := s.buildDocumentXML(doc.Outline)
	contentTypes := s.buildContentTypes()
	rels := s.buildRels()

	for name, data := range map[string]string{
		"[Content_Types].xml":               contentTypes,
		"word/document.xml":                 docContent,
		"word/_rels/document.xml.rels":      rels,
	} {
		f, err := w.Create(name)
		if err != nil {
			return nil, fmt.Errorf("failed to create %s in zip: %w", name, err)
		}
		if _, err := f.Write([]byte(data)); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", name, err)
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
		s.writeSection(&b, &sec)
	}
	b.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/></w:sectPr>`)
	b.WriteString(`</w:body></w:document>`)
	return b.String()
}

func (s *DocxService) writeSection(b *strings.Builder, sec *model.Section) {
	writeParagraph(b, sec.Title, fmt.Sprintf("Heading %d", sec.Level))

	if sec.Content != "" {
		for _, line := range strings.Split(strings.TrimSpace(sec.Content), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			writeParagraph(b, line, "")
		}
	}

	for _, child := range sec.Children {
		s.writeSection(b, &child)
	}
}

func writeParagraph(b *strings.Builder, text, style string) {
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

func NowUTC() time.Time {
	return time.Now().UTC()
}
