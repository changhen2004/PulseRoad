# PulseRoad

PulseRoad 是一个面向产品团队的反馈与路线图管理后端服务。当前代码已经完成基础设施、用户认证、团队管理和产品管理，后续规划会继续扩展反馈、需求、路线图、发布日志、功能开关、事件和缓存能力。

## 当前能力

- 用户注册、登录和当前用户查询。
- JWT 登录态校验，并将当前用户 ID 注入 Gin Context。
- 团队创建、团队列表和团队详情查询。
- 团队创建者自动成为 `owner`。
- 非团队成员不能访问团队详情。
- 团队下产品创建、产品列表和产品详情查询。
- 只有团队成员可以创建和查看团队产品。
- 统一 API 响应结构、请求日志、MySQL 连接和自动迁移入口。

## 技术栈

- Go 1.25.6
- Gin
- Gorm
- MySQL
- JWT: `github.com/golang-jwt/jwt/v5`
- 密码哈希：`golang.org/x/crypto/bcrypt`

## 项目结构

```text
.
├── cmd
│   ├── api       # HTTP API 服务入口
│   ├── migrate   # Gorm AutoMigrate 迁移入口
│   └── worker    # Worker 进程入口，目前仅完成基础启动骨架
├── docs
│   └── development-tasks.md
├── internal
│   ├── auth       # 用户注册、登录、JWT、当前用户
│   ├── middleware # 认证中间件
│   ├── pkg
│   │   ├── config   # 配置加载
│   │   ├── database # MySQL / Gorm 初始化与迁移注册
│   │   ├── logger   # 请求日志中间件
│   │   └── response # 统一响应结构
│   ├── product    # 团队下产品管理
│   └── team       # 团队与成员管理
├── go.mod
└── go.sum
```

## 业务模型

### User

用户通过邮箱注册和登录。密码不会明文存储，服务端使用 bcrypt 生成 `password_hash`。

核心字段：

- `id`
- `email`
- `name`
- `password_hash`
- `created_at`
- `updated_at`

### Team

团队是产品和后续反馈、需求、路线图的组织边界。

核心字段：

- `id`
- `name`
- `description`
- `created_by`
- `created_at`
- `updated_at`

### TeamMember

团队成员关系用于权限控制。`team_members(team_id, user_id)` 有唯一索引，避免同一用户重复加入同一团队。

核心字段：

- `id`
- `team_id`
- `user_id`
- `role`
- `created_at`

当前已实现角色：

- `owner`

### Product

产品必须归属于一个团队，团队成员才能创建和查看产品。

核心字段：

- `id`
- `team_id`
- `name`
- `description`
- `created_by`
- `created_at`
- `updated_at`

## 配置

默认配置文件路径为：

```text
internal/pkg/config/config.yaml
```

配置项：

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

kafka:
  brokers:
    - "127.0.0.1:9092"

jwt:
  secret: "change-me-in-production"
```

支持通过环境变量覆盖配置：

| 环境变量 | 说明 |
| --- | --- |
| `PULSEROAD_APP_NAME` | 应用名称 |
| `PULSEROAD_APP_ENV` | 运行环境 |
| `PULSEROAD_SERVER_PORT` | HTTP 服务端口 |
| `PULSEROAD_MYSQL_DSN` | MySQL DSN |
| `PULSEROAD_REDIS_ADDR` | Redis 地址 |
| `PULSEROAD_KAFKA_BROKERS` | Kafka broker 列表，逗号分隔 |
| `PULSEROAD_JWT_SECRET` | JWT 签名密钥 |

生产环境注意事项：

- 不要使用 `change-me-in-production` 作为 JWT 密钥。
- `app.env` 非 `development` 时，JWT 密钥不能使用默认占位值，且长度至少为 32 个字符。
- Redis 和 Kafka 目前只完成配置项预留，业务代码尚未接入。

## 数据库迁移

先确保 MySQL 可连接，并已创建 `pulseroad` 数据库。

```bash
go run ./cmd/migrate
```

迁移模型由各模块通过 `database.RegisterModel` 注册，目前包括：

- `auth.User`
- `team.Team`
- `team.TeamMember`
- `product.Product`

## 启动服务

启动 HTTP API：

```bash
go run ./cmd/api
```

默认监听：

```text
:8080
```

健康检查：

```bash
curl http://localhost:8080/health
```

启动 Worker：

```bash
go run ./cmd/worker
```

当前 Worker 只完成基础进程骨架和数据库连接校验，后台任务尚未实现。

## API 响应格式

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

失败响应：

```json
{
  "code": 401,
  "message": "unauthorized"
}
```

受保护接口需要携带：

```http
Authorization: Bearer <token>
```

## API 列表

### 健康检查

```http
GET /health
```

### 用户认证

#### 注册

```http
POST /api/auth/register
Content-Type: application/json
```

请求示例：

```json
{
  "email": "ada@example.com",
  "password": "password123",
  "name": "Ada"
}
```

#### 登录

```http
POST /api/auth/login
Content-Type: application/json
```

请求示例：

```json
{
  "email": "ada@example.com",
  "password": "password123"
}
```

响应中的 `data.token` 用于后续受保护接口。

#### 当前用户

```http
GET /api/auth/me
Authorization: Bearer <token>
```

### 团队管理

#### 创建团队

```http
POST /api/teams
Authorization: Bearer <token>
Content-Type: application/json
```

请求示例：

```json
{
  "name": "Core Team",
  "description": "Product and roadmap team"
}
```

创建成功后，当前用户会自动成为团队 `owner`。

#### 查看我的团队

```http
GET /api/teams
Authorization: Bearer <token>
```

#### 查看团队详情

```http
GET /api/teams/:id
Authorization: Bearer <token>
```

非团队成员访问会返回 403。

### 产品管理

#### 创建产品

```http
POST /api/teams/:team_id/products
Authorization: Bearer <token>
Content-Type: application/json
```

请求示例：

```json
{
  "name": "PulseRoad",
  "description": "Feedback and roadmap platform"
}
```

只有团队成员可以创建产品。

#### 查看团队产品列表

```http
GET /api/teams/:team_id/products
Authorization: Bearer <token>
```

#### 查看产品详情

```http
GET /api/products/:id
Authorization: Bearer <token>
```

只有产品所属团队的成员可以查看。

## 本地开发建议流程

1. 修改 `internal/pkg/config/config.yaml` 或设置 `PULSEROAD_*` 环境变量。
2. 启动 MySQL，并创建 `pulseroad` 数据库。
3. 执行迁移：

   ```bash
   go run ./cmd/migrate
   ```

4. 启动 API：

   ```bash
   go run ./cmd/api
   ```

5. 运行测试：

   ```bash
   go test ./...
   ```

6. 运行静态检查：

   ```bash
   go vet ./...
   ```

## 测试

当前已有单元测试覆盖：

- 用户注册、bcrypt 密码哈希、登录、JWT 解析和 `/me`。
- JWT 认证中间件的缺失、非法、过期和有效 token 行为。
- 团队创建、成员列表、成员详情权限和唯一成员约束。
- 产品创建、团队成员权限、产品列表和产品详情权限。

运行：

```bash
go test ./...
```

禁用缓存运行：

```bash
go test -count=1 ./...
```

## 已知边界

- 当前没有独立的版本化迁移工具，迁移入口使用 Gorm `AutoMigrate`。
- Redis 和 Kafka 目前只在配置中预留，业务逻辑尚未使用。
- Worker 目前只完成启动骨架，尚未接入后台任务。
- 配置文件路径在入口中仍是固定相对路径 `internal/pkg/config/config.yaml`。
- 生产部署前需要替换 JWT 密钥，并使用安全的 MySQL 账号和密码。

## 后续规划

详见 [开发任务拆分](docs/development-tasks.md)。后续计划包括：

- 反馈提交与列表。
- 评论与投票。
- 反馈转需求。
- 路线图管理。
- 发布日志。
- 功能开关和灰度策略。
- Kafka 事件和 Redis 缓存、限流、热榜。
