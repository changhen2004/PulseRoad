# PulseRoad 开发备忘

## 当前业务主线

```text
用户 -> 团队 -> 产品 -> 反馈
用户 -> 团队 -> 产品 -> 功能开关
```

## 已完成模块

### 基础设施
- Gin HTTP API + 请求日志中间件
- MySQL + Gorm 自动迁移
- Redis：登录失败限制 + 功能开关缓存
- RabbitMQ：反馈/开关事件发布与 Worker 消费
- 配置文件 + 环境变量覆盖
- Docker Compose 开发环境
- 后端测试覆盖 service、handler、publisher、cache、中间件

### 认证
- 注册、登录、JWT 签发与解析
- bcrypt 密码哈希
- 登录失败限流

### 团队与成员
- 创建团队（创建者为 owner）
- 邀请成员加入、接受邀请
- 成员列表、角色调整、移除成员
- 保护最后一个 owner 不被降级或移除

### 产品
- 团队下创建/查看产品
- 产品详情聚合摘要（反馈、评论、投票、开关数量）

### 反馈
- 产品下创建/查看/筛选反馈
- 反馈状态：open / resolved
- 评论和投票

### 功能开关
- 产品下创建/编辑/启停开关
- 按 user_key 灰度命中计算
- Redis 缓存 + RabbitMQ 事件发布

## 待办

- 版本化数据库迁移
- Token 刷新、登出、黑名单
- 邀请过期和撤销
- 产品编辑和归档
- 反馈优先级、标签、状态流转扩展（planned/in_progress/closed）
- 开关多环境配置、规则引擎、变更历史
- 通知、审计日志、数据看板
- 路线图、发布日志

## 维护规范

- 新增模块沿用 `model → repository → service → handler → router` 分层
- 涉及团队、产品、反馈和开关的数据访问必须先走成员权限校验
- Redis 和 RabbitMQ 依赖保持可替换接口
- 每个模块至少覆盖 service 核心规则和 handler HTTP 行为
