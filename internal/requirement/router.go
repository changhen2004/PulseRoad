package requirement

import (
	"github.com/gin-gonic/gin"

	"pulseroad/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, parser middleware.TokenParser, service *Service) {
	handler := NewHandler(service)
	auth := r.Group("", middleware.AuthRequired(parser))
	auth.POST("/products/:product_id/requirements", handler.Create)
	auth.GET("/products/:product_id/requirements", handler.ListByProduct)
	auth.GET("/requirements/:id", handler.Get)
	auth.PATCH("/requirements/:id", handler.Update)
	auth.DELETE("/requirements/:id", handler.Delete)
}
