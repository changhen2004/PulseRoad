# PulseRoad 精简开发任务

这份文档只保留当前项目需要掌握和维护的核心模块，目标是让项目更容易完成、运行和理解。

## 当前主线

```text
用户 -> 团队 -> 产品
```

## 已完成模块

### 1. 基础设施

- Go module。
- Gin HTTP 服务。
- MySQL + Gorm 初始化。
- Gorm `AutoMigrate` 迁移入口。
- 统一响应结构。
- 请求日志中间件。
- 配置文件和环境变量覆盖。

入口：

- `cmd/api`
- `cmd/migrate`
- `cmd/worker`

### 2. 用户认证

接口：

```http
POST /api/auth/register
POST /api/auth/login
GET  /api/auth/me
```

能力：

- bcrypt 密码哈希。
- JWT 登录态。
- 当前用户查询。

### 3. 登录态中间件

能力：

- 校验 `Authorization: Bearer <token>`。
- 非法或过期 token 返回 401。
- 将当前用户 ID 注入 Gin Context。

### 4. 团队管理

接口：

```http
POST /api/teams
GET  /api/teams
GET  /api/teams/:id
```

能力：

- 用户创建团队。
- 创建者自动成为团队 `owner`。
- 用户查看自己加入的团队。
- 非团队成员不能访问团队详情。
- `team_members(team_id, user_id)` 有唯一约束。

### 5. 产品管理

接口：

```http
POST /api/teams/:team_id/products
GET  /api/teams/:team_id/products
GET  /api/products/:id
```

能力：

- 团队成员可以在团队下创建产品。
- 产品必须归属于团队。
- 非团队成员不能创建或查看团队产品。

## 保留但暂不深入的基础设施

### Redis

当前只保留配置项：

```yaml
redis:
  addr: "127.0.0.1:6379"
```

暂不实现缓存、限流或排行榜。

### RabbitMQ

当前只保留配置和 Worker 骨架：

```yaml
rabbitmq:
  url: "amqp://guest:guest@127.0.0.1:5672/"
```

Worker 会校验 RabbitMQ URL 格式，但暂不连接和消费消息。

## 暂不实现的模块

为了降低项目复杂度，以下模块先不做：

- 反馈提交与列表。
- 评论与投票。
- 反馈转需求。
- 路线图。
- 发布日志。
- 功能开关。
- 事件日志。
- 通知。
- Redis 缓存、限流、热榜。
- RabbitMQ 生产者和消费者。

## 推荐下一步

1. 先跑通 `go test ./...`。
2. 启动 MySQL，执行 `go run ./cmd/migrate`。
3. 启动 `go run ./cmd/api`。
4. 依次调用认证、团队、产品接口。
5. 理解每个业务模块中的 `model -> repository -> service -> handler -> router` 分层。

完成以上步骤后，再考虑新增反馈模块。
