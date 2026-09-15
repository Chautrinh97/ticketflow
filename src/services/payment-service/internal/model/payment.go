// Package model holds payment-service's GORM model for the `payments`
// table it owns (docs/03-data/postgres-schema.md#payment-service--payments).
// raw_payload JSONB is omitted from Phase 1's model — the mock flow never
// receives a real provider payload to store (see payment_service.go).
package model

import "time"

const (
	StatusInitiated = "initiated"
	StatusSuccess   = "success"
	StatusFailed    = "failed"
)

type Payment struct {
	ID            string `gorm:"column:id;primaryKey"`
	OrderID       string
	Provider      *string
	ProviderTxnID *string
	Amount        float64
	Status        string
	CreatedAt     time.Time
}

func (Payment) TableName() string { return "payments" }
