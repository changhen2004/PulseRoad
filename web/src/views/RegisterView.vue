<script setup lang="ts">
import { Activity } from '@lucide/vue';
import { computed, reactive, ref } from 'vue';
import { RouterLink, useRouter } from 'vue-router';
import { NButton, NForm, NFormItem, NInput, useMessage } from 'naive-ui';

import { useSessionStore } from '../stores/session';

const router = useRouter();
const message = useMessage();
const session = useSessionStore();
const loading = ref(false);

const form = reactive({
  name: '',
  email: '',
  password: ''
});

const canSubmit = computed(
  () => form.name.trim().length > 0 && form.email.trim().includes('@') && form.password.length >= 8
);

async function submit() {
  if (!canSubmit.value || loading.value) return;
  loading.value = true;
  try {
    await session.register({
      name: form.name.trim(),
      email: form.email.trim(),
      password: form.password
    });
    router.push('/app/teams');
  } catch (error) {
    message.error(error instanceof Error ? error.message : '注册失败');
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
          <p class="brand-subtitle">创建账号并开始管理产品</p>
        </div>
      </div>

      <n-form label-placement="top" @submit.prevent="submit">
        <n-form-item label="姓名">
          <n-input v-model:value="form.name" placeholder="你的名字" autocomplete="name" />
        </n-form-item>
        <n-form-item label="邮箱">
          <n-input v-model:value="form.email" placeholder="you@example.com" autocomplete="email" />
        </n-form-item>
        <n-form-item label="密码">
          <n-input
            v-model:value="form.password"
            type="password"
            placeholder="至少 8 位"
            autocomplete="new-password"
            show-password-on="click"
          />
        </n-form-item>
        <div class="form-actions">
          <RouterLink to="/login">已有账号</RouterLink>
          <n-button type="primary" attr-type="submit" :disabled="!canSubmit" :loading="loading">
            注册
          </n-button>
        </div>
      </n-form>
    </section>
  </main>
</template>
