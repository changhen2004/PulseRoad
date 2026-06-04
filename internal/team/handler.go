package team

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"pulseroad/internal/middleware"
	"pulseroad/internal/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var input CreateTeamInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	team, err := h.service.CreateTeam(c.Request.Context(), userID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, team)
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	teams, err := h.service.ListTeams(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, teams)
}

func (h *Handler) Get(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil || teamID == 0 {
		response.BadRequest(c, "invalid team id")
		return
	}

	team, err := h.service.GetTeam(c.Request.Context(), userID, uint(teamID))
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, team)
}

func (h *Handler) ListMembers(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	teamID, ok := parseUintParam(c, "team_id")
	if !ok {
		response.BadRequest(c, "invalid team id")
		return
	}
	members, err := h.service.ListMembers(c.Request.Context(), userID, teamID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, members)
}

func (h *Handler) InviteMember(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	teamID, ok := parseUintParam(c, "team_id")
	if !ok {
		response.BadRequest(c, "invalid team id")
		return
	}
	var input InviteMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	invitation, err := h.service.InviteMember(c.Request.Context(), userID, teamID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, invitation)
}

func (h *Handler) ListInvitations(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	invitations, err := h.service.ListInvitations(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, invitations)
}

func (h *Handler) AcceptInvitation(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	invitationID, ok := parseUintParam(c, "id")
	if !ok {
		response.BadRequest(c, "invalid invitation id")
		return
	}
	invitation, err := h.service.AcceptInvitationForUser(c.Request.Context(), userID, invitationID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, invitation)
}

func (h *Handler) UpdateMemberRole(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	teamID, ok := parseUintParam(c, "team_id")
	if !ok {
		response.BadRequest(c, "invalid team id")
		return
	}
	memberUserID, ok := parseUintParam(c, "user_id")
	if !ok {
		response.BadRequest(c, "invalid user id")
		return
	}
	var input UpdateMemberRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	member, err := h.service.UpdateMemberRole(c.Request.Context(), userID, teamID, memberUserID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, member)
}

func (h *Handler) RemoveMember(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	teamID, ok := parseUintParam(c, "team_id")
	if !ok {
		response.BadRequest(c, "invalid team id")
		return
	}
	memberUserID, ok := parseUintParam(c, "user_id")
	if !ok {
		response.BadRequest(c, "invalid user id")
		return
	}
	if err := h.service.RemoveMember(c.Request.Context(), userID, teamID, memberUserID); err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, gin.H{"removed": true})
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalid), errors.Is(err, ErrInvitationExists), errors.Is(err, ErrLastOwner):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrForbidden):
		response.Fail(c, 403, 403, "forbidden")
	case errors.Is(err, ErrTeamNotFound):
		response.Fail(c, 404, 404, "team not found")
	case errors.Is(err, ErrInvitationNotFound):
		response.Fail(c, 404, 404, "invitation not found")
	case errors.Is(err, ErrMemberNotFound):
		response.Fail(c, 404, 404, "member not found")
	case errors.Is(err, ErrUserNotFound):
		response.Fail(c, 404, 404, "user not found")
	default:
		response.InternalError(c, "internal server error")
	}
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		return 0, false
	}
	return uint(value), true
}
