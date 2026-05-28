# 反馈模块实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实现此计划。步骤使用复选框（`- [ ]`）语法来跟踪进度。

**目标：** 在产品详情下新增最小可用反馈流，团队成员可以创建、查看和处理产品反馈。

**架构：** 后端新增 `internal/feedback`，继续沿用 `model -> repository -> service -> handler -> router` 分层；反馈权限通过产品访问能力复用现有产品团队成员校验。前端在 `ProductDetailView` 内嵌反馈列表、创建表单和详情抽屉，通过新的 `feedbackApi` 调用 RESTful 接口。

**技术栈：** Go、Gin、Gorm、MySQL、Vue 3、TypeScript、Naive UI、Axios、Vitest。

---

## 文件结构

- 创建：`internal/feedback/model.go`  
  定义 `Feedback`、`FeedbackResponse`、状态常量和 `database.RegisterModel(&Feedback{})`。
- 创建：`internal/feedback/repository.go`  
  实现 Gorm 查询：创建反馈、按产品倒序列表、按 ID 查询、更新状态。
- 创建：`internal/feedback/service.go`  
  校验输入和状态，调用产品访问接口确认当前用户是产品所属团队成员。
- 创建：`internal/feedback/service_test.go`  
  用 fake repository 和 fake product access 覆盖核心权限和状态规则。
- 创建：`internal/feedback/handler.go`  
  解析请求、参数和错误，返回统一响应。
- 创建：`internal/feedback/router.go`  
  注册反馈路由。注意 Gin 通配符冲突，产品下反馈路由内部使用 `/products/:id/feedback`，语义上这个 `id` 是 `product_id`。
- 创建：`internal/feedback/handler_test.go`  
  覆盖 HTTP 创建、列表、详情、状态更新、401、403。
- 修改：`cmd/api/main.go`  
  初始化 feedback repository/service 并注册路由。
- 修改：`cmd/api/routes_test.go`  
  把 feedback 路由加入路由注册防 panic 回归测试。
- 修改：`cmd/migrate/main.go`  
  blank import `pulseroad/internal/feedback`，确保 `feedbacks` 表自动迁移。
- 修改：`cmd/migrate/main_test.go`  
  将注册模型数量断言提升到至少 5 个模型。
- 修改：`web/src/api/http.ts`、`web/src/api/http.test.ts`  
  增加 `PATCH` 支持。
- 修改：`web/src/api/types.ts`  
  增加 `Feedback`、`CreateFeedbackPayload`、`UpdateFeedbackStatusPayload`。
- 创建：`web/src/api/feedback.ts`、`web/src/api/feedback.test.ts`  
  封装反馈接口并测试 URL、方法和 payload。
- 修改：`web/src/views/ProductDetailView.vue`  
  在产品详情页嵌入反馈列表、创建抽屉、详情抽屉和状态按钮。
- 修改：`web/src/styles.css`  
  增加反馈列表和详情区域所需样式。
- 修改：`README.md`、`docs/development-tasks.md`  
  记录反馈模块接口和运行迁移要求。

## 实现注意

- 产品下反馈路由不要写成 `/products/:product_id/feedback`，否则会和已有 `/products/:id` 再次触发 Gin wildcard 名称冲突。使用 `/products/:id/feedback`，handler 中把 `id` 当作产品 ID。
- `feedback.Service` 不直接依赖 team 包，避免复制成员权限逻辑；它依赖最小接口 `ProductAccess`，实际注入 `product.Service`。
- `product.Service.GetProduct(ctx, userID, productID)` 已经完成产品存在性和团队成员校验，feedback 服务把 `product.ErrForbidden` 映射为 `feedback.ErrForbidden`，把 `product.ErrProductNotFound` 映射为 `feedback.ErrProductNotFound`。
- 状态只允许 `open` 和 `resolved`，创建时强制 `open`。

---

### 任务 1：反馈服务领域规则

**文件：**
- 创建：`internal/feedback/service_test.go`
- 创建：`internal/feedback/model.go`
- 创建：`internal/feedback/service.go`

- [ ] **步骤 1：编写失败的服务测试**

创建 `internal/feedback/service_test.go`：

```go
package feedback

import (
	"context"
	"errors"
	"testing"
	"time"

	"pulseroad/internal/product"
)

type fakeFeedbackRepository struct {
	nextID    uint
	feedback map[uint]*Feedback
}

func newFakeFeedbackRepository() *fakeFeedbackRepository {
	return &fakeFeedbackRepository{nextID: 1, feedback: make(map[uint]*Feedback)}
}

func (r *fakeFeedbackRepository) Create(_ context.Context, item *Feedback) error {
	item.ID = r.nextID
	item.CreatedAt = time.Now()
	item.UpdatedAt = item.CreatedAt
	r.nextID++
	copy := *item
	r.feedback[item.ID] = &copy
	return nil
}

func (r *fakeFeedbackRepository) ListByProduct(_ context.Context, productID uint) ([]Feedback, error) {
	items := make([]Feedback, 0)
	for _, item := range r.feedback {
		if item.ProductID == productID {
			items = append(items, *item)
		}
	}
	return items, nil
}

func (r *fakeFeedbackRepository) FindByID(_ context.Context, id uint) (*Feedback, error) {
	item, ok := r.feedback[id]
	if !ok {
		return nil, ErrFeedbackNotFound
	}
	copy := *item
	return &copy, nil
}

func (r *fakeFeedbackRepository) UpdateStatus(_ context.Context, id uint, status string) error {
	item, ok := r.feedback[id]
	if !ok {
		return ErrFeedbackNotFound
	}
	item.Status = status
	item.UpdatedAt = time.Now()
	return nil
}

type fakeProductAccess struct {
	products map[uint]product.ProductResponse
	denied   map[uint]bool
}

func (a fakeProductAccess) GetProduct(_ context.Context, userID uint, productID uint) (*product.ProductResponse, error) {
	if a.denied[userID] {
		return nil, product.ErrForbidden
	}
	item, ok := a.products[productID]
	if !ok {
		return nil, product.ErrProductNotFound
	}
	return &item, nil
}

func newTestService() (*Service, *fakeFeedbackRepository) {
	repo := newFakeFeedbackRepository()
	access := fakeProductAccess{
		products: map[uint]product.ProductResponse{10: {ID: 10, TeamID: 3, Name: "PulseRoad"}},
		denied:   map[uint]bool{8: true},
	}
	return NewService(repo, access), repo
}

func TestCreateFeedbackRequiresProductMember(t *testing.T) {
	svc, _ := newTestService()
	created, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{
		Title:   "Need roadmap",
		Content: "Please add a roadmap view.",
	})
	if err != nil {
		t.Fatalf("create feedback: %v", err)
	}
	if created.ID == 0 || created.ProductID != 10 || created.Status != StatusOpen || created.CreatedBy != 7 {
		t.Fatalf("unexpected feedback: %#v", created)
	}

	_, err = svc.CreateFeedback(context.Background(), 8, 10, CreateFeedbackInput{Title: "No", Content: "Denied"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestListAndGetFeedbackRequireProductMember(t *testing.T) {
	svc, _ := newTestService()
	created, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{Title: "A", Content: "B"})
	if err != nil {
		t.Fatalf("create feedback: %v", err)
	}

	list, err := svc.ListFeedback(context.Background(), 7, 10)
	if err != nil {
		t.Fatalf("list feedback: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("unexpected list: %#v", list)
	}

	got, err := svc.GetFeedback(context.Background(), 7, created.ID)
	if err != nil {
		t.Fatalf("get feedback: %v", err)
	}
	if got.ID != created.ID || got.Title != "A" {
		t.Fatalf("unexpected detail: %#v", got)
	}

	_, err = svc.GetFeedback(context.Background(), 8, created.ID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdateFeedbackStatusValidatesStatus(t *testing.T) {
	svc, _ := newTestService()
	created, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{Title: "A", Content: "B"})
	if err != nil {
		t.Fatalf("create feedback: %v", err)
	}

	updated, err := svc.UpdateStatus(context.Background(), 7, created.ID, UpdateFeedbackStatusInput{Status: StatusResolved})
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if updated.Status != StatusResolved {
		t.Fatalf("expected resolved, got %#v", updated)
	}

	_, err = svc.UpdateStatus(context.Background(), 7, created.ID, UpdateFeedbackStatusInput{Status: "closed"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：

```bash
go test -count=1 ./internal/feedback
```

预期：FAIL，报错包含 `undefined: Feedback`、`undefined: Service` 或包还没有实现。

- [ ] **步骤 3：编写模型和服务最小实现**

创建 `internal/feedback/model.go`：

```go
package feedback

import (
	"time"

	"pulseroad/internal/pkg/database"
)

const (
	StatusOpen     = "open"
	StatusResolved = "resolved"
)

type Feedback struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProductID uint      `json:"product_id" gorm:"not null;index"`
	Title     string    `json:"title" gorm:"type:varchar(160);not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Status    string    `json:"status" gorm:"type:varchar(32);not null;index"`
	CreatedBy uint      `json:"created_by" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FeedbackResponse struct {
	ID        uint      `json:"id"`
	ProductID uint      `json:"product_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedBy uint      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (f Feedback) ToResponse() FeedbackResponse {
	return FeedbackResponse{
		ID: f.ID, ProductID: f.ProductID, Title: f.Title, Content: f.Content,
		Status: f.Status, CreatedBy: f.CreatedBy, CreatedAt: f.CreatedAt, UpdatedAt: f.UpdatedAt,
	}
}

func init() {
	database.RegisterModel(&Feedback{})
}
```

创建 `internal/feedback/service.go`：

```go
package feedback

import (
	"context"
	"errors"
	"strings"

	"pulseroad/internal/product"
)

var (
	ErrForbidden       = errors.New("forbidden")
	ErrInvalid         = errors.New("invalid input")
	ErrFeedbackNotFound = errors.New("feedback not found")
	ErrProductNotFound = errors.New("product not found")
)

type CreateFeedbackInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdateFeedbackStatusInput struct {
	Status string `json:"status"`
}

type RepositoryPort interface {
	Create(ctx context.Context, item *Feedback) error
	ListByProduct(ctx context.Context, productID uint) ([]Feedback, error)
	FindByID(ctx context.Context, id uint) (*Feedback, error)
	UpdateStatus(ctx context.Context, id uint, status string) error
}

type ProductAccess interface {
	GetProduct(ctx context.Context, userID uint, productID uint) (*product.ProductResponse, error)
}

type Service struct {
	repo    RepositoryPort
	product ProductAccess
}

func NewService(repo RepositoryPort, productAccess ProductAccess) *Service {
	return &Service{repo: repo, product: productAccess}
}

func (s *Service) CreateFeedback(ctx context.Context, userID uint, productID uint, input CreateFeedbackInput) (*FeedbackResponse, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if userID == 0 || productID == 0 || title == "" || content == "" {
		return nil, ErrInvalid
	}
	if err := s.requireProductAccess(ctx, userID, productID); err != nil {
		return nil, err
	}
	item := &Feedback{ProductID: productID, Title: title, Content: content, Status: StatusOpen, CreatedBy: userID}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	response := item.ToResponse()
	return &response, nil
}

func (s *Service) ListFeedback(ctx context.Context, userID uint, productID uint) ([]FeedbackResponse, error) {
	if userID == 0 || productID == 0 {
		return nil, ErrForbidden
	}
	if err := s.requireProductAccess(ctx, userID, productID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	response := make([]FeedbackResponse, 0, len(items))
	for _, item := range items {
		response = append(response, item.ToResponse())
	}
	return response, nil
}

func (s *Service) GetFeedback(ctx context.Context, userID uint, feedbackID uint) (*FeedbackResponse, error) {
	item, err := s.feedbackForUser(ctx, userID, feedbackID)
	if err != nil {
		return nil, err
	}
	response := item.ToResponse()
	return &response, nil
}

func (s *Service) UpdateStatus(ctx context.Context, userID uint, feedbackID uint, input UpdateFeedbackStatusInput) (*FeedbackResponse, error) {
	status := strings.TrimSpace(input.Status)
	if status != StatusOpen && status != StatusResolved {
		return nil, ErrInvalid
	}
	item, err := s.feedbackForUser(ctx, userID, feedbackID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateStatus(ctx, item.ID, status); err != nil {
		return nil, err
	}
	item.Status = status
	response := item.ToResponse()
	return &response, nil
}

func (s *Service) feedbackForUser(ctx context.Context, userID uint, feedbackID uint) (*Feedback, error) {
	if userID == 0 || feedbackID == 0 {
		return nil, ErrForbidden
	}
	item, err := s.repo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, err
	}
	if err := s.requireProductAccess(ctx, userID, item.ProductID); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) requireProductAccess(ctx context.Context, userID uint, productID uint) error {
	if _, err := s.product.GetProduct(ctx, userID, productID); err != nil {
		switch {
		case errors.Is(err, product.ErrForbidden):
			return ErrForbidden
		case errors.Is(err, product.ErrProductNotFound):
			return ErrProductNotFound
		default:
			return err
		}
	}
	return nil
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：

```bash
go test -count=1 ./internal/feedback
```

预期：PASS。

- [ ] **步骤 5：Commit**

```bash
git add internal/feedback/model.go internal/feedback/service.go internal/feedback/service_test.go
git commit -m "feat: add feedback service rules"
```

---

### 任务 2：反馈 HTTP 接口

**文件：**
- 创建：`internal/feedback/handler_test.go`
- 创建：`internal/feedback/handler.go`
- 创建：`internal/feedback/router.go`
- 创建：`internal/feedback/repository.go`

- [ ] **步骤 1：编写失败的 HTTP 测试**

创建 `internal/feedback/handler_test.go`：

```go
package feedback

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"

	"pulseroad/internal/product"
)

type staticTokenParser struct{ users map[string]uint }

func (p staticTokenParser) ParseToken(token string) (uint, error) {
	userID, ok := p.users[token]
	if !ok {
		return 0, errors.New("invalid token")
	}
	return userID, nil
}

func newFeedbackTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	repo := newFakeFeedbackRepository()
	access := fakeProductAccess{
		products: map[uint]product.ProductResponse{10: {ID: 10, TeamID: 3, Name: "PulseRoad"}},
		denied:   map[uint]bool{8: true},
	}
	svc := NewService(repo, access)
	parser := staticTokenParser{users: map[string]uint{"user-7": 7, "user-8": 8}}

	r := gin.New()
	RegisterRoutes(r.Group("/api"), parser, svc)
	return r
}

func performFeedbackJSON(r http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			panic(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeFeedbackResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	return payload
}

func TestFeedbackHTTPCreateListGetAndResolve(t *testing.T) {
	r := newFeedbackTestRouter()

	createResp := performFeedbackJSON(r, http.MethodPost, "/api/products/10/feedback", gin.H{
		"title": "Need roadmap", "content": "Please add a roadmap view.",
	}, "user-7")
	if createResp.Code != http.StatusOK {
		t.Fatalf("create status = %d, body = %s", createResp.Code, createResp.Body.String())
	}
	created := decodeFeedbackResponse(t, createResp)["data"].(map[string]any)
	if created["product_id"] != float64(10) || created["status"] != StatusOpen {
		t.Fatalf("unexpected create response: %#v", created)
	}
	feedbackID := uint(created["id"].(float64))

	listResp := performFeedbackJSON(r, http.MethodGet, "/api/products/10/feedback", nil, "user-7")
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listResp.Code, listResp.Body.String())
	}
	list := decodeFeedbackResponse(t, listResp)["data"].([]any)
	if len(list) != 1 {
		t.Fatalf("expected one feedback, got %#v", list)
	}

	getResp := performFeedbackJSON(r, http.MethodGet, "/api/feedback/"+strconvID(feedbackID), nil, "user-7")
	if getResp.Code != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", getResp.Code, getResp.Body.String())
	}

	updateResp := performFeedbackJSON(r, http.MethodPatch, "/api/feedback/"+strconvID(feedbackID)+"/status", gin.H{
		"status": StatusResolved,
	}, "user-7")
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", updateResp.Code, updateResp.Body.String())
	}
	updated := decodeFeedbackResponse(t, updateResp)["data"].(map[string]any)
	if updated["status"] != StatusResolved {
		t.Fatalf("expected resolved, got %#v", updated)
	}
}

func TestFeedbackHTTPRequiresAuthentication(t *testing.T) {
	r := newFeedbackTestRouter()
	w := performFeedbackJSON(r, http.MethodGet, "/api/products/10/feedback", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d with body %s", w.Code, w.Body.String())
	}
}

func TestFeedbackHTTPRejectsNonMember(t *testing.T) {
	r := newFeedbackTestRouter()
	w := performFeedbackJSON(r, http.MethodPost, "/api/products/10/feedback", gin.H{
		"title": "Denied", "content": "Denied",
	}, "user-8")
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d with body %s", w.Code, w.Body.String())
	}
}

func TestFeedbackHTTPRejectsInvalidStatus(t *testing.T) {
	r := newFeedbackTestRouter()
	createResp := performFeedbackJSON(r, http.MethodPost, "/api/products/10/feedback", gin.H{
		"title": "Need roadmap", "content": "Please add a roadmap view.",
	}, "user-7")
	created := decodeFeedbackResponse(t, createResp)["data"].(map[string]any)
	feedbackID := uint(created["id"].(float64))

	w := performFeedbackJSON(r, http.MethodPatch, "/api/feedback/"+strconvID(feedbackID)+"/status", gin.H{
		"status": "closed",
	}, "user-7")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d with body %s", w.Code, w.Body.String())
	}
}

func strconvID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：

```bash
go test -count=1 ./internal/feedback
```

预期：FAIL，报错包含 `undefined: RegisterRoutes` 或 `undefined: NewHandler`。

- [ ] **步骤 3：实现 handler、router、repository**

创建 `internal/feedback/router.go`：

```go
package feedback

import (
	"github.com/gin-gonic/gin"

	"pulseroad/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, parser middleware.TokenParser, service *Service) {
	handler := NewHandler(service)
	auth := r.Group("", middleware.AuthRequired(parser))
	auth.POST("/products/:id/feedback", handler.Create)
	auth.GET("/products/:id/feedback", handler.ListByProduct)
	auth.GET("/feedback/:id", handler.Get)
	auth.PATCH("/feedback/:id/status", handler.UpdateStatus)
}
```

创建 `internal/feedback/handler.go`，结构与 product handler 保持一致：

```go
package feedback

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"pulseroad/internal/middleware"
	"pulseroad/internal/pkg/response"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok { response.Unauthorized(c, "unauthorized"); return }
	productID, ok := parseUintParam(c, "id")
	if !ok { response.BadRequest(c, "invalid product id"); return }
	var input CreateFeedbackInput
	if err := c.ShouldBindJSON(&input); err != nil { response.BadRequest(c, "invalid request body"); return }
	item, err := h.service.CreateFeedback(c.Request.Context(), userID, productID, input)
	if err != nil { h.writeError(c, err); return }
	response.Success(c, item)
}

func (h *Handler) ListByProduct(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok { response.Unauthorized(c, "unauthorized"); return }
	productID, ok := parseUintParam(c, "id")
	if !ok { response.BadRequest(c, "invalid product id"); return }
	items, err := h.service.ListFeedback(c.Request.Context(), userID, productID)
	if err != nil { h.writeError(c, err); return }
	response.Success(c, items)
}

func (h *Handler) Get(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok { response.Unauthorized(c, "unauthorized"); return }
	feedbackID, ok := parseUintParam(c, "id")
	if !ok { response.BadRequest(c, "invalid feedback id"); return }
	item, err := h.service.GetFeedback(c.Request.Context(), userID, feedbackID)
	if err != nil { h.writeError(c, err); return }
	response.Success(c, item)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok { response.Unauthorized(c, "unauthorized"); return }
	feedbackID, ok := parseUintParam(c, "id")
	if !ok { response.BadRequest(c, "invalid feedback id"); return }
	var input UpdateFeedbackStatusInput
	if err := c.ShouldBindJSON(&input); err != nil { response.BadRequest(c, "invalid request body"); return }
	item, err := h.service.UpdateStatus(c.Request.Context(), userID, feedbackID, input)
	if err != nil { h.writeError(c, err); return }
	response.Success(c, item)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrForbidden):
		response.Fail(c, 403, 403, "forbidden")
	case errors.Is(err, ErrFeedbackNotFound):
		response.Fail(c, 404, 404, "feedback not found")
	case errors.Is(err, ErrProductNotFound):
		response.Fail(c, 404, 404, "product not found")
	default:
		response.InternalError(c, "internal server error")
	}
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 { return 0, false }
	return uint(value), true
}
```

创建 `internal/feedback/repository.go`：

```go
package feedback

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, item *Feedback) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *Repository) ListByProduct(ctx context.Context, productID uint) ([]Feedback, error) {
	var items []Feedback
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*Feedback, error) {
	var item Feedback
	err := r.db.WithContext(ctx).First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFeedbackNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db.WithContext(ctx).Model(&Feedback{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFeedbackNotFound
	}
	return nil
}
```

- [ ] **步骤 4：运行测试验证通过**

运行：

```bash
go test -count=1 ./internal/feedback
```

预期：PASS。

- [ ] **步骤 5：Commit**

```bash
git add internal/feedback
git commit -m "feat: add feedback HTTP API"
```

---

### 任务 3：接入 API、迁移和路由回归

**文件：**
- 修改：`cmd/api/main.go`
- 修改：`cmd/api/routes_test.go`
- 修改：`cmd/migrate/main.go`
- 修改：`cmd/migrate/main_test.go`

- [ ] **步骤 1：编写失败的接入测试**

修改 `cmd/api/routes_test.go`，加入 feedback 路由注册：

```go
import (
	"testing"

	"github.com/gin-gonic/gin"

	"pulseroad/internal/feedback"
	"pulseroad/internal/product"
	"pulseroad/internal/team"
)

func TestProtectedRoutesRegisterWithoutPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	parser := testTokenParser{}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("register protected routes panic: %v", recovered)
		}
	}()

	team.RegisterRoutes(r.Group("/api"), parser, nil)
	product.RegisterRoutes(r.Group("/api"), parser, nil)
	feedback.RegisterRoutes(r.Group("/api"), parser, nil)
}
```

修改 `cmd/migrate/main_test.go`，确认迁移命令注册 5 个模型：

```go
func TestMigrateCommandRegistersApplicationModels(t *testing.T) {
	if got := database.RegisteredModelCount(); got < 5 {
		t.Fatalf("expected at least 5 registered application models, got %d", got)
	}
}
```

- [ ] **步骤 2：运行测试验证失败**

运行：

```bash
go test -count=1 ./cmd/api ./cmd/migrate
```

预期：`cmd/migrate` FAIL，模型数量仍是 4；如果 feedback 路由内部错误使用 `:product_id`，`cmd/api` 会 panic。

- [ ] **步骤 3：接入 main 和 migrate**

修改 `cmd/api/main.go`：

```go
import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"pulseroad/internal/auth"
	"pulseroad/internal/feedback"
	"pulseroad/internal/pkg/config"
	"pulseroad/internal/pkg/database"
	"pulseroad/internal/pkg/logger"
	"pulseroad/internal/pkg/response"
	"pulseroad/internal/product"
	"pulseroad/internal/team"
)
```

在 `product.RegisterRoutes(...)` 后添加：

```go
feedbackService := feedback.NewService(feedback.NewRepository(db), productService)
feedback.RegisterRoutes(r.Group("/api"), authService, feedbackService)
```

修改 `cmd/migrate/main.go`：

```go
import (
	"log"

	_ "pulseroad/internal/auth"
	_ "pulseroad/internal/feedback"
	"pulseroad/internal/pkg/config"
	"pulseroad/internal/pkg/database"
	_ "pulseroad/internal/product"
	_ "pulseroad/internal/team"
)
```

- [ ] **步骤 4：运行测试验证通过**

运行：

```bash
go test -count=1 ./cmd/api ./cmd/migrate
```

预期：PASS。

- [ ] **步骤 5：Commit**

```bash
git add cmd/api/main.go cmd/api/routes_test.go cmd/migrate/main.go cmd/migrate/main_test.go
git commit -m "feat: wire feedback module"
```

---

### 任务 4：前端反馈 API

**文件：**
- 修改：`web/src/api/http.ts`
- 修改：`web/src/api/http.test.ts`
- 修改：`web/src/api/types.ts`
- 创建：`web/src/api/feedback.ts`
- 创建：`web/src/api/feedback.test.ts`

- [ ] **步骤 1：编写失败的前端 API 测试**

在 `web/src/api/http.test.ts` 中增加 PATCH 测试：

```ts
it('supports patch requests', async () => {
  const adapter = responseAdapter(200, { code: 0, message: 'ok', data: { status: 'resolved' } });
  const api = createApiClient({ adapter });

  await api.patch('/feedback/1/status', { status: 'resolved' });

  const config = vi.mocked(adapter).mock.calls[0][0] as InternalAxiosRequestConfig;
  expect(config.method?.toLowerCase()).toBe('patch');
  expect(config.url).toBe('/feedback/1/status');
});
```

创建 `web/src/api/feedback.test.ts`：

```ts
import { describe, expect, it, vi } from 'vitest';

import type { ApiClient } from './http';
import { createFeedbackApi } from './feedback';

function fakeClient(): ApiClient {
  return {
    get: vi.fn(async () => undefined),
    post: vi.fn(async () => undefined),
    patch: vi.fn(async () => undefined)
  };
}

describe('feedback api', () => {
  it('creates and lists product feedback', async () => {
    const client = fakeClient();
    const api = createFeedbackApi(client);

    await api.create(10, { title: 'Need roadmap', content: 'Please add roadmap.' });
    await api.listByProduct(10);

    expect(client.post).toHaveBeenCalledWith('/products/10/feedback', {
      title: 'Need roadmap',
      content: 'Please add roadmap.'
    });
    expect(client.get).toHaveBeenCalledWith('/products/10/feedback');
  });

  it('gets feedback detail and updates status', async () => {
    const client = fakeClient();
    const api = createFeedbackApi(client);

    await api.get(5);
    await api.updateStatus(5, { status: 'resolved' });

    expect(client.get).toHaveBeenCalledWith('/feedback/5');
    expect(client.patch).toHaveBeenCalledWith('/feedback/5/status', { status: 'resolved' });
  });
});
```

- [ ] **步骤 2：运行测试验证失败**

运行：

```bash
npm --prefix web test -- --run src/api/http.test.ts src/api/feedback.test.ts
```

预期：FAIL，报错包含 `api.patch is not a function` 或找不到 `./feedback`。

- [ ] **步骤 3：实现前端 API**

修改 `web/src/api/http.ts`：

```ts
export interface ApiClient {
  get<T>(url: string, config?: AxiosRequestConfig): Promise<T>;
  post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>;
  patch<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T>;
}
```

在返回对象中添加：

```ts
patch<T>(url: string, data?: unknown, config?: AxiosRequestConfig) {
  return request<T>(instance, { ...config, method: 'PATCH', url, data }, unauthorized);
}
```

修改 `web/src/api/types.ts`：

```ts
export type FeedbackStatus = 'open' | 'resolved';

export interface Feedback {
  id: number;
  product_id: number;
  title: string;
  content: string;
  status: FeedbackStatus;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface CreateFeedbackPayload {
  title: string;
  content: string;
}

export interface UpdateFeedbackStatusPayload {
  status: FeedbackStatus;
}
```

创建 `web/src/api/feedback.ts`：

```ts
import { api, type ApiClient } from './http';
import type { CreateFeedbackPayload, Feedback, UpdateFeedbackStatusPayload } from './types';

export function createFeedbackApi(client: ApiClient = api) {
  return {
    create(productID: number, payload: CreateFeedbackPayload) {
      return client.post<Feedback>(`/products/${productID}/feedback`, payload);
    },
    listByProduct(productID: number) {
      return client.get<Feedback[]>(`/products/${productID}/feedback`);
    },
    get(id: number) {
      return client.get<Feedback>(`/feedback/${id}`);
    },
    updateStatus(id: number, payload: UpdateFeedbackStatusPayload) {
      return client.patch<Feedback>(`/feedback/${id}/status`, payload);
    }
  };
}

export const feedbackApi = createFeedbackApi();
```

- [ ] **步骤 4：运行测试验证通过**

运行：

```bash
npm --prefix web test -- --run src/api/http.test.ts src/api/feedback.test.ts
```

预期：PASS。

- [ ] **步骤 5：Commit**

```bash
git add web/src/api/http.ts web/src/api/http.test.ts web/src/api/types.ts web/src/api/feedback.ts web/src/api/feedback.test.ts
git commit -m "feat: add feedback frontend api"
```

---

### 任务 5：产品详情页嵌入反馈流

**文件：**
- 修改：`web/src/views/ProductDetailView.vue`
- 修改：`web/src/styles.css`

- [ ] **步骤 1：扩展产品详情页状态和方法**

修改 `web/src/views/ProductDetailView.vue` 的脚本区，引入反馈 API 和 UI 组件：

```ts
import { ArrowLeft, CheckCircle, Plus, RotateCcw } from '@lucide/vue';
import {
  NButton,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NIcon,
  NInput,
  NList,
  NListItem,
  NSpace,
  NSpin,
  NTag,
  useMessage
} from 'naive-ui';
import { computed, onMounted, reactive, ref, watch } from 'vue';

import { feedbackApi } from '../api/feedback';
import { productsApi } from '../api/products';
import type { Feedback, FeedbackStatus, Product } from '../api/types';
```

增加状态：

```ts
const feedbackLoading = ref(false);
const feedbackSaving = ref(false);
const feedbackDrawerOpen = ref(false);
const selectedFeedback = ref<Feedback | null>(null);
const feedbackItems = ref<Feedback[]>([]);
const feedbackForm = reactive({ title: '', content: '' });
const canCreateFeedback = computed(
  () => feedbackForm.title.trim().length > 0 && feedbackForm.content.trim().length > 0
);
```

将 `loadProduct` 扩展为同时加载反馈：

```ts
async function loadProduct() {
  if (!Number.isFinite(productID.value)) return;
  loading.value = true;
  try {
    product.value = await productsApi.get(productID.value);
    await loadFeedback();
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载产品详情失败');
  } finally {
    loading.value = false;
  }
}
```

新增方法：

```ts
async function loadFeedback() {
  if (!Number.isFinite(productID.value)) return;
  feedbackLoading.value = true;
  try {
    feedbackItems.value = await feedbackApi.listByProduct(productID.value);
  } catch (error) {
    message.error(error instanceof Error ? error.message : '加载反馈失败');
  } finally {
    feedbackLoading.value = false;
  }
}

async function createFeedback() {
  if (!canCreateFeedback.value || feedbackSaving.value) return;
  feedbackSaving.value = true;
  try {
    const created = await feedbackApi.create(productID.value, {
      title: feedbackForm.title.trim(),
      content: feedbackForm.content.trim()
    });
    feedbackForm.title = '';
    feedbackForm.content = '';
    selectedFeedback.value = created;
    feedbackDrawerOpen.value = false;
    await loadFeedback();
  } catch (error) {
    message.error(error instanceof Error ? error.message : '创建反馈失败');
  } finally {
    feedbackSaving.value = false;
  }
}

async function setFeedbackStatus(item: Feedback, status: FeedbackStatus) {
  try {
    const updated = await feedbackApi.updateStatus(item.id, { status });
    selectedFeedback.value = updated;
    await loadFeedback();
  } catch (error) {
    message.error(error instanceof Error ? error.message : '更新反馈状态失败');
  }
}

function selectFeedback(item: Feedback) {
  selectedFeedback.value = item;
}

function feedbackStatusType(status: FeedbackStatus) {
  return status === 'resolved' ? 'success' : 'warning';
}

function closeFeedbackDetail(show: boolean) {
  if (!show) selectedFeedback.value = null;
}
```

- [ ] **步骤 2：扩展模板**

在基础信息 `section` 后增加反馈区域：

```vue
<section class="content-panel feedback-panel">
  <div class="feedback-header">
    <div>
      <h3>产品反馈</h3>
      <p>记录团队成员对这个产品的反馈和处理状态。</p>
    </div>
    <n-button type="primary" @click="feedbackDrawerOpen = true">
      <template #icon>
        <n-icon><Plus /></n-icon>
      </template>
      新建反馈
    </n-button>
  </div>

  <n-spin :show="feedbackLoading">
    <n-list v-if="feedbackItems.length > 0" clickable hoverable>
      <n-list-item v-for="item in feedbackItems" :key="item.id" @click="selectFeedback(item)">
        <div class="feedback-row">
          <div>
            <strong>{{ item.title }}</strong>
            <p>{{ item.content }}</p>
          </div>
          <n-space align="center">
            <n-tag :type="feedbackStatusType(item.status)" bordered>
              {{ item.status }}
            </n-tag>
            <span class="feedback-date">{{ formatDate(item.created_at) }}</span>
          </n-space>
        </div>
      </n-list-item>
    </n-list>
    <div v-else class="empty-state">这个产品还没有反馈。</div>
  </n-spin>
</section>

<n-drawer v-model:show="feedbackDrawerOpen" :width="420" placement="right">
  <n-drawer-content title="新建反馈" closable>
    <n-form label-placement="top" @submit.prevent="createFeedback">
      <n-form-item label="标题">
        <n-input v-model:value="feedbackForm.title" placeholder="简短描述反馈" />
      </n-form-item>
      <n-form-item label="内容">
        <n-input
          v-model:value="feedbackForm.content"
          type="textarea"
          placeholder="说明具体问题或建议"
          :autosize="{ minRows: 5, maxRows: 10 }"
        />
      </n-form-item>
      <n-space justify="end">
        <n-button @click="feedbackDrawerOpen = false">取消</n-button>
        <n-button type="primary" attr-type="submit" :disabled="!canCreateFeedback" :loading="feedbackSaving">
          创建
        </n-button>
      </n-space>
    </n-form>
  </n-drawer-content>
</n-drawer>

<n-drawer :show="Boolean(selectedFeedback)" :width="460" placement="right" @update:show="closeFeedbackDetail">
  <n-drawer-content v-if="selectedFeedback" title="反馈详情" closable>
    <n-space vertical size="large">
      <div>
        <n-tag :type="feedbackStatusType(selectedFeedback.status)" bordered>
          {{ selectedFeedback.status }}
        </n-tag>
        <h3 class="feedback-detail-title">{{ selectedFeedback.title }}</h3>
        <p class="feedback-detail-content">{{ selectedFeedback.content }}</p>
      </div>
      <n-space>
        <n-button
          v-if="selectedFeedback.status === 'open'"
          type="primary"
          @click="setFeedbackStatus(selectedFeedback, 'resolved')"
        >
          <template #icon><n-icon><CheckCircle /></n-icon></template>
          标记已解决
        </n-button>
        <n-button
          v-else
          @click="setFeedbackStatus(selectedFeedback, 'open')"
        >
          <template #icon><n-icon><RotateCcw /></n-icon></template>
          重新打开
        </n-button>
      </n-space>
    </n-space>
  </n-drawer-content>
</n-drawer>
```

- [ ] **步骤 3：增加样式**

修改 `web/src/styles.css`：

```css
.feedback-panel {
  margin-top: 18px;
  padding: 22px;
}

.feedback-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.feedback-header h3 {
  margin: 0;
  font-size: 18px;
}

.feedback-header p {
  margin: 6px 0 0;
  color: #687386;
}

.feedback-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  width: 100%;
}

.feedback-row strong,
.feedback-row p {
  overflow-wrap: anywhere;
}

.feedback-row p {
  margin: 6px 0 0;
  color: #687386;
  line-height: 1.5;
}

.feedback-date {
  color: #687386;
  font-size: 12px;
  white-space: nowrap;
}

.feedback-detail-title {
  margin: 14px 0 8px;
  font-size: 20px;
  letter-spacing: 0;
}

.feedback-detail-content {
  margin: 0;
  line-height: 1.7;
  overflow-wrap: anywhere;
}

@media (max-width: 760px) {
  .feedback-header,
  .feedback-row {
    flex-direction: column;
  }
}
```

- [ ] **步骤 4：运行前端构建验证**

运行：

```bash
npm --prefix web run build
```

预期：PASS。

- [ ] **步骤 5：Commit**

```bash
git add web/src/views/ProductDetailView.vue web/src/styles.css
git commit -m "feat: show feedback in product detail"
```

---

### 任务 6：文档、迁移和全量验证

**文件：**
- 修改：`README.md`
- 修改：`docs/development-tasks.md`

- [ ] **步骤 1：更新 README API 概览**

在 `README.md` 的 API 概览中加入：

````markdown
### 反馈

```http
POST  /api/products/:product_id/feedback
GET   /api/products/:product_id/feedback
GET   /api/feedback/:id
PATCH /api/feedback/:id/status
```
````

在已实现功能中加入：

```markdown
- 在产品下创建反馈、查看产品反馈、查看反馈详情、更新反馈状态。
```

- [ ] **步骤 2：更新开发任务文档**

在 `docs/development-tasks.md` 新增“反馈管理”小节：

````markdown
### 6. 反馈管理

接口：

```http
POST  /api/products/:product_id/feedback
GET   /api/products/:product_id/feedback
GET   /api/feedback/:id
PATCH /api/feedback/:id/status
```

能力：

- 团队成员可以在产品下创建反馈。
- 团队成员可以查看产品反馈列表和反馈详情。
- 团队成员可以把反馈状态设置为 `open` 或 `resolved`。
- 非团队成员不能访问产品反馈。
````

- [ ] **步骤 3：运行全量验证**

运行：

```bash
go test -count=1 ./...
go vet ./...
npm --prefix web test -- --run
npm --prefix web run build
```

预期：全部 exit 0。

- [ ] **步骤 4：运行数据库迁移**

运行：

```bash
go run ./cmd/migrate
```

预期：输出包含：

```text
AutoMigrate completed (5 models)
Migration completed successfully
```

- [ ] **步骤 5：启动服务并验证健康检查**

运行：

```bash
go run ./cmd/api
```

另一个终端运行：

```bash
curl http://127.0.0.1:8080/health
```

预期：

```json
{"code":0,"message":"ok","data":{"status":"ok"}}
```

- [ ] **步骤 6：Commit**

```bash
git add README.md docs/development-tasks.md
git commit -m "docs: document feedback module"
```

---

## 自检清单

- 规格中的所有接口都有对应后端任务：创建、列表、详情、状态更新。
- 规格中的权限规则都有服务测试或 HTTP 测试覆盖。
- 迁移注册问题有 `cmd/migrate` 测试覆盖。
- Gin 路由 wildcard 冲突有 `cmd/api` 路由注册测试覆盖。
- 前端使用 RESTful API，不绕过现有 Axios token 注入和 401 处理。
- 最终验证命令覆盖后端测试、后端 vet、前端测试和前端构建。
