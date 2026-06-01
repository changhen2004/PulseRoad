import { api, type ApiClient } from './http';
import type {
  CreateFeatureFlagPayload,
  EvaluateFeatureFlagPayload,
  EvaluateFeatureFlagResult,
  FeatureFlag,
  ToggleFeatureFlagPayload,
  UpdateFeatureFlagPayload
} from './types';

export function createFlagflowApi(client: ApiClient = api) {
  return {
    create(productID: number, payload: CreateFeatureFlagPayload) {
      return client.post<FeatureFlag>(`/products/${productID}/flags`, payload);
    },
    listByProduct(productID: number, environment?: string) {
      const query = environment ? `?environment=${encodeURIComponent(environment)}` : '';
      return client.get<FeatureFlag[]>(`/products/${productID}/flags${query}`);
    },
    get(id: number) {
      return client.get<FeatureFlag>(`/flags/${id}`);
    },
    update(id: number, payload: UpdateFeatureFlagPayload) {
      return client.patch<FeatureFlag>(`/flags/${id}`, payload);
    },
    toggle(id: number, payload: ToggleFeatureFlagPayload) {
      return client.patch<FeatureFlag>(`/flags/${id}/toggle`, payload);
    },
    evaluate(payload: EvaluateFeatureFlagPayload) {
      return client.post<EvaluateFeatureFlagResult>('/flags/evaluate', payload);
    }
  };
}

export const flagflowApi = createFlagflowApi();
