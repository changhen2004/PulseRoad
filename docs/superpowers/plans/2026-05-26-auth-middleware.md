# Auth Middleware 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:test-driven-development 执行此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 实现 JWT 登录态校验中间件，并将当前用户 ID 注入 Gin Context。

**架构：** `auth.Service` 继续作为 JWT 规则来源，对外暴露 token 解析方法。`internal/middleware/auth.go` 只负责 Bearer token 提取、401 响应、context 注入和 helper 读取。`/api/auth/me` 从中间件注入的 user id 查询用户。

**技术栈：** Gin、github.com/golang-jwt/jwt/v5、现有统一 response 包。

---

## 文件结构

- 创建：`internal/middleware/auth.go`，实现 `AuthRequired` 和 `CurrentUserID`。
- 创建：`internal/middleware/auth_test.go`，覆盖缺失、非法、过期、有效 token。
- 修改：`internal/auth/service.go`，公开 `ParseToken` 和按用户 ID 查询当前用户。
- 修改：`internal/auth/handler.go`，`Me` 从 Gin Context 读取 user id。
- 修改：`internal/auth/router.go`，`GET /me` 挂载 `middleware.AuthRequired(service)`。
- 修改：`internal/auth/handler_test.go`、`internal/auth/service_test.go`，覆盖新行为。

## 任务

### 任务 1：Token 解析能力

- [ ] 编写失败测试：有效 token 可解析 user id，过期 token 返回未授权错误。
- [ ] 运行 `go test ./internal/auth` 验证失败。
- [ ] 将 `parseToken` 暴露为 `ParseToken`，添加 `CurrentUserByID`。
- [ ] 运行 `go test ./internal/auth` 验证通过。

### 任务 2：Gin 中间件

- [ ] 编写失败测试：无 token、非法 token、过期 token 都 401；有效 token 注入 `current_user_id`。
- [ ] 运行 `go test ./internal/middleware` 验证失败。
- [ ] 实现 `internal/middleware/auth.go`。
- [ ] 运行 `go test ./internal/middleware` 验证通过。

### 任务 3：接入 `/api/auth/me`

- [ ] 修改路由让 `/me` 使用 `AuthRequired`。
- [ ] 修改 handler 从 context 读取 user id。
- [ ] 运行 `go test ./...` 和 `go vet ./...`。
