import apiClient from './client';

export interface ExtractionRule {
  id: string;
  fieldName: string;
  pattern: string;
  enabled: boolean;
  category: string;
  groupName: string;
  aliasFor?: string;
}

export async function getRules(group?: string): Promise<ExtractionRule[]> {
  const params = group ? { group } : {};
  const response = await apiClient.get<ExtractionRule[]>('/rules', { params });
  return response.data;
}

export async function addRule(data: Omit<ExtractionRule, 'id'>): Promise<ExtractionRule> {
  const response = await apiClient.post<ExtractionRule>('/rules', data);
  return response.data;
}

export async function updateRule(id: string, data: Partial<ExtractionRule>): Promise<ExtractionRule> {
  const response = await apiClient.put<ExtractionRule>(`/rules/${id}`, data);
  return response.data;
}

export async function deleteRule(id: string): Promise<{ deleted: boolean }> {
  const response = await apiClient.delete(`/rules/${id}`);
  return response.data;
}

export async function seedUniversalRules(): Promise<{ ok: boolean; added: number; total: number }> {
  const response = await apiClient.post<{ ok: boolean; added: number; total: number }>('/rules/universal');
  return response.data;
}


