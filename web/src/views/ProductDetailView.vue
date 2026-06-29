<script setup lang="ts">
import { ArrowLeft, CheckCircle, Plus, RotateCcw } from '@lucide/vue';
import {
  NButton,
  NDrawer,
  NDrawerContent,
  NDropdown,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NInputNumber,
  NList,
  NListItem,
  NSelect,
  NSpace,
  NSpin,
  NTag,
  useMessage
} from 'naive-ui';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { feedbackApi } from '../api/feedback';
import { flagflowApi } from '../api/flagflow';
import { requirementApi } from '../api/requirements';
import { productsApi } from '../api/products';
import type {
  FeatureFlag,
  Feedback,
  FeedbackComment,
  FeedbackStatus,
  Product,
  ProductSummary,
  Requirement,
  RequirementPriority,
  RequirementStatus,
  CreateRequirementPayload,
  UpdateRequirementPayload
} from '../api/types';

const route = useRoute();
const router = useRouter();
const message = useMessage();
const loading = ref(false);
const feedbackLoading = ref(false);
const feedbackSaving = ref(false);
const feedbackDrawerOpen = ref(false);
const selectedFeedback = ref<Feedback | null>(null);
const feedbackComments = ref<FeedbackComment[]>([]);
const commentSaving = ref(false);
const flagLoading = ref(false);
const flagSaving = ref(false);
const flagDrawerOpen = ref(false);
const product = ref<Product | null>(null);
const productSummary = ref<ProductSummary | null>(null);
const feedbackItems = ref<Feedback[]>([]);
const feedbackStatusFilter = ref<FeedbackStatus | ''>('');
const feedbackPage = ref(1);
const feedbackPageSize = 10;
const feedbackTotal = ref(0);
const flagItems = ref<FeatureFlag[]>([]);
const productID = computed(() => Number(route.params.id));
let productLoadVersion = 0;
let feedbackLoadVersion = 0;
let flagLoadVersion = 0;
const feedbackForm = reactive({
  title: '',
  content: ''
});
const commentForm = reactive({
  content: ''
});
const flagForm = reactive({
  key: '',
  name: '',
  description: '',
  environment: 'production',
  rolloutPercentage: 0
});
const requirementLoading = ref(false);
const requirementSaving = ref(false);
const requirementItems = ref<Requirement[]>([]);
const requirementDrawerOpen = ref(false);
const requirementStatusFilter = ref<RequirementStatus | ''>('');
const requirementPage = ref(1);
const requirementPageSize = 10;
const requirementTotal = ref(0);
let requirementLoadVersion = 0;
const requirementForm = reactive({
  title: '',
  description: '',
  priority: 'p2' as RequirementPriority,
  sourceFeedbackId: null as number | null
});
const canCreateRequirement = computed(
  () => requirementForm.title.trim().length > 0
);
const requirementTotalPages = computed(() =>
  Math.max(1, Math.ceil(requirementTotal.value / requirementPageSize))
);
type LoadFeedbackOptions = {
  preserveFeedback?: Feedback;
  clearOnError?: boolean;
};
type LoadFlagOptions = {
  preserveFlag?: FeatureFlag;
  clearOnError?: boolean;
};
const canCreateFeedback = computed(
  () => feedbackForm.title.trim().length > 0 && feedbackForm.content.trim().length > 0
);
const canCreateComment = computed(() => commentForm.content.trim().length > 0);
const feedbackTotalPages = computed(() =>
  Math.max(1, Math.ceil(feedbackTotal.value / feedbackPageSize))
);
const canCreateFlag = computed(() => {
  const rollout = Number(flagForm.rolloutPercentage);
  return (
    flagForm.key.trim().length > 0 &&
    flagForm.name.trim().length > 0 &&
    flagForm.environment.trim().length > 0 &&
    Number.isFinite(rollout) &&
    rollout >= 0 &&
    rollout <= 100
  );
});

onMounted(loadProduct);
watch(() => route.params.id, loadProduct);

async function loadProduct() {
  const requestedProductID = productID.value;
  const requestVersion = ++productLoadVersion;
  selectedFeedback.value = null;
  feedbackComments.value = [];
  product.value = null;
  productSummary.value = null;
  feedbackItems.value = [];
  flagItems.value = [];
  requirementItems.value = [];
  feedbackLoadVersion++;
  flagLoadVersion++;
  requirementLoadVersion++;
  feedbackLoading.value = false;
  flagLoading.value = false;

  if (!Number.isFinite(requestedProductID)) return;
  loading.value = true;
  try {
    const [loadedProduct, loadedSummary] = await Promise.all([
      productsApi.get(requestedProductID),
      productsApi.summary(requestedProductID)
    ]);
    if (!isCurrentProductRequest(requestedProductID, requestVersion)) return;
    product.value = loadedProduct;
    productSummary.value = loadedSummary;
    await Promise.all([
      loadFeedback(requestedProductID, { clearOnError: true }),
      loadFlags(requestedProductID, { clearOnError: true }),
      loadRequirements(requestedProductID)
    ]);
  } catch (error) {
    if (!isCurrentProductRequest(requestedProductID, requestVersion)) return;
    product.value = null;
    productSummary.value = null;
    feedbackItems.value = [];
    flagItems.value = [];
    requirementItems.value = [];
    message.error(error instanceof Error ? error.message : '加载产品详情失败');
  } finally {
    if (isCurrentProductRequest(requestedProductID, requestVersion)) {
      loading.value = false;
    }
  }
}

async function loadFeedback(requestedProductID: number, options: LoadFeedbackOptions = {}) {
  if (!Number.isFinite(requestedProductID)) return;
  const requestVersion = ++feedbackLoadVersion;
  feedbackLoading.value = true;
  try {
    const page = await feedbackApi.listByProduct(requestedProductID, {
      page: feedbackPage.value,
      page_size: feedbackPageSize,
      status: feedbackStatusFilter.value
    });
    if (!isCurrentFeedbackRequest(requestedProductID, requestVersion)) return;
    feedbackTotal.value = page.total;
    feedbackItems.value = options.preserveFeedback
      ? preserveFeedbackItem(page.items, options.preserveFeedback)
      : page.items;
  } catch (error) {
    if (!isCurrentFeedbackRequest(requestedProductID, requestVersion)) return;
    if (options.preserveFeedback) {
      feedbackItems.value = preserveFeedbackItem(feedbackItems.value, options.preserveFeedback);
    } else if (options.clearOnError ?? true) {
      feedbackItems.value = [];
      feedbackTotal.value = 0;
    }
    message.error(error instanceof Error ? error.message : '加载产品反馈失败');
  } finally {
    if (isCurrentFeedbackRequest(requestedProductID, requestVersion)) {
      feedbackLoading.value = false;
    }
  }
}

async function changeFeedbackFilter(status: FeedbackStatus | '') {
  feedbackStatusFilter.value = status;
  feedbackPage.value = 1;
  await loadFeedback(productID.value, { clearOnError: true });
}

async function changeFeedbackPage(page: number) {
  feedbackPage.value = Math.min(Math.max(1, page), feedbackTotalPages.value);
  await loadFeedback(productID.value, { clearOnError: true });
}

async function loadFlags(requestedProductID: number, options: LoadFlagOptions = {}) {
  if (!Number.isFinite(requestedProductID)) return;
  const requestVersion = ++flagLoadVersion;
  flagLoading.value = true;
  try {
    const items = await flagflowApi.listByProduct(requestedProductID);
    if (!isCurrentFlagRequest(requestedProductID, requestVersion)) return;
    flagItems.value = options.preserveFlag ? preserveFlagItem(items, options.preserveFlag) : items;
  } catch (error) {
    if (!isCurrentFlagRequest(requestedProductID, requestVersion)) return;
    if (options.preserveFlag) {
      flagItems.value = preserveFlagItem(flagItems.value, options.preserveFlag);
    } else if (options.clearOnError ?? true) {
      flagItems.value = [];
    }
    message.error(error instanceof Error ? error.message : '加载功能开关失败');
  } finally {
    if (isCurrentFlagRequest(requestedProductID, requestVersion)) {
      flagLoading.value = false;
    }
  }
}

async function createFeedback() {
  if (feedbackSaving.value) return;
  const requestedProductID = productID.value;
  const title = feedbackForm.title.trim();
  const content = feedbackForm.content.trim();
  if (!Number.isFinite(requestedProductID) || !title || !content) {
    message.warning('请填写反馈标题和内容');
    return;
  }

  feedbackSaving.value = true;
  try {
    const created = await feedbackApi.create(requestedProductID, { title, content });
    if (requestedProductID !== productID.value) return;
    feedbackForm.title = '';
    feedbackForm.content = '';
    feedbackDrawerOpen.value = false;
    selectedFeedback.value = created;
    await reloadSummary(requestedProductID);
    await loadFeedback(requestedProductID, { preserveFeedback: created, clearOnError: false });
  } catch (error) {
    if (requestedProductID !== productID.value) return;
    message.error(error instanceof Error ? error.message : '创建反馈失败');
  } finally {
    feedbackSaving.value = false;
  }
}

async function reloadSummary(requestedProductID: number) {
  try {
    productSummary.value = await productsApi.summary(requestedProductID);
  } catch {
    // 摘要失败不阻塞主流程。
  }
}

async function loadComments(feedbackID: number) {
  try {
    feedbackComments.value = await feedbackApi.listComments(feedbackID);
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载评论失败');
  }
}

async function createFlag() {
  if (flagSaving.value) return;
  const requestedProductID = productID.value;
  const key = flagForm.key.trim();
  const name = flagForm.name.trim();
  const environment = flagForm.environment.trim();
  const rolloutPercentage = Number(flagForm.rolloutPercentage);
  if (!Number.isFinite(requestedProductID) || !canCreateFlag.value) {
    message.warning('请填写开关键、名称和发布比例');
    return;
  }

  flagSaving.value = true;
  try {
    const created = await flagflowApi.create(requestedProductID, {
      key,
      name,
      description: flagForm.description.trim(),
      environment,
      rollout_percentage: rolloutPercentage
    });
    if (requestedProductID !== productID.value) return;
    flagForm.key = '';
    flagForm.name = '';
    flagForm.description = '';
    flagForm.environment = 'production';
    flagForm.rolloutPercentage = 0;
    flagDrawerOpen.value = false;
    await loadFlags(requestedProductID, { preserveFlag: created, clearOnError: false });
  } catch (error) {
    if (requestedProductID !== productID.value) return;
    message.error(error instanceof Error ? error.message : '创建功能开关失败');
  } finally {
    flagSaving.value = false;
  }
}

async function toggleFlag(flag: FeatureFlag) {
  if (flagSaving.value) return;
  const requestedProductID = productID.value;
  flagSaving.value = true;
  try {
    const updated = await flagflowApi.toggle(flag.id, { enabled: !flag.enabled });
    if (requestedProductID !== productID.value) return;
    replaceFlagItem(updated);
    await loadFlags(requestedProductID, { preserveFlag: updated, clearOnError: false });
  } catch (error) {
    if (requestedProductID !== productID.value) return;
    message.error(error instanceof Error ? error.message : '更新功能开关失败');
  } finally {
    flagSaving.value = false;
  }
}

async function setFeedbackStatus(status: FeedbackStatus) {
  if (!selectedFeedback.value || feedbackSaving.value) return;
  const requestedProductID = productID.value;
  feedbackSaving.value = true;
  try {
    const updated = await feedbackApi.updateStatus(selectedFeedback.value.id, { status });
    if (requestedProductID !== productID.value) return;
    replaceFeedbackItem(updated);
    selectedFeedback.value = updated;
    await reloadSummary(requestedProductID);
    await loadFeedback(requestedProductID, { preserveFeedback: updated, clearOnError: false });
  } catch (error) {
    if (requestedProductID !== productID.value) return;
    message.error(error instanceof Error ? error.message : '更新反馈状态失败');
  } finally {
    feedbackSaving.value = false;
  }
}

async function toggleFeedbackVote(feedback: Feedback) {
  if (feedbackSaving.value) return;
  const requestedProductID = productID.value;
  feedbackSaving.value = true;
  try {
    const result = feedback.voted
      ? await feedbackApi.unvote(feedback.id)
      : await feedbackApi.vote(feedback.id);
    const updated = { ...feedback, voted: result.voted, vote_count: result.vote_count };
    replaceFeedbackItem(updated);
    if (selectedFeedback.value?.id === feedback.id) selectedFeedback.value = updated;
    await reloadSummary(requestedProductID);
  } catch (error) {
    message.error(error instanceof Error ? error.message : '更新投票失败');
  } finally {
    feedbackSaving.value = false;
  }
}

async function createComment() {
  if (!selectedFeedback.value || commentSaving.value) return;
  const content = commentForm.content.trim();
  if (!content) {
    message.warning('请填写评论内容');
    return;
  }
  commentSaving.value = true;
  try {
    const comment = await feedbackApi.createComment(selectedFeedback.value.id, { content });
    feedbackComments.value = [...feedbackComments.value, comment];
    commentForm.content = '';
    const updated = {
      ...selectedFeedback.value,
      comment_count: selectedFeedback.value.comment_count + 1
    };
    selectedFeedback.value = updated;
    replaceFeedbackItem(updated);
    await reloadSummary(productID.value);
  } catch (error) {
    message.error(error instanceof Error ? error.message : '创建评论失败');
  } finally {
    commentSaving.value = false;
  }
}

function selectFeedback(feedback: Feedback) {
  selectedFeedback.value = feedback;
  feedbackComments.value = [];
  void loadComments(feedback.id);
}

function replaceFeedbackItem(updated: Feedback) {
  feedbackItems.value = feedbackItems.value.map((feedback) =>
    feedback.id === updated.id ? updated : feedback
  );
}

function replaceFlagItem(updated: FeatureFlag) {
  flagItems.value = flagItems.value.map((flag) => (flag.id === updated.id ? updated : flag));
}

function preserveFlagItem(items: FeatureFlag[], preserved: FeatureFlag) {
  const hasPreservedItem = items.some((flag) => flag.id === preserved.id);
  if (!hasPreservedItem) return [preserved, ...items];
  return items.map((flag) => (flag.id === preserved.id ? preserved : flag));
}

function preserveFeedbackItem(items: Feedback[], preserved: Feedback) {
  const hasPreservedItem = items.some((feedback) => feedback.id === preserved.id);
  if (!hasPreservedItem) return [preserved, ...items];
  return items.map((feedback) => (feedback.id === preserved.id ? preserved : feedback));
}

function isCurrentProductRequest(requestedProductID: number, requestVersion: number) {
  return requestedProductID === productID.value && requestVersion === productLoadVersion;
}

function isCurrentFeedbackRequest(requestedProductID: number, requestVersion: number) {
  return requestedProductID === productID.value && requestVersion === feedbackLoadVersion;
}

function isCurrentFlagRequest(requestedProductID: number, requestVersion: number) {
  return requestedProductID === productID.value && requestVersion === flagLoadVersion;
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

function flagStatusType(enabled: boolean) {
  return enabled ? 'success' : 'default';
}

function flagStatusLabel(enabled: boolean) {
  return enabled ? '已开启' : '已关闭';
}

async function loadRequirements(requestedProductID: number) {
  if (!Number.isFinite(requestedProductID)) return;
  const requestVersion = ++requirementLoadVersion;
  requirementLoading.value = true;
  try {
    const page = await requirementApi.listByProduct(requestedProductID, {
      page: requirementPage.value,
      page_size: requirementPageSize,
      status: requirementStatusFilter.value
    });
    if (!isCurrentRequirementRequest(requestedProductID, requestVersion)) return;
    requirementTotal.value = page.total;
    requirementItems.value = page.items;
  } catch (error) {
    if (!isCurrentRequirementRequest(requestedProductID, requestVersion)) return;
    requirementItems.value = [];
    requirementTotal.value = 0;
    message.error(error instanceof Error ? error.message : '加载需求失败');
  } finally {
    if (isCurrentRequirementRequest(requestedProductID, requestVersion)) {
      requirementLoading.value = false;
    }
  }
}

async function changeRequirementFilter(status: RequirementStatus | '') {
  requirementStatusFilter.value = status;
  requirementPage.value = 1;
  await loadRequirements(productID.value);
}

async function changeRequirementPage(page: number) {
  requirementPage.value = Math.min(Math.max(1, page), requirementTotalPages.value);
  await loadRequirements(productID.value);
}

async function createRequirement() {
  if (requirementSaving.value) return;
  const requestedProductID = productID.value;
  const title = requirementForm.title.trim();
  if (!Number.isFinite(requestedProductID) || !title) {
    message.warning('请填写需求标题');
    return;
  }
  requirementSaving.value = true;
  try {
    const payload: CreateRequirementPayload = {
      title,
      description: requirementForm.description.trim(),
      priority: requirementForm.priority
    };
    if (requirementForm.sourceFeedbackId) {
      payload.source_feedback_id = requirementForm.sourceFeedbackId;
    }
    await requirementApi.create(requestedProductID, payload);
    if (requestedProductID !== productID.value) return;
    requirementForm.title = '';
    requirementForm.description = '';
    requirementForm.priority = 'p2';
    requirementForm.sourceFeedbackId = null;
    requirementDrawerOpen.value = false;
    requirementPage.value = 1;
    await loadRequirements(requestedProductID);
  } catch (error) {
    if (requestedProductID !== productID.value) return;
    message.error(error instanceof Error ? error.message : '创建需求失败');
  } finally {
    requirementSaving.value = false;
  }
}

async function updateRequirementStatus(req: Requirement, status: RequirementStatus) {
  if (requirementSaving.value) return;
  const requestedProductID = productID.value;
  requirementSaving.value = true;
  try {
    const payload: UpdateRequirementPayload = {
      title: req.title,
      description: req.description,
      status,
      priority: req.priority
    };
    await requirementApi.update(req.id, payload);
    if (requestedProductID !== productID.value) return;
    await loadRequirements(requestedProductID);
  } catch (error) {
    if (requestedProductID !== productID.value) return;
    message.error(error instanceof Error ? error.message : '更新需求状态失败');
  } finally {
    requirementSaving.value = false;
  }
}

async function deleteRequirement(id: number) {
  if (requirementSaving.value) return;
  const requestedProductID = productID.value;
  requirementSaving.value = true;
  try {
    await requirementApi.delete(id);
    if (requestedProductID !== productID.value) return;
    await loadRequirements(requestedProductID);
  } catch (error) {
    if (requestedProductID !== productID.value) return;
    message.error(error instanceof Error ? error.message : '删除需求失败');
  } finally {
    requirementSaving.value = false;
  }
}

function isCurrentRequirementRequest(requestedProductID: number, requestVersion: number) {
  return requestedProductID === productID.value && requestVersion === requirementLoadVersion;
}

function requirementStatusType(status: RequirementStatus) {
  if (status === 'released') return 'success';
  if (status === 'in_progress') return 'info';
  return 'default';
}

function requirementStatusLabel(status: RequirementStatus) {
  if (status === 'in_progress') return '开发中';
  if (status === 'released') return '已上线';
  return '待规划';
}

function requirementPriorityType(priority: RequirementPriority) {
  if (priority === 'p0') return 'error';
  if (priority === 'p1') return 'warning';
  if (priority === 'p2') return 'info';
  return 'default';
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

      <section class="detail-grid summary-grid">
        <div class="metric">
          <p class="metric-label">反馈总数</p>
          <p class="metric-value">{{ productSummary?.feedback_total ?? 0 }}</p>
        </div>
        <div class="metric">
          <p class="metric-label">待处理 / 已解决</p>
          <p class="metric-value">
            {{ productSummary?.feedback_open ?? 0 }} / {{ productSummary?.feedback_resolved ?? 0 }}
          </p>
        </div>
        <div class="metric">
          <p class="metric-label">评论 / 投票</p>
          <p class="metric-value">
            {{ productSummary?.comment_total ?? 0 }} / {{ productSummary?.vote_total ?? 0 }}
          </p>
        </div>
        <div class="metric">
          <p class="metric-label">功能开关</p>
          <p class="metric-value">
            {{ productSummary?.flag_enabled ?? 0 }} / {{ productSummary?.flag_total ?? 0 }}
          </p>
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

      <section class="content-panel flag-panel">
        <div class="flag-panel-header">
          <div>
            <h3>功能开关</h3>
            <p>按产品维护灰度开关，支持环境隔离、启停和百分比发布。</p>
          </div>
          <n-button type="primary" @click="flagDrawerOpen = true">
            <template #icon>
              <n-icon><Plus /></n-icon>
            </template>
            新建开关
          </n-button>
        </div>

        <n-spin :show="flagLoading">
          <n-list v-if="flagItems.length > 0" hoverable :bordered="false">
            <n-list-item v-for="flag in flagItems" :key="flag.id" class="flag-row">
              <div class="flag-row-main">
                <div class="flag-row-title">
                  <strong>{{ flag.name }}</strong>
                  <code>{{ flag.key }}</code>
                </div>
                <p>{{ flag.description || '暂无说明' }}</p>
                <div class="flag-meta">
                  <n-tag size="small" :bordered="false">{{ flag.environment }}</n-tag>
                  <n-tag :type="flagStatusType(flag.enabled)" size="small" :bordered="false">
                    {{ flagStatusLabel(flag.enabled) }}
                  </n-tag>
                  <span>{{ flag.rollout_percentage }}% 发布</span>
                </div>
              </div>
              <n-button
                size="small"
                :type="flag.enabled ? 'warning' : 'primary'"
                :loading="flagSaving"
                @click="toggleFlag(flag)"
              >
                {{ flag.enabled ? '关闭' : '开启' }}
              </n-button>
            </n-list-item>
          </n-list>
          <div v-if="!flagLoading && flagItems.length === 0" class="empty-state">
            暂无功能开关。
          </div>
        </n-spin>
      </section>

      <section class="content-panel flag-panel">
        <div class="flag-panel-header">
          <div>
            <h3>需求</h3>
            <p>管理产品需求规划，跟踪开发状态。</p>
          </div>
          <n-button type="primary" @click="requirementDrawerOpen = true">
            <template #icon>
              <n-icon><Plus /></n-icon>
            </template>
            新建需求
          </n-button>
        </div>

        <n-spin :show="requirementLoading">
          <div class="panel-actions" style="padding: 0 22px; margin-top: 12px;">
            <n-space>
              <n-button
                size="small"
                :type="requirementStatusFilter === '' ? 'primary' : 'default'"
                @click="changeRequirementFilter('')"
              >
                全部
              </n-button>
              <n-button
                size="small"
                :type="requirementStatusFilter === 'planned' ? 'primary' : 'default'"
                @click="changeRequirementFilter('planned')"
              >
                待规划
              </n-button>
              <n-button
                size="small"
                :type="requirementStatusFilter === 'in_progress' ? 'primary' : 'default'"
                @click="changeRequirementFilter('in_progress')"
              >
                开发中
              </n-button>
              <n-button
                size="small"
                :type="requirementStatusFilter === 'released' ? 'primary' : 'default'"
                @click="changeRequirementFilter('released')"
              >
                已上线
              </n-button>
            </n-space>
          </div>
          <n-list v-if="requirementItems.length > 0" hoverable :bordered="false">
            <n-list-item v-for="req in requirementItems" :key="req.id" class="feedback-row">
              <div class="feedback-row-main">
                <div class="feedback-row-title">
                  <strong>{{ req.title }}</strong>
                  <n-tag :type="requirementStatusType(req.status)" size="small" :bordered="false">
                    {{ requirementStatusLabel(req.status) }}
                  </n-tag>
                  <n-tag :type="requirementPriorityType(req.priority)" size="small" :bordered="false">
                    {{ req.priority.toUpperCase() }}
                  </n-tag>
                </div>
                <p v-if="req.description">{{ req.description }}</p>
                <div v-if="req.source_feedback_id" class="feedback-meta">
                  <span>来自反馈 #{{ req.source_feedback_id }}</span>
                </div>
                <div class="feedback-meta">
                  <span>{{ formatDate(req.created_at) }}</span>
                </div>
              </div>
              <div class="feedback-row-side">
                <n-dropdown trigger="click" :options="[
                  { label: '待规划', key: 'planned', disabled: req.status === 'planned' },
                  { label: '开发中', key: 'in_progress', disabled: req.status === 'in_progress' },
                  { label: '已上线', key: 'released', disabled: req.status === 'released' }
                ]" @select="(key: string) => updateRequirementStatus(req, key as RequirementStatus)">
                  <n-button size="small" :loading="requirementSaving">状态</n-button>
                </n-dropdown>
                <n-button
                  size="small"
                  type="error"
                  :loading="requirementSaving"
                  @click="deleteRequirement(req.id)"
                >
                  删除
                </n-button>
              </div>
            </n-list-item>
          </n-list>
          <div v-if="!requirementLoading && requirementItems.length === 0" class="empty-state">
            暂无需求。
          </div>
          <div v-if="requirementTotal > requirementPageSize" class="pager-row">
            <n-button
              size="small"
              :disabled="requirementPage <= 1"
              @click="changeRequirementPage(requirementPage - 1)"
            >
              上一页
            </n-button>
            <span>第 {{ requirementPage }} / {{ requirementTotalPages }} 页</span>
            <n-button
              size="small"
              :disabled="requirementPage >= requirementTotalPages"
              @click="changeRequirementPage(requirementPage + 1)"
            >
              下一页
            </n-button>
          </div>
        </n-spin>
      </section>

      <section class="content-panel feedback-panel">
        <div class="feedback-panel-header">
          <div>
            <h3>产品反馈</h3>
            <p>收集团队对这个产品的反馈，并跟踪处理状态。</p>
          </div>
          <div class="panel-actions">
            <n-space>
              <n-button
                size="small"
                :type="feedbackStatusFilter === '' ? 'primary' : 'default'"
                @click="changeFeedbackFilter('')"
              >
                全部
              </n-button>
              <n-button
                size="small"
                :type="feedbackStatusFilter === 'open' ? 'primary' : 'default'"
                @click="changeFeedbackFilter('open')"
              >
                待处理
              </n-button>
              <n-button
                size="small"
                :type="feedbackStatusFilter === 'resolved' ? 'primary' : 'default'"
                @click="changeFeedbackFilter('resolved')"
              >
                已解决
              </n-button>
            </n-space>
            <n-button type="primary" @click="feedbackDrawerOpen = true">
              <template #icon>
                <n-icon><Plus /></n-icon>
              </template>
              新建反馈
            </n-button>
          </div>
        </div>

        <n-spin :show="feedbackLoading">
          <n-list v-if="feedbackItems.length > 0" hoverable clickable :bordered="false">
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
                <div class="feedback-meta">
                  <span>{{ feedback.vote_count }} 票</span>
                  <span>{{ feedback.comment_count }} 条评论</span>
                </div>
              </div>
              <div class="feedback-row-side">
                <n-button
                  size="small"
                  :type="feedback.voted ? 'primary' : 'default'"
                  :loading="feedbackSaving"
                  @click.stop="toggleFeedbackVote(feedback)"
                >
                  {{ feedback.voted ? '已投票' : '投票' }}
                </n-button>
                <time class="feedback-date">{{ formatDate(feedback.created_at) }}</time>
              </div>
            </n-list-item>
          </n-list>
          <div v-if="!feedbackLoading && feedbackItems.length === 0" class="empty-state">
            暂无产品反馈。
          </div>
          <div v-if="feedbackTotal > feedbackPageSize" class="pager-row">
            <n-button
              size="small"
              :disabled="feedbackPage <= 1"
              @click="changeFeedbackPage(feedbackPage - 1)"
            >
              上一页
            </n-button>
            <span>第 {{ feedbackPage }} / {{ feedbackTotalPages }} 页</span>
            <n-button
              size="small"
              :disabled="feedbackPage >= feedbackTotalPages"
              @click="changeFeedbackPage(feedbackPage + 1)"
            >
              下一页
            </n-button>
          </div>
        </n-spin>
      </section>
    </n-spin>

    <n-drawer v-model:show="flagDrawerOpen" width="min(420px, 100vw)" placement="right">
      <n-drawer-content title="新建功能开关" closable>
        <n-form label-placement="top" @submit.prevent="createFlag">
          <n-form-item label="开关键">
            <n-input v-model:value="flagForm.key" placeholder="例如：new_dashboard" />
          </n-form-item>
          <n-form-item label="名称">
            <n-input v-model:value="flagForm.name" placeholder="例如：新版仪表盘" />
          </n-form-item>
          <n-form-item label="环境">
            <n-input v-model:value="flagForm.environment" placeholder="production" />
          </n-form-item>
          <n-form-item label="发布比例">
            <n-input-number
              v-model:value="flagForm.rolloutPercentage"
              :min="0"
              :max="100"
              :step="5"
              style="width: 100%"
            />
          </n-form-item>
          <n-form-item label="说明">
            <n-input
              v-model:value="flagForm.description"
              type="textarea"
              placeholder="说明这个开关控制的功能或发布计划"
              :autosize="{ minRows: 4, maxRows: 8 }"
            />
          </n-form-item>
          <n-space justify="end">
            <n-button @click="flagDrawerOpen = false">取消</n-button>
            <n-button
              type="primary"
              attr-type="submit"
              :disabled="!canCreateFlag"
              :loading="flagSaving"
            >
              创建
            </n-button>
          </n-space>
        </n-form>
      </n-drawer-content>
    </n-drawer>

    <n-drawer v-model:show="feedbackDrawerOpen" width="min(420px, 100vw)" placement="right">
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

    <n-drawer v-model:show="requirementDrawerOpen" width="min(420px, 100vw)" placement="right">
      <n-drawer-content title="新建需求" closable>
        <n-form label-placement="top" @submit.prevent="createRequirement">
          <n-form-item label="标题">
            <n-input v-model:value="requirementForm.title" placeholder="简要描述需求内容" />
          </n-form-item>
          <n-form-item label="描述">
            <n-input
              v-model:value="requirementForm.description"
              type="textarea"
              placeholder="补充需求背景、目标和验收标准"
              :autosize="{ minRows: 5, maxRows: 10 }"
            />
          </n-form-item>
          <n-form-item label="优先级">
            <n-select
              v-model:value="requirementForm.priority"
              :options="[
                { label: 'P0 紧急', value: 'p0' },
                { label: 'P1 高', value: 'p1' },
                { label: 'P2 中', value: 'p2' },
                { label: 'P3 低', value: 'p3' }
              ]"
            />
          </n-form-item>
          <n-space justify="end">
            <n-button @click="requirementDrawerOpen = false">取消</n-button>
            <n-button
              type="primary"
              attr-type="submit"
              :disabled="!canCreateRequirement"
              :loading="requirementSaving"
            >
              创建
            </n-button>
          </n-space>
        </n-form>
      </n-drawer-content>
    </n-drawer>

    <n-drawer
      :show="Boolean(selectedFeedback)"
      width="min(460px, 100vw)"
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
          <div class="feedback-meta">
            <span>{{ selectedFeedback.vote_count }} 票</span>
            <span>{{ selectedFeedback.comment_count }} 条评论</span>
          </div>
          <time>{{ formatDate(selectedFeedback.created_at) }}</time>

          <n-button
            :type="selectedFeedback.voted ? 'primary' : 'default'"
            :loading="feedbackSaving"
            @click="toggleFeedbackVote(selectedFeedback)"
          >
            {{ selectedFeedback.voted ? '取消投票' : '投票支持' }}
          </n-button>

          <div class="comment-section">
            <h4>评论</h4>
            <div v-if="feedbackComments.length === 0" class="comment-empty">暂无评论。</div>
            <div v-for="comment in feedbackComments" :key="comment.id" class="comment-item">
              <p>{{ comment.content }}</p>
              <time>{{ formatDate(comment.created_at) }}</time>
            </div>
            <n-form label-placement="top" @submit.prevent="createComment">
              <n-form-item label="新增评论">
                <n-input
                  v-model:value="commentForm.content"
                  type="textarea"
                  placeholder="补充上下文、处理进展或讨论结论"
                  :autosize="{ minRows: 3, maxRows: 6 }"
                />
              </n-form-item>
              <n-space justify="end">
                <n-button
                  type="primary"
                  attr-type="submit"
                  :disabled="!canCreateComment"
                  :loading="commentSaving"
                >
                  发送评论
                </n-button>
              </n-space>
            </n-form>
          </div>

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
