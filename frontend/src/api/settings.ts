import apiClient from './client';

export async function getTheme(): Promise<{ type: string }> {
  const response = await apiClient.get('/settings/theme');
  return response.data;
}

export async function setTheme(type: string): Promise<{ type: string }> {
  const response = await apiClient.put('/settings/theme', { type });
  return response.data;
}

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

export async function getExportSettings(): Promise<ExportSettings & { format: string }> {
  const response = await apiClient.get('/settings/export');
  return response.data;
}

export async function setExportSettings(data: ExportSettings): Promise<void> {
  await apiClient.put('/settings/export', data);
}

export async function addApiKey(data: { provider: string; model: string; api_key: string; region?: string; base_url?: string }): Promise<void> {
  await apiClient.post('/settings/apikeys', data);
}

// Template management API
export interface TemplateData {
  id?: number;
  name: string;
  type: 'bidding' | 'proposal' | 'custom';
  category: string;
  description: string;
  keywords: string[];
  createdAt?: number;
}

export async function getTemplateList(): Promise<{ templates: TemplateData[] }> {
  const response = await apiClient.get('/settings/templates');
  return response.data;
}

export async function createTemplate(data: Omit<TemplateData, 'id' | 'createdAt'>): Promise<void> {
  await apiClient.post('/settings/templates', data);
}

export async function updateTemplate(id: number, data: Omit<TemplateData, 'id' | 'createdAt'>): Promise<void> {
  await apiClient.put(`/settings/templates/${id}`, data);
}

export async function deleteTemplate(id: number): Promise<void> {
  await apiClient.delete(`/settings/templates/${id}`);
}

export async function resetTemplates(): Promise<{ templates: TemplateData[] }> {
  const response = await apiClient.post('/settings/templates/reset');
  return response.data;
}

export async function toggleApiKey(id: string): Promise<{ enabled: number }> {
  const res = await apiClient.put(`/settings/apikeys/${id}/toggle`);
  return res.data;
}

export async function deleteApiKey(id: string): Promise<void> {
  await apiClient.delete(`/settings/apikeys/${id}`);
}

export async function testApiKey(data: { id?: string; provider?: string; model?: string; key?: string; endpoint?: string; format?: string }): Promise<{ available: boolean; error?: string }> {
  const res = await apiClient.post('/settings/test-key', data);
  return res.data;
}

// ── Profile settings ────────────────────────────────────────────
export interface ProfileSettings {
  id: number;
  auto_redirect_after_extract: number;
  default_skill_id: string;
  extract_preview_mode: string;
}

export async function getProfileSettings(): Promise<ProfileSettings> {
  const res = await apiClient.get('/settings/profile');
  return res.data;
}

export async function setProfileSettings(data: Partial<ProfileSettings>): Promise<ProfileSettings> {
  const res = await apiClient.put<ProfileSettings>('/settings/profile', data);
  return res.data;
}

