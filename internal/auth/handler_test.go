package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	svc, _ := newTestService(t)
	return newTestRouterWithService(svc)
}

func newTestRouterWithService(svc *Service) *gin.Engine {
	r := gin.New()
	RegisterRoutes(r.Group("/api"), svc)
	return r
}

func performJSON(r http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
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

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	return payload
}

func TestRegisterLoginAndCurrentUserHTTP(t *testing.T) {
	r := newTestRouter(t)

	registerResp := performJSON(r, http.MethodPost, "/api/auth/register", gin.H{
		"email":    "ada@example.com",
		"password": "password123",
		"name":     "Ada",
	}, "")
	if registerResp.Code != http.StatusOK {
		t.Fatalf("register status = %d, body = %s", registerResp.Code, registerResp.Body.String())
	}

	loginResp := performJSON(r, http.MethodPost, "/api/auth/login", gin.H{
		"email":    "ada@example.com",
		"password": "password123",
	}, "")
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", loginResp.Code, loginResp.Body.String())
	}
	loginPayload := decodeResponse(t, loginResp)
	data := loginPayload["data"].(map[string]any)
	token, ok := data["token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected token in login response, got %#v", data)
	}

	meResp := performJSON(r, http.MethodGet, "/api/auth/me", nil, token)
	if meResp.Code != http.StatusOK {
		t.Fatalf("me status = %d, body = %s", meResp.Code, meResp.Body.String())
	}
	mePayload := decodeResponse(t, meResp)
	user := mePayload["data"].(map[string]any)
	if user["email"] != "ada@example.com" || user["name"] != "Ada" {
		t.Fatalf("unexpected me response: %#v", user)
	}
}

func TestCurrentUserRequiresToken(t *testing.T) {
	r := newTestRouter(t)

	w := performJSON(r, http.MethodGet, "/api/auth/me", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d with body %s", w.Code, w.Body.String())
	}
	payload := decodeResponse(t, w)
	if payload["code"] != float64(401) {
		t.Fatalf("expected response code 401, got %#v", payload)
	}
}

func TestLoginRateLimitReturnsTooManyRequests(t *testing.T) {
	repo := newFakeUserRepository()
	limiter := newFakeLoginLimiter()
	limiter.blocked["ada@example.com"] = true
	svc := NewServiceWithLoginLimiter(repo, testJWTSecret, limiter)
	r := newTestRouterWithService(svc)

	w := performJSON(r, http.MethodPost, "/api/auth/login", gin.H{
		"email":    "ada@example.com",
		"password": "password123",
	}, "")
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d with body %s", w.Code, w.Body.String())
	}
	payload := decodeResponse(t, w)
	if payload["code"] != float64(429) {
		t.Fatalf("expected response code 429, got %#v", payload)
	}
}
