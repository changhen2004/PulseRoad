import { api, type ApiClient } from './http';
import type {
  CreateFeedbackCommentPayload,
  CreateFeedbackPayload,
  Feedback,
  FeedbackComment,
  FeedbackPage,
  FeedbackStatus,
  FeedbackVoteResult,
  UpdateFeedbackStatusPayload
} from './types';

export function createFeedbackApi(client: ApiClient = api) {
  return {
    create(productID: number, payload: CreateFeedbackPayload) {
      return client.post<Feedback>(`/products/${productID}/feedback`, payload);
    },
    listByProduct(productID: number, options: { page?: number; page_size?: number; status?: FeedbackStatus | '' } = {}) {
      const params = new URLSearchParams();
      if (options.page) params.set('page', String(options.page));
      if (options.page_size) params.set('page_size', String(options.page_size));
      if (options.status) params.set('status', options.status);
      const query = params.toString();
      return client.get<FeedbackPage>(`/products/${productID}/feedback${query ? `?${query}` : ''}`);
    },
    get(id: number) {
      return client.get<Feedback>(`/feedback/${id}`);
    },
    updateStatus(id: number, payload: UpdateFeedbackStatusPayload) {
      return client.patch<Feedback>(`/feedback/${id}/status`, payload);
    },
    createComment(id: number, payload: CreateFeedbackCommentPayload) {
      return client.post<FeedbackComment>(`/feedback/${id}/comments`, payload);
    },
    listComments(id: number) {
      return client.get<FeedbackComment[]>(`/feedback/${id}/comments`);
    },
    vote(id: number) {
      return client.post<FeedbackVoteResult>(`/feedback/${id}/vote`);
    },
    unvote(id: number) {
      return client.delete<FeedbackVoteResult>(`/feedback/${id}/vote`);
    }
  };
}

export const feedbackApi = createFeedbackApi();
