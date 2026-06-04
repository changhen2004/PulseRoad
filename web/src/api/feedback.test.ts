import { describe, expect, it, vi } from 'vitest';

import { createFeedbackApi } from './feedback';
import type { ApiClient } from './http';

function fakeApiClient(): ApiClient {
  return {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn()
  };
}

describe('feedback api', () => {
  it('creates feedback for a product', () => {
    const client = fakeApiClient();
    const api = createFeedbackApi(client);
    const payload = { title: 'Slow search', content: 'Search takes too long' };

    api.create(7, payload);

    expect(client.post).toHaveBeenCalledWith('/products/7/feedback', payload);
  });

  it('lists feedback by product', () => {
    const client = fakeApiClient();
    const api = createFeedbackApi(client);

    api.listByProduct(7);

    expect(client.get).toHaveBeenCalledWith('/products/7/feedback');
  });

  it('lists feedback with filters and pagination', () => {
    const client = fakeApiClient();
    const api = createFeedbackApi(client);

    api.listByProduct(7, { status: 'open', page: 2, page_size: 10 });

    expect(client.get).toHaveBeenCalledWith('/products/7/feedback?page=2&page_size=10&status=open');
  });

  it('gets feedback by id', () => {
    const client = fakeApiClient();
    const api = createFeedbackApi(client);

    api.get(11);

    expect(client.get).toHaveBeenCalledWith('/feedback/11');
  });

  it('updates feedback status', () => {
    const client = fakeApiClient();
    const api = createFeedbackApi(client);
    const payload = { status: 'resolved' as const };

    api.updateStatus(11, payload);

    expect(client.patch).toHaveBeenCalledWith('/feedback/11/status', payload);
  });

  it('creates comments and toggles votes', () => {
    const client = fakeApiClient();
    const api = createFeedbackApi(client);

    api.createComment(11, { content: 'I need this too' });
    api.vote(11);
    api.unvote(11);

    expect(client.post).toHaveBeenCalledWith('/feedback/11/comments', { content: 'I need this too' });
    expect(client.post).toHaveBeenCalledWith('/feedback/11/vote');
    expect(client.delete).toHaveBeenCalledWith('/feedback/11/vote');
  });
});
