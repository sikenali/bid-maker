import apiClient from './client';

export interface TableRow {
  cells: string[];
}

export interface DocTable {
  rows: TableRow[];
}

export interface MarkedItem {
  symbol: string;
  text: string;
  page: number;
}

export interface UploadResult {
  id: string;
  filename: string;
  size: number;
  result: {
    status: string;
    text?: string;
    extracts?: Record<string, unknown>;
    groups?: Record<string, string>;
    chapters?: Array<{ title: string; content: string[]; page: number }>;
    tables?: DocTable[];
    markedItems?: MarkedItem[];
    fieldParaMap?: Record<string, number>;
    fieldGroups?: Record<string, string>;
    pageCount?: number;
    paraToPage?: number[];
    error?: string;
  };
}

export interface ParseStatus {
  id: string;
  status: string;
  progress: number;
  filename: string;
  fileSize?: number;
  error?: string;
  result?: {
    status: string;
    text?: string;
    extracts?: Record<string, unknown>;
    groups?: Record<string, string>;
    chapters?: Array<{ title: string; content: string[]; page: number }>;
    tables?: DocTable[];
    markedItems?: MarkedItem[];
    fieldParaMap?: Record<string, number>;
    fieldGroups?: Record<string, string>;
    pageCount?: number;
    paraToPage?: number[];
    error?: string;
  };
}

export async function uploadFile(file: File, skillId?: string): Promise<UploadResult> {
  const formData = new FormData();
  formData.append('file', file);
  if (skillId) {
    formData.append('skill_id', skillId);
  }
  const response = await apiClient.post<UploadResult>('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    // 大文件同步解析 + LLM 增强耗时较长，单独放宽超时（全局默认 60s）
    timeout: 180000,
  });
  return response.data;
}

export async function getParseStatus(id: string): Promise<ParseStatus> {
  const response = await apiClient.get<ParseStatus>(`/upload/${id}/status`);
  return response.data;
}