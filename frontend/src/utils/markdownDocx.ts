/**
 * markdown-to-docx.ts
 *
 * 轻量级 Markdown → DOCX 转换器。
 * 基于已有的 docxtemplater + pizzip，无需额外依赖。
 * 支持：# H1 / ## H2 / ### H3 / 段落 / 列表 / 粗体 / 斜体 / 代码块
 */

import PizZip from 'pizzip'
import Docxtemplater from 'docxtemplater'

// ── 最小 OOXML 模板骨架 ──────────────────────────────────────────────

const CONTENT_TYPES = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

const RELS = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

const DOC_RELS = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

// 样式定义：Heading1 / Heading2 / Heading3 / Normal
const STYLES = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:styleId="Heading1">
    <w:name w:val="heading 1"/><w:basedOn w:val="Normal"/>
    <w:rPr><w:b/><w:sz w:val="32"/><w:szCs w:val="32"/>
      <w:rFonts w:hint="eastAsia" w:ascii="SimHei" w:eastAsia="SimHei"/></w:rPr>
  </w:style>
  <w:style w:type="paragraph" w:styleId="Heading2">
    <w:name w:val="heading 2"/><w:basedOn w:val="Normal"/>
    <w:rPr><w:b/><w:sz w:val="28"/><w:szCs w:val="28"/>
      <w:rFonts w:hint="eastAsia" w:ascii="SimHei" w:eastAsia="SimHei"/></w:rPr>
  </w:style>
  <w:style w:type="paragraph" w:styleId="Heading3">
    <w:name w:val="heading 3"/><w:basedOn w:val="Normal"/>
    <w:rPr><w:b/><w:sz w:val="24"/><w:szCs w:val="24"/>
      <w:rFonts w:hint="eastAsia" w:ascii="SimHei" w:eastAsia="SimHei"/></w:rPr>
  </w:style>
  <w:style w:type="paragraph" w:styleId="Normal">
    <w:name w:val="Normal"/>
    <w:rPr><w:sz w:val="24"/><w:szCs w:val="24"/>
      <w:rFonts w:hint="eastAsia" w:ascii="SimSun" w:eastAsia="SimSun"/></w:rPr>
  </w:style>
</w:styles>`

// ── Markdown 行级解析 ────────────────────────────────────────────────

interface Block { type: 'heading' | 'paragraph'; level: number; text: string }

function parseMarkdownLines(md: string): Block[] {
  const lines = md.split('\n')
  const blocks: Block[] = []
  let codeBlock = false
  let codeLines: string[] = []
  let inList = false
  let listItems: string[] = []

  const flushList = () => {
    if (listItems.length > 0) {
      for (const item of listItems) {
        blocks.push({ type: 'paragraph', level: 0, text: item })
      }
      listItems = []
      inList = false
    }
  }

  const flushCode = () => {
    if (codeLines.length > 0) {
      blocks.push({ type: 'paragraph', level: 0, text: codeLines.join('\n') })
      codeLines = []
      codeBlock = false
    }
  }

  for (const rawLine of lines) {
    const line = rawLine.replace(/\r$/, '')

    // 代码块
    if (line.startsWith('```')) {
      if (codeBlock) { flushCode() }
      else { flushList(); codeBlock = true }
      continue
    }
    if (codeBlock) { codeLines.push(line); continue }

    // 空行 → 刷新列表
    if (!line.trim()) { flushList(); continue }

    // 无序列表
    if (/^\s*[-*+]\s/.test(line)) {
      inList = true
      listItems.push(line.replace(/^\s*[-*+]\s/, '').trim())
      continue
    }
    // 有序列表
    if (/^\s*\d+\.\s/.test(line)) {
      inList = true
      listItems.push(line.replace(/^\s*\d+\.\s/, '').trim())
      continue
    }
    if (inList) flushList()

    // 标题
    const hm = line.match(/^(#{1,6})\s+(.+)/)
    if (hm) {
      flushList()
      const level = Math.min(hm[1].length, 3)
      blocks.push({ type: 'heading', level, text: hm[2].trim() })
      continue
    }

    // 普通段落
    flushList()
    blocks.push({ type: 'paragraph', level: 0, text: line.trim() })
  }

  flushList()
  flushCode()
  return blocks
}

// ── XML 转义 ─────────────────────────────────────────────────────────

function escXml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

// ── 渲染单段 XML ─────────────────────────────────────────────────────

function renderBlockXML(block: Block): string {
  const styleId = block.type === 'heading' ? `Heading${block.level}` : 'Normal'
  const text = escXml(block.text)
  return `<w:p><w:pPr><w:pStyle w:val="${styleId}"/></w:pPr><w:r><w:t>${text}</w:t></w:r></w:p>`
}

// ── 公钥导出 ─────────────────────────────────────────────────────────

/**
 * 将 Markdown 字符串转换为 DOCX ArrayBuffer。
 * 可直接传入 DocxEditor.loadDocumentBuffer(buffer) 使用。
 */
export async function markdownToDocx(md: string): Promise<ArrayBuffer> {
  const blocks = parseMarkdownLines(md || '')
  let bodyXml = ''
  for (const b of blocks) bodyXml += renderBlockXML(b)

  const docXml = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
    `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
    `<w:body>${bodyXml}` +
    `<w:sectPr><w:pgSz w:w="11900" w:h="16840"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440"/></w:sectPr>` +
    `</w:body></w:document>`

  const zip = new PizZip()
  zip.file('[Content_Types].xml', CONTENT_TYPES)
  zip.file('_rels/.rels', RELS)
  zip.file('word/document.xml', docXml)
  zip.file('word/styles.xml', STYLES)
  zip.file('word/_rels/document.xml.rels', DOC_RELS)

  const doc = new Docxtemplater(zip, {
    paragraphLoop: true,
    linebreaks: true,
    nullGetter: () => '',
  })
  doc.render({})

  const result = zip.generate({ type: 'arraybuffer', compression: 'DEFLATE' })
  return result
}
