package product

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

	teamID, ok := parseUintParam(c, "team_id")
	if !ok {
		response.BadRequest(c, "invalid team id")
		return
	}

	var input CreateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	product, err := h.service.CreateProduct(c.Request.Context(), userID, teamID, input)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, product)
}

func (h *Handler) ListByTeam(c *gin.Context) {
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

	products, err := h.service.ListProducts(c.Request.Context(), userID, teamID)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, products)
}

func (h *Handler) Get(c *gin.Context) {
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

	product, err := h.service.GetProduct(c.Request.Context(), userID, productID)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, product)
}

func (h *Handler) Summary(c *gin.Context) {
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

	summary, err := h.service.GetProductSummary(c.Request.Context(), userID, productID)
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.Success(c, summary)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrForbidden):
		response.Fail(c, 403, 403, "forbidden")
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
