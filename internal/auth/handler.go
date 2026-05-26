package auth

import (
	"errors"

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

func (h *Handler) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	user, err := h.service.Register(c.Request.Context(), input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, user)
}

func (h *Handler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	result, err := h.service.Login(c.Request.Context(), input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, result)
}

func (h *Handler) Me(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	user, err := h.service.CurrentUserByID(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, user)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrEmailAlreadyExists):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrInvalidCredentials), errors.Is(err, ErrUnauthorized):
		response.Unauthorized(c, "unauthorized")
	default:
		response.InternalError(c, "internal server error")
	}
}
