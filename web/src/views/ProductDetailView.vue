<script setup lang="ts">
import { ArrowLeft, CheckCircle, Plus, RotateCcw } from '@lucide/vue';
import {
  NButton,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NList,
  NListItem,
  NSpace,
  NSpin,
  NTag,
  useMessage
} from 'naive-ui';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { feedbackApi } from '../api/feedback';
import { productsApi } from '../api/products';
import type { Feedback, FeedbackStatus, Product } from '../api/types';

const route = useRoute();
const router = useRouter();
const message = useMessage();
const loading = ref(false);
const feedbackLoading = ref(false);
const feedbackSaving = ref(false);
const feedbackDrawerOpen = ref(false);
const selectedFeedback = ref<Feedback | null>(null);
const product = ref<Product | null>(null);
const feedbackItems = ref<Feedback[]>([]);
const productID = computed(() => Number(route.params.id));
const feedbackForm = reactive({
  title: '',
  content: ''
});
const canCreateFeedback = computed(
  () => feedbackForm.title.trim().length > 0 && feedbackForm.content.trim().length > 0
);

onMounted(loadProduct);
watch(() => route.params.id, loadProduct);

async function loadProduct() {
  if (!Number.isFinite(productID.value)) return;
  loading.value = true;
  try {
    product.value = await productsApi.get(productID.value);
    selectedFeedback.value = null;
    await loadFeedback();
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载产品详情失败');
  } finally {
    loading.value = false;
  }
}

async function loadFeedback() {
  if (!Number.isFinite(productID.value)) return;
  feedbackLoading.value = true;
  try {
    feedbackItems.value = await feedbackApi.listByProduct(productID.value);
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载产品反馈失败');
  } finally {
    feedbackLoading.value = false;
  }
}

async function createFeedback() {
  if (feedbackSaving.value) return;
  const title = feedbackForm.title.trim();
  const content = feedbackForm.content.trim();
  if (!title || !content) {
    message.warning('请填写反馈标题和内容');
    return;
  }

  feedbackSaving.value = true;
  try {
    const created = await feedbackApi.create(productID.value, { title, content });
    feedbackForm.title = '';
    feedbackForm.content = '';
    feedbackDrawerOpen.value = false;
    selectedFeedback.value = created;
    await loadFeedback();
  } catch (error) {
    message.error(error instanceof Error ? error.message : '创建反馈失败');
  } finally {
    feedbackSaving.value = false;
  }
}

async function setFeedbackStatus(status: FeedbackStatus) {
  if (!selectedFeedback.value || feedbackSaving.value) return;
  feedbackSaving.value = true;
  try {
    const updated = await feedbackApi.updateStatus(selectedFeedback.value.id, { status });
    selectedFeedback.value = updated;
    await loadFeedback();
  } catch (error) {
    message.error(error instanceof Error ? error.message : '更新反馈状态失败');
  } finally {
    feedbackSaving.value = false;
  }
}

function selectFeedback(feedback: Feedback) {
  selectedFeedback.value = feedback;
}

function closeFeedbackDetail(show: boolean) {
  if (!show) selectedFeedback.value = null;
}

function feedbackStatusType(status?: FeedbackStatus) {
  return status === 'resolved' ? 'success' : 'warning';
}

function feedbackStatusLabel(status?: FeedbackStatus) {
  return status === 'resolved' ? '已解决' : '待处理';
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

      <section class="content-panel feedback-panel">
        <div class="feedback-panel-header">
          <div>
            <h3>产品反馈</h3>
            <p>收集团队对这个产品的反馈，并跟踪处理状态。</p>
          </div>
          <n-button type="primary" @click="feedbackDrawerOpen = true">
            <template #icon>
              <n-icon><Plus /></n-icon>
            </template>
            新建反馈
          </n-button>
        </div>

        <n-spin :show="feedbackLoading">
          <n-list v-if="feedbackItems.length > 0" hoverable clickable>
            <n-list-item
              v-for="feedback in feedbackItems"
              :key="feedback.id"
              class="feedback-row"
              @click="selectFeedback(feedback)"
            >
              <div class="feedback-row-main">
                <div class="feedback-row-title">
                  <strong>{{ feedback.title }}</strong>
                  <n-tag :type="feedbackStatusType(feedback.status)" size="small" :bordered="false">
                    {{ feedbackStatusLabel(feedback.status) }}
                  </n-tag>
                </div>
                <p>{{ feedback.content }}</p>
              </div>
              <time class="feedback-date">{{ formatDate(feedback.created_at) }}</time>
            </n-list-item>
          </n-list>
          <div v-if="!feedbackLoading && feedbackItems.length === 0" class="empty-state">
            暂无产品反馈。
          </div>
        </n-spin>
      </section>
    </n-spin>

    <n-drawer v-model:show="feedbackDrawerOpen" :width="420" placement="right">
      <n-drawer-content title="新建反馈" closable>
        <n-form label-placement="top" @submit.prevent="createFeedback">
          <n-form-item label="标题">
            <n-input v-model:value="feedbackForm.title" placeholder="简要描述反馈主题" />
          </n-form-item>
          <n-form-item label="内容">
            <n-input
              v-model:value="feedbackForm.content"
              type="textarea"
              placeholder="说明问题、建议或背景信息"
              :autosize="{ minRows: 5, maxRows: 10 }"
            />
          </n-form-item>
          <n-space justify="end">
            <n-button @click="feedbackDrawerOpen = false">取消</n-button>
            <n-button
              type="primary"
              attr-type="submit"
              :disabled="!canCreateFeedback"
              :loading="feedbackSaving"
            >
              创建
            </n-button>
          </n-space>
        </n-form>
      </n-drawer-content>
    </n-drawer>

    <n-drawer
      :show="Boolean(selectedFeedback)"
      :width="460"
      placement="right"
      @update:show="closeFeedbackDetail"
    >
      <n-drawer-content title="反馈详情" closable>
        <div v-if="selectedFeedback" class="feedback-detail">
          <n-tag :type="feedbackStatusType(selectedFeedback.status)" :bordered="false">
            {{ feedbackStatusLabel(selectedFeedback.status) }}
          </n-tag>
          <h3>{{ selectedFeedback.title }}</h3>
          <p>{{ selectedFeedback.content }}</p>
          <time>{{ formatDate(selectedFeedback.created_at) }}</time>

          <n-space justify="end">
            <n-button
              v-if="selectedFeedback.status === 'open'"
              type="primary"
              :loading="feedbackSaving"
              @click="setFeedbackStatus('resolved')"
            >
              <template #icon>
                <n-icon><CheckCircle /></n-icon>
              </template>
              标记已解决
            </n-button>
            <n-button
              v-else
              :loading="feedbackSaving"
              @click="setFeedbackStatus('open')"
            >
              <template #icon>
                <n-icon><RotateCcw /></n-icon>
              </template>
              重新打开
            </n-button>
          </n-space>
        </div>
      </n-drawer-content>
    </n-drawer>
  </main>
</template>
