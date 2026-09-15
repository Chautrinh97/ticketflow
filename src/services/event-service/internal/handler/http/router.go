package http

import (
	"github.com/gin-gonic/gin"

	"ticketflow/pkg/httpauth"
)

// NewRouter mounts api-docs/openapi/event-service.yaml's Phase 1 endpoints,
// plus the Phase 1 addition GET /organizer/events/{id} (see organizer_event_handler.go).
func NewRouter(jwtSecret []byte, public *PublicEventHandler, organizer *OrganizerEventHandler, ticketTypes *TicketTypeHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	v1 := r.Group("/api/v1")
	{
		v1.GET("/events", public.ListEvents)
		v1.GET("/events/:slug", public.GetBySlug)

		org := v1.Group("/organizer/events")
		org.Use(httpauth.RequireAuth(jwtSecret), httpauth.RequireRole("organizer", "super_admin"))
		org.GET("", organizer.List)
		org.POST("", organizer.Create)
		org.GET("/:id", httpauth.RequireOwnership(organizer.OwnerLookup), organizer.GetByID)
		org.PATCH("/:id", httpauth.RequireOwnership(organizer.OwnerLookup), organizer.Update)
		org.DELETE("/:id", httpauth.RequireOwnership(organizer.OwnerLookup), organizer.Delete)
		org.POST("/:id/publish", httpauth.RequireOwnership(organizer.OwnerLookup), organizer.Publish)
		org.POST("/:id/ticket-types", httpauth.RequireOwnership(organizer.OwnerLookup), ticketTypes.Create)
	}
	return r
}
