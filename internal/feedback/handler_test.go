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
)

type staticTokenParser struct {
	users map[string]uint
}

func (p staticTokenParser) ParseToken(token string) (uint, error) {
	userID, ok := p.users[token]
	if !ok {
		return 0, errors.New("invalid token")
	}
	return userID, nil
}

func newFeedbackTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	access := newFakeProductAccess()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	svc := NewService(newFakeFeedbackRepository(), access)
	parser := staticTokenParser{users: map[string]uint{
		"user-7": 7,
		"user-8": 8,
	}}

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

func TestFeedbackHTTPCreateListGetAndUpdateStatus(t *testing.T) {
	r := newFeedbackTestRouter()

	createResp := performFeedbackJSON(r, http.MethodPost, "/api/products/10/feedback", gin.H{
		"title":   "Missing export",
		"content": "CSV export would help.",
	}, "user-7")
	if createResp.Code != http.StatusOK {
		t.Fatalf("create status = %d, body = %s", createResp.Code, createResp.Body.String())
	}
	createPayload := decodeFeedbackResponse(t, createResp)
	created := createPayload["data"].(map[string]any)
	if created["status"] != StatusOpen || created["product_id"] != float64(10) || created["content"] != "CSV export would help." {
		t.Fatalf("unexpected create response: %#v", created)
	}
	feedbackID := uint(created["id"].(float64))

	listResp := performFeedbackJSON(r, http.MethodGet, "/api/products/10/feedback", nil, "user-7")
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listResp.Code, listResp.Body.String())
	}
	listPayload := decodeFeedbackResponse(t, listResp)
	list := listPayload["data"].([]any)
	if len(list) != 1 {
		t.Fatalf("expected one feedback item, got %#v", list)
	}

	getResp := performFeedbackJSON(r, http.MethodGet, "/api/feedback/"+strconvFeedbackID(feedbackID), nil, "user-7")
	if getResp.Code != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", getResp.Code, getResp.Body.String())
	}
	getPayload := decodeFeedbackResponse(t, getResp)
	detail := getPayload["data"].(map[string]any)
	if detail["id"] != float64(feedbackID) || detail["product_id"] != float64(10) {
		t.Fatalf("unexpected detail response: %#v", detail)
	}

	updateResp := performFeedbackJSON(r, http.MethodPatch, "/api/feedback/"+strconvFeedbackID(feedbackID)+"/status", gin.H{
		"status": StatusResolved,
	}, "user-7")
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", updateResp.Code, updateResp.Body.String())
	}
	updatePayload := decodeFeedbackResponse(t, updateResp)
	updated := updatePayload["data"].(map[string]any)
	if updated["status"] != StatusResolved {
		t.Fatalf("expected status resolved, got %#v", updated)
	}
}

func TestFeedbackHTTPRequiresAuthentication(t *testing.T) {
	r := newFeedbackTestRouter()

	w := performFeedbackJSON(r, http.MethodGet, "/api/products/10/feedback", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d with body %s", w.Code, w.Body.String())
	}
}

func TestFeedbackHTTPRejectsNonMemberCreate(t *testing.T) {
	r := newFeedbackTestRouter()

	w := performFeedbackJSON(r, http.MethodPost, "/api/products/10/feedback", gin.H{
		"title":   "Missing export",
		"content": "CSV export would help.",
	}, "user-8")
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d with body %s", w.Code, w.Body.String())
	}
}

func TestFeedbackHTTPRejectsMissingProductCreate(t *testing.T) {
	r := newFeedbackTestRouter()

	w := performFeedbackJSON(r, http.MethodPost, "/api/products/999/feedback", gin.H{
		"title":   "Missing export",
		"content": "CSV export would help.",
	}, "user-7")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d with body %s", w.Code, w.Body.String())
	}

	payload := decodeFeedbackResponse(t, w)
	if payload["code"] != float64(404) || payload["message"] != "product not found" {
		t.Fatalf("unexpected error response: %#v", payload)
	}
}

func TestFeedbackHTTPRejectsInvalidStatus(t *testing.T) {
	r := newFeedbackTestRouter()

	createResp := performFeedbackJSON(r, http.MethodPost, "/api/products/10/feedback", gin.H{
		"title":   "Missing export",
		"content": "CSV export would help.",
	}, "user-7")
	if createResp.Code != http.StatusOK {
		t.Fatalf("create status = %d, body = %s", createResp.Code, createResp.Body.String())
	}
	payload := decodeFeedbackResponse(t, createResp)
	feedbackID := uint(payload["data"].(map[string]any)["id"].(float64))

	w := performFeedbackJSON(r, http.MethodPatch, "/api/feedback/"+strconvFeedbackID(feedbackID)+"/status", gin.H{
		"status": "closed",
	}, "user-7")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d with body %s", w.Code, w.Body.String())
	}
}

func strconvFeedbackID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
