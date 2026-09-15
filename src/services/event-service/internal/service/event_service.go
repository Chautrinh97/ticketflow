package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/pagination"
	"ticketflow/services/event-service/internal/model"
	"ticketflow/services/event-service/internal/repository"
)

type EventService struct {
	events      *repository.EventPostgresRepo
	ticketTypes *repository.TicketTypePostgresRepo
	catalog     *repository.EventCatalogMongoRepo
}

func NewEventService(events *repository.EventPostgresRepo, ticketTypes *repository.TicketTypePostgresRepo, catalog *repository.EventCatalogMongoRepo) *EventService {
	return &EventService{events: events, ticketTypes: ticketTypes, catalog: catalog}
}

type EventDetail struct {
	Event       *model.Event
	TicketTypes []model.TicketType
	Catalog     *model.EventCatalog
}

type CreateEventInput struct {
	Title       string
	Category    string
	VenueName   *string
	Address     *string
	City        *string
	StartTime   time.Time
	EndTime     *time.Time
	Description *string
	Attributes  map[string]any
	Tags        []string
}

// Create implements the two-phase Postgres+Mongo write mandated by
// docs/03-data/mongodb-schema.md: insert the events row inside a Postgres
// transaction, write the event_catalog Mongo document, and only commit
// Postgres if the Mongo write succeeds (no 2PC needed — a Mongo failure
// just rolls back the Postgres insert, so the event is simply "not created").
func (s *EventService) Create(ctx context.Context, organizerID string, in CreateEventInput) (*EventDetail, error) {
	event := &model.Event{
		ID:          uuid.NewString(),
		OrganizerID: organizerID,
		Title:       in.Title,
		Slug:        s.uniqueSlug(ctx, in.Title),
		Category:    in.Category,
		VenueName:   in.VenueName,
		Address:     in.Address,
		City:        in.City,
		StartTime:   in.StartTime,
		EndTime:     in.EndTime,
		Description: in.Description,
		Status:      model.StatusDraft,
	}

	attributes := in.Attributes
	if attributes == nil {
		attributes = map[string]any{}
	}
	catalog := &model.EventCatalog{
		EventID:    event.ID,
		Category:   in.Category,
		Attributes: attributes,
		Tags:       in.Tags,
	}

	err := s.events.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.events.CreateTx(tx, event); err != nil {
			return err
		}
		if err := s.catalog.Upsert(ctx, catalog); err != nil {
			return fmt.Errorf("write event_catalog: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &EventDetail{Event: event, TicketTypes: nil, Catalog: catalog}, nil
}

func (s *EventService) GetDetailBySlug(ctx context.Context, slug string) (*EventDetail, error) {
	event, err := s.events.GetPublishedBySlug(ctx, slug)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.assembleDetail(ctx, event)
}

func (s *EventService) GetDetailByID(ctx context.Context, id string) (*EventDetail, error) {
	event, err := s.events.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.assembleDetail(ctx, event)
}

// GetOwnerID backs httpauth.RequireOwnership for the PUT/DELETE/publish/
// ticket-type endpoints — it queries the real resource, never trusting
// client-supplied ownership claims, per docs/04-security/authorization.md.
func (s *EventService) GetOwnerID(ctx context.Context, eventID string) (string, error) {
	event, err := s.events.GetByID(ctx, eventID)
	if err != nil {
		return "", err
	}
	return event.OrganizerID, nil
}

func (s *EventService) assembleDetail(ctx context.Context, event *model.Event) (*EventDetail, error) {
	ticketTypes, err := s.ticketTypes.ListByEventID(ctx, event.ID)
	if err != nil {
		return nil, err
	}
	catalog, err := s.catalog.GetByEventID(ctx, event.ID)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		catalog = &model.EventCatalog{EventID: event.ID, Category: event.Category}
	}
	return &EventDetail{Event: event, TicketTypes: ticketTypes, Catalog: catalog}, nil
}

func (s *EventService) ListPublished(ctx context.Context, filter repository.EventFilter, p pagination.Params) ([]repository.EventSummaryRow, int64, error) {
	return s.events.ListPublished(ctx, filter, p)
}

func (s *EventService) ListByOrganizer(ctx context.Context, organizerID, status string, p pagination.Params) ([]repository.EventSummaryRow, int64, error) {
	return s.events.ListByOrganizer(ctx, organizerID, status, p)
}

// Publish transitions draft->published — the only transition Phase 1 needs.
func (s *EventService) Publish(ctx context.Context, id string) (*model.Event, error) {
	event, err := s.events.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if event.Status != model.StatusDraft {
		return nil, apperr.WithMessage(apperr.ErrConflict, "Chỉ có thể publish sự kiện đang ở trạng thái draft")
	}
	event.Status = model.StatusPublished
	if err := s.events.Update(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

type UpdateEventInput struct {
	Title       *string
	VenueName   *string
	Address     *string
	City        *string
	StartTime   *time.Time
	EndTime     *time.Time
	BannerURL   *string
	Description *string
	Attributes  map[string]any
	Tags        []string
}

func (s *EventService) Update(ctx context.Context, id string, in UpdateEventInput) (*EventDetail, error) {
	event, err := s.events.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if in.Title != nil {
		event.Title = *in.Title
	}
	if in.VenueName != nil {
		event.VenueName = in.VenueName
	}
	if in.Address != nil {
		event.Address = in.Address
	}
	if in.City != nil {
		event.City = in.City
	}
	if in.StartTime != nil {
		event.StartTime = *in.StartTime
	}
	if in.EndTime != nil {
		event.EndTime = in.EndTime
	}
	if in.BannerURL != nil {
		event.BannerURL = in.BannerURL
	}
	if in.Description != nil {
		event.Description = in.Description
	}
	if err := s.events.Update(ctx, event); err != nil {
		return nil, err
	}

	if in.Attributes != nil || in.Tags != nil {
		catalog, err := s.catalog.GetByEventID(ctx, event.ID)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err
			}
			catalog = &model.EventCatalog{EventID: event.ID, Category: event.Category}
		}
		if in.Attributes != nil {
			catalog.Attributes = in.Attributes
		}
		if in.Tags != nil {
			catalog.Tags = in.Tags
		}
		if err := s.catalog.Upsert(ctx, catalog); err != nil {
			return nil, err
		}
	}

	return s.assembleDetail(ctx, event)
}

// Cancel soft-deletes an event (status='cancelled'). Phase 1 does not check
// for existing paid orders before cancelling (that check needs a
// cross-service read into booking-service's data, out of Phase 1 scope —
// see implementation plan) — this is a deliberate scope trim, not an
// oversight.
func (s *EventService) Cancel(ctx context.Context, id string) error {
	event, err := s.events.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.ErrNotFound
	}
	if err != nil {
		return err
	}
	event.Status = model.StatusCancelled
	return s.events.Update(ctx, event)
}

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(title string) string {
	s := slugNonAlnum.ReplaceAllString(strings.ToLower(title), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "event"
	}
	return s
}

func (s *EventService) uniqueSlug(ctx context.Context, title string) string {
	base := slugify(title)
	slug := base
	for i := 0; i < 5; i++ {
		exists, err := s.events.SlugExists(ctx, slug)
		if err != nil || !exists {
			return slug
		}
		slug = fmt.Sprintf("%s-%s", base, uuid.NewString()[:8])
	}
	return fmt.Sprintf("%s-%s", base, uuid.NewString())
}
