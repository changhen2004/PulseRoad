<script setup lang="ts">
import { ArrowLeft, Plus } from '@lucide/vue';
import {
  NButton,
  NDataTable,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NSpace,
  type DataTableColumns,
  useMessage
} from 'naive-ui';
import { computed, h, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { productsApi } from '../api/products';
import { teamsApi } from '../api/teams';
import type { Product, Team } from '../api/types';

const route = useRoute();
const router = useRouter();
const message = useMessage();
const loading = ref(false);
const saving = ref(false);
const drawerOpen = ref(false);
const team = ref<Team | null>(null);
const products = ref<Product[]>([]);

const teamID = computed(() => Number(route.params.id));
const productForm = reactive({
  name: '',
  description: ''
});
const canCreateProduct = computed(() => productForm.name.trim().length > 0);

const columns: DataTableColumns<Product> = [
  {
    title: '产品',
    key: 'name',
    render(row) {
      return h('div', { class: 'table-title-cell' }, [
        h('strong', row.name),
        row.description ? h('span', row.description) : null
      ]);
    }
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 180,
    render(row) {
      return formatDate(row.created_at);
    }
  },
  {
    title: '',
    key: 'actions',
    width: 110,
    align: 'right',
    render(row) {
      return h(
        NButton,
        {
          size: 'small',
          onClick: () => router.push(`/app/products/${row.id}`)
        },
        () => '查看'
      );
    }
  }
];

onMounted(loadTeam);
watch(() => route.params.id, loadTeam);

async function loadTeam() {
  if (!Number.isFinite(teamID.value)) return;
  loading.value = true;
  try {
    const [teamResult, productResult] = await Promise.all([
      teamsApi.get(teamID.value),
      productsApi.listByTeam(teamID.value)
    ]);
    team.value = teamResult;
    products.value = productResult;
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载团队详情失败');
  } finally {
    loading.value = false;
  }
}

async function createProduct() {
  if (!canCreateProduct.value || saving.value) return;
  saving.value = true;
  try {
    const product = await productsApi.create(teamID.value, {
      name: productForm.name.trim(),
      description: productForm.description.trim()
    });
    drawerOpen.value = false;
    productForm.name = '';
    productForm.description = '';
    await loadTeam();
    router.push(`/app/products/${product.id}`);
  } catch (error) {
    message.error(error instanceof Error ? error.message : '创建产品失败');
  } finally {
    saving.value = false;
  }
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  }).format(new Date(value));
}

function productRowKey(row: Product) {
  return row.id;
}
</script>

<template>
  <main class="page">
    <div class="toolbar">
      <n-button quaternary @click="router.push('/app/teams')">
        <template #icon>
          <n-icon><ArrowLeft /></n-icon>
        </template>
        返回团队
      </n-button>
    </div>

    <div class="page-header">
      <div>
        <h2 class="page-title">{{ team?.name || '团队详情' }}</h2>
        <p class="page-description">{{ team?.description || '团队下的产品统一在这里管理。' }}</p>
      </div>
      <n-button type="primary" @click="drawerOpen = true">
        <template #icon>
          <n-icon><Plus /></n-icon>
        </template>
        新建产品
      </n-button>
    </div>

    <section class="detail-grid">
      <div class="metric">
        <p class="metric-label">团队 ID</p>
        <p class="metric-value">{{ team?.id || '-' }}</p>
      </div>
      <div class="metric">
        <p class="metric-label">我的角色</p>
        <p class="metric-value">{{ team?.role || '-' }}</p>
      </div>
      <div class="metric">
        <p class="metric-label">产品数量</p>
        <p class="metric-value">{{ products.length }}</p>
      </div>
    </section>

    <section class="content-panel">
      <n-data-table
        :columns="columns"
        :data="products"
        :loading="loading"
        :bordered="false"
        :single-line="false"
        :row-key="productRowKey"
      />
      <div v-if="!loading && products.length === 0" class="empty-state">
        这个团队还没有产品。
      </div>
    </section>

    <n-drawer v-model:show="drawerOpen" :width="420" placement="right">
      <n-drawer-content title="新建产品" closable>
        <n-form label-placement="top" @submit.prevent="createProduct">
          <n-form-item label="产品名称">
            <n-input v-model:value="productForm.name" placeholder="例如：PulseRoad Console" />
          </n-form-item>
          <n-form-item label="描述">
            <n-input
              v-model:value="productForm.description"
              type="textarea"
              placeholder="产品用途、目标用户或当前阶段"
              :autosize="{ minRows: 4, maxRows: 8 }"
            />
          </n-form-item>
          <n-space justify="end">
            <n-button @click="drawerOpen = false">取消</n-button>
            <n-button
              type="primary"
              attr-type="submit"
              :disabled="!canCreateProduct"
              :loading="saving"
            >
              创建
            </n-button>
          </n-space>
        </n-form>
      </n-drawer-content>
    </n-drawer>
  </main>
</template>
