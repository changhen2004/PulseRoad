package product

import (
	"github.com/gin-gonic/gin"

	"pulseroad/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, parser middleware.TokenParser, service *Service) {
	handler := NewHandler(service)
	auth := r.Group("", middleware.AuthRequired(parser))
	auth.POST("/teams/:team_id/products", handler.Create)
	auth.GET("/teams/:team_id/products", handler.ListByTeam)
	auth.GET("/products/:id", handler.Get)
}
