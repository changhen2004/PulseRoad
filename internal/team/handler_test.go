package team

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

func newTeamTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	svc := NewService(newFakeTeamRepository())
	parser := staticTokenParser{users: map[string]uint{
		"user-7": 7,
		"user-8": 8,
	}}

	r := gin.New()
	RegisterRoutes(r.Group("/api"), parser, svc)
	return r
}

func performTeamJSON(r http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
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

func decodeTeamResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response %q: %v", w.Body.String(), err)
	}
	return payload
}

func TestTeamHTTPCreateListAndGet(t *testing.T) {
	r := newTeamTestRouter()

	createResp := performTeamJSON(r, http.MethodPost, "/api/teams", gin.H{
		"name":        "Core",
		"description": "Roadmap work",
	}, "user-7")
	if createResp.Code != http.StatusOK {
		t.Fatalf("create status = %d, body = %s", createResp.Code, createResp.Body.String())
	}
	createPayload := decodeTeamResponse(t, createResp)
	created := createPayload["data"].(map[string]any)
	if created["name"] != "Core" || created["role"] != RoleOwner {
		t.Fatalf("unexpected create response: %#v", created)
	}
	teamID := uint(created["id"].(float64))

	listResp := performTeamJSON(r, http.MethodGet, "/api/teams", nil, "user-7")
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listResp.Code, listResp.Body.String())
	}
	listPayload := decodeTeamResponse(t, listResp)
	list := listPayload["data"].([]any)
	if len(list) != 1 {
		t.Fatalf("expected one team, got %#v", list)
	}

	getResp := performTeamJSON(r, http.MethodGet, "/api/teams/"+strconvID(teamID), nil, "user-7")
	if getResp.Code != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", getResp.Code, getResp.Body.String())
	}
	getPayload := decodeTeamResponse(t, getResp)
	detail := getPayload["data"].(map[string]any)
	if detail["id"] != float64(teamID) || detail["role"] != RoleOwner {
		t.Fatalf("unexpected detail response: %#v", detail)
	}
}

func TestTeamHTTPRequiresAuthentication(t *testing.T) {
	r := newTeamTestRouter()

	w := performTeamJSON(r, http.MethodGet, "/api/teams", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d with body %s", w.Code, w.Body.String())
	}
}

func TestTeamHTTPRejectsNonMemberDetail(t *testing.T) {
	r := newTeamTestRouter()

	createResp := performTeamJSON(r, http.MethodPost, "/api/teams", gin.H{"name": "Core"}, "user-7")
	if createResp.Code != http.StatusOK {
		t.Fatalf("create status = %d, body = %s", createResp.Code, createResp.Body.String())
	}
	payload := decodeTeamResponse(t, createResp)
	teamID := uint(payload["data"].(map[string]any)["id"].(float64))

	getResp := performTeamJSON(r, http.MethodGet, "/api/teams/"+strconvID(teamID), nil, "user-8")
	if getResp.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d with body %s", getResp.Code, getResp.Body.String())
	}
}

func strconvID(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}
