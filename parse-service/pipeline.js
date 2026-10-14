'use strict';

/** 提取管线:技术方案 §3.2
 *
 *   const elements = await extract(path);   // officeparser 语义
 *   return {
 *     template: extractTemplate(elements),
 *     technical_indicators: extractIndicators(elements),
 *     commercial_reqs: extractCommercial(elements),
 *     scoring_points: extractScoring(elements),
 *     disqualify_clauses: extractDisqualify(elements),
 *   };
 *
 * 输出对齐 Go 端 model.BidParsingOutput;空块返回 [],废标条款独立且必填.
 */

const crypto = require('crypto');
const path = require('path');
const { extract } = require('./extractor');

const BID_NO_RE = /(招标编号|项目编号|采购编号)[：:\s]*([A-Za-z0-9\-_]+)/;
const ISSUER_RE = /(招标人|采购人|招标代理机构|采购代理机构)[：:\s]*([^\n，。；;]{2,40})/;
const MANDATORY_MARK = ['★', '▲', '实质性', '必须', '不得偏离', '不可偏离', '否则废标', '否则作废标'];
const DISQUALIFY_KW = ['废标', '否决', '无效标', '无效投标', '拒绝', '不予受理', '取消投标资格'];
const SCORE_RE = /(\d+(?:\.\d+)?)\s*分/;
const NUM_UNIT_RE = /(不低于|不小于|不少于|大于等于|≥|>=|至少)\s*(\d+(?:\.\d+)?)\s*(TFlops|TOPS|GHz|MHz|Gbps|Mbps|TB|GB|MB|KB|核|路|个|台|套|年|月|天|小时|分钟|秒|%|mm|cm|m|km|V|A|W|kW|MW|元|万元|人|户|份|项|册|本)?/;
const MAX_UNIT_RE = /(不得大于|不得超过|不超过|不大于|小于|低于|≤|<)\s*(\d+(?:\.\d+)?)\s*(TFlops|TOPS|GHz|MHz|Gbps|Mbps|TB|GB|MB|KB|核|路|个|台|套|年|月|天|小时|分钟|秒|%|mm|cm|m|km|V|A|W|kW|MW|元|万元|人|户|份|项|册|本)?/;
const DATE_RE = /20\d{2}[-年/.\/]\d{1,2}[-月/.\/]\d{1,2}日?(?:\s*\d{1,2}[：:]\d{2}(?::\d{2})?)?/;

const KNOWN_SECTIONS = [
  '投标函', '法定代表人', '授权委托书', '投标保证金', '报价一览表', '分项报价',
  '技术方案', '技术规格', '技术参数', '技术标', '技术偏离表', '商务偏离表', '商务条款',
  '商务标', '商务响应', '投标报价', '报价',
  '实施计划', '项目计划', '售后服务', '技术支持', '培训方案',
  '质量保证', '资格证明', '营业执照', '业绩', '人员配置',
  '评分标准', '评标办法',
];
const REQUIRED_SECTIONS = ['投标函', '报价', '技术方案', '资格', '保证金'];

// 句子特征:出现这些词的是要求正文而非章节标题(均为多字词,避免单字误伤如"响应"中的"应")
const SENTENCE_MARK = ['必须', '应当', '应该', '不得', '严禁', '禁止', '要求', '需要',
  '按照', '否则', '废标', '盖章', '签字', '提交', '提供',
  '不少于', '不低于', '不小于', '截止', '逾期'];

const COMMERCIAL_CATS = {
  training: ['培训'],
  technical_support: ['售后', '技术支持', '运维', '保修期内服务'],
  project_plan: ['工期', '进度计划', '实施计划', '交付计划', '里程碑'],
  quality_management: ['质量保证', '质量管理体系', '质保', 'ISO'],
  delivery: ['交货', '交付', '到货', '安装调试'],
  warranty: ['保修期', '质保期', '免费维保'],
};
const ALL_COMMERCIAL_KW = Object.values(COMMERCIAL_CATS).flat();

const SCORING_SECTION_KW = ['评分标准', '评标办法', '评分细则', '分值', '评审标准'];
const SEAL_KW = ['盖章', '公章', '签字', '签章'];
const DOC_KW = ['证书', '证明', '报告', '材料', '文件', '执照', '资质'];

const includesAny = (t, kws) => kws.some((k) => t.includes(k));
const clauseRef = (p, idx) => (p.section ? p.section.slice(0, 40) : `正文第${idx + 1}段`);

function looksLikeSectionTitle(text) {
  const norm = text.replace(/^[第\d一二三四五六七八九十章节\d.、\s]+/, '').trim();
  if (!norm || norm.length > 12) return null;
  if (!KNOWN_SECTIONS.some((k) => norm.includes(k))) return null;
  if (includesAny(text, SENTENCE_MARK)) return null;
  if (/\d+\s*分/.test(text)) return null;
  return norm;
}

// -- 1. template ---------------------------------------------------------------
function extractTemplate(elements) {
  const full = elements.full_text;
  let hasPreset = ['投标文件格式', '投标文件组成', '投标文件目录', '投标一览表'].some((k) => full.includes(k));
  const sections = [];
  let order = 0;
  const seen = new Set();
  let lastL1Id = null;
  for (const p of elements.paragraphs) {
    if (p.heading_level !== 1 && p.heading_level !== 2) continue;
    const norm = looksLikeSectionTitle(p.text);
    if (!norm || seen.has(norm)) continue;
    seen.add(norm);
    order += 1;
    const id = `sec-${String(order).padStart(2, '0')}`;
    const level = p.heading_level;
    // L2 挂到最近的 L1 下,前端/后端据 parent_id 还原嵌套树
    const parentId = level === 2 ? lastL1Id : null;
    sections.push({
      section_id: id,
      title: norm,
      level,
      order,
      required: REQUIRED_SECTIONS.some((k) => norm.includes(k)),
      parent_id: parentId,
    });
    if (level === 1) lastL1Id = id;
  }
  if (!sections.length) {
    ['投标函', '技术方案', '商务响应', '资格证明文件'].forEach((t, i) => {
      sections.push({ section_id: `sec-${String(i + 1).padStart(2, '0')}`, title: t, level: 1, order: i + 1, required: true, parent_id: null });
    });
    hasPreset = false;
  }
  return {
    has_preset_structure: hasPreset,
    preset_sections: sections,
    format_rules: {
      font: full.includes('宋体') ? '宋体' : '',
      font_size: full.includes('小四') ? '小四' : '',
      has_cover: full.includes('封面'),
      has_toc: full.includes('目录'),
      seal_pages: elements.paragraphs.filter((p) => includesAny(p.text, SEAL_KW)).slice(0, 10).map((p) => p.text.slice(0, 30)),
      other: {},
    },
  };
}

// -- 2. 技术指标 ----------------------------------------------------------------
function extractIndicators(elements) {
  const out = [];
  let n = 0;
  elements.paragraphs.forEach((p, idx) => {
    const t = p.text;
    if (t.length > 200) return;
    if (includesAny(t, ALL_COMMERCIAL_KW)) return; // 商务句不重复计入
    const m = NUM_UNIT_RE.exec(t);
    const mx = MAX_UNIT_RE.exec(t);
    const star = t.includes('★') || t.includes('▲');
    const techHint = ['技术要求', '技术规格', '技术参数', '规格', '参数', '配置'].some((k) => (p.section + t).includes(k));
    if (!m && !mx && !(star && techHint)) return;
    n += 1;
    // 名称取自包含数值约束的分句,避免整行多要求时名实不符(如"响应时间…并发数…")
    const numMatch = m || mx;
    const segOf = (numMatch ? t.split(/[，,。；;]/).find((s) => s.includes(numMatch[0])) : null) || t;
    out.push({
      indicator_id: `tech-${String(n).padStart(2, '0')}`,
      name: segOf.split(/[：:，,。；;]/)[0].trim().slice(0, 20) || t.split(/[：:，,。；;]/)[0].slice(0, 20),
      requirement: t.slice(0, 200),
      clause_ref: clauseRef(p, idx),
      is_mandatory: star || includesAny(t, MANDATORY_MARK),
      unit: (m && m[3]) || (mx && mx[3]) || null,
      min_value: m && m[1] ? parseFloat(m[2]) : null,
      max_value: mx && mx[1] ? parseFloat(mx[2]) : null,
      response_type: p.table_idx >= 0 ? 'table' : 'narrative',
    });
  });
  return out;
}

// -- 3. 商务要求 -----------------------------------------------------------------
function extractCommercial(elements) {
  const out = [];
  let n = 0;
  elements.paragraphs.forEach((p, idx) => {
    const t = p.text;
    // 标题行本身不是商务要求(如"6.3 售后服务方案"),跳过以免产生空洞条目
    if (p.heading_level === 1 || p.heading_level === 2) return;
    for (const [cat, kws] of Object.entries(COMMERCIAL_CATS)) {
      if (!includesAny(t, kws)) continue;
      if (t.length < 6 || t.length > 300) continue;
      n += 1;
      const docs = [...t.matchAll(/《([^》]{2,30})》/g)].map((m) => m[1]);
      for (const k of DOC_KW) {
        if (!t.includes(k)) continue;
        if (docs.some((d) => d.includes(k))) continue;
        docs.push(/证明|证书$/.test(k) ? k : `${k}证明`);
      }
      // 去掉被更长条目覆盖的短项(如"证明"被"资质证明"覆盖)
      const deduped = docs.filter((d, i) => !docs.some((o, j) => j !== i && o.includes(d) && o.length > d.length));
      out.push({
        req_id: `biz-${String(n).padStart(2, '0')}`,
        category: cat,
        title: t.slice(0, 24),
        description: t.slice(0, 300),
        clause_ref: clauseRef(p, idx),
        deliverable: deduped.slice(0, 5),
      });
      break;
    }
  });
  return out;
}

// -- 4. 评分点 -------------------------------------------------------------------
function extractScoring(elements) {
  const out = [];
  let n = 0;
  let inScoring = false;
  elements.paragraphs.forEach((p, idx) => {
    const t = p.text;
    if (includesAny(t, SCORING_SECTION_KW) && t.length < 60) inScoring = true;
    const matches = [...t.matchAll(/(\d+(?:\.\d+)?)\s*分/g)];
    if (!matches.length) return;
    if (!inScoring && p.table_idx < 0 && t.length > 120) return;
    // 同一行多个分值逐个拆分，按就近分句还原评分项名称
    const segs = t.split(/[；;。]/);
    for (const m of matches) {
      n += 1;
      const seg = segs.find((s) => s.includes(m[0])) || t;
      const firstCell = seg.split(' | ')[0];
      const item = (firstCell.split(/[：:，,。；;]/)[0].trim().slice(0, 30)) || `评分项${n}`;
      out.push({
        scoring_id: `score-${String(n).padStart(2, '0')}`,
        scoring_item: item,
        full_score: parseFloat(m[1]),
        description: t.slice(0, 300),
        clause_ref: clauseRef(p, idx),
        grading_tiers: ['优', '良', '一般', '差'].filter((k) => t.includes(k))
          .map((k) => ({ tier: k, range: '', criteria: t.slice(0, 120) })),
        required_docs: [...t.matchAll(/《([^》]{2,30})》/g)].map((x) => x[1]).slice(0, 5),
      });
    }
  });
  return out;
}

// -- 5. 废标条款(单独清单) ---------------------------------------------------------------
function extractDisqualify(elements) {
  const out = [];
  let n = 0;
  elements.paragraphs.forEach((p, idx) => {
    const t = p.text;
    const isStarred = t.includes('★') || t.includes('▲');
    // 带 ★▲ 标记的行豁免最小长度检查(短强制条款如"★必须盖章"不应丢弃)
    if ((!isStarred && t.length < 8) || t.length > 400) return;
    const hitDq = includesAny(t, DISQUALIFY_KW);
    // 仅多字词，避免单字"应/须"误伤"响应/须知"等正常词
    const hitStar = isStarred && ['实质性', '必须', '应当', '应该', '要求', '否则', '不得'].some((k) => t.includes(k));
    if (!hitDq && !hitStar) return;
    n += 1;
    const severity = hitDq ? 'disqualify' : (['实质性', '★', '▲'].some((k) => t.includes(k)) ? 'substantive' : 'warning');
    const dm = DATE_RE.exec(t);
    const check = includesAny(t, SEAL_KW) ? 'seal'
      : (t.includes('保证金') || t.includes('截止') || dm) ? 'deadline'
      : includesAny(t, DOC_KW) ? 'doc_required'
      : ['格式', '装订', '份数', '密封'].some((k) => t.includes(k)) ? 'format' : 'presence';
    out.push({
      clause_id: `dq-${String(n).padStart(2, '0')}`,
      clause_text: t.slice(0, 300),
      clause_ref: clauseRef(p, idx),
      severity,
      check_type: check,
      required_docs: [...t.matchAll(/《([^》]{2,30})》/g)].map((x) => x[1]).slice(0, 5),
      deadline: dm ? dm[0] : null,
    });
  });
  return out;
}

// -- 总装 -------------------------------------------------------------------------
async function parseBidFile(input, { titleHint = '', bidNoHint = '', filenameHint = '' } = {}) {
  // input: 文件路径或 Buffer
  const filename = typeof input === 'string' ? input : (filenameHint || '');
  const elements = await extract(input, filenameHint);
  const full = elements.full_text;

  const mNo = BID_NO_RE.exec(full);
  const mIs = ISSUER_RE.exec(full);
  const firstHeading = (elements.paragraphs.find((p) => p.heading_level === 1) || {}).text || '';
  const base = typeof input === 'string' ? path.basename(input) : '';
  const looksTmp = /^tmp[a-zA-Z0-9_-]{4,}/.test(base.replace(path.extname(base), ''));
  const title = (titleHint || firstHeading.slice(0, 60) || filenameHint || (looksTmp ? '' : base) || '未命名招标文件').slice(0, 120);

  const hash = crypto.createHash('sha256');
  if (Buffer.isBuffer(input)) {
    hash.update(input);
  } else {
    try { hash.update(require('fs').readFileSync(input)); }
    catch { hash.update(String(input)); }
  }

  return {
    source_meta: {
      title,
      bid_no: bidNoHint || (mNo ? mNo[2] : ''),
      issuer: mIs ? mIs[2].trim() : '',
      parse_time: new Date().toISOString(),
      file_hash: 'sha256:' + hash.digest('hex'),
      total_pages: Math.max(1, Math.floor(elements.paragraphs.length / 45) + 1),
    },
    template: extractTemplate(elements),
    technical_indicators: extractIndicators(elements),
    commercial_reqs: extractCommercial(elements),
    scoring_points: extractScoring(elements),
    disqualify_clauses: extractDisqualify(elements),
  };
}

module.exports = { parseBidFile, extractTemplate, extractIndicators, extractCommercial, extractScoring, extractDisqualify };
