import type { AxiosAdapter, InternalAxiosRequestConfig } from 'axios';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { clearStoredToken, setStoredToken } from '../stores/auth';
import { createApiClient } from './http';

function responseAdapter(status: number, data: unknown): AxiosAdapter {
  return vi.fn(async (config: InternalAxiosRequestConfig) => ({
    data,
    status,
    statusText: status === 200 ? 'OK' : 'Error',
    headers: {},
    config
  }));
}

describe('api client', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('adds the bearer token to outgoing requests', async () => {
    setStoredToken('token-123');
    const adapter = responseAdapter(200, { code: 0, message: 'ok', data: { ok: true } });
    const api = createApiClient({ adapter });

    await api.get('/auth/me');

    const config = vi.mocked(adapter).mock.calls[0][0] as InternalAxiosRequestConfig;
    expect(config.headers.Authorization).toBe('Bearer token-123');
  });

  it('unwraps successful backend responses', async () => {
    const api = createApiClient({
      adapter: responseAdapter(200, { code: 0, message: 'ok', data: { name: 'Core' } })
    });

    await expect(api.get('/teams/1')).resolves.toEqual({ name: 'Core' });
  });

  it('supports patch requests', async () => {
    const adapter = responseAdapter(200, { code: 0, message: 'ok', data: { id: 1 } });
    const api = createApiClient({ adapter });

    await api.patch('/feedback/1/status', { status: 'resolved' });

    const config = vi.mocked(adapter).mock.calls[0][0] as InternalAxiosRequestConfig;
    expect(config.method).toBe('patch');
    expect(config.url).toBe('/feedback/1/status');
    expect(config.data).toBe(JSON.stringify({ status: 'resolved' }));
  });

  it('supports delete requests', async () => {
    const adapter = responseAdapter(200, { code: 0, message: 'ok', data: { removed: true } });
    const api = createApiClient({ adapter });

    await api.delete('/teams/1/members/2');

    const config = vi.mocked(adapter).mock.calls[0][0] as InternalAxiosRequestConfig;
    expect(config.method).toBe('delete');
    expect(config.url).toBe('/teams/1/members/2');
  });

  it('clears token and notifies callers when backend returns unauthorized', async () => {
    setStoredToken('token-123');
    const onUnauthorized = vi.fn();
    const api = createApiClient({
      adapter: responseAdapter(401, { code: 401, message: 'unauthorized' }),
      onUnauthorized
    });

    await expect(api.get('/auth/me')).rejects.toThrow('unauthorized');

    expect(clearStoredToken()).toBeUndefined();
    expect(localStorage.getItem('pulseroad.auth.token')).toBeNull();
    expect(onUnauthorized).toHaveBeenCalledTimes(1);
  });

  it('reports a clear message when the dev proxy cannot reach the backend', async () => {
    const api = createApiClient({
      adapter: responseAdapter(500, 'Error occurred while trying to proxy: localhost:5173/api/auth/register')
    });

    await expect(api.post('/auth/register', {})).rejects.toThrow('后端 API 服务不可用');
  });
});
