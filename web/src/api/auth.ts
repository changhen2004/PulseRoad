import { api } from './http';
import type { AuthResult, LoginPayload, RegisterPayload, User } from './types';

export const authApi = {
  register(payload: RegisterPayload) {
    return api.post<User>('/auth/register', payload);
  },
  login(payload: LoginPayload) {
    return api.post<AuthResult>('/auth/login', payload);
  },
  currentUser() {
    return api.get<User>('/auth/me');
  }
};
