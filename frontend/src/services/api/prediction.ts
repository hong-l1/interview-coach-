import apiClient from './client';

export const predictionService = {
  getEvaluationReport: async (recordId: number) => {
    return apiClient.post('/evaluation/report', { record_id: recordId });
  },

  getEvaluationPrediction: async (recordId: number) => {
    return apiClient.post('/evaluation/predict', { record_id: recordId });
  },

  getPredictionList: async () => {
    return { list: [], total: 0, page: 1, size: 10 };
  },

  getPredictionDetail: async (id: number) => {
    return apiClient.post('/evaluation/predict', { record_id: id });
  },
};
