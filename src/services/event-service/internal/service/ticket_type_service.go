package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ticketflow/pkg/apperr"
	"ticketflow/services/event-service/internal/model"
	"ticketflow/services/event-service/internal/repository"
)

type TicketTypeService struct {
	events      *repository.EventPostgresRepo
	ticketTypes *repository.TicketTypePostgresRepo
}

func NewTicketTypeService(events *repository.EventPostgresRepo, ticketTypes *repository.TicketTypePostgresRepo) *TicketTypeService {
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
