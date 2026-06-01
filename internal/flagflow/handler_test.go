package flagflow

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

type staticTokenParser struct {
	userID uint
	ok     bool
}

func (p staticTokenParser) ParseToken(string) (uint, error) {
	if !p.ok {
		return 0, nil
	}
	return p.userID, nil
}

func newFlagRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	access := newFakeProductAccess()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	svc := NewService(newFakeFlagRepository(), access, newFakeCache(), &fakeEventPublisher{})
	r := gin.New()
	RegisterRoutes(r.Group("/api"), staticTokenParser{userID: 7, ok: true}, svc)
	return r
}

func performFlagJSON(r http.Handler, method string, path string, body any, token string) *httptest.ResponseRecorder {
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

func decodeFlagPayload(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	return payload
}

func TestFlagHTTPFlow(t *testing.T) {
	r := newFlagRouter()

	createResp := performFlagJSON(r, http.MethodPost, "/api/products/10/flags", gin.H{
		"key":                "new_dashboard",
		"name":               "New Dashboard",
		"description":        "Roll out dashboard",
		"environment":        "production",
		"rollout_percentage": 100,
	}, "user-7")
	if createResp.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	created := decodeFlagPayload(t, createResp)["data"].(map[string]any)
	flagID := uint(created["id"].(float64))

	listResp := performFlagJSON(r, http.MethodGet, "/api/products/10/flags?environment=production", nil, "user-7")
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listResp.Code, listResp.Body.String())
	}
	list := decodeFlagPayload(t, listResp)["data"].([]any)
	if len(list) != 1 {
		t.Fatalf("expected one flag, got %#v", list)
	}

	toggleResp := performFlagJSON(r, http.MethodPatch, "/api/flags/"+strconvID(flagID)+"/toggle", gin.H{"enabled": true}, "user-7")
	if toggleResp.Code != http.StatusOK {
		t.Fatalf("toggle status=%d body=%s", toggleResp.Code, toggleResp.Body.String())
	}

	evaluateResp := performFlagJSON(r, http.MethodPost, "/api/flags/evaluate", gin.H{
		"product_id":  10,
		"key":         "new_dashboard",
		"environment": "production",
		"user_key":    "user-10001",
	}, "user-7")
	if evaluateResp.Code != http.StatusOK {
		t.Fatalf("evaluate status=%d body=%s", evaluateResp.Code, evaluateResp.Body.String())
	}
	result := decodeFlagPayload(t, evaluateResp)["data"].(map[string]any)
	if result["enabled"] != true {
		t.Fatalf("expected enabled result, got %#v", result)
	}
}

func strconvID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
