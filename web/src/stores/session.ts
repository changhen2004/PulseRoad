import { defineStore } from 'pinia';

import { authApi } from '../api/auth';
import type { LoginPayload, RegisterPayload, User } from '../api/types';
import { clearStoredToken, getStoredToken, setStoredToken } from './auth';

interface SessionState {
  user: User | null;
  bootstrapped: boolean;
}

export const useSessionStore = defineStore('session', {
  state: (): SessionState => ({
    user: null,
    bootstrapped: false
  }),
  getters: {
    isAuthenticated: () => Boolean(getStoredToken())
  },
  actions: {
    async login(payload: LoginPayload) {
      const result = await authApi.login(payload);
      setStoredToken(result.token);
      this.user = result.user;
      this.bootstrapped = true;
    },
    async register(payload: RegisterPayload) {
      await authApi.register(payload);
      await this.login({ email: payload.email, password: payload.password });
    },
    async loadCurrentUser() {
      if (!getStoredToken()) {
        this.clearSession();
        return;
      }
      this.user = await authApi.currentUser();
      this.bootstrapped = true;
    },
    clearSession() {
      clearStoredToken();
      this.user = null;
      this.bootstrapped = true;
    }
  }
});
