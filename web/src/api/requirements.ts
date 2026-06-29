import { api, type ApiClient } from './http';
import type {
  CreateRequirementPayload,
  Requirement,
  RequirementPage,
  UpdateRequirementPayload
} from './types';

export function createRequirementApi(client: ApiClient = api) {
  return {
    create(productID: number, payload: CreateRequirementPayload) {
      return client.post<Requirement>(`/products/${productID}/requirements`, payload);
    },
    listByProduct(productID: number, options: { page?: number; page_size?: number; status?: string } = {}) {
      const params = new URLSearchParams();
      if (options.page) params.set('page', String(options.page));
      if (options.page_size) params.set('page_size', String(options.page_size));
      if (options.status) params.set('status', options.status);
      const query = params.toString();
      return client.get<RequirementPage>(`/products/${productID}/requirements${query ? `?${query}` : ''}`);
    },
    get(id: number) {
      return client.get<Requirement>(`/requirements/${id}`);
    },
    update(id: number, payload: UpdateRequirementPayload) {
      return client.patch<Requirement>(`/requirements/${id}`, payload);
    },
    delete(id: number) {
      return client.delete<void>(`/requirements/${id}`);
    }
  };
}

export const requirementApi = createRequirementApi();
