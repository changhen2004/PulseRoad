# Auth Register Login 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:test-driven-development 执行此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 实现用户注册、登录、获取当前用户。

**架构：** `internal/auth` 按模型、仓储、服务、HTTP 处理、路由拆分。密码由 bcrypt 哈希存储，登录成功签发 HS256 JWT，`/api/auth/me` 在本任务内直接校验 Bearer token。

**技术栈：** Gin、Gorm、bcrypt、github.com/golang-jwt/jwt/v5、内存 fake repository 测试。

---

## 文件结构

- 创建：`internal/auth/repository.go`，封装用户查询和创建。
- 创建：`internal/auth/service.go`，封装注册、登录、JWT 签发和当前用户查询。
- 创建：`internal/auth/handler.go`，封装 HTTP JSON 输入输出。
- 创建：`internal/auth/router.go`，注册 `/api/auth/*` 路由。
- 修改：`internal/auth/model.go`，定义 User 模型并注册迁移模型。
- 修改：`cmd/api/main.go`，初始化 auth 模块并挂载路由。
- 修改：`go.mod`/`go.sum`，加入 JWT 和 bcrypt 依赖。
- 测试：`internal/auth/service_test.go`、`internal/auth/handler_test.go`。

## 任务

### 任务 1：Service 行为

- [ ] 编写失败测试：注册存储 bcrypt hash、重复邮箱失败、登录返回 JWT、错误密码失败。
- [ ] 运行 `go test ./internal/auth` 验证失败。
- [ ] 实现 `User`、repository、service 的最少代码。
- [ ] 运行 `go test ./internal/auth` 验证通过。

### 任务 2：HTTP 路由行为

- [ ] 编写失败测试：`POST /api/auth/register`、`POST /api/auth/login`、`GET /api/auth/me`，未登录 `/me` 返回 401。
- [ ] 运行 `go test ./internal/auth` 验证失败。
- [ ] 实现 handler/router 并返回统一响应格式。
- [ ] 运行 `go test ./internal/auth` 验证通过。

### 任务 3：接入 API 入口

- [ ] 修改 `cmd/api/main.go`，用已初始化 DB 和 JWT secret 构建 auth service 并注册路由。
- [ ] 运行 `go test ./...` 和 `go vet ./...`。
