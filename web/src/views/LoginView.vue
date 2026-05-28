<script setup lang="ts">
import { Activity } from '@lucide/vue';
import { computed, reactive, ref } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import { NButton, NForm, NFormItem, NInput, useMessage } from 'naive-ui';

import { useSessionStore } from '../stores/session';

const router = useRouter();
const route = useRoute();
const message = useMessage();
const session = useSessionStore();
const loading = ref(false);

const form = reactive({
  email: '',
  password: ''
});

const canSubmit = computed(() => form.email.trim().includes('@') && form.password.length >= 8);

async function submit() {
  if (!canSubmit.value || loading.value) return;
  loading.value = true;
  try {
    await session.login({
      email: form.email.trim(),
      password: form.password
    });
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/app/teams';
    router.push(redirect);
  } catch (error) {
    message.error(error instanceof Error ? error.message : '登录失败');
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <main class="auth-shell">
    <section class="auth-panel">
      <div class="brand-row">
        <div class="brand-mark">
          <Activity :size="20" />
        </div>
        <div>
          <h1 class="brand-title">PulseRoad</h1>
          <p class="brand-subtitle">登录后进入团队工作台</p>
        </div>
      </div>

      <n-form label-placement="top" @submit.prevent="submit">
        <n-form-item label="邮箱">
          <n-input v-model:value="form.email" placeholder="you@example.com" autocomplete="email" />
        </n-form-item>
        <n-form-item label="密码">
          <n-input
            v-model:value="form.password"
            type="password"
            placeholder="至少 8 位"
            autocomplete="current-password"
            show-password-on="click"
          />
        </n-form-item>
        <div class="form-actions">
          <RouterLink to="/register">创建账号</RouterLink>
          <n-button type="primary" attr-type="submit" :disabled="!canSubmit" :loading="loading">
            登录
          </n-button>
        </div>
      </n-form>
    </section>
  </main>
</template>
