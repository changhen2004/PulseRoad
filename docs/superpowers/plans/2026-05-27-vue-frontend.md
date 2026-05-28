# Vue Frontend 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 新增一个轻量 Vue 3 前端，用 RESTful API 对接现有认证、团队、产品模块。

**架构：** 前端放在 `web/`，使用 Vite 独立开发和构建，通过 dev proxy 访问后端 `/api`。API、token、状态管理集中封装，页面只处理表单和展示。

**技术栈：** Vue 3、TypeScript、Vite、Vue Router、Pinia、Axios、Naive UI、Lucide Vue、Vitest。

---

### 任务 1：前端工程骨架和 API 行为测试

**文件：**
- 创建：`web/package.json`
- 创建：`web/vite.config.ts`
- 创建：`web/tsconfig.json`
- 创建：`web/tsconfig.node.json`
- 创建：`web/index.html`
- 创建：`web/src/api/http.test.ts`
- 创建：`web/src/stores/auth.test.ts`

- [ ] **步骤 1：编写失败的测试**

```ts
import { describe, expect, it, vi } from 'vitest';
import { createApiClient } from './http';
import { setStoredToken } from '../stores/auth';

it('adds bearer token to requests', async () => {
  setStoredToken('token-123');
  const adapter = vi.fn(async (config) => ({
    data: { code: 0, message: 'ok', data: { ok: true } },
    status: 200,
    statusText: 'OK',
    headers: {},
    config
  }));
  const api = createApiClient({ adapter });
  await api.get('/auth/me');
  expect(adapter.mock.calls[0][0].headers.Authorization).toBe('Bearer token-123');
});
```

- [ ] **步骤 2：运行测试验证失败**

运行：`npm test -- --run`
预期：FAIL，原因是 API/token 模块尚未实现。

- [ ] **步骤 3：实现最少代码**

创建 `src/api/http.ts`、`src/stores/auth.ts`，实现 token 存储、Authorization 注入、统一响应解包、401 清理登录态。

- [ ] **步骤 4：运行测试验证通过**

运行：`npm test -- --run`
预期：PASS。

### 任务 2：业务 API、路由和应用状态

**文件：**
- 创建：`web/src/api/types.ts`
- 创建：`web/src/api/auth.ts`
- 创建：`web/src/api/teams.ts`
- 创建：`web/src/api/products.ts`
- 创建：`web/src/router/index.ts`
- 创建：`web/src/stores/session.ts`

- [ ] **步骤 1：定义后端响应类型**

包含 `User`、`Team`、`Product`、`AuthResult`、创建团队/产品请求类型。

- [ ] **步骤 2：封装 RESTful API**

认证接口对接 `/api/auth/register`、`/api/auth/login`、`/api/auth/me`；团队接口对接 `/api/teams`；产品接口对接 `/api/teams/:team_id/products` 和 `/api/products/:id`。

- [ ] **步骤 3：实现路由守卫**

未登录访问 `/app/**` 跳转 `/login`，已登录访问 `/login` 或 `/register` 跳转 `/app/teams`。

### 任务 3：页面和交互

**文件：**
- 创建：`web/src/App.vue`
- 创建：`web/src/main.ts`
- 创建：`web/src/styles.css`
- 创建：`web/src/layouts/AppLayout.vue`
- 创建：`web/src/views/LoginView.vue`
- 创建：`web/src/views/RegisterView.vue`
- 创建：`web/src/views/TeamsView.vue`
- 创建：`web/src/views/TeamDetailView.vue`
- 创建：`web/src/views/ProductDetailView.vue`

- [ ] **步骤 1：实现登录和注册页面**

表单提交后保存 JWT，加载当前用户，进入 `/app/teams`。

- [ ] **步骤 2：实现应用布局**

左侧导航、顶部当前用户、退出登录，保持简洁的工作台风格。

- [ ] **步骤 3：实现团队和产品页面**

团队列表支持创建；团队详情支持查看信息、创建产品、查看产品；产品详情展示基础信息。

### 任务 4：启动脚本、文档和验证

**文件：**
- 创建：`scripts/dev-web.sh`
- 修改：`README.md`

- [ ] **步骤 1：添加启动脚本**

`scripts/dev-web.sh` 执行 `npm --prefix web run dev -- --host 0.0.0.0`。

- [ ] **步骤 2：更新 README**

说明前端目录、启动方式、Vite proxy 和后端依赖。

- [ ] **步骤 3：运行验证**

运行：`npm --prefix web run build`、`npm --prefix web test -- --run`、`go test -count=1 ./...`。
