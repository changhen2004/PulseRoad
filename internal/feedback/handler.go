package feedback

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

	productID, ok := parseUintParam(c, "id")
	if !ok {
		response.BadRequest(c, "invalid product id")
		return
	}

	var input CreateFeedbackInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	feedback, err := h.service.CreateFeedback(c.Request.Context(), userID, productID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, feedback)
}

func (h *Handler) ListByProduct(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	productID, ok := parseUintParam(c, "id")
	if !ok {
		response.BadRequest(c, "invalid product id")
		return
	}

	feedbackItems, err := h.service.ListFeedback(c.Request.Context(), userID, productID)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, feedbackItems)
}

func (h *Handler) Get(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	feedbackID, ok := parseUintParam(c, "id")
	if !ok {
		response.BadRequest(c, "invalid feedback id")
		return
	}

	feedback, err := h.service.GetFeedback(c.Request.Context(), userID, feedbackID)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, feedback)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	feedbackID, ok := parseUintParam(c, "id")
	if !ok {
		response.BadRequest(c, "invalid feedback id")
		return
	}

	var input UpdateFeedbackStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	feedback, err := h.service.UpdateStatus(c.Request.Context(), userID, feedbackID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, feedback)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrForbidden):
		response.Fail(c, 403, 403, "forbidden")
	case errors.Is(err, ErrFeedbackNotFound):
		response.Fail(c, 404, 404, "feedback not found")
	case errors.Is(err, ErrProductNotFound):
		response.Fail(c, 404, 404, "product not found")
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
