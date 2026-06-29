package requirement

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
	productID, ok := parseUintParam(c, "product_id")
	if !ok {
		response.BadRequest(c, "invalid product id")
		return
	}
	var input CreateRequirementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	req, err := h.service.Create(c.Request.Context(), userID, productID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, req)
}

func (h *Handler) ListByProduct(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	productID, ok := parseUintParam(c, "product_id")
	if !ok {
		response.BadRequest(c, "invalid product id")
		return
	}
	page := queryIntDefault(c, "page", 1)
	pageSize := queryIntDefault(c, "page_size", 20)
	items, err := h.service.ListByProduct(c.Request.Context(), userID, productID, c.Query("status"), page, pageSize)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, items)
}

func (h *Handler) Get(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		response.BadRequest(c, "invalid requirement id")
		return
	}
	req, err := h.service.Get(c.Request.Context(), userID, id)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, req)
}

func (h *Handler) Update(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		response.BadRequest(c, "invalid requirement id")
		return
	}
	var input UpdateRequirementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	req, err := h.service.Update(c.Request.Context(), userID, id, input)
	if err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, req)
}

func (h *Handler) Delete(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		response.BadRequest(c, "invalid requirement id")
		return
	}
	if err := h.service.Delete(c.Request.Context(), userID, id); err != nil {
		h.writeError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrForbidden):
		response.Fail(c, 403, 403, "forbidden")
	case errors.Is(err, ErrNotOwner):
		response.Fail(c, 403, 403, "forbidden")
	case errors.Is(err, ErrRequirementNotFound):
		response.Fail(c, 404, 404, "requirement not found")
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

func queryIntDefault(c *gin.Context, name string, defaultVal int) int {
	value, err := strconv.Atoi(c.Query(name))
	if err != nil {
		return defaultVal
	}
	return value
}
