package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/httpauth"
	"ticketflow/pkg/pagination"
	"ticketflow/services/booking-service/internal/model"
	"ticketflow/services/booking-service/internal/service"
)

type BookingHandler struct {
	bookings *service.BookingService
}

func NewBookingHandler(bookings *service.BookingService) *BookingHandler {
	return &BookingHandler{bookings: bookings}
}

type bookingItemRequest struct {
	TicketTypeID string `json:"ticket_type_id" binding:"required"`
	Quantity     int    `json:"quantity" binding:"required,min=1"`
}

type createBookingRequest struct {
	Items []bookingItemRequest `json:"items" binding:"required,min=1,dive"`
}

func (h *BookingHandler) Create(c *gin.Context) {
	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Respond(c, apperr.WithMessage(apperr.ErrValidation, "items là bắt buộc và cần ít nhất 1 dòng {ticket_type_id, quantity}"))
		return
	}
	items := make([]model.BookingItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, model.BookingItem{TicketTypeID: it.TicketTypeID, Quantity: it.Quantity})
	}
	order, err := h.bookings.CreateOrder(c.Request.Context(), httpauth.UserID(c), items)
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, toOrderDTO(order))
}

func (h *BookingHandler) GetByID(c *gin.Context) {
	order, err := h.bookings.GetOrderByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, toOrderDTO(order))
}

func (h *BookingHandler) ListMine(c *gin.Context) {
	p := pagination.ParseParams(c.Query("page"), c.Query("page_size"))
	orders, total, err := h.bookings.ListMyOrders(c.Request.Context(), httpauth.UserID(c), c.Query("status"), p)
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	items := make([]orderDTO, 0, len(orders))
	for _, o := range orders {
		items = append(items, toOrderDTO(o))
	}
	c.JSON(http.StatusOK, pagination.New(items, total))
}

// OwnerLookup backs httpauth.RequireOwnership for GET /bookings/:id.
func (h *BookingHandler) OwnerLookup(c *gin.Context) (string, error) {
	return h.bookings.GetOwnerID(c.Request.Context(), c.Param("id"))
}
