// Package router mounts every Phase 1 client-facing path from
// api-docs/openapi/{identity,event,booking,payment}-service.yaml (plus the
// Phase 1 addition GET /organizer/events/{id}), routing each to the owning
// service's REST API unchanged. Routes not implemented in Phase 1 (search,
// cancel, buyers/stats, admin/*, organizer-request) are intentionally not
// registered here — see the implementation plan's Phase 1 endpoint trim.
package router

import (
	"github.com/gin-gonic/gin"

	"ticketflow/services/api-gateway/internal/proxy"
)

type Targets struct {
	Identity string
	Event    string
	Booking  string
	Payment  string
}

func New(corsMW, authMW gin.HandlerFunc, t Targets) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger(), corsMW, authMW)

	identityProxy := proxy.New(t.Identity)
	eventProxy := proxy.New(t.Event)
	bookingProxy := proxy.New(t.Booking)
	paymentProxy := proxy.New(t.Payment)

	v1 := r.Group("/api/v1")
	{
		v1.Any("/auth/login", identityProxy)
		v1.Any("/auth/refresh", identityProxy)
		v1.Any("/auth/logout", identityProxy)
		v1.Any("/users/me", identityProxy)
		v1.Any("/users/me/bookings", bookingProxy) // owned by booking-service, not identity-service

		v1.Any("/events", eventProxy)
		v1.Any("/events/:slug", eventProxy)
		v1.Any("/organizer/events", eventProxy)
		v1.Any("/organizer/events/:id", eventProxy)
		v1.Any("/organizer/events/:id/publish", eventProxy)
		v1.Any("/organizer/events/:id/ticket-types", eventProxy)

		v1.Any("/bookings", bookingProxy)
		v1.Any("/bookings/:id", bookingProxy)

		v1.Any("/payments/:orderId/checkout", paymentProxy)
		v1.Any("/payments/webhook", paymentProxy)
	}
	return r
}
