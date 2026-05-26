package auth

import (
	"github.com/gin-gonic/gin"

	"pulseroad/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, service *Service) {
	handler := NewHandler(service)
	auth := r.Group("/auth")
	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Login)
	auth.GET("/me", middleware.AuthRequired(service), handler.Me)
}
