package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ticketflow/pkg/pagination"
	"ticketflow/services/event-service/internal/model"
)

type EventPostgresRepo struct {
	db *gorm.DB
}

func NewEventPostgresRepo(db *gorm.DB) *EventPostgresRepo {
	return &EventPostgresRepo{db: db}
}

// DB exposes the underlying *gorm.DB so EventService can open a
// Postgres+Mongo two-phase write transaction (see mongodb-schema.md's
// "không cần 2-phase commit" note: insert events row, write Mongo doc,
// commit only if Mongo succeeds).
func (r *EventPostgresRepo) DB() *gorm.DB { return r.db }

func (r *EventPostgresRepo) CreateTx(tx *gorm.DB, e *model.Event) error {
	return tx.Create(e).Error
}

func (r *EventPostgresRepo) GetByID(ctx context.Context, id string) (*model.Event, error) {
	var e model.Event
	if err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EventPostgresRepo) GetPublishedBySlug(ctx context.Context, slug string) (*model.Event, error) {
	var e model.Event
	if err := r.db.WithContext(ctx).First(&e, "slug = ? AND status = ?", slug, model.StatusPublished).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EventPostgresRepo) Update(ctx context.Context, e *model.Event) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *EventPostgresRepo) SlugExists(ctx context.Context, slug string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Event{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

type EventFilter struct {
	Category string
	City     string
	From     *time.Time
}

// EventSummaryRow scans exactly the columns EventSummary needs, plus a
// computed min_price (MIN over the event's ticket_types.price) — explicit
// column list rather than embedding model.Event to keep the raw-SQL
// projection unambiguous.
type EventSummaryRow struct {
	ID        string    `gorm:"column:id"`
	Title     string    `gorm:"column:title"`
	Slug      string    `gorm:"column:slug"`
	Category  string    `gorm:"column:category"`
	City      *string   `gorm:"column:city"`
	StartTime time.Time `gorm:"column:start_time"`
	BannerURL *string   `gorm:"column:banner_url"`
	MinPrice  *float64  `gorm:"column:min_price"`
}

func (r *EventPostgresRepo) filteredQuery(ctx context.Context, status string, filter EventFilter) *gorm.DB {
	q := r.db.WithContext(ctx).Table("events").Where("events.status = ?", status)
	if filter.Category != "" {
		q = q.Where("events.category = ?", filter.Category)
	}
	if filter.City != "" {
		q = q.Where("events.city = ?", filter.City)
	}
	if filter.From != nil {
		q = q.Where("events.start_time >= ?", *filter.From)
	}
	return q
}

func (r *EventPostgresRepo) ListPublished(ctx context.Context, filter EventFilter, p pagination.Params) ([]EventSummaryRow, int64, error) {
	var total int64
	if err := r.filteredQuery(ctx, model.StatusPublished, filter).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []EventSummaryRow
	err := r.filteredQuery(ctx, model.StatusPublished, filter).
		Select(`events.id, events.title, events.slug, events.category, events.city, events.start_time, events.banner_url,
			(SELECT MIN(price) FROM ticket_types WHERE ticket_types.event_id = events.id) AS min_price`).
		Order("events.start_time ASC").
		Offset(p.Offset()).Limit(p.Limit()).
		Scan(&rows).Error
	return rows, total, err
}

// ListByOrganizer lists an organizer's own events (any status) — status="" means all.
func (r *EventPostgresRepo) ListByOrganizer(ctx context.Context, organizerID, status string, p pagination.Params) ([]EventSummaryRow, int64, error) {
	base := r.db.WithContext(ctx).Table("events").Where("events.organizer_id = ?", organizerID)
	if status != "" {
		base = base.Where("events.status = ?", status)
	}

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []EventSummaryRow
	err := base.Select(`events.id, events.title, events.slug, events.category, events.city, events.start_time, events.banner_url,
			(SELECT MIN(price) FROM ticket_types WHERE ticket_types.event_id = events.id) AS min_price`).
		Order("events.start_time DESC").
		Offset(p.Offset()).Limit(p.Limit()).
		Scan(&rows).Error
	return rows, total, err
}
