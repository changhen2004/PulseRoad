import { createApp, defineComponent, h, nextTick, reactive, type App } from 'vue';
import { afterEach, describe, expect, it, vi } from 'vitest';

import type {
  FeatureFlag,
  Feedback,
  FeedbackPage,
  Product,
  ProductSummary,
  UpdateFeedbackStatusPayload
} from '../api/types';

function product(id: number): Product {
  return {
    id,
    team_id: 1,
    name: `Product ${id}`,
    description: '',
    created_by: 1,
    created_at: '2026-01-01T00:00:00Z'
  };
}

function productSummary(id: number): ProductSummary {
  return {
    product: product(id),
    feedback_total: 0,
    feedback_open: 0,
    feedback_resolved: 0,
    comment_total: 0,
    vote_total: 0,
    flag_total: 0,
    flag_enabled: 0
  };
}

function feedback(id: number, overrides: Partial<Feedback> = {}): Feedback {
  return {
    id,
    product_id: 1,
    title: 'Old feedback',
    content: 'stale feedback content',
    status: 'open',
    created_by: 1,
    vote_count: 0,
    comment_count: 0,
    voted: false,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  };
}

function featureFlag(id: number, overrides: Partial<FeatureFlag> = {}): FeatureFlag {
  return {
    id,
    product_id: 1,
    key: 'new_dashboard',
    name: 'New Dashboard',
    description: 'Roll out dashboard',
    environment: 'production',
    enabled: false,
    rollout_percentage: 25,
    created_by: 1,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  };
}

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}

function asyncStubComponent(tag = 'div') {
  return defineComponent({
    inheritAttrs: false,
    props: {
      show: {
        type: Boolean,
        default: false
      }
    },
    setup(props, { attrs, slots }) {
      return () => h(tag, attrs, [props.show ? h('span', 'loading') : null, slots.default?.()]);
    }
  });
}

function stubComponent(tag = 'div') {
  return defineComponent({
    inheritAttrs: false,
    setup(_, { attrs, slots }) {
      return () => h(tag, attrs, slots.default?.());
    }
  });
}

async function flushView() {
  await Promise.resolve();
  await nextTick();
  await Promise.resolve();
  await nextTick();
  await Promise.resolve();
  await nextTick();
}

function feedbackPage(items: Feedback[]): FeedbackPage {
  return {
    items,
    page: 1,
    page_size: 10,
    total: items.length
  };
}

interface ViewMocks {
  productGet: ReturnType<typeof vi.fn<(id: number) => Promise<Product>>>;
  productSummaryGet: ReturnType<typeof vi.fn<(id: number) => Promise<ProductSummary>>>;
  listByProduct: ReturnType<typeof vi.fn<(productID: number) => Promise<FeedbackPage>>>;
  listFlagsByProduct: ReturnType<typeof vi.fn<(productID: number) => Promise<FeatureFlag[]>>>;
  toggleFlag: ReturnType<typeof vi.fn<(id: number, payload: { enabled: boolean }) => Promise<FeatureFlag>>>;
  updateStatus: ReturnType<
    typeof vi.fn<(id: number, payload: UpdateFeedbackStatusPayload) => Promise<Feedback>>
  >;
}

async function mountView(configure: (mocks: ViewMocks) => void) {
  vi.resetModules();

  const route = reactive({ params: { id: '1' } });
  const router = { push: vi.fn() };
  const message = { error: vi.fn(), warning: vi.fn() };
  const productGet = vi.fn<(id: number) => Promise<Product>>();
  const productSummaryGet = vi.fn<(id: number) => Promise<ProductSummary>>();
  const listByProduct = vi.fn<(productID: number) => Promise<FeedbackPage>>();
  const listFlagsByProduct = vi.fn<(productID: number) => Promise<FeatureFlag[]>>();
  const toggleFlag = vi.fn<(id: number, payload: { enabled: boolean }) => Promise<FeatureFlag>>();
  const updateStatus =
    vi.fn<(id: number, payload: UpdateFeedbackStatusPayload) => Promise<Feedback>>();
  productSummaryGet.mockImplementation(async (id) => productSummary(id));
  listFlagsByProduct.mockResolvedValue([]);
  configure({ productGet, productSummaryGet, listByProduct, listFlagsByProduct, toggleFlag, updateStatus });

  vi.doMock('vue-router', () => ({
    useRoute: () => route,
    useRouter: () => router
  }));
  vi.doMock('../api/products', () => ({
    productsApi: { get: productGet, summary: productSummaryGet }
  }));
  vi.doMock('../api/feedback', () => ({
    feedbackApi: {
      create: vi.fn(),
      get: vi.fn(),
      listByProduct,
      listComments: vi.fn().mockResolvedValue([]),
      vote: vi.fn(),
      unvote: vi.fn(),
      updateStatus
    }
  }));
  vi.doMock('../api/flagflow', () => ({
    flagflowApi: {
      create: vi.fn(),
      evaluate: vi.fn(),
      get: vi.fn(),
      listByProduct: listFlagsByProduct,
      toggle: toggleFlag,
      update: vi.fn()
    }
  }));
  vi.doMock('naive-ui', () => ({
    NButton: stubComponent('button'),
    NDropdown: stubComponent(),
    NDrawer: stubComponent(),
    NDrawerContent: stubComponent(),
    NForm: stubComponent('form'),
    NFormItem: stubComponent(),
    NIcon: stubComponent('span'),
    NInput: stubComponent('input'),
    NInputNumber: stubComponent('input'),
    NList: stubComponent(),
    NListItem: stubComponent(),
    NSelect: stubComponent(),
    NSpace: stubComponent(),
    NSpin: asyncStubComponent(),
    NTag: stubComponent('span'),
    useMessage: () => message
  }));
  vi.doMock('../api/requirements', () => ({
    requirementApi: {
      create: vi.fn(),
      delete: vi.fn(),
      get: vi.fn(),
      listByProduct: vi.fn().mockResolvedValue({ items: [], page: 1, page_size: 10, total: 0 }),
      update: vi.fn()
    }
  }));

  const { default: ProductDetailView } = await import('./ProductDetailView.vue');
  const root = document.createElement('div');
  document.body.appendChild(root);

  const app = createApp(ProductDetailView);
  app.mount(root);
  await flushView();

  return {
    app,
    root,
    route,
    productGet,
    productSummaryGet,
    listByProduct,
    listFlagsByProduct,
    toggleFlag,
    updateStatus
  };
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.doUnmock('vue-router');
  vi.doUnmock('../api/products');
  vi.doUnmock('../api/feedback');
  vi.doUnmock('../api/flagflow');
  vi.doUnmock('naive-ui');
  document.body.innerHTML = '';
});

describe('ProductDetailView feedback state', () => {
  it('clears stale feedback when switched product feedback fails to load', async () => {
    const mounted = await mountView(({ productGet, listByProduct }) => {
      productGet.mockImplementation(async (id) => product(id));
      listByProduct.mockImplementation(async (productID) => {
        if (productID === 1) return feedbackPage([feedback(10)]);
        throw new Error('feedback unavailable');
      });
    });
    let app: App<Element> | undefined = mounted.app;

    try {
      expect(mounted.root.textContent).toContain('Old feedback');

      mounted.route.params.id = '2';
      await flushView();

      expect(mounted.root.textContent).toContain('Product 2');
      expect(mounted.root.textContent).not.toContain('Old feedback');
      expect(mounted.root.textContent).not.toContain('stale feedback content');
    } finally {
      app?.unmount();
      app = undefined;
    }
  });

  it('ignores late product responses after route changes', async () => {
    const firstProduct = deferred<Product>();
    const mounted = await mountView(({ productGet, listByProduct }) => {
      productGet.mockImplementation((id) => {
        if (id === 1) return firstProduct.promise;
        return Promise.resolve(product(id));
      });
      listByProduct.mockImplementation(async (productID) =>
        feedbackPage([feedback(productID, {
          product_id: productID,
          title: `Feedback ${productID}`,
          content: `feedback for product ${productID}`
        })])
      );
    });
    let app: App<Element> | undefined = mounted.app;

    try {
      mounted.route.params.id = '2';
      await flushView();

      expect(mounted.root.textContent).toContain('Product 2');
      expect(mounted.root.textContent).toContain('Feedback 2');

      firstProduct.resolve(product(1));
      await flushView();

      expect(mounted.root.textContent).toContain('Product 2');
      expect(mounted.root.textContent).toContain('Feedback 2');
      expect(mounted.root.textContent).not.toContain('Product 1');
      expect(mounted.root.textContent).not.toContain('Feedback 1');
    } finally {
      app?.unmount();
      app = undefined;
    }
  });

  it('keeps the updated status when the follow-up refresh returns stale data', async () => {
    const openFeedback = feedback(10);
    const resolvedFeedback = feedback(10, { status: 'resolved' });
    const mounted = await mountView(({ productGet, listByProduct, updateStatus }) => {
      productGet.mockImplementation(async (id) => product(id));
      listByProduct.mockResolvedValue(feedbackPage([openFeedback]));
      updateStatus.mockResolvedValue(resolvedFeedback);
    });
    let app: App<Element> | undefined = mounted.app;

    try {
      mounted.root.querySelector<HTMLElement>('.feedback-row')?.click();
      await flushView();

      const resolveButton = Array.from(mounted.root.querySelectorAll('button')).find((button) =>
        button.textContent?.includes('标记已解决')
      );
      resolveButton?.click();
      await flushView();

      expect(mounted.updateStatus).toHaveBeenCalledWith(10, { status: 'resolved' });
      const rowText = mounted.root.querySelector<HTMLElement>('.feedback-row')?.textContent ?? '';
      expect(rowText).toContain('已解决');
      expect(rowText).not.toContain('待处理');
    } finally {
      app?.unmount();
      app = undefined;
    }
  });

  it('keeps the updated feedback in the list when the follow-up refresh fails', async () => {
    const openFeedback = feedback(10);
    const resolvedFeedback = feedback(10, { status: 'resolved' });
    const mounted = await mountView(({ productGet, listByProduct, updateStatus }) => {
      productGet.mockImplementation(async (id) => product(id));
      listByProduct
        .mockResolvedValueOnce(feedbackPage([openFeedback]))
        .mockRejectedValueOnce(new Error('offline'));
      updateStatus.mockResolvedValue(resolvedFeedback);
    });
    let app: App<Element> | undefined = mounted.app;

    try {
      mounted.root.querySelector<HTMLElement>('.feedback-row')?.click();
      await flushView();

      const resolveButton = Array.from(mounted.root.querySelectorAll('button')).find((button) =>
        button.textContent?.includes('标记已解决')
      );
      resolveButton?.click();
      await flushView();

      const rowText = mounted.root.querySelector<HTMLElement>('.feedback-row')?.textContent ?? '';
      expect(rowText).toContain('已解决');
      expect(rowText).not.toContain('待处理');
    } finally {
      app?.unmount();
      app = undefined;
    }
  });

  it('keeps the updated flag status when the follow-up refresh fails', async () => {
    const disabledFlag = featureFlag(11);
    const enabledFlag = featureFlag(11, { enabled: true });
    const mounted = await mountView(({ productGet, listByProduct, listFlagsByProduct, toggleFlag }) => {
      productGet.mockImplementation(async (id) => product(id));
      listByProduct.mockResolvedValue(feedbackPage([]));
      listFlagsByProduct.mockResolvedValueOnce([disabledFlag]).mockRejectedValueOnce(new Error('offline'));
      toggleFlag.mockResolvedValue(enabledFlag);
    });
    let app: App<Element> | undefined = mounted.app;

    try {
      const enableButton = Array.from(mounted.root.querySelectorAll('button')).find((button) =>
        button.textContent?.includes('开启')
      );
      enableButton?.click();
      await flushView();

      expect(mounted.toggleFlag).toHaveBeenCalledWith(11, { enabled: true });
      expect(mounted.root.textContent).toContain('已开启');
      expect(mounted.root.textContent).not.toContain('已关闭');
    } finally {
      app?.unmount();
      app = undefined;
    }
  });

  it('stops stale feedback loading when switched product fails to load', async () => {
    const firstFeedback = deferred<FeedbackPage>();
    const mounted = await mountView(({ productGet, listByProduct }) => {
      productGet.mockImplementation((id) => {
        if (id === 2) return Promise.reject(new Error('product unavailable'));
        return Promise.resolve(product(id));
      });
      listByProduct.mockImplementation((productID) => {
        if (productID === 1) return firstFeedback.promise;
        return Promise.resolve(feedbackPage([]));
      });
    });
    let app: App<Element> | undefined = mounted.app;

    try {
      expect(mounted.root.textContent).toContain('loading');

      mounted.route.params.id = '2';
      await flushView();

      expect(mounted.root.textContent).not.toContain('loading');
    } finally {
      app?.unmount();
      app = undefined;
      firstFeedback.resolve(feedbackPage([]));
    }
  });
});
