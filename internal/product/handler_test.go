package product

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

func newProductTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	membership := newFakeTeamMembership()
	membership.add(10, 7)
	svc := NewService(newFakeProductRepository(), membership)
	parser := staticTokenParser{users: map[string]uint{
		"user-7": 7,
		"user-8": 8,
	}}

	r := gin.New()
	RegisterRoutes(r.Group("/api"), parser, svc)
	return r
}

func performProductJSON(r http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
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

func decodeProductResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	return payload
}

func TestProductHTTPCreateListAndGet(t *testing.T) {
	r := newProductTestRouter()

	createResp := performProductJSON(r, http.MethodPost, "/api/teams/10/products", gin.H{
		"name":        "PulseRoad",
		"description": "Feedback platform",
	}, "user-7")
	if createResp.Code != http.StatusOK {
		t.Fatalf("create status = %d, body = %s", createResp.Code, createResp.Body.String())
	}
	createPayload := decodeProductResponse(t, createResp)
	created := createPayload["data"].(map[string]any)
	if created["name"] != "PulseRoad" || created["team_id"] != float64(10) {
		t.Fatalf("unexpected create response: %#v", created)
	}
	productID := uint(created["id"].(float64))

	listResp := performProductJSON(r, http.MethodGet, "/api/teams/10/products", nil, "user-7")
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listResp.Code, listResp.Body.String())
	}
	listPayload := decodeProductResponse(t, listResp)
	list := listPayload["data"].([]any)
	if len(list) != 1 {
		t.Fatalf("expected one product, got %#v", list)
	}

	getResp := performProductJSON(r, http.MethodGet, "/api/products/"+strconvID(productID), nil, "user-7")
	if getResp.Code != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", getResp.Code, getResp.Body.String())
	}
	getPayload := decodeProductResponse(t, getResp)
	detail := getPayload["data"].(map[string]any)
	if detail["id"] != float64(productID) || detail["team_id"] != float64(10) {
		t.Fatalf("unexpected detail response: %#v", detail)
	}
}

func TestProductHTTPRequiresAuthentication(t *testing.T) {
	r := newProductTestRouter()

	w := performProductJSON(r, http.MethodGet, "/api/teams/10/products", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d with body %s", w.Code, w.Body.String())
	}
}

func TestProductHTTPRejectsNonMember(t *testing.T) {
	r := newProductTestRouter()

	w := performProductJSON(r, http.MethodPost, "/api/teams/10/products", gin.H{"name": "PulseRoad"}, "user-8")
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d with body %s", w.Code, w.Body.String())
	}
}

func strconvID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
