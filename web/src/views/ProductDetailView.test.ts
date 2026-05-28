import { createApp, defineComponent, h, nextTick, reactive, type App } from 'vue';
import { afterEach, describe, expect, it, vi } from 'vitest';

import type { Feedback, Product, UpdateFeedbackStatusPayload } from '../api/types';

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

function feedback(id: number, overrides: Partial<Feedback> = {}): Feedback {
  return {
    id,
    product_id: 1,
    title: 'Old feedback',
    content: 'stale feedback content',
    status: 'open',
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
}

interface ViewMocks {
  productGet: ReturnType<typeof vi.fn<(id: number) => Promise<Product>>>;
  listByProduct: ReturnType<typeof vi.fn<(productID: number) => Promise<Feedback[]>>>;
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
  const listByProduct = vi.fn<(productID: number) => Promise<Feedback[]>>();
  const updateStatus =
    vi.fn<(id: number, payload: UpdateFeedbackStatusPayload) => Promise<Feedback>>();
  configure({ productGet, listByProduct, updateStatus });

  vi.doMock('vue-router', () => ({
    useRoute: () => route,
    useRouter: () => router
  }));
  vi.doMock('../api/products', () => ({
    productsApi: { get: productGet }
  }));
  vi.doMock('../api/feedback', () => ({
    feedbackApi: {
      create: vi.fn(),
      get: vi.fn(),
      listByProduct,
      updateStatus
    }
  }));
  vi.doMock('naive-ui', () => ({
    NButton: stubComponent('button'),
    NDrawer: stubComponent(),
    NDrawerContent: stubComponent(),
    NForm: stubComponent('form'),
    NFormItem: stubComponent(),
    NIcon: stubComponent('span'),
    NInput: stubComponent('input'),
    NList: stubComponent(),
    NListItem: stubComponent(),
    NSpace: stubComponent(),
    NSpin: stubComponent(),
    NTag: stubComponent('span'),
    useMessage: () => message
  }));

  const { default: ProductDetailView } = await import('./ProductDetailView.vue');
  const root = document.createElement('div');
  document.body.appendChild(root);

  const app = createApp(ProductDetailView);
  app.mount(root);
  await flushView();

  return { app, root, route, productGet, listByProduct, updateStatus };
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.doUnmock('vue-router');
  vi.doUnmock('../api/products');
  vi.doUnmock('../api/feedback');
  vi.doUnmock('naive-ui');
  document.body.innerHTML = '';
});

describe('ProductDetailView feedback state', () => {
  it('clears stale feedback when switched product feedback fails to load', async () => {
    const mounted = await mountView(({ productGet, listByProduct }) => {
      productGet.mockImplementation(async (id) => product(id));
      listByProduct.mockImplementation(async (productID) => {
        if (productID === 1) return [feedback(10)];
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
      listByProduct.mockImplementation(async (productID) => [
        feedback(productID, {
          product_id: productID,
          title: `Feedback ${productID}`,
          content: `feedback for product ${productID}`
        })
      ]);
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
      listByProduct.mockResolvedValue([openFeedback]);
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
      expect(mounted.root.textContent).toContain('已解决');
      expect(mounted.root.textContent).not.toContain('待处理');
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
      listByProduct.mockResolvedValueOnce([openFeedback]).mockRejectedValueOnce(new Error('offline'));
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
});
