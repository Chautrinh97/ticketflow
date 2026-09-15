package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/pagination"
	"ticketflow/services/event-service/internal/repository"
	"ticketflow/services/event-service/internal/service"
)

type PublicEventHandler struct {
	events *service.EventService
}

func NewPublicEventHandler(events *service.EventService) *PublicEventHandler {
	return &PublicEventHandler{events: events}
}

func (h *PublicEventHandler) ListEvents(c *gin.Context) {
	filter := repository.EventFilter{
		Category: c.Query("category"),
		City:     c.Query("city"),
	}
	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.From = &t
		}
	}
	p := pagination.ParseParams(c.Query("page"), c.Query("page_size"))

	rows, total, err := h.events.ListPublished(c.Request.Context(), filter, p)
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	items := make([]eventSummaryDTO, 0, len(rows))
	for _, r := range rows {
		items = append(items, toEventSummaryDTO(r))
	}
	c.JSON(http.StatusOK, pagination.New(items, total))
}

func (h *PublicEventHandler) GetBySlug(c *gin.Context) {
	detail, err := h.events.GetDetailBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, toEventDetailDTO(detail))
}
