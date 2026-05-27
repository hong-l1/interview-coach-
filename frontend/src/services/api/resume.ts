import apiClient from './client';

export type ResumeItem = {
  id: number;
  user_id?: number;
  file_name: string;
  file_size: number;
  file_type: string;
  is_default: number;
  created_at: number;
  updated_at?: number;
};

export function formatFileSize(bytes = 0) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

export function formatTimestamp(timestamp?: number) {
  if (!timestamp) return '-';
  return new Date(timestamp * 1000).toLocaleDateString('zh-CN');
}

export const resumeService = {
  async list() {
    const data: any = await apiClient.get('/resume/list');
    return {
      resumes: (data?.resumes || []) as ResumeItem[],
      total: Number(data?.total || data?.resumes?.length || 0),
    };
  },

  async upload(file: File) {
    const formData = new FormData();
    formData.append('file', file);
    return apiClient.post('/resume/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 300000,
    });
  },

  async remove(id: number) {
    return apiClient.delete(`/resume/${id}`);
  },

  async setDefault(id: number) {
    return apiClient.put(`/resume/${id}/default`);
  },
};
