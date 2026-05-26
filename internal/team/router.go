package team

import (
	"github.com/gin-gonic/gin"

	"pulseroad/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, parser middleware.TokenParser, service *Service) {
	handler := NewHandler(service)
	teams := r.Group("/teams", middleware.AuthRequired(parser))
	teams.POST("", handler.Create)
	teams.GET("", handler.List)
	teams.GET("/:id", handler.Get)
}
