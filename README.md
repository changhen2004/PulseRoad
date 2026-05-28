# PulseRoad

PulseRoad 是一个精简版产品反馈管理项目。当前项目只保留最核心、最容易理解的一条业务线：

```text
用户 -> 团队 -> 产品
```

后续反馈、需求、路线图、发布日志、功能开关等模块暂不实现，避免项目过早变复杂。

## 已实现功能

- 用户注册、登录、获取当前用户。
- JWT 登录态校验。
- 创建团队、查看我的团队、查看团队详情。
- 创建者自动成为团队 `owner`。
- 在团队下创建产品、查看团队产品、查看产品详情。
- 团队和产品接口都带成员权限校验。
- MySQL 连接和 Gorm 自动迁移。
- Redis 配置预留。
- RabbitMQ 配置预留和 Worker 骨架。
- Vue 前端工作台，对接当前 RESTful API。

## 技术栈

- Go 1.25.6
- Gin
- Gorm
- MySQL
- JWT
- bcrypt
- Redis（配置预留）
- RabbitMQ（配置预留）
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
│   └── worker    # Worker 骨架，预留 RabbitMQ 消费能力
├── docs
│   └── development-tasks.md
├── internal
│   ├── auth       # 注册、登录、JWT、当前用户
│   ├── middleware # 登录态中间件
│   ├── pkg        # 配置、数据库、日志、响应、RabbitMQ 工具
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

1. 创建 MySQL 数据库：

   ```sql
   CREATE DATABASE pulseroad DEFAULT CHARACTER SET utf8mb4;
   ```

2. 修改 `internal/pkg/config/config.yaml` 中的 MySQL DSN。

3. 执行迁移：

   ```bash
   go run ./cmd/migrate
   ```

4. 启动 API：

   ```bash
   go run ./cmd/api
   ```

5. 健康检查：

   ```bash
   curl http://localhost:8080/health
   ```

6. 启动前端：

   ```bash
   npm --prefix web install
   ./scripts/dev-web.sh
   ```

   前端默认运行在：

   ```text
   http://localhost:5173
   ```

   Vite 已配置 `/api` 代理到 `http://127.0.0.1:8080`，所以本地开发时需要先启动 API 服务。

Worker 目前只校验 MySQL 和 RabbitMQ 配置，不消费消息：

```bash
go run ./cmd/worker
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
POST /api/teams
GET  /api/teams
GET  /api/teams/:id
```

### 产品

```http
POST /api/teams/:team_id/products
GET  /api/teams/:team_id/products
GET  /api/products/:id
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

- 没有真实 Redis 业务逻辑。
- 没有真实 RabbitMQ 消费者。
- 没有反馈、需求、路线图和发布日志模块。
- 数据库迁移使用 Gorm `AutoMigrate`，没有版本化迁移文件。

这个版本的目标是让你先完整理解后端基础结构和核心权限链路，再逐步扩展业务功能。
