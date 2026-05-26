# Team Management 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:test-driven-development 执行此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 实现团队创建、当前用户团队列表、成员可见的团队详情。

**架构：** `internal/team` 按 model/repository/service/handler/router 分层。所有团队路由复用 `middleware.AuthRequired` 获取当前用户 ID；服务层负责创建团队时同步创建 owner 成员，并在查看详情前校验成员身份。

**技术栈：** Gin、Gorm、现有 auth middleware、现有统一 response 包。

---

## 文件结构

- 创建：`internal/team/model.go`，定义 `Team`、`TeamMember` 并注册迁移模型。
- 创建：`internal/team/repository.go`，封装团队创建、成员关系、列表和详情查询。
- 创建：`internal/team/service.go`，封装创建团队、列出团队、按 ID 获取团队的业务规则。
- 创建：`internal/team/handler.go`，封装 HTTP 请求解析和统一响应。
- 创建：`internal/team/router.go`，注册 `/api/teams` 路由并挂认证中间件。
- 创建：`internal/team/service_test.go`、`internal/team/handler_test.go`，覆盖核心验收标准。
- 修改：`cmd/api/main.go`，初始化并挂载 team 模块。

## 任务

### 任务 1：Service 行为

- [ ] 编写失败测试：创建团队会创建 owner 成员；成员能列出自己的团队；非成员查看详情返回 forbidden；`TeamMember` 模型有 `(team_id,user_id)` 唯一索引。
- [ ] 运行 `go test ./internal/team` 验证失败。
- [ ] 实现 model/repository/service 的最少代码。
- [ ] 运行 `go test ./internal/team` 验证通过。

### 任务 2：HTTP 路由行为

- [ ] 编写失败测试：`POST /api/teams`、`GET /api/teams`、`GET /api/teams/:id`，未登录 401，非成员详情 403。
- [ ] 运行 `go test ./internal/team` 验证失败。
- [ ] 实现 handler/router。
- [ ] 运行 `go test ./internal/team` 验证通过。

### 任务 3：API 接入

- [ ] 修改 `cmd/api/main.go`，创建 team service 并注册团队路由。
- [ ] 运行 `go test ./...`、`go vet ./...`、`go test -count=1 ./...`。
