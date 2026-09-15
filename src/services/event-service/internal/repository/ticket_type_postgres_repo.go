package repository

import (
	"context"

	"gorm.io/gorm"

	"ticketflow/services/event-service/internal/model"
)

type TicketTypePostgresRepo struct {
	db *gorm.DB
}

func NewTicketTypePostgresRepo(db *gorm.DB) *TicketTypePostgresRepo {
	return &TicketTypePostgresRepo{db: db}
}

func (r *TicketTypePostgresRepo) Create(ctx context.Context, tt *model.TicketType) error {
	return r.db.WithContext(ctx).Create(tt).Error
}

func (r *TicketTypePostgresRepo) ListByEventID(ctx context.Context, eventID string) ([]model.TicketType, error) {
	var rows []model.TicketType
	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Order("price ASC").Find(&rows).Error
	return rows, err
}
