# 反馈模块设计规格

## 目标

在当前 `用户 -> 团队 -> 产品` 主线之后，新增一个最小可用的反馈模块。用户在产品详情页下创建和查看反馈，团队成员可以把反馈标记为已解决。实现重点是保持项目容易理解，并保证已有认证、团队、产品功能继续正常运行。

## 范围

本次只实现产品内嵌反馈流：

- 团队成员在产品下创建反馈。
- 团队成员查看产品下的反馈列表。
- 团队成员查看反馈详情。
- 团队成员将反馈状态从 `open` 标记为 `resolved`。
- 前端在产品详情页内嵌反馈区域。

本次不实现：

- 外部用户公开提交反馈。
- 评论。
- 投票。
- 分类、优先级、标签。
- 团队级反馈收件箱。
- RabbitMQ 事件发布。

## 数据模型

新增 `internal/feedback` 模块，模型为 `Feedback`：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `uint` | 主键 |
| `product_id` | `uint` | 所属产品 |
| `title` | `string` | 反馈标题，必填 |
| `content` | `string` | 反馈内容，必填 |
| `status` | `string` | `open` 或 `resolved` |
| `created_by` | `uint` | 创建人用户 ID |
| `created_at` | `time.Time` | 创建时间 |
| `updated_at` | `time.Time` | 更新时间 |

默认状态为 `open`。状态只允许 `open` 和 `resolved`。

## 后端接口

所有接口都需要登录态。

```http
POST  /api/products/:product_id/feedback
GET   /api/products/:product_id/feedback
GET   /api/feedback/:id
PATCH /api/feedback/:id/status
```

### 创建反馈

```http
POST /api/products/:product_id/feedback
```

请求：

```json
{
  "title": "希望支持路线图视图",
  "content": "当前只能管理产品，希望后续能看到路线图。"
}
```

成功返回反馈详情，状态为 `open`。

### 查看产品反馈列表

```http
GET /api/products/:product_id/feedback
```

返回该产品下的反馈列表，按创建时间倒序。

### 查看反馈详情

```http
GET /api/feedback/:id
```

返回单条反馈。

### 更新反馈状态

```http
PATCH /api/feedback/:id/status
```

请求：

```json
{
  "status": "resolved"
}
```

本次允许将状态设置为 `open` 或 `resolved`，方便误操作后恢复。非法状态返回 400。

## 权限规则

反馈继承产品的团队权限。

- 创建反馈：当前用户必须是产品所属团队成员。
- 查看反馈列表：当前用户必须是产品所属团队成员。
- 查看反馈详情：当前用户必须是反馈所属产品的团队成员。
- 更新反馈状态：当前用户必须是反馈所属产品的团队成员。

非成员访问返回 403，找不到产品或反馈返回 404。错误响应沿用当前统一响应结构。

## 后端结构

新增目录：

```text
internal/feedback
```

文件职责：

- `model.go`：定义 `Feedback`、状态常量和响应结构，并注册迁移模型。
- `repository.go`：封装反馈创建、列表、详情、状态更新查询。
- `service.go`：校验输入、校验产品团队成员权限、执行业务逻辑。
- `handler.go`：解析 HTTP 请求和参数，转换错误响应。
- `router.go`：注册反馈接口。

依赖关系：

- `feedback.Service` 依赖 `feedback.Repository`。
- `feedback.Service` 依赖一个产品读取和成员校验能力，用于确认产品存在并判断用户是否能访问。
- 复用现有 `product.Service` 或抽象最小接口，避免复制团队成员校验逻辑。

## 前端设计

在产品详情页中增加反馈区域：

- 反馈列表：标题、状态、创建时间。
- 创建反馈：标题、内容。
- 反馈详情：点击列表项后在当前页面或抽屉中显示。
- 状态操作：`open` 状态显示“标记已解决”，`resolved` 状态显示“重新打开”。

前端 API 新增：

```text
web/src/api/feedback.ts
```

前端类型扩展：

```text
Feedback
- id
- product_id
- title
- content
- status
- created_by
- created_at
- updated_at
```

## 测试策略

后端测试：

- 团队成员可以在产品下创建反馈。
- 非团队成员不能创建反馈。
- 团队成员可以查看产品反馈列表。
- 非团队成员不能查看产品反馈列表。
- 团队成员可以查看反馈详情。
- 团队成员可以把反馈标记为 `resolved`。
- 非法状态返回 400。
- 路由注册不与已有 `/api/products/:id` 冲突。

前端测试：

- API 客户端能正确调用反馈接口。
- 原有认证 token 注入和 401 处理继续通过。

回归验证：

```bash
go test -count=1 ./...
go vet ./...
npm --prefix web test -- --run
npm --prefix web run build
```

## 迁移与运行

`Feedback` 模型需要注册到 Gorm `AutoMigrate`。新增模块后，`cmd/migrate` 必须 blank import `internal/feedback`，否则 `feedbacks` 表不会创建。

实现完成后需要执行：

```bash
go run ./cmd/migrate
```

然后启动：

```bash
go run ./cmd/api
./scripts/dev-web.sh
```
