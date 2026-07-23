package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
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

// ---------- unioffice helpers ----------

func paragraphText(para document.Paragraph) string {
	var sb strings.Builder
	for _, r := range para.Runs() {
		sb.WriteString(r.Text())
	}
	return sb.String()
}

func isHeadingUnioffice(para document.Paragraph) (bool, int) {
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

func (s *DocxService) extractSectionsWithKeywordUnioffice(paras []document.Paragraph, keyword string) []model.Section {
	var sections []model.Section
	var parentStack []*model.Section
	keywordFound := false

	for _, p := range paras {
		text := strings.TrimSpace(paragraphText(p))
		if text == "" {
			continue
		}

		isH, level := isHeadingUnioffice(p)

		if !keywordFound {
			if isH && strings.Contains(text, keyword) {
				keywordFound = true
				continue
			}
			if !isH && strings.Contains(text, keyword) {
				keywordFound = true
				continue
			}
			continue
		}

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

	if len(sections) == 0 {
		return nil
	}
	return sections
}

func (s *DocxService) extractSectionsUnioffice(paras []document.Paragraph) []model.Section {
	var sections []model.Section
	var parentStack []*model.Section

	for _, p := range paras {
		text := strings.TrimSpace(paragraphText(p))
		if text == "" {
			continue
		}

		isH, level := isHeadingUnioffice(p)
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

// ---------- manual XML fallback types ----------

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

func extractTextXml(runs []wRun) string {
	var b strings.Builder
	for _, r := range runs {
		for _, t := range r.Texts {
			b.WriteString(t.Text)
		}
	}
	return b.String()
}

func headingLevelXml(p wPara) int {
	if p.Props == nil || p.Props.Style == nil {
		return 0
	}
	style := p.Props.Style.Val
	var level int
	if n, _ := fmt.Sscanf(style, "Heading %d", &level); n == 1 && level >= 1 && level <= 9 {
		return level
	}
	if n, _ := fmt.Sscanf(style, "Heading%d", &level); n == 1 && level >= 1 && level <= 9 {
		return level
	}
	return 0
}

func (s *DocxService) extractSectionsXml(paras []wPara, keyword string) []model.Section {
	var sections []model.Section
	var parentStack []*model.Section
	keywordFound := false

	for _, p := range paras {
		text := strings.TrimSpace(extractTextXml(p.Runs))
		if text == "" {
			continue
		}

		level := headingLevelXml(p)

		if !keywordFound {
			if level > 0 && strings.Contains(text, keyword) {
				keywordFound = true
				continue
			}
			if level == 0 && strings.Contains(text, keyword) {
				keywordFound = true
				continue
			}
			continue
		}

		if level > 0 {
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
			last.Content += text + "\n"
		}
	}

	if len(sections) == 0 {
		// Fallback: treat entire doc as one section with content
		var allText strings.Builder
		for _, p := range paras {
			text := strings.TrimSpace(extractTextXml(p.Runs))
			if text == "" {
				continue
			}
			if allText.Len() > 0 {
				allText.WriteRune('\n')
			}
			allText.WriteString(text)
		}
		if allText.Len() > 0 {
			return []model.Section{{
				ID:      "sec-1",
				Title:   "全文",
				Level:   1,
				Content: allText.String(),
			}}
		}
	}

	return sections
}

func (s *DocxService) filterKeywordOutline(sections []model.Section, keyword string) []model.Section {
	for i, sec := range sections {
		if strings.Contains(sec.Title, keyword) {
			return sec.Children
		}
		if strings.Contains(sec.Content, keyword) {
			return sections[i:]
		}
		if found := s.filterKeywordOutline(sec.Children, keyword); found != nil {
			return found
		}
	}
	return nil
}

func findLeaf(sec *model.Section) *model.Section {
	if len(sec.Children) == 0 || sec.Content != "" {
		return sec
	}
	return findLeaf(&sec.Children[len(sec.Children)-1])
}

// ---------- main entry point ----------

func (s *DocxService) ParseDocument(data []byte) (*model.Document, error) {
	// Try unioffice first
	doc, err := s.parseWithUnioffice(data)
	if err != nil {
		return nil, err
	}

	// If unioffice returned nothing, fall back to manual XML parsing
	if len(doc.Outline) == 0 {
		doc, err = s.parseWithXmlFallback(data)
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}

func (s *DocxService) parseWithUnioffice(data []byte) (*model.Document, error) {
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
	filtered := s.extractSectionsWithKeywordUnioffice(paras, s.Keyword)
	if filtered == nil {
		sections := s.extractSectionsUnioffice(paras)
		filtered = sections
	}

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

func (s *DocxService) parseWithXmlFallback(data []byte) (*model.Document, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open docx as zip: %w", err)
	}

	var docXML []byte
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			docXML, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
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

	sections := s.extractSectionsXml(doc.Body.Paragraphs, s.Keyword)
	if sections == nil {
		sections = []model.Section{}
	}

	id := fmt.Sprintf("doc-%d", time.Now().Unix())
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

// ---------- Document generation (unchanged) ----------

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
		"[Content_Types].xml":          contentTypes,
		"word/document.xml":            docContent,
		"word/_rels/document.xml.rels": rels,
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
