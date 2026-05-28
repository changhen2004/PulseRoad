import { api, type ApiClient } from './http';
import type { CreateFeedbackPayload, Feedback, UpdateFeedbackStatusPayload } from './types';

export function createFeedbackApi(client: ApiClient = api) {
  return {
    create(productID: number, payload: CreateFeedbackPayload) {
      return client.post<Feedback>(`/products/${productID}/feedback`, payload);
    },
    listByProduct(productID: number) {
      return client.get<Feedback[]>(`/products/${productID}/feedback`);
    },
    get(id: number) {
      return client.get<Feedback>(`/feedback/${id}`);
    },
    updateStatus(id: number, payload: UpdateFeedbackStatusPayload) {
      return client.patch<Feedback>(`/feedback/${id}/status`, payload);
    }
  };
}

export const feedbackApi = createFeedbackApi();
