const PizZip = require('pizzip');
const Docxtemplater = require('docxtemplater');
const fs = require('fs');

const CT = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>';
const RL = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>';
const DOC_RELS = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>';
const STYLES = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:style w:type="paragraph" w:styleId="Heading1"><w:name w:val="heading 1"/></w:style><w:style w:type="paragraph" w:styleId="Heading2"><w:name w:val="heading 2"/></w:style><w:style w:type="paragraph" w:styleId="Normal"><w:name w:val="Normal"/></w:style></w:styles>';

const lines = [
  { level: 1, text: '智慧城市综合管理平台项目招标文件' },
  { level: 0, text: '招标编号：ZBCG-2024-0892' },
  { level: 0, text: '招标人：市政务服务数据管理局' },
  { level: 0, text: '投标截止时间：2024年12月20日 14:00' },
  { level: 0, text: '本项目要求投标人具备电子与智能化工程专业承包二级及以上资质。' },
  { level: 0, text: '★投标报价不得高于预算金额，否则作废标处理。' },
  { level: 1, text: '第一章 投标函' },
  { level: 2, text: '1.1 投标承诺' },
  { level: 0, text: '我方完全接受招标文件的全部条款和条件。' },
  { level: 1, text: '第二章 法定代表人身份证明' },
  { level: 1, text: '第三章 授权委托书' },
  { level: 1, text: '第四章 投标保证金' },
  { level: 1, text: '第五章 商务标' },
  { level: 2, text: '5.1 投标报价表' },
  { level: 0, text: '本项目总报价不低于人民币500万元，单项设备报价须逐项列出。' },
  { level: 2, text: '5.2 企业资质证书' },
  { level: 2, text: '5.3 业绩清单' },
  { level: 1, text: '第六章 技术标' },
  { level: 2, text: '6.1 技术方案' },
  { level: 0, text: '技术方案应包含系统架构设计、关键技术选型、数据接口规范。' },
  { level: 0, text: '▲系统响应时间不得大于2秒，并发用户数不低于1000人。' },
  { level: 2, text: '6.2 质量保障措施' },
  { level: 2, text: '6.3 售后服务方案' },
  { level: 0, text: '提供7x24小时技术支持，保修期不少于3年。' },
  { level: 1, text: '第七章 评分标准' },
  { level: 0, text: '技术方案：40分；商务报价：30分；企业资质：20分；售后服务：10分。' },
  { level: 0, text: '★未按实质性要求提供资质证明的，否决其投标。' },
];

let bodyXml = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>';
for (const line of lines) {
  const styleId = line.level === 1 ? 'Heading1' : line.level === 2 ? 'Heading2' : 'Normal';
  const text = line.text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  bodyXml += '<w:p><w:pPr><w:pStyle w:val="' + styleId + '"/></w:pPr><w:r><w:t>' + text + '</w:t></w:r></w:p>';
}
bodyXml += '<w:sectPr><w:pgSz w:w="11900" w:h="16840"/></w:sectPr></w:body></w:document>';

const zip = new PizZip();
zip.file('[Content_Types].xml', CT);
zip.file('_rels/.rels', RL);
zip.file('word/document.xml', bodyXml);
zip.file('word/styles.xml', STYLES);
zip.file('word/_rels/document.xml.rels', DOC_RELS);
const doc = new Docxtemplater(zip, { paragraphLoop: true, linebreaks: true, nullGetter: () => '' });
doc.render({});
const buf = zip.generate({ type: 'nodebuffer', compression: 'DEFLATE' });
fs.writeFileSync('/home/jingle/opc/bid-maker/bid-doc.docx', buf);
console.log('Created bid-doc.docx, size:', buf.length, 'bytes');
