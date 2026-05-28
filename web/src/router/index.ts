import { createRouter, createWebHistory } from 'vue-router';

import { getStoredToken } from '../stores/auth';
import { useSessionStore } from '../stores/session';

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/app/teams' },
    { path: '/login', component: () => import('../views/LoginView.vue'), meta: { guest: true } },
    { path: '/register', component: () => import('../views/RegisterView.vue'), meta: { guest: true } },
    {
      path: '/app',
      component: () => import('../layouts/AppLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/app/teams' },
        { path: 'teams', component: () => import('../views/TeamsView.vue') },
        { path: 'teams/:id', component: () => import('../views/TeamDetailView.vue') },
        { path: 'products/:id', component: () => import('../views/ProductDetailView.vue') }
      ]
    }
  ]
});

router.beforeEach(async (to) => {
  const session = useSessionStore();
  const token = getStoredToken();

  if (!session.bootstrapped && token) {
    try {
      await session.loadCurrentUser();
    } catch {
      session.clearSession();
    }
  }

  if (to.meta.requiresAuth && !getStoredToken()) {
    return { path: '/login', query: { redirect: to.fullPath } };
  }

  if (to.meta.guest && getStoredToken()) {
    return '/app/teams';
  }

  return true;
});
