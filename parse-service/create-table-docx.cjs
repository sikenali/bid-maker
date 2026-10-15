const PizZip = require('pizzip');
const Docxtemplater = require('docxtemplater');
const fs = require('fs');

const CT = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>';
const RL = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>';
const DOC_RELS = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>';
const STYLES = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:style w:type="paragraph" w:styleId="Heading1"><w:name w:val="heading 1"/></w:style><w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style></w:styles>';

const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
const h1 = (t) => `<w:p><w:pPr><w:pStyle w:val="Heading1"/></w:pPr><w:r><w:t>${esc(t)}</w:t></w:r></w:p>`;
const para = (t) => `<w:p><w:r><w:t>${esc(t)}</w:t></w:r></w:p>`;
const cell = (t) => `<w:tc><w:p><w:r><w:t>${esc(t)}</w:t></w:r></w:p></w:tc>`;
const table = (rows) =>
  '<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/></w:tblPr><w:tblGrid>' +
  rows[0].map(() => '<w:gridCol w:w="2000"/>').join('') + '</w:tblGrid>' +
  rows.map((r) => '<w:tr>' + r.map(cell).join('') + '</w:tr>').join('') + '</w:tbl>';

let body = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>';
body += h1('第七章 评分标准');
body += table([
  ['评分项', '分值', '评分标准'],
  ['技术方案', '40分', '优：方案完整得40分；良：得30分'],
  ['商务报价', '30分', '最低价得满分，每高1%扣1分'],
  ['企业资质', '20分', '具有相关资质得20分'],
]);
body += h1('第六章 技术标');
body += table([
  ['参数名称', '技术要求'],
  ['并发用户数', '不低于1000人'],
  ['系统响应时间', '不得大于2秒'],
]);
body += para('★投标文件份数正本一份副本两份，密封递交。');
body += '</w:body></w:document>';
body = body.replace('</w:body>', '<w:sectPr><w:pgSz w:w="11900" w:h="16840"/></w:sectPr></w:body>');

const zip = new PizZip();
zip.file('[Content_Types].xml', CT);
zip.file('_rels/.rels', RL);
zip.file('word/document.xml', body);
zip.file('word/styles.xml', STYLES);
zip.file('word/_rels/document.xml.rels', DOC_RELS);
const doc = new Docxtemplater(zip, { paragraphLoop: true, linebreaks: true, nullGetter: () => '' });
doc.render({});
fs.writeFileSync('/tmp/bid-table.docx', zip.generate({ type: 'nodebuffer', compression: 'DEFLATE' }));
console.log('created /tmp/bid-table.docx');
