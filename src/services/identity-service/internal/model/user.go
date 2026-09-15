// Package model holds identity-service's GORM models for the `users` table
// it owns (docs/03-data/postgres-schema.md#identity-service--users).
package model

import "time"

const (
	RoleSuperAdmin = "super_admin"
	RoleOrganizer  = "organizer"
	RoleUser       = "user"

	StatusActive  = "active"
	StatusPending = "pending"
	StatusBanned  = "banned"
)

type User struct {
	ID          string `gorm:"column:id;primaryKey"`
	FirebaseUID *string
	Email       string
	FullName    *string
	AvatarURL   *string
	Role        string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (User) TableName() string { return "users" }
