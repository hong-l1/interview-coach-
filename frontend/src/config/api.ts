export function getApiBaseUrl() {
  if (process.env.NEXT_PUBLIC_API_BASE_URL) {
    return process.env.NEXT_PUBLIC_API_BASE_URL;
  }

  return '/api';
}

export const API_BASE_URL = getApiBaseUrl();

export const INTERVIEW_API = {
  get START_STREAM() {
    return `${getApiBaseUrl()}/mianshi/stream/start`;
  },
  get RESUME_STREAM() {
    return `${getApiBaseUrl()}/mianshi/stream/resume`;
  },
  get SESSION_INFO() {
    return `${getApiBaseUrl()}/mianshi/session/info`;
  },
  get END_INTERVIEW() {
    return `${getApiBaseUrl()}/mianshi/interview/end`;
  },
  get SUBMIT_ANSWER() {
    return `${getApiBaseUrl()}/mianshi/answer/submit`;
  },
};

export const USER_API = {
  get LOGIN() {
    return `${getApiBaseUrl()}/user/login`;
  },
  get REGISTER() {
    return `${getApiBaseUrl()}/user/register`;
  },
  get MODELS() {
    return `${getApiBaseUrl()}/user/models`;
  },
};

export const RESUME_API = {
  get UPLOAD() {
    return `${getApiBaseUrl()}/resume/upload`;
  },
  detail(id: number | string) {
    return `${getApiBaseUrl()}/resume/${id}`;
  },
  setDefault(id: number | string) {
    return `${getApiBaseUrl()}/resume/${id}/default`;
  },
};

export const EVALUATION_API = {
  get REPORT() {
    return `${getApiBaseUrl()}/evaluation/report`;
  },
  get PREDICT() {
    return `${getApiBaseUrl()}/evaluation/predict`;
  },
};
