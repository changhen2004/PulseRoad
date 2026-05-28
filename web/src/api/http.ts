import axios, {
  type AxiosAdapter,
  type AxiosError,
  type AxiosInstance,
  type AxiosRequestConfig
} from 'axios';

import { clearStoredToken, getStoredToken } from '../stores/auth';

export interface ApiEnvelope<T> {
  code: number;
  message: string;
  data?: T;
}

interface ApiClientOptions {
  adapter?: AxiosAdapter;
  onUnauthorized?: () => void;
}

export interface ApiClient {
  get<T>(url: string, config?: AxiosRequestConfig): Promise<T>;
  post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>;
}

export const apiEvents = {
  onUnauthorized: undefined as undefined | (() => void)
};

export function createApiClient(options: ApiClientOptions = {}): ApiClient {
  const instance = axios.create({
    baseURL: '/api',
    adapter: options.adapter
  });

  instance.interceptors.request.use((config) => {
    const token = getStoredToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  });

  const unauthorized = () => {
    clearStoredToken();
    options.onUnauthorized?.();
    apiEvents.onUnauthorized?.();
  };

  return {
    get<T>(url: string, config?: AxiosRequestConfig) {
      return request<T>(instance, { ...config, method: 'GET', url }, unauthorized);
    },
    post<T>(url: string, data?: unknown, config?: AxiosRequestConfig) {
      return request<T>(instance, { ...config, method: 'POST', url, data }, unauthorized);
    }
  };
}

export const api = createApiClient();

async function request<T>(
  instance: AxiosInstance,
  config: AxiosRequestConfig,
  onUnauthorized: () => void
): Promise<T> {
  try {
    const response = await instance.request<ApiEnvelope<T>>(config);
    return unwrapResponse(response.status, response.data, onUnauthorized);
  } catch (error) {
    const axiosError = error as AxiosError<ApiEnvelope<unknown>>;
    if (axiosError.response) {
      unwrapResponse(axiosError.response.status, axiosError.response.data, onUnauthorized);
    }
    if (axiosError.request) {
      throw new Error('无法连接后端 API 服务，请确认 API 服务已启动');
    }
    throw error;
  }
}

function unwrapResponse<T>(
  status: number,
  payload: ApiEnvelope<T>,
  onUnauthorized: () => void
): T {
  if (!isApiEnvelope(payload)) {
    if (status >= 500) {
      throw new Error('后端 API 服务不可用，请确认 API 服务已启动');
    }
    throw new Error(`HTTP ${status} 请求失败`);
  }

  if (status === 401 || payload.code === 401) {
    onUnauthorized();
    throw new Error(payload.message || 'unauthorized');
  }

  if (payload.code !== 0) {
    throw new Error(payload.message || 'request failed');
  }

  return payload.data as T;
}

function isApiEnvelope<T>(payload: unknown): payload is ApiEnvelope<T> {
  return (
    typeof payload === 'object' &&
    payload !== null &&
    'code' in payload &&
    typeof (payload as ApiEnvelope<T>).code === 'number'
  );
}
