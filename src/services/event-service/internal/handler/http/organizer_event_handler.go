package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/httpauth"
	"ticketflow/pkg/pagination"
	"ticketflow/services/event-service/internal/model"
	"ticketflow/services/event-service/internal/repository"
	"ticketflow/services/event-service/internal/service"
)

// OrganizerEventService is the subset of *service.EventService that
// OrganizerEventHandler actually calls — defined here (the consuming
// package) so it can be mocked with mockery, per
// docs/01-architecture/backend-conventions.md's mock convention.
// *service.EventService satisfies this automatically.
type OrganizerEventService interface {
	ListByOrganizer(ctx context.Context, organizerID, status string, p pagination.Params) ([]repository.EventSummaryRow, int64, error)
	Create(ctx context.Context, organizerID string, in service.CreateEventInput) (*service.EventDetail, error)
	GetDetailByID(ctx context.Context, id string) (*service.EventDetail, error)
	Update(ctx context.Context, id string, in service.UpdateEventInput) (*service.EventDetail, error)
	Cancel(ctx context.Context, id string) error
	Publish(ctx context.Context, id string) (*model.Event, error)
	GetOwnerID(ctx context.Context, eventID string) (string, error)
}

type OrganizerEventHandler struct {
	events OrganizerEventService
}

func NewOrganizerEventHandler(events OrganizerEventService) *OrganizerEventHandler {
	return &OrganizerEventHandler{events: events}
}

func (h *OrganizerEventHandler) List(c *gin.Context) {
	p := pagination.ParseParams(c.Query("page"), c.Query("page_size"))
	rows, total, err := h.events.ListByOrganizer(c.Request.Context(), httpauth.UserID(c), c.Query("status"), p)
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

type createEventRequest struct {
	Title       string         `json:"title" binding:"required"`
	Category    string         `json:"category" binding:"required"`
	VenueName   *string        `json:"venue_name"`
	Address     *string        `json:"address"`
	City        *string        `json:"city"`
	StartTime   time.Time      `json:"start_time" binding:"required"`
	EndTime     *time.Time     `json:"end_time"`
	Description *string        `json:"description"`
	Attributes  map[string]any `json:"attributes"`
	Tags        []string       `json:"tags"`
}

func (h *OrganizerEventHandler) Create(c *gin.Context) {
	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Respond(c, apperr.WithMessage(apperr.ErrValidation, "Dữ liệu tạo sự kiện không hợp lệ"))
		return
	}
	detail, err := h.events.Create(c.Request.Context(), httpauth.UserID(c), service.CreateEventInput{
		Title: req.Title, Category: req.Category, VenueName: req.VenueName, Address: req.Address,
		City: req.City, StartTime: req.StartTime, EndTime: req.EndTime, Description: req.Description,
		Attributes: req.Attributes, Tags: req.Tags,
	})
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, toEventDetailDTO(detail))
}

// GetByID backs the Phase 1 addition GET /organizer/events/{id} (not in the
// original OpenAPI file — added per the implementation plan so
// event-manage.md has a reloadable single-event read; update
// api-docs/openapi/event-service.yaml + event-catalog/spec.md alongside).
func (h *OrganizerEventHandler) GetByID(c *gin.Context) {
	detail, err := h.events.GetDetailByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, toEventDetailDTO(detail))
}

type updateEventRequest struct {
	Title       *string        `json:"title"`
	VenueName   *string        `json:"venue_name"`
	Address     *string        `json:"address"`
	City        *string        `json:"city"`
	StartTime   *time.Time     `json:"start_time"`
	EndTime     *time.Time     `json:"end_time"`
	BannerURL   *string        `json:"banner_url"`
	Description *string        `json:"description"`
	Attributes  map[string]any `json:"attributes"`
	Tags        []string       `json:"tags"`
}

func (h *OrganizerEventHandler) Update(c *gin.Context) {
	var req updateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperr.Respond(c, apperr.WithMessage(apperr.ErrValidation, "Dữ liệu cập nhật không hợp lệ"))
		return
	}
	detail, err := h.events.Update(c.Request.Context(), c.Param("id"), service.UpdateEventInput{
		Title: req.Title, VenueName: req.VenueName, Address: req.Address, City: req.City,
		StartTime: req.StartTime, EndTime: req.EndTime, BannerURL: req.BannerURL, Description: req.Description,
		Attributes: req.Attributes, Tags: req.Tags,
	})
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, toEventDetailDTO(detail))
}

func (h *OrganizerEventHandler) Delete(c *gin.Context) {
	if err := h.events.Cancel(c.Request.Context(), c.Param("id")); err != nil {
		apperr.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *OrganizerEventHandler) Publish(c *gin.Context) {
	event, err := h.events.Publish(c.Request.Context(), c.Param("id"))
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	detail, err := h.events.GetDetailByID(c.Request.Context(), event.ID)
	if err != nil {
		apperr.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, toEventDetailDTO(detail))
}

// OwnerLookup backs httpauth.RequireOwnership for PUT/DELETE/publish/ticket-types —
// it queries the real event row, never trusting client input.
func (h *OrganizerEventHandler) OwnerLookup(c *gin.Context) (string, error) {
	return h.events.GetOwnerID(c.Request.Context(), c.Param("id"))
}
