'use strict';

/** 格式层:md 承载内容,Word 模板承载格式(技术方案 §9).
 *
 * 用 docxtemplater + pizzip 渲染:模板以代码内嵌的最小 OOXML 构建,
 * 内容经 {title}/{bid_no}/{export_time} + {#sections} 循环注入,
 * 一级标题 -> Heading1,二级 -> Heading2,其余正文 Normal,
 * content 内 \n 经 linebreaks 转为真实换行.
 */

const PizZip = require('pizzip');
const Docxtemplater = require('docxtemplater');

const CONTENT_TYPES = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/></Types>`;

const RELS = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`;

const STYLES = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:style w:type="paragraph" w:styleId="Heading1"><w:name w:val="heading 1"/><w:pPr><w:keepNext/><w:keepLines/><w:spacing w:after="120"/><w:jc w:val="center"/></w:pPr><w:rPr><w:b/><w:sz w:val="32"/><w:szCs w:val="32"/><w:rFonts w:hint="eastAsia" w:ascii="SimHei" w:eastAsia="SimHei"/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="Heading2"><w:name w:val="heading 2"/><w:pPr><w:keepNext/><w:keepLines/><w:spacing w:before="120" w:after="60"/></w:pPr><w:rPr><w:b/><w:sz w:val="28"/><w:szCs w:val="28"/><w:rFonts w:hint="eastAsia" w:ascii="SimHei" w:eastAsia="SimHei"/></w:rPr></w:style><w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/><w:pPr><w:spacing w:after="100" w:line="360" w:lineRule="auto"/><w:ind w:firstLine="480"/></w:pPr><w:rPr><w:sz w:val="24"/><w:szCs w:val="24"/><w:rFonts w:hint="eastAsia" w:ascii="SimSun" w:eastAsia="SimSun"/></w:rPr></w:style></w:styles>`;

const DOC_RELS = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>`;

const para = (inner) => `<w:p>${inner}</w:p>`;
const textRun = (text) => `<w:r><w:t xml:space="preserve">${text}</w:t></w:r>`;
const titlePara = (style, tag) =>
  para(`<w:pPr><w:pStyle w:val="${style}"/></w:pPr>${textRun(`{${tag}}`)}`);
const loopPara = (tag) => para(textRun(`{${tag}}`));

// 模板正文:封面标题 + 元信息 + 章节循环(h1/h2 条件段落 + 正文)
const DOCUMENT_XML =
  `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
  `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>` +
  titlePara('Heading1', 'title') +
  para(`<w:pPr><w:jc w:val="center"/></w:pPr>` + textRun('招标编号:{bid_no}　导出时间:{export_time}')) +
  loopPara('#sections') +
  para(`<w:pPr><w:pStyle w:val="Heading1"/></w:pPr>` + textRun('{#h1}{title}{/h1}')) +
  para(`<w:pPr><w:pStyle w:val="Heading2"/></w:pPr>` + textRun('{#h2}{title}{/h2}')) +
  para(textRun('{content}')) +
  loopPara('/sections') +
  `<w:sectPr><w:pgSz w:w="11900" w:h="16840"/><w:pgMar w:top="1440" w:right="1800" w:bottom="1440" w:left="1800"/></w:sectPr>` +
  `</w:body></w:document>`;

function buildTemplateZip() {
  const zip = new PizZip();
  zip.file('[Content_Types].xml', CONTENT_TYPES);
  zip.file('_rels/.rels', RELS);
  zip.file('word/document.xml', DOCUMENT_XML);
  zip.file('word/styles.xml', STYLES);
  zip.file('word/_rels/document.xml.rels', DOC_RELS);
  return zip;
}

/**
 * @param data {title, bid_no, export_time?, sections: [{title, content, level?}]}
 * @returns Buffer docx 二进制
 */
function renderDocx(data) {
  const sections = (data.sections || []).map((s) => ({
    title: String(s.title || ''),
    content: String(s.content || ''),
    h1: (s.level || 1) === 1,
    h2: (s.level || 1) === 2,
  }));
  const doc = new Docxtemplater(buildTemplateZip(), {
    paragraphLoop: true,
    linebreaks: true,
    nullGetter: () => '',
  });
  doc.render({
    title: String(data.title || '投标文件'),
    bid_no: String(data.bid_no || ''),
    export_time: data.export_time || new Date().toISOString().slice(0, 10),
    sections,
  });
  // 清掉条件未命中留下的空段落(仅含空 w:t 的 w:p,含自闭合 <w:t/>)
  const zip = doc.getZip();
  const name = 'word/document.xml';
  const xml = zip.file(name).asText()
    .replace(/<w:p><w:pPr>((?!<w:p[ >]).)*?<\/w:pPr>\s*<w:r>\s*<w:t(\s[^>]*)?(\/\s*>|>\s*<\/w:t>)\s*<\/w:r>\s*<\/w:p>/g, '');
  zip.file(name, xml);
  return zip.generate({ type: 'nodebuffer', compression: 'DEFLATE' });
}

module.exports = { renderDocx };
