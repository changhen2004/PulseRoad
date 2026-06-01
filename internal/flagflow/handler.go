package flagflow

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
	var input CreateFlagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	flag, err := h.service.CreateFlag(c.Request.Context(), userID, productID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, flag)
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
	flags, err := h.service.ListFlags(c.Request.Context(), userID, productID, c.Query("environment"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, flags)
}

func (h *Handler) Get(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	flagID, ok := parseUintParam(c, "id")
	if !ok {
		response.BadRequest(c, "invalid flag id")
		return
	}
	flag, err := h.service.GetFlag(c.Request.Context(), userID, flagID)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, flag)
}

func (h *Handler) Update(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	flagID, ok := parseUintParam(c, "id")
	if !ok {
		response.BadRequest(c, "invalid flag id")
		return
	}
	var input UpdateFlagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	flag, err := h.service.UpdateFlag(c.Request.Context(), userID, flagID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, flag)
}

func (h *Handler) Toggle(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	flagID, ok := parseUintParam(c, "id")
	if !ok {
		response.BadRequest(c, "invalid flag id")
		return
	}
	var input ToggleFlagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	flag, err := h.service.ToggleFlag(c.Request.Context(), userID, flagID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, flag)
}

func (h *Handler) Evaluate(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	var input EvaluateFlagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	result, err := h.service.EvaluateFlag(c.Request.Context(), userID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalid), errors.Is(err, ErrFlagAlreadyExists):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrForbidden):
		response.Fail(c, 403, 403, "forbidden")
	case errors.Is(err, ErrFlagNotFound):
		response.Fail(c, 404, 404, "flag not found")
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
