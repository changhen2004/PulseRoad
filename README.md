# PulseRoad

PulseRoad 是一个精简版产品反馈管理项目。当前项目围绕两条容易理解的业务流展开：

```text
用户 -> 团队 -> 产品 -> 反馈
用户 -> 团队 -> 产品 -> 功能开关
```

需求、路线图、发布日志等模块暂不实现，避免项目过早变复杂。

## 已实现功能

- 用户注册、登录、获取当前用户。
- JWT 登录态校验。
- 创建团队、查看我的团队、查看团队详情。
- 创建者自动成为团队 `owner`。
- 支持团队成员邀请、接受邀请、成员列表、角色调整和移除成员。
- 在团队下创建产品、查看团队产品、查看产品详情。
- 产品详情提供反馈、评论、投票和功能开关聚合摘要。
- 在产品下创建反馈、筛选分页查看反馈、查看反馈详情、更新反馈状态。
- 支持反馈评论和反馈投票。
- 在产品下创建功能开关、查看开关、开启/关闭开关、按用户键计算灰度结果。
- 团队、产品、反馈和功能开关接口都带成员权限校验。
- MySQL 连接和 Gorm 自动迁移。
- Redis 登录失败次数限制。
- Redis 功能开关缓存。
- RabbitMQ 反馈和功能开关事件发布，Worker 统一消费。
- Vue 前端工作台，对接当前 RESTful API。

## 技术栈

- Go 1.25.6
- Gin
- Gorm
- MySQL
- JWT
- bcrypt
- Redis
- RabbitMQ
- Vue 3
- Vite
- Naive UI
- Pinia
- Axios

## 目录结构

```text
.
├── cmd
│   ├── api       # HTTP API 服务
│   ├── migrate   # 数据库自动迁移
│   └── worker    # RabbitMQ 业务事件消费者
├── docs
│   └── development-tasks.md
├── internal
│   ├── auth       # 注册、登录、JWT、当前用户
│   ├── feedback   # 产品反馈管理
│   ├── flagflow   # 功能开关、灰度发布、缓存和事件
│   ├── middleware # 登录态中间件
│   ├── pkg        # 配置、数据库、日志、响应、Redis、RabbitMQ 工具
│   ├── product    # 产品管理
│   └── team       # 团队和成员权限
├── scripts
│   └── dev-web.sh # 前端开发服务启动脚本
├── web            # Vue 前端工作台
├── go.mod
└── go.sum
```

## 配置

默认配置文件：

```text
internal/pkg/config/config.yaml
```

示例：

```yaml
app:
  name: "pulseroad"
  env: "development"

server:
  port: 8080

mysql:
  dsn: "user:password@tcp(127.0.0.1:3306)/pulseroad?charset=utf8mb4&parseTime=True&loc=Local"

redis:
  addr: "127.0.0.1:6379"

rabbitmq:
  url: "amqp://guest:guest@127.0.0.1:5672/"

jwt:
  secret: "change-me-in-production"
```

支持环境变量覆盖：

| 环境变量 | 说明 |
| --- | --- |
| `PULSEROAD_APP_NAME` | 应用名称 |
| `PULSEROAD_APP_ENV` | 运行环境 |
| `PULSEROAD_SERVER_PORT` | HTTP 端口 |
| `PULSEROAD_MYSQL_DSN` | MySQL DSN |
| `PULSEROAD_REDIS_ADDR` | Redis 地址 |
| `PULSEROAD_RABBITMQ_URL` | RabbitMQ URL |
| `PULSEROAD_JWT_SECRET` | JWT 密钥 |

生产环境不要使用默认 JWT 密钥。

## 运行项目

推荐使用 Docker Compose 一键启动开发环境：

```bash
docker compose up
```

首次启动会下载 Go 和 NPM 依赖，前端服务可能需要等待一两分钟才会显示 Vite 启动日志。

Compose 会启动：

- MySQL
- Redis
- RabbitMQ
- 数据库迁移任务
- API 服务
- Worker
- Vue 前端

访问地址：

```text
前端：http://localhost:5173
API：http://localhost:8080
RabbitMQ 管理台：http://localhost:15672
```

RabbitMQ 管理台默认账号密码：

```text
guest / guest
```

停止服务：

```bash
docker compose down
```

如需清空数据库和中间件数据：

```bash
docker compose down -v
```

也可以手动运行本地开发环境：

1. 创建 MySQL 数据库：

   ```sql
   CREATE DATABASE pulseroad DEFAULT CHARACTER SET utf8mb4;
   ```

2. 修改 `internal/pkg/config/config.yaml` 中的 MySQL DSN，并确认 Redis、RabbitMQ 地址可连接。

3. 分别启动迁移、API、worker 和前端：

   ```bash
   go run ./cmd/migrate
   go run ./cmd/api
   go run ./cmd/worker
   ./scripts/dev-web.sh
   ```

## API 概览

### 认证

```http
POST /api/auth/register
POST /api/auth/login
GET  /api/auth/me
```

### 团队

```http
POST   /api/teams
GET    /api/teams
GET    /api/teams/:id
GET    /api/teams/invitations
POST   /api/teams/invitations/:id/accept
GET    /api/teams/:id/members
POST   /api/teams/:id/invitations
PATCH  /api/teams/:id/members/:user_id/role
DELETE /api/teams/:id/members/:user_id
```

### 产品

```http
POST /api/teams/:team_id/products
GET  /api/teams/:team_id/products
GET  /api/products/:id
GET  /api/products/:id/summary
```

### 反馈

```http
POST  /api/products/:product_id/feedback
GET   /api/products/:product_id/feedback?page=1&page_size=20&status=open
GET   /api/feedback/:id
PATCH /api/feedback/:id/status
POST  /api/feedback/:id/comments
GET   /api/feedback/:id/comments
POST  /api/feedback/:id/vote
DELETE /api/feedback/:id/vote
```

### 功能开关

```http
POST  /api/products/:product_id/flags
GET   /api/products/:product_id/flags
GET   /api/flags/:id
PATCH /api/flags/:id
PATCH /api/flags/:id/toggle
POST  /api/flags/evaluate
```

受保护接口需要：

```http
Authorization: Bearer <token>
```

统一响应格式：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

## 测试

运行后端测试：

```bash
go test ./...
```

运行静态检查：

```bash
go vet ./...
```

运行前端测试和构建：

```bash
npm --prefix web test -- --run
npm --prefix web run build
```

## 当前边界

- 没有需求、路线图和发布日志模块。
- 数据库迁移使用 Gorm `AutoMigrate`，没有版本化迁移文件。

这个版本的目标是让你先完整理解后端基础结构、核心权限链路、产品反馈流和功能开关流，再逐步扩展业务功能。
