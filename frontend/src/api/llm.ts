import apiClient from './client';

export interface LlmSkill {
  id: string;
  skill_name: string;
  skill_code: string;
  biz_type: string;
  prompt_template: string;
  temperature: number;
  top_p: number;
  max_tokens: number;
  enable: number;
  enable_llm: number;
  bind_prompt_template_id: string;
  bind_rules_json: string;
  llm_api_config_id: string;
  version: number;
  create_time: string;
  update_time: string;
}

export interface LlmPromptTemplate {
  id: string;
  name: string;
  description: string;
  biz_type: string;
  template_text: string;
  variables: string[];
  enable: number;
  version: number;
  created_at: string;
  updated_at: string;
}

export async function getLlmSkills(): Promise<LlmSkill[]> {
  const response = await apiClient.get<LlmSkill[]>('/rules/llm-skills');
  return response.data;
}

export async function createLlmSkill(data: Partial<LlmSkill>): Promise<LlmSkill> {
  const response = await apiClient.post<LlmSkill>('/rules/llm-skills', data);
  return response.data;
}

export async function updateLlmSkill(id: string, data: Partial<LlmSkill>): Promise<LlmSkill> {
  const response = await apiClient.put<LlmSkill>(`/rules/llm-skills/${id}`, data);
  return response.data;
}

export async function deleteLlmSkill(id: string): Promise<{ deleted: boolean }> {
  const response = await apiClient.delete(`/rules/llm-skills/${id}`);
  return response.data;
}

export async function getLlmPrompts(): Promise<LlmPromptTemplate[]> {
  const response = await apiClient.get<LlmPromptTemplate[]>('/llm-prompts');
  return response.data;
}

export async function getLlmApiConfigs(): Promise<Array<{ id: string; provider: string; model: string }>> {
  const response = await apiClient.get<Array<{ id: string; provider: string; model: string }>>('/rules/llm-api-configs');
  return response.data;
}
