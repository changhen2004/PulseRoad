# Product Management 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:test-driven-development 执行此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 实现在团队下创建产品、按团队列出产品、按 ID 查看产品。

**架构：** `internal/product` 按 model/repository/service/handler/router 分层。产品服务通过一个团队成员校验接口判断当前用户是否能访问团队产品；`team.Service` 暴露 `IsMember` 供 API 装配时注入。

**技术栈：** Gin、Gorm、现有 auth middleware、现有统一 response 包。

---

## 文件结构

- 创建：`internal/product/model.go`，定义 `Product` 并注册迁移模型。
- 创建：`internal/product/repository.go`，封装创建、团队列表、按 ID 查询。
- 创建：`internal/product/service.go`，封装成员权限、创建和查询逻辑。
- 创建：`internal/product/handler.go`，封装 HTTP 参数解析和响应。
- 创建：`internal/product/router.go`，注册产品路由并挂认证中间件。
- 创建：`internal/product/service_test.go`、`internal/product/handler_test.go`，覆盖验收行为。
- 修改：`internal/team/service.go`，新增 `IsMember`。
- 修改：`cmd/api/main.go`，初始化并挂载 product 模块。

## 任务

### 任务 1：Service 行为

- [ ] 编写失败测试：团队成员可以创建产品；非成员创建返回 forbidden；产品有非零 team id；成员可列出产品；非成员不能查看产品详情。
- [ ] 运行 `go test ./internal/product` 验证失败。
- [ ] 实现 product model/repository/service 的最少代码。
- [ ] 运行 `go test ./internal/product` 验证通过。

### 任务 2：HTTP 路由行为

- [ ] 编写失败测试：`POST /api/teams/:team_id/products`、`GET /api/teams/:team_id/products`、`GET /api/products/:id`；未登录 401，非成员 403。
- [ ] 运行 `go test ./internal/product` 验证失败。
- [ ] 实现 handler/router。
- [ ] 运行 `go test ./internal/product` 验证通过。

### 任务 3：跨模块接入

- [ ] 为 `team.Service` 添加 `IsMember(ctx,userID,teamID)`。
- [ ] 修改 `cmd/api/main.go`，创建 product service 并注册路由。
- [ ] 运行 `go test ./...`、`go vet ./...`、`go test -count=1 ./...`。
