package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"ticketflow/pkg/apperr"
	"ticketflow/services/payment-service/internal/service"
)

// PaymentUseCase is the narrow slice of *service.PaymentService's methods
// PaymentHandler actually calls, extracted here at the consuming package
// per docs/01-architecture/backend-conventions.md's "Unit test có mock"
// convention — mockery can only generate a mock from an interface, and
// *service.PaymentService already implements this implicitly, so no
// service-side change beyond what payment_service.go already needed.
type PaymentUseCase interface {
	Checkout(ctx context.Context, orderID, provider string) (*service.CheckoutResult, error)
	HandleWebhook(ctx context.Context, orderID, status string) error
}

type PaymentHandler struct {
	payments PaymentUseCase
}

func NewPaymentHandler(payments PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{payments: payments}
}

type checkoutRequest struct {
	Provider string `json:"provider"`
}

func (h *PaymentHandler) Checkout(c *gin.Context) {
	var req checkoutRequest
	_ = c.ShouldBindJSON(&req) // body is optional per api-docs/openapi/payment-service.yaml
	result, err := h.payments.Checkout(c.Request.Context(), c.Param("orderId"), req.Provider)
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, checkoutResponseDTO{PaymentID: result.PaymentID, CheckoutURL: result.CheckoutURL})
}

type webhookRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Status  string `json:"status" binding:"required"`
}

// Webhook is a thin Phase 1 pass-through kept for contract stability and
// manual testing of the failure path — the Phase 1 checkout flow already
// confirms success synchronously (see PaymentService.Checkout).
func (h *PaymentHandler) Webhook(c *gin.Context) {
	var req webhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Respond(c, apperr.WithMessage(apperr.ErrValidation, "order_id và status là bắt buộc"))
		return
	}
	if err := h.payments.HandleWebhook(c.Request.Context(), req.OrderID, req.Status); err != nil {
		apperr.Respond(c, err)
		return
	}
	c.Status(http.StatusOK)
}
