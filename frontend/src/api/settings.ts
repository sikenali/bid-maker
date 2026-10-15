// All settings use localStorage for persistence
const STORAGE_KEY = 'bid-maker-settings';

function loadSettings(): Record<string, any> {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}');
  } catch {
    return {};
  }
}

function saveSettings(data: Record<string, any>) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
  } catch {}
}

function getSetting<T>(key: string, defaultVal: T): T {
  const settings = loadSettings();
  return settings[key] ?? defaultVal;
}

function setSetting(key: string, value: any) {
  const settings = loadSettings();
  settings[key] = value;
  saveSettings(settings);
}

// ── Theme ──────────────────────────────────────────────────────
export async function getTheme(): Promise<{ type: string }> {
  return { type: getSetting('theme.type', 'parchment') };
}

export async function setTheme(type: string): Promise<{ type: string }> {
  setSetting('theme.type', type);
  return { type };
}

// ── Export Settings ────────────────────────────────────────────
export interface ExportSettings {
  format: string;
  include_table_of_contents?: number;
  page_numbers?: number;
  header_footer?: number;
  word_keep_style?: number;
  word_toc?: number;
  word_embed_font?: number;
  pdf_high_res?: number;
  pdf_bookmarks?: number;
  pdf_watermark?: number;
  md_keep_tables?: number;
  md_demote_heading?: number;
}

const DEFAULT_EXPORT: ExportSettings = {
  format: 'docx',
  include_table_of_contents: 1,
  page_numbers: 1,
  header_footer: 1,
  word_keep_style: 1,
  word_toc: 1,
  word_embed_font: 0,
  pdf_high_res: 1,
  pdf_bookmarks: 1,
  pdf_watermark: 0,
  md_keep_tables: 1,
  md_demote_heading: 0,
};

export async function getExportSettings(): Promise<ExportSettings> {
  return getSetting<ExportSettings>('export', DEFAULT_EXPORT);
}

export async function setExportSettings(data: ExportSettings): Promise<void> {
  setSetting('export', data);
}

// ── API Keys ───────────────────────────────────────────────────
export interface ApiKeyData {
  id: string;
  provider: string;
  model: string;
  api_key: string;
  base_url?: string;
  format?: string;
  enabled?: number;
  createdAt?: number;
}

export function getApiKeyList(): ApiKeyData[] {
  return getSetting<ApiKeyData[]>('apikeys', []);
}

export async function addApiKey(data: { provider: string; model: string; api_key: string; base_url?: string; format?: string }): Promise<void> {
  const keys = getSetting<ApiKeyData[]>('apikeys', []);
  keys.push({
    ...data,
    id: Date.now().toString(),
    enabled: 1,
    createdAt: Date.now(),
  });
  setSetting('apikeys', keys);
}

export async function toggleApiKey(id: string): Promise<{ enabled: number }> {
  const keys = getSetting<ApiKeyData[]>('apikeys', []);
  const key = keys.find(k => k.id === id);
  if (key) {
    key.enabled = key.enabled === 1 ? 0 : 1;
    setSetting('apikeys', keys);
    return { enabled: key.enabled };
  }
  return { enabled: 0 };
}

export async function deleteApiKey(id: string): Promise<void> {
  const keys = getSetting<ApiKeyData[]>('apikeys', []);
  setSetting('apikeys', keys.filter(k => k.id !== id));
}

export async function testApiKey(data: { id?: string; provider?: string; model?: string; key?: string; endpoint?: string; format?: string }): Promise<{ available: boolean; error?: string }> {
  const keys = getSetting<ApiKeyData[]>('apikeys', []);
  const key = data.id ? keys.find(k => k.id === data.id) : null;
  const apiKey = data.key || key?.api_key || '';
  const provider = data.provider || key?.provider || '';
  const format = data.format || (key as any)?.format || 'openai';
  
  if (!apiKey) return { available: false, error: '未配置 API Key' };
  
  try {
    const resp = await fetch('/api/test-key', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        provider,
        model: data.model || key?.model || '',
        key: apiKey,
        endpoint: data.endpoint || key?.base_url || '',
        format,
      }),
    });
    return await resp.json();
  } catch (err: any) {
    return { available: false, error: err?.message || '测试失败' };
  }
}

export async function updateApiKey(id: string, data: { provider?: string; model?: string; api_key?: string; base_url?: string; format?: string }): Promise<void> {
  const keys = getSetting<ApiKeyData[]>('apikeys', []);
  const key = keys.find(k => k.id === id);
  if (key) {
    if (data.provider) key.provider = data.provider;
    if (data.model) key.model = data.model;
    if (data.api_key) key.api_key = data.api_key;
    if (data.base_url !== undefined) key.base_url = data.base_url;
    if (data.format !== undefined) key.format = data.format;
    setSetting('apikeys', keys);
  }
}

// ── Templates ──────────────────────────────────────────────────
export interface TemplateSection {
  title: string;
  level: number;
  children?: TemplateSection[];
}

export interface TemplateData {
  id?: number;
  name: string;
  type: 'bidding' | 'proposal' | 'custom';
  category: string;
  description: string;
  keywords: string[];
  content: TemplateSection[];
  enabled?: number;
  createdAt?: number;
}

const DEFAULT_TEMPLATES: TemplateData[] = [
  {
    id: 1,
    name: '标准投标模板',
    type: 'bidding',
    category: '综合类',
    description: '适用于大多数招标项目的标准投标文件模板',
    keywords: ['标准', '通用', '综合'],
    content: [
      { title: '第一章 投标函', level: 1 },
      { title: '第二章 法定代表人身份证明', level: 1 },
      { title: '第三章 授权委托书', level: 1 },
      { title: '第四章 投标保证金', level: 1 },
      { title: '第五章 商务标', level: 1, children: [
        { title: '5.1 投标报价表', level: 2 },
        { title: '5.2 企业资质证书', level: 2 },
        { title: '5.3 业绩清单', level: 2 },
      ]},
      { title: '第六章 技术标', level: 1, children: [
        { title: '6.1 技术方案', level: 2 },
        { title: '6.2 质量保障措施', level: 2 },
        { title: '6.3 售后服务方案', level: 2 },
      ]},
    ],
    createdAt: Date.now(),
  },
  {
    id: 2,
    name: '技术方案模板',
    type: 'proposal',
    category: '技术类',
    description: '专注于技术方案部分的专业模板',
    keywords: ['技术', '方案', '专业'],
    content: [
      { title: '一、项目概述', level: 1 },
      { title: '二、技术需求分析', level: 1 },
      { title: '三、总体技术方案', level: 1, children: [
        { title: '3.1 系统架构设计', level: 2 },
        { title: '3.2 关键技术选型', level: 2 },
      ]},
      { title: '四、实施方案', level: 1, children: [
        { title: '4.1 实施计划', level: 2 },
        { title: '4.2 人员配置', level: 2 },
        { title: '4.3 质量保证', level: 2 },
      ]},
      { title: '五、售后服务', level: 1 },
    ],
    createdAt: Date.now(),
  },
  {
    id: 3,
    name: '商务标模板',
    type: 'bidding',
    category: '商务类',
    description: '商务标专用模板，包含资质、业绩等模块',
    keywords: ['商务', '资质', '业绩'],
    content: [
      { title: '第一章 投标报价', level: 1 },
      { title: '第二章 企业资质', level: 1, children: [
        { title: '2.1 营业执照', level: 2 },
        { title: '2.2 资质证书', level: 2 },
        { title: '2.3 质量管理体系认证', level: 2 },
      ]},
      { title: '第三章 业绩证明', level: 1 },
      { title: '第四章 财务状况', level: 1 },
      { title: '第五章 售后服务承诺', level: 1 },
    ],
    createdAt: Date.now(),
  },
];

export function getTemplateList(): { templates: TemplateData[] } {
  return { templates: getSetting<TemplateData[]>('templates', DEFAULT_TEMPLATES) };
}

export async function createTemplate(data: Omit<TemplateData, 'id' | 'createdAt'>): Promise<void> {
  const templates = getSetting<TemplateData[]>('templates', DEFAULT_TEMPLATES);
  const maxId = templates.reduce((max, t) => Math.max(max, t.id || 0), 0);
  templates.push({
    ...data,
    id: maxId + 1,
    createdAt: Date.now(),
  });
  setSetting('templates', templates);
}

export async function updateTemplate(id: number, data: Omit<TemplateData, 'id' | 'createdAt'>): Promise<void> {
  const templates = getSetting<TemplateData[]>('templates', DEFAULT_TEMPLATES);
  const idx = templates.findIndex(t => t.id === id);
  if (idx >= 0) {
    templates[idx] = { ...templates[idx], ...data };
    setSetting('templates', templates);
  }
}

export async function deleteTemplate(id: number): Promise<void> {
  const templates = getSetting<TemplateData[]>('templates', DEFAULT_TEMPLATES);
  setSetting('templates', templates.filter(t => t.id !== id));
}

export async function resetTemplates(): Promise<{ templates: TemplateData[] }> {
  setSetting('templates', DEFAULT_TEMPLATES);
  return { templates: DEFAULT_TEMPLATES };
}

// ── Skills ─────────────────────────────────────────────────────
export interface SkillItem {
  id: string;
  name: string;
  description: string;
  group: string;
  type: 'rules' | 'prompt';
  fields: string[];
  content?: string;
  enabled: boolean;
  importedAt: string;
}

const DEFAULT_SKILLS: SkillItem[] = [
  {
    id: '1',
    name: '招标文件要点提取',
    description: '自动抽取资质门槛、工期与评分办法',
    group: 'info',
    type: 'rules',
    fields: ['项目名称', '招标编号', '投标截止', '资质要求', '评分标准'],
    enabled: true,
    importedAt: new Date().toISOString().slice(0, 10),
  },
  {
    id: '2',
    name: '商务标生成',
    description: '基于资质与业绩生成商务标内容',
    group: 'business',
    type: 'prompt',
    fields: ['公司简介', '资质证照', '业绩清单'],
    enabled: true,
    importedAt: new Date().toISOString().slice(0, 10),
  },
  {
    id: '3',
    name: '技术方案生成',
    description: '针对技术要求生成响应方案',
    group: 'tech',
    type: 'prompt',
    fields: ['技术方案', '质量保障', '售后服务'],
    enabled: true,
    importedAt: new Date().toISOString().slice(0, 10),
  },
];

export function getSkillList(): SkillItem[] {
  return getSetting<SkillItem[]>('skills', DEFAULT_SKILLS);
}

export async function updateSkill(name: string, data: { name?: string; description?: string; enabled?: number }): Promise<void> {
  const skills = getSetting<SkillItem[]>('skills', DEFAULT_SKILLS);
  const skill = skills.find(s => s.name === name);
  if (skill) {
    if (data.name) skill.name = data.name;
    if (data.description !== undefined) skill.description = data.description;
    if (data.enabled !== undefined) skill.enabled = data.enabled === 1;
    setSetting('skills', skills);
  }
}

export async function deleteSkill(name: string): Promise<void> {
  const skills = getSetting<SkillItem[]>('skills', DEFAULT_SKILLS);
  setSetting('skills', skills.filter(s => s.name !== name));
}

export async function resetSkills(): Promise<void> {
  setSetting('skills', DEFAULT_SKILLS);
}

export async function syncSkills(): Promise<{ synced: number; message: string }> {
  try {
    const resp = await fetch('/api/local-skills');
    if (!resp.ok) return { synced: 0, message: '同步失败：后端不可用' };
    const data = await resp.json();
    const remoteSkills: any[] = data.skills || [];
    if (remoteSkills.length === 0) return { synced: 0, message: '未发现可同步的本地技能' };

    const existing = getSkillList();
    let synced = 0;
    for (const rs of remoteSkills) {
      const name = rs.name || rs.path || '';
      if (!name) continue;
      const exists = existing.find(s => s.name === name);
      if (exists) continue;
      existing.push({
        id: `sync_${Date.now()}_${synced}`,
        name,
        description: rs.description || '',
        group: 'info',
        type: 'prompt',
        fields: [],
        content: '',
        enabled: true,
        importedAt: new Date().toISOString().slice(0, 10),
      });
      synced++;
    }
    if (synced > 0) setSetting('skills', existing);
    return { synced, message: `已同步 ${synced} 个技能` };
  } catch (err: any) {
    return { synced: 0, message: `同步失败: ${err?.message || '网络错误'}` };
  }
}
