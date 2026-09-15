// Package model holds event-service's GORM models for the tables it owns
// (docs/03-data/postgres-schema.md#event-service--events-ticket_types).
package model

import "time"

const (
	CategoryConcert  = "concert"
	CategoryWorkshop = "workshop"
	CategorySport    = "sport"

	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusCancelled = "cancelled"
)

type Event struct {
	ID          string `gorm:"column:id;primaryKey"`
	OrganizerID string
	Title       string
	Slug        string
	Category    string
	VenueName   *string
	Address     *string
	City        *string
	StartTime   time.Time
	EndTime     *time.Time
	BannerURL   *string
	Description *string
	Status      string
	CreatedAt   time.Time
}

func (Event) TableName() string { return "events" }
