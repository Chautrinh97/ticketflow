package http

import (
	"github.com/gin-gonic/gin"

	"ticketflow/pkg/httpauth"
)

func NewRouter(jwtSecret []byte, payments *PaymentHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	v1 := r.Group("/api/v1")
	{
		v1.POST("/payments/:orderId/checkout", httpauth.RequireAuth(jwtSecret), payments.Checkout)
		// security: [] per api-docs/openapi/payment-service.yaml — public,
		// verified via HMAC header in Phase 2 instead of a bearer token.
		v1.POST("/payments/webhook", payments.Webhook)
	}
	return r
}
