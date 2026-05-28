package main

import (
	"testing"

	"github.com/gin-gonic/gin"

	"pulseroad/internal/product"
	"pulseroad/internal/team"
)

type testTokenParser struct{}

func (testTokenParser) ParseToken(_ string) (uint, error) {
	return 1, nil
}

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
}
