import axios, { AxiosInstance, InternalAxiosRequestConfig, AxiosResponse } from 'axios';
import { getApiBaseUrl } from '@/config/api';

const apiClient: AxiosInstance = axios.create({
  baseURL: getApiBaseUrl(),
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

function readLocalJSON(key: string) {
  if (typeof window === 'undefined') return null;
  try {
    const raw = window.localStorage.getItem(key);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function readUserIDFromToken(token: string) {
  try {
    const payload = token.split('.')[1];
    if (!payload) return '';
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/');
    const decoded = JSON.parse(window.atob(normalized));
    return String(decoded?.userid || decoded?.user_id || decoded?.id || '');
  } catch {
    return '';
  }
}

apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const url = config.url || '';
    const isAuthFree = url.includes('/user/register') || url.includes('/user/login');
    const token = typeof window !== 'undefined' ? window.localStorage.getItem('token') : '';
    if (token && !isAuthFree) {
      config.headers = (config.headers || {}) as any;
      (config.headers as any).Authorization = `Bearer ${token}`;
      (config.headers as any)['X-Auth-Token'] = token;
    }

    const user = readLocalJSON('user');
    const userID = user?.id || user?.user_id || user?.ID || (token ? readUserIDFromToken(token) : '');
    if (userID && !isAuthFree) {
      config.headers = (config.headers || {}) as any;
      (config.headers as any).userid = String(userID);
    }

    if (url.includes('/evaluation/report') || url.includes('/evaluation/predict')) {
      config.timeout = 180000;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

apiClient.interceptors.response.use(
  (response: AxiosResponse<any>) => {
    const payload = response?.data;
    if (payload && typeof payload === 'object' && 'code' in payload) {
      if (payload.code === 0) {
        return payload.data;
      }
      if (payload.code === 401) {
        window.localStorage.removeItem('token');
      }
      return Promise.reject({ response, message: payload.message, code: payload.code });
    }
    return payload;
  },
  (error) => {
    if (error.response?.status === 401) {
      window.localStorage.removeItem('token');
    }
    return Promise.reject(error);
  }
);

export default apiClient;
