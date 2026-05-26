# PulseRoad + FlagFlow 开发任务拆分

## 使用方式

这份任务清单按“先跑通主链路，再逐步加中间件和技术亮点”的顺序设计。

执行原则：

- 每次只做一个 task，不要并行铺太多模块。
- 每个 task 完成后都要能运行、能测试、能解释。
- 先保证后端 API 和业务闭环，再考虑前端页面。
- Redis 和 Kafka 要在业务流程跑通后加入，避免一开始复杂化。

最终目标：

```text
用户反馈 -> 评论/投票 -> 反馈转需求 -> 路线图 -> 发布日志
  -> 功能开关灰度 -> Kafka 事件 -> Redis 缓存/限流/热榜
```

## 阶段 0：项目基础设施

### TASK-0001：初始化 Go 项目

目标：

- 初始化 Go module。
- 准备 Gin、Gorm、MySQL Driver、Redis Client、Kafka Client 等基础依赖。

建议依赖：

```text
github.com/gin-gonic/gin
gorm.io/gorm
gorm.io/driver/mysql
github.com/redis/go-redis/v9
github.com/segmentio/kafka-go
github.com/spf13/viper
github.com/golang-jwt/jwt/v5
golang.org/x/crypto
```

产出：

- `go.mod`
- `go.sum`
- `cmd/api/main.go`
- `cmd/worker/main.go`

验收标准：

- `go run ./cmd/api` 可以启动 HTTP 服务。
- `GET /health` 返回成功。
- `go run ./cmd/worker` 可以启动 worker 进程并打印启动日志。

### TASK-0002：配置加载

目标：

- 支持从配置文件和环境变量读取服务配置。

产出：

- `internal/pkg/config`
- `config.yaml` 或 `.env.example`

配置项：

```text
app.name
app.env
server.port
mysql.dsn
redis.addr
kafka.brokers
jwt.secret
```

验收标准：

- API 和 worker 都使用同一套配置加载逻辑。
- 缺少关键配置时启动失败，并输出明确错误。

### TASK-0003：数据库连接与迁移入口

目标：

- 建立 MySQL 连接。
- 封装 Gorm 初始化逻辑。
- 预留自动迁移入口。

产出：

- `internal/pkg/database`
- `scripts/migrate.sh` 或 `cmd/migrate/main.go`

验收标准：

- 服务启动时可以连接 MySQL。
- 数据库连接失败时服务启动失败。
- 后续模块可以注册自己的模型迁移。

### TASK-0004：统一响应、错误与日志

目标：

- 统一 API 返回结构。
- 统一错误码。
- 加入请求日志中间件。

产出：

- `internal/pkg/response`
- `internal/pkg/logger`
- `internal/middleware`

响应格式：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

验收标准：

- 成功、参数错误、未登录、服务异常都有统一响应。
- 每个请求能打印 method、path、status、latency。

## 阶段 1：认证与团队产品基础

### TASK-0101：用户注册与登录

目标：

- 实现用户注册、登录、获取当前用户。

产出：

- `internal/auth/model.go`
- `internal/auth/repository.go`
- `internal/auth/service.go`
- `internal/auth/handler.go`
- `internal/auth/router.go`

接口：

```http
POST /api/auth/register
POST /api/auth/login
GET  /api/auth/me
```

验收标准：

- 密码使用 bcrypt 哈希存储。
- 登录成功返回 JWT。
- 未登录访问 `/api/auth/me` 返回 401。

### TASK-0102：JWT 认证中间件

目标：

- 实现登录态校验。
- 将当前用户 ID 注入 Gin Context。

产出：

- `internal/middleware/auth.go`

验收标准：

- 受保护接口必须携带合法 Token。
- Token 过期或非法时返回 401。

### TASK-0103：团队与成员管理

目标：

- 用户可以创建团队。
- 用户可以查看自己加入的团队。
- 创建者自动成为团队 owner。

产出：

- `internal/team`

接口：

```http
POST /api/teams
GET  /api/teams
GET  /api/teams/:id
```

验收标准：

- `team_members(team_id, user_id)` 有唯一约束。
- 非团队成员不能访问团队详情。

### TASK-0104：产品管理

目标：

- 在团队下创建和管理产品。

产出：

- `internal/product`

接口：

```http
POST /api/teams/:team_id/products
GET  /api/teams/:team_id/products
GET  /api/products/:id
```

验收标准：

- 只有团队成员可以创建产品。
- 产品必须归属于一个团队。

## 阶段 2：反馈业务闭环

### TASK-0201：反馈提交与列表

目标：

- 用户可以在产品下提交反馈。
- 支持按产品查看反馈列表。

产出：

- `internal/feedback/model.go`
- `internal/feedback/repository.go`
- `internal/feedback/service.go`
- `internal/feedback/handler.go`
- `internal/feedback/router.go`

接口：

```http
POST /api/products/:product_id/feedbacks
GET  /api/products/:product_id/feedbacks
GET  /api/feedbacks/:id
```

字段建议：

```text
title
content
status
priority
product_id
created_by
vote_count
comment_count
```

验收标准：

- 创建反馈后状态为 `pending`。
- 列表支持分页。
- 非产品所属团队成员不能访问内部产品反馈；如果你想做公开反馈页，需要单独设计公开权限。

### TASK-0202：反馈评论

目标：

- 用户可以对反馈发表评论。

产出：

- `feedback_comments` 表。
- 评论创建和列表接口。

接口：

```http
POST /api/feedbacks/:id/comments
GET  /api/feedbacks/:id/comments
```

验收标准：

- 评论创建后反馈的 `comment_count` 增加。
- 评论内容不能为空。
- 评论列表支持分页。

### TASK-0203：反馈投票

目标：

- 用户可以对反馈投票和取消投票。

产出：

- `feedback_votes` 表。

接口：

```http
POST   /api/feedbacks/:id/vote
DELETE /api/feedbacks/:id/vote
```

验收标准：

- `feedback_votes(feedback_id, user_id)` 有唯一约束。
- 重复投票不会导致 `vote_count` 重复增加。
- 取消投票后 `vote_count` 正确减少。

### TASK-0204：反馈状态流转

目标：

- 产品经理可以更新反馈状态。

接口：

```http
PATCH /api/feedbacks/:id/status
```

状态：

```text
pending
accepted
linked
rejected
closed
```

验收标准：

- 状态只能从允许的路径流转。
- 普通用户不能随意修改反馈状态。

## 阶段 3：需求池、路线图与发布日志

### TASK-0301：需求创建

目标：

- 产品经理可以创建需求。
- 需求可以来自一个或多个反馈。

产出：

- `internal/requirement`
- `requirements` 表。
- `requirement_feedback_links` 表。

接口：

```http
POST /api/products/:product_id/requirements
GET  /api/products/:product_id/requirements
GET  /api/requirements/:id
```

验收标准：

- 需求初始状态为 `backlog`。
- 需求必须归属于产品。

### TASK-0302：反馈转需求

目标：

- 将已有反馈关联到需求。
- 反馈状态变为 `linked`。

接口：

```http
POST /api/requirements/:id/link-feedback
```

验收标准：

- 同一个反馈不能重复关联同一个需求。
- 关联后反馈状态更新为 `linked`。

### TASK-0303：需求状态机

目标：

- 支持需求从需求池流转到已计划、开发中、已发布。

接口：

```http
PATCH /api/requirements/:id/status
```

状态：

```text
backlog
planned
developing
released
closed
```

验收标准：

- 状态流转合法。
- 状态变更记录更新时间和操作者。

### TASK-0304：路线图接口

目标：

- 按产品展示路线图。
- 按需求状态分组。

产出：

- `internal/roadmap`

接口：

```http
GET /api/products/:product_id/roadmap
```

验收标准：

- 返回 `planned`、`developing`、`released` 三组数据。
- 每个需求包含标题、状态、优先级、关联反馈数。

### TASK-0305：发布日志

目标：

- 产品经理或开发人员可以创建发布日志。
- 发布日志可以关联已发布需求。

产出：

- `internal/changelog`

接口：

```http
POST /api/products/:product_id/changelogs
GET  /api/products/:product_id/changelogs
```

验收标准：

- 发布日志支持分页。
- 发布日志创建后可以关联多个 requirement。

## 阶段 4：Redis 接入

### TASK-0401：Redis 客户端封装

目标：

- 初始化 Redis 客户端。
- 提供健康检查。

产出：

- `internal/pkg/redis`

验收标准：

- 服务启动时可以连接 Redis。
- Redis 不可用时给出明确错误。

### TASK-0402：反馈详情缓存

目标：

- 对反馈详情接口加入缓存。

Key：

```text
feedback:detail:{feedback_id}
```

验收标准：

- 第一次查询走 MySQL。
- 后续查询命中 Redis。
- 更新反馈、评论、投票后能删除或刷新缓存。

### TASK-0403：热门反馈排行榜

目标：

- 使用 Redis Sorted Set 维护产品下热门反馈。

Key：

```text
feedback:hot:{product_id}
```

评分建议：

```text
score = vote_count * 10 + comment_count * 2
```

验收标准：

- 投票或评论后更新热榜分数。
- 提供热门反馈接口。

接口：

```http
GET /api/products/:product_id/feedbacks/hot
```

### TASK-0404：基础限流

目标：

- 对登录、提交反馈、发表评论做简单限流。

Key：

```text
rate_limit:{action}:{user_id_or_ip}
```

验收标准：

- 高频请求返回 429。
- 限流窗口和次数可配置。

## 阶段 5：Kafka 与异步事件

### TASK-0501：Kafka Producer/Consumer 封装

目标：

- 封装 Kafka 生产者和消费者。
- worker 可以订阅 topic。

产出：

- `internal/pkg/kafka`
- `internal/event`

验收标准：

- API 能发送测试事件。
- worker 能消费测试事件。

### TASK-0502：反馈创建事件

Topic：

```text
feedback.created
```

目标：

- 创建反馈后发送事件。
- worker 消费后写入 `event_logs`。

验收标准：

- 创建反馈接口不依赖 worker 是否成功。
- worker 重启后仍能继续消费。

### TASK-0503：通知事件

Topic：

```text
notification.dispatch
```

目标：

- 评论、需求状态变更、发布日志创建后生成通知事件。
- worker 消费事件后写入 `notifications`。

产出：

- `internal/notification`

验收标准：

- 用户可以查看未读通知。
- 用户可以标记通知已读。

接口：

```http
GET   /api/notifications
PATCH /api/notifications/:id/read
```

### TASK-0504：统计聚合事件

Topic：

```text
stats.aggregate
```

目标：

- 异步统计反馈数、投票数、需求状态分布。

验收标准：

- 有统计看板接口。
- 统计数据可以接受短暂延迟。

接口：

```http
GET /api/products/:product_id/stats
```

## 阶段 6：FlagFlow 功能开关

### TASK-0601：功能开关基础模型

目标：

- 支持创建应用、环境和功能开关。

产出：

- `internal/flagflow`

接口：

```http
POST /api/flagflow/apps
POST /api/flagflow/apps/:app_id/flags
GET  /api/flagflow/apps/:app_id/flags
```

验收标准：

- 功能开关包含 `flag_key`、`enabled`、`env`、`version`。
- 同一应用同一环境下 `flag_key` 唯一。

### TASK-0602：灰度规则

目标：

- 支持白名单和百分比灰度。

接口：

```http
POST /api/flagflow/flags/:id/rules
```

规则：

```text
user_id_whitelist
percentage
```

验收标准：

- 白名单用户一定命中。
- 百分比灰度使用稳定哈希，同一用户结果稳定。

### TASK-0603：开关判断接口

目标：

- 业务系统可以判断某用户是否命中功能开关。

接口：

```http
POST /api/flagflow/evaluate
```

请求示例：

```json
{
  "app_key": "pulseroad",
  "env": "prod",
  "flag_key": "new_roadmap_view",
  "user_id": "1001"
}
```

验收标准：

- 返回是否命中。
- 返回命中的规则类型。
- 关闭状态下永远不命中。

### TASK-0604：配置缓存

目标：

- 使用 Redis 缓存功能开关配置。

Key：

```text
flag:config:{app_key}:{env}
flag:version:{app_key}:{env}
```

验收标准：

- 判断接口优先读 Redis。
- 修改配置后删除或刷新缓存。

### TASK-0605：发布与审计事件

Topic：

```text
flag.published
audit.logged
```

目标：

- 修改功能开关后记录审计日志。
- 发布配置后发送 Kafka 事件。

验收标准：

- 可以查询操作审计记录。
- worker 能消费 `flag.published`。

## 阶段 7：PulseRoad 接入 FlagFlow

### TASK-0701：新版路线图开关

目标：

- 在路线图接口中接入 `new_roadmap_view` 功能开关。

逻辑：

```text
命中开关 -> 返回新版路线图结构
未命中 -> 返回旧版路线图结构
```

验收标准：

- 白名单用户看到新版结构。
- 非命中用户看到旧版结构。
- 开关关闭后所有用户都看到旧版结构。

### TASK-0702：曝光事件

Topic：

```text
flag.exposed
```

目标：

- 用户命中功能开关时发送曝光事件。

验收标准：

- worker 能消费曝光事件。
- 曝光事件可以写入 `exposure_events`。

## 阶段 8：Docker 与本地演示

### TASK-0801：Docker Compose

目标：

- 一键启动 MySQL、Redis、Kafka。

产出：

- `deploy/docker-compose.yml`
- `deploy/.env.example`

验收标准：

- `docker compose up -d` 可以启动依赖。
- API 服务可以连接所有依赖。

### TASK-0802：接口文档

目标：

- 编写核心接口文档。

产出：

- `docs/api.md`

验收标准：

- 每个核心接口包含路径、方法、请求示例、响应示例。
- 文档覆盖认证、反馈、需求、路线图、发布日志、通知、FlagFlow。

### TASK-0803：数据库文档

目标：

- 记录核心表结构和关系。

产出：

- `docs/db-schema.md`

验收标准：

- 每个表说明用途。
- 标明关键索引和唯一约束。

### TASK-0804：Redis 与 Kafka 文档

目标：

- 记录 Redis Key 和 Kafka Topic。

产出：

- `docs/redis-keys.md`
- `docs/kafka-topics.md`

验收标准：

- 每个 Key 说明用途、结构、TTL、失效策略。
- 每个 Topic 说明事件来源、消费者、是否需要幂等。

### TASK-0805：README 演示流程

目标：

- 写出项目启动方式和演示路径。

产出：

- `README.md`

演示流程：

```text
1. 注册登录
2. 创建团队和产品
3. 提交反馈
4. 评论和投票
5. 反馈转需求
6. 更新需求状态到 released
7. 创建发布日志
8. 创建功能开关 new_roadmap_view
9. 白名单用户访问新版路线图
10. 查看 Kafka 消费后的通知、事件和曝光记录
```

验收标准：

- 新人按 README 可以启动项目。
- 面试时可以按演示流程稳定复现。

## 阶段 9：测试与质量

### TASK-0901：核心 Service 单元测试

目标：

- 覆盖核心业务逻辑。

重点：

- 投票幂等。
- 反馈转需求。
- 需求状态流转。
- FlagFlow 百分比灰度。

验收标准：

- 核心 service 有单元测试。
- `go test ./...` 通过。

### TASK-0902：集成测试

目标：

- 覆盖核心 API 流程。

流程：

```text
注册 -> 登录 -> 创建团队 -> 创建产品 -> 提交反馈 -> 投票 -> 转需求 -> 路线图
```

验收标准：

- 集成测试可以在本地依赖启动后运行。
- 测试数据不会污染开发数据。

### TASK-0903：错误处理与幂等梳理

目标：

- 检查重复请求、重复消息、缓存失效、权限错误。

验收标准：

- 重复投票不会出错。
- 重复消费 Kafka 消息不会产生重复通知或重复统计。
- Redis 缓存失效后系统仍能从 MySQL 恢复。

## 建议从哪里开始

第一周只做阶段 0 和阶段 1：

```text
TASK-0001 -> TASK-0002 -> TASK-0003 -> TASK-0004
TASK-0101 -> TASK-0102 -> TASK-0103 -> TASK-0104
```

做到这里时，项目应该具备：

- 服务能启动。
- 配置能加载。
- MySQL 能连接。
- 用户能注册登录。
- JWT 能保护接口。
- 能创建团队和产品。

这一步完成后，再进入反馈业务闭环。不要一开始就碰 Kafka 和 FlagFlow，否则项目会变成很多基础设施都没跑稳的半成品。

