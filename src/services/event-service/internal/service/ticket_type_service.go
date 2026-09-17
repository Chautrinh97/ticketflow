package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ticketflow/pkg/apperr"
	"ticketflow/services/event-service/internal/model"
)

// EventLookup is the subset of *repository.EventPostgresRepo that
// TicketTypeService actually calls (existence check before creating a
// ticket type) — defined here (the consuming package) per
// docs/01-architecture/backend-conventions.md's mock convention.
// *repository.EventPostgresRepo satisfies this automatically.
type EventLookup interface {
	GetByID(ctx context.Context, id string) (*model.Event, error)
}

// TicketTypeWriter is the subset of *repository.TicketTypePostgresRepo that
// TicketTypeService actually calls.
type TicketTypeWriter interface {
	Create(ctx context.Context, tt *model.TicketType) error
}

type TicketTypeService struct {
	events      EventLookup
	ticketTypes TicketTypeWriter
}

func NewTicketTypeService(events EventLookup, ticketTypes TicketTypeWriter) *TicketTypeService {
	return &TicketTypeService{events: events, ticketTypes: ticketTypes}
}

type CreateTicketTypeInput struct {
	Name     string
	Price    float64
	Currency string
	Quota    int
}

func (s *TicketTypeService) Create(ctx context.Context, eventID string, in CreateTicketTypeInput) (*model.TicketType, error) {
	if _, err := s.events.GetByID(ctx, eventID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	currency := in.Currency
	if currency == "" {
		currency = "VND"
	}
	tt := &model.TicketType{
		ID:       uuid.NewString(),
		EventID:  eventID,
		Name:     in.Name,
		Price:    in.Price,
		Currency: currency,
		Quota:    in.Quota,
	}
	if err := s.ticketTypes.Create(ctx, tt); err != nil {
		return nil, err
	}
	return tt, nil
}
