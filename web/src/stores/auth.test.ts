import { beforeEach, describe, expect, it } from 'vitest';

import { clearStoredToken, getStoredToken, setStoredToken } from './auth';

describe('auth token storage', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('persists and reads the current auth token', () => {
    setStoredToken('token-123');

    expect(getStoredToken()).toBe('token-123');
    expect(localStorage.getItem('pulseroad.auth.token')).toBe('token-123');
  });

  it('clears the current auth token', () => {
    setStoredToken('token-123');
    clearStoredToken();

    expect(getStoredToken()).toBeNull();
    expect(localStorage.getItem('pulseroad.auth.token')).toBeNull();
  });
});
