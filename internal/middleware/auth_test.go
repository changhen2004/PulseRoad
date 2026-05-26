package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "test-secret-that-is-long-enough-for-hs256"

type testTokenParser struct {
	secret []byte
}

func (p testTokenParser) ParseToken(tokenString string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return p.secret, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}
	userID, ok := claims["user_id"].(float64)
	if !ok || userID == 0 {
		return 0, errors.New("missing user id")
	}
	return uint(userID), nil
}

func signedToken(t *testing.T, userID uint, expiresAt time.Time) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	})
	signed, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func newMiddlewareRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", AuthRequired(testTokenParser{secret: []byte(testJWTSecret)}), func(c *gin.Context) {
		userID, ok := CurrentUserID(c)
		if !ok {
			c.String(http.StatusInternalServerError, "missing user id")
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})
	return r
}

func requestProtected(r http.Handler, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAuthRequiredRejectsMissingToken(t *testing.T) {
	r := newMiddlewareRouter()

	w := requestProtected(r, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d with body %s", w.Code, w.Body.String())
	}
}

func TestAuthRequiredRejectsInvalidToken(t *testing.T) {
	r := newMiddlewareRouter()

	w := requestProtected(r, "not-a-token")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d with body %s", w.Code, w.Body.String())
	}
}

func TestAuthRequiredRejectsExpiredToken(t *testing.T) {
	r := newMiddlewareRouter()
	token := signedToken(t, 42, time.Now().Add(-time.Hour))

	w := requestProtected(r, token)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d with body %s", w.Code, w.Body.String())
	}
}

func TestAuthRequiredInjectsCurrentUserID(t *testing.T) {
	r := newMiddlewareRouter()
	token := signedToken(t, 42, time.Now().Add(time.Hour))

	w := requestProtected(r, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %s", w.Code, w.Body.String())
	}
	if w.Body.String() != `{"user_id":42}` {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}
