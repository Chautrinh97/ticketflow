package model

// Note: booking-service is the only writer of SoldCount, within its own
// ACID transaction — event-service only reads it here (see
// docs/02-domains/booking/spec.md).
type TicketType struct {
	ID        string `gorm:"column:id;primaryKey"`
	EventID   string
	Name      string
	Price     float64
	Currency  string
	Quota     int
	SoldCount int
	Version   int
}

func (TicketType) TableName() string { return "ticket_types" }
