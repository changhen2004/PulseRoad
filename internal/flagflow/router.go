package flagflow

import (
	"github.com/gin-gonic/gin"

	"pulseroad/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, parser middleware.TokenParser, service *Service) {
	handler := NewHandler(service)
	auth := r.Group("", middleware.AuthRequired(parser))
	auth.POST("/products/:id/flags", handler.Create)
	auth.GET("/products/:id/flags", handler.ListByProduct)
	auth.POST("/flags/evaluate", handler.Evaluate)
	auth.GET("/flags/:id", handler.Get)
	auth.PATCH("/flags/:id", handler.Update)
	auth.PATCH("/flags/:id/toggle", handler.Toggle)
}
