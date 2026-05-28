import { api } from './http';
import type { CreateProductPayload, Product } from './types';

export const productsApi = {
  create(teamID: number, payload: CreateProductPayload) {
    return api.post<Product>(`/teams/${teamID}/products`, payload);
  },
  listByTeam(teamID: number) {
    return api.get<Product[]>(`/teams/${teamID}/products`);
  },
  get(id: number) {
    return api.get<Product>(`/products/${id}`);
  }
};
