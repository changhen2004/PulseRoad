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

	teamID, err := strconv.ParseUint(c.Param("id"), 10, 64)
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

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrForbidden):
		response.Fail(c, 403, 403, "forbidden")
	case errors.Is(err, ErrTeamNotFound):
		response.Fail(c, 404, 404, "team not found")
	default:
		response.InternalError(c, "internal server error")
	}
}
