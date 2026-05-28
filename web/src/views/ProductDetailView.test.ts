import { createApp, defineComponent, h, nextTick, reactive, type App } from 'vue';
import { afterEach, describe, expect, it, vi } from 'vitest';

import type { Feedback, Product } from '../api/types';

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

function feedback(id: number): Feedback {
  return {
    id,
    product_id: 1,
    title: 'Old feedback',
    content: 'stale feedback content',
    status: 'open',
    created_by: 1,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z'
  };
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
}

async function mountView(configure: (mocks: ViewMocks) => void) {
  vi.resetModules();

  const route = reactive({ params: { id: '1' } });
  const router = { push: vi.fn() };
  const message = { error: vi.fn(), warning: vi.fn() };
  const productGet = vi.fn<(id: number) => Promise<Product>>();
  const listByProduct = vi.fn<(productID: number) => Promise<Feedback[]>>();
  configure({ productGet, listByProduct });

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
      updateStatus: vi.fn()
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

  return { app, root, route, productGet, listByProduct };
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
});
