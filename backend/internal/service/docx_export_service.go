package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var exportHeadingRe = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

const (
	contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
 <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
 <Default Extension="xml" ContentType="application/xml"/>
 <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
 <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`

	relsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
 <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	stylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
 <w:style w:type="paragraph" w:styleId="Heading1"><w:name w:val="heading 1"/><w:pPr><w:keepNext/><w:keepLines/></w:pPr></w:style>
 <w:style w:type="paragraph" w:styleId="Heading2"><w:name w:val="heading 2"/><w:pPr><w:keepNext/><w:keepLines/></w:pPr></w:style>
 <w:style w:type="paragraph" w:styleId="Heading3"><w:name w:val="heading 3"/><w:pPr><w:keepNext/><w:keepLines/></w:pPr></w:style>
 <w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
</w:styles>`

	documentRelXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
 <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`
)

func escapeXML(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&apos;")
	return r.Replace(s)
}

func paraXML(text string, style string) string {
	if style != "" {
		return fmt.Sprintf(`<w:p><w:pPr><w:pStyle w:val="%s"/></w:pPr><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, style, escapeXML(text))
	}
	return fmt.Sprintf(`<w:p><w:r><w:t xml:space="preserve">%s</w:t></w:r></w:p>`, escapeXML(text))
}

func BuildDocxFromMarkdown(md string) ([]byte, error) {
	var body strings.Builder
	body.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	body.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)

	for _, raw := range strings.Split(md, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		m := exportHeadingRe.FindStringSubmatch(line)
		if m != nil {
			level := len(m[1])
			style := ""
			switch {
			case level == 1:
				style = "Heading1"
			case level == 2:
				style = "Heading2"
			case level == 3:
				style = "Heading3"
			}
			body.WriteString(paraXML(strings.TrimSpace(m[2]), style))
			continue
		}
		body.WriteString(paraXML(line, "Normal"))
	}
	body.WriteString("</w:body></w:document>")

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	add := func(name, content string) error {
		f, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(content))
		return err
	}
	for _, item := range []struct{ name, content string }{
		{"[Content_Types].xml", contentTypesXML},
		{"_rels/.rels", relsXML},
		{"word/document.xml", body.String()},
		{"word/styles.xml", stylesXML},
		{"word/_rels/document.xml.rels", documentRelXML},
	} {
		if err := add(item.name, item.content); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
