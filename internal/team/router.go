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
	teams.GET("/invitations", handler.ListInvitations)
	teams.POST("/invitations/:id/accept", handler.AcceptInvitation)
	teams.GET("/:team_id", handler.Get)
	teams.GET("/:team_id/members", handler.ListMembers)
	teams.POST("/:team_id/invitations", handler.InviteMember)
	teams.PATCH("/:team_id/members/:user_id/role", handler.UpdateMemberRole)
	teams.DELETE("/:team_id/members/:user_id", handler.RemoveMember)
}
