package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/example/bid-maker-backend/internal/model"
	"github.com/unidoc/unioffice/v2/document"
)

var (
	reChapter    = regexp.MustCompile(`^第[一二三四五六七八九十]章`)
	reSection    = regexp.MustCompile(`^第[一二三四五六七八九十]节`)
	reNumbered   = regexp.MustCompile(`^\d+[、.]\d+`)
	reChineseNum = regexp.MustCompile(`^[一二三四五六七八九十]+[、]`)

)

const maxHeadingLen = 80

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
	text := paragraphText(para)
	text = strings.TrimSpace(text)
	if text == "" {
		return false, 0
	}
	if utf8.RuneCountInString(text) > maxHeadingLen {
		return false, 0
	}

	props := para.X().PPr
	if props != nil && props.PStyle != nil {
		styleVal := props.PStyle.ValAttr
		// Standard patterns (Heading1, 标题1, etc.)
		for i := 1; i <= 9; i++ {
			patterns := []string{
				fmt.Sprintf("Heading%d", i), fmt.Sprintf("Heading %d", i),
				fmt.Sprintf("heading%d", i), fmt.Sprintf("heading %d", i),
				fmt.Sprintf("标题%d", i), fmt.Sprintf("标题 %d", i),
			}
			for _, p := range patterns {
				if styleVal == p {
					return true, i
				}
			}
		}
		// 第X章 always level 1
		for i := 1; i <= 9; i++ {
			if styleVal == fmt.Sprintf("第%s章", chineseNum(i)) {
				return true, 1
			}
		}
		// 第X节 always level 2
		for i := 1; i <= 9; i++ {
			if styleVal == fmt.Sprintf("第%s节", chineseNum(i)) {
				return true, 2
			}
		}
		// Chinese level patterns: 一级标题, 二级标题...
		for level := 1; level <= 9; level++ {
			levelStr := chineseNum(level)
			if styleVal == levelStr+"级标题" {
				return true, level
			}
			if styleVal == "第"+levelStr+"章" {
				return true, 1
			}
			if styleVal == "第"+levelStr+"节" {
				return true, 2
			}
		}
		// Check outlineLvl attribute
		if props.OutlineLvl != nil {
			val := props.OutlineLvl.ValAttr
			if val >= 0 && val <= 8 {
				return true, int(val + 1)
			}
		}
	}

	// Fallback: detect numbered headings by text content via regex
	if reChapter.MatchString(text) {
		return true, 1
	}
	if reSection.MatchString(text) {
		return true, 2
	}
	if reNumbered.MatchString(text) {
		endIdx := strings.IndexFunc(text, func(r rune) bool {
			return r != '.' && r != '、' && r != ' ' && (r < '0' || r > '9')
		})
		if endIdx == -1 {
			endIdx = len(text)
		}
		prefix := text[:endIdx]
		delimCount := strings.Count(prefix, ".") + strings.Count(prefix, "、")
		return true, delimCount + 1
	}
	if reChineseNum.MatchString(text) {
		return true, 1
	}

	return false, 0
}

func chineseNum(n int) string {
	nums := []string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九"}
	if n >= 1 && n <= 9 {
		return nums[n]
	}
	return ""
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

func (s *DocxService) Reparse(doc *model.Document, keyword string) {
	if len(doc.RawBuffer) == 0 {
		return
	}
	kwd := s.Keyword
	if keyword != "" {
		kwd = keyword
	}
	// temporarily swap keyword for parsing
	orig := s.Keyword
	s.Keyword = kwd
	parsed, err := s.ParseDocument(doc.RawBuffer)
	s.Keyword = orig
	if err != nil {
		return
	}
	doc.Outline = parsed.Outline
	doc.Title = parsed.Title
	doc.UpdatedAt = time.Now().UTC()
}

func NowUTC() time.Time {
	return time.Now().UTC()
}
