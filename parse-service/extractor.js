'use strict';

/** 底层抽取层:对齐技术方案 officeparser.extract(path) 的语义.
 *
 * 使用 npm officeparser 解析为 AST,再拍平成:
 *   { paragraphs: [{text, style, heading_level, section, table_idx}],
 *     tables: [[[...]]], full_text: str }
 * section 为最近的上级标题,用于 clause_ref 溯源.
 */

const HEADING_STYLE_RE = /heading\s*(\d)/i;
const CN_CHAPTER_RE = /^第[一二三四五六七八九十\d]+[章节篇]/;
const CN_ITEM_RE = /^[一二三四五六七八九十]+[、．.]/;
const NUM_ITEM_RE = /^(\d+(\.\d+)*)([、．。\s]|\.(?!\d))/;

function guessLevel(text, style) {
  const m = HEADING_STYLE_RE.exec(style || '');
  if (m) return parseInt(m[1], 10);
  if (CN_CHAPTER_RE.test(text)) return 1;
  if (CN_ITEM_RE.test(text)) return text.length < 30 ? 2 : 0;
  const m2 = NUM_ITEM_RE.exec(text);
  if (m2 && text.length < 40) {
    return Math.min(m2[1].split('.').length, 3);
  }
  // 无分隔符的多段数字编号(如 5.1投标报价表):点分段数即层级;
  // 要求含小数点,避免"2024年…"这类年份行被误判为标题
  const m3 = /^(\d+(?:\.\d+)+)(?=[\u4e00-\u9fa5A-Za-z])/.exec(text);
  if (m3 && text.length < 40) {
    return Math.min(m3[1].split('.').length, 3);
  }
  return 0;
}

/** 递归收集节点文本(叶子 text 节点拼接). */
function nodeText(node) {
  if (!node) return '';
  if (typeof node.text === 'string' && (!node.children || node.children.length === 0)) {
    return node.text;
  }
  if (Array.isArray(node.children)) {
    return node.children.map(nodeText).join('');
  }
  return typeof node.text === 'string' ? node.text : '';
}

function extractFromAst(ast) {
  const paragraphs = [];
  const tables = [];
  let currentSection = '';

  const pushPara = (text, style, level, tableIdx) => {
    text = (text || '').trim();
    if (!text) return;
    // L1/L2 均更新溯源节,使 clause_ref 指向最近标题而非仅章标题
    if (level === 1 || level === 2) currentSection = text.slice(0, 60);
    paragraphs.push({ text, style: style || '', heading_level: level, section: currentSection, table_idx: tableIdx });
  };

  const walkCell = (cell) => {
    // cell -> paragraph -> text,全部叶子文本拼接
    const parts = [];
    for (const k of cell.children || []) {
      const t = nodeText(k).trim();
      if (t) parts.push(t);
    }
    return parts.join(' ');
  };

  const walk = (node, tableIdx) => {
    if (!node) return;
    switch (node.type) {
      case 'heading': {
        const level = (node.metadata && node.metadata.level) || guessLevel(nodeText(node), (node.metadata && node.metadata.style) || '');
        pushPara(nodeText(node), (node.metadata && node.metadata.style) || 'Heading', level, -1);
        break;
      }
      case 'paragraph': {
        const style = (node.metadata && node.metadata.style) || '';
        pushPara(nodeText(node), style, guessLevel(nodeText(node), style), tableIdx);
        break;
      }
      case 'table': {
        const idx = tables.length;
        const grid = [];
        for (const row of node.rows || node.children || []) {
          const cells = (row.cells || row.children || []).map(walkCell);
          if (cells.some((c) => c)) {
            grid.push(cells);
            pushPara(cells.filter(Boolean).join(' | '), 'Table', 0, idx);
          }
        }
        if (grid.length) tables.push(grid);
        break;
      }
      case 'sheet':
      case 'slide':
      case 'page': {
        const label = (node.title || nodeText(node) || `${node.type}`).slice(0, 60);
        if (label) currentSection = label;
        for (const k of node.children || []) walk(k, tableIdx);
        break;
      }
      case 'text': {
        const t = (node.text || '').trim();
        if (t) pushPara(t, '', 0, tableIdx);
        break;
      }
      default: {
        if (Array.isArray(node.children)) {
          for (const k of node.children) walk(k, tableIdx);
        } else if (typeof node.text === 'string' && node.text.trim()) {
          pushPara(node.text, '', 0, tableIdx);
        }
      }
    }
  };

  for (const n of ast.content || []) walk(n, -1);
  return { paragraphs, tables, full_text: paragraphs.map((p) => p.text).join('\n') };
}

const EXT_BY_NAME = { '.docx': 'docx', '.xlsx': 'xlsx', '.xlsm': 'xlsx', '.pptx': 'pptx', '.pdf': 'pdf' };

async function extract(input, filename) {
  // input: 文件路径或 Buffer;filename 用于 Buffer 时的类型推断
  const officeParser = require('officeparser');
  const name = (typeof input === 'string' ? input : filename || '').toLowerCase();
  const ext = Object.keys(EXT_BY_NAME).find((e) => name.endsWith(e));
  if (!name || !ext) {
    throw new Error(`不支持的文件格式:仅支持 ${Object.keys(EXT_BY_NAME).join('/')}`);
  }
  const opts = { fileType: EXT_BY_NAME[ext] };
  const ast = await officeParser.parseOffice(input, opts);
  return extractFromAst(ast);
}

module.exports = { extract, extractFromAst, guessLevel };
