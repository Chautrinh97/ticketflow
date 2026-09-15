package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticketflow/pkg/apperr"
	"ticketflow/services/event-service/internal/service"
)

type TicketTypeHandler struct {
	ticketTypes *service.TicketTypeService
}

func NewTicketTypeHandler(ticketTypes *service.TicketTypeService) *TicketTypeHandler {
	return &TicketTypeHandler{ticketTypes: ticketTypes}
}

type createTicketTypeRequest struct {
	Name     string  `json:"name" binding:"required"`
	Price    float64 `json:"price" binding:"required"`
	Currency string  `json:"currency"`
	Quota    int     `json:"quota" binding:"required,min=1"`
}

func (h *TicketTypeHandler) Create(c *gin.Context) {
	var req createTicketTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Respond(c, apperr.WithMessage(apperr.ErrValidation, "Dữ liệu loại vé không hợp lệ"))
		return
	}
	tt, err := h.ticketTypes.Create(c.Request.Context(), c.Param("id"), service.CreateTicketTypeInput{
		Name: req.Name, Price: req.Price, Currency: req.Currency, Quota: req.Quota,
	})
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, toTicketTypeDTO(*tt))
}
