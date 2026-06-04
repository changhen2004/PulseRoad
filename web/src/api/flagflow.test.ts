import { describe, expect, it, vi } from 'vitest';

import { createFlagflowApi } from './flagflow';
import type { ApiClient } from './http';

function fakeApiClient(): ApiClient {
  return {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn()
  };
}

describe('flagflow api', () => {
  it('creates a feature flag for a product', () => {
    const client = fakeApiClient();
    const api = createFlagflowApi(client);
    const payload = {
      key: 'new_dashboard',
      name: 'New Dashboard',
      description: '',
      environment: 'production',
      rollout_percentage: 25
    };

    api.create(7, payload);

    expect(client.post).toHaveBeenCalledWith('/products/7/flags', payload);
  });

  it('lists flags by product and environment', () => {
    const client = fakeApiClient();
    const api = createFlagflowApi(client);

    api.listByProduct(7, 'production');

    expect(client.get).toHaveBeenCalledWith('/products/7/flags?environment=production');
  });

  it('lists flags by product without environment filter', () => {
    const client = fakeApiClient();
    const api = createFlagflowApi(client);

    api.listByProduct(7);

    expect(client.get).toHaveBeenCalledWith('/products/7/flags');
  });

  it('toggles a flag', () => {
    const client = fakeApiClient();
    const api = createFlagflowApi(client);

    api.toggle(11, { enabled: true });

    expect(client.patch).toHaveBeenCalledWith('/flags/11/toggle', { enabled: true });
  });

  it('evaluates a flag', () => {
    const client = fakeApiClient();
    const api = createFlagflowApi(client);
    const payload = {
      product_id: 7,
      key: 'new_dashboard',
      environment: 'production',
      user_key: 'user-1'
    };

    api.evaluate(payload);

    expect(client.post).toHaveBeenCalledWith('/flags/evaluate', payload);
  });
});
