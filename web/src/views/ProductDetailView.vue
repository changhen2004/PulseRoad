<script setup lang="ts">
import { ArrowLeft } from '@lucide/vue';
import { NButton, NIcon, NSpin, useMessage } from 'naive-ui';
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { productsApi } from '../api/products';
import type { Product } from '../api/types';

const route = useRoute();
const router = useRouter();
const message = useMessage();
const loading = ref(false);
const product = ref<Product | null>(null);
const productID = computed(() => Number(route.params.id));

onMounted(loadProduct);
watch(() => route.params.id, loadProduct);

async function loadProduct() {
  if (!Number.isFinite(productID.value)) return;
  loading.value = true;
  try {
    product.value = await productsApi.get(productID.value);
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载产品详情失败');
  } finally {
    loading.value = false;
  }
}

function formatDate(value?: string) {
  if (!value) return '-';
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(new Date(value));
}
</script>

<template>
  <main class="page">
    <div class="toolbar">
      <n-button quaternary @click="router.push(`/app/teams/${product?.team_id || ''}`)">
        <template #icon>
          <n-icon><ArrowLeft /></n-icon>
        </template>
        返回团队
      </n-button>
    </div>

    <n-spin :show="loading">
      <div class="page-header">
        <div>
          <h2 class="page-title">{{ product?.name || '产品详情' }}</h2>
          <p class="page-description">{{ product?.description || '暂无产品描述。' }}</p>
        </div>
      </div>

      <section class="detail-grid">
        <div class="metric">
          <p class="metric-label">产品 ID</p>
          <p class="metric-value">{{ product?.id || '-' }}</p>
        </div>
        <div class="metric">
          <p class="metric-label">所属团队</p>
          <p class="metric-value">{{ product?.team_id || '-' }}</p>
        </div>
        <div class="metric">
          <p class="metric-label">创建时间</p>
          <p class="metric-value">{{ formatDate(product?.created_at) }}</p>
        </div>
      </section>

      <section class="content-panel product-summary">
        <h3>基础信息</h3>
        <dl>
          <div>
            <dt>名称</dt>
            <dd>{{ product?.name || '-' }}</dd>
          </div>
          <div>
            <dt>描述</dt>
            <dd>{{ product?.description || '-' }}</dd>
          </div>
          <div>
            <dt>创建者 ID</dt>
            <dd>{{ product?.created_by || '-' }}</dd>
          </div>
        </dl>
      </section>
    </n-spin>
  </main>
</template>
