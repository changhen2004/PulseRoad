package feedback

import (
	"github.com/gin-gonic/gin"

	"pulseroad/internal/middleware"
)

func RegisterRoutes(r gin.IRouter, parser middleware.TokenParser, service *Service) {
	handler := NewHandler(service)
	auth := r.Group("", middleware.AuthRequired(parser))
	auth.POST("/products/:id/feedback", handler.Create)
	auth.GET("/products/:id/feedback", handler.ListByProduct)
	auth.GET("/feedback/:id", handler.Get)
	auth.PATCH("/feedback/:id/status", handler.UpdateStatus)
	auth.POST("/feedback/:id/comments", handler.CreateComment)
	auth.GET("/feedback/:id/comments", handler.ListComments)
	auth.POST("/feedback/:id/vote", handler.Vote)
	auth.DELETE("/feedback/:id/vote", handler.Unvote)
}
