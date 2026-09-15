// Package model holds booking-service's domain structs for the tables it
// owns (docs/03-data/postgres-schema.md#booking-service--orders-order_items-tickets).
// Booking-service uses raw pgx (no ORM), so these are plain structs, not
// GORM models.
package model

import "time"

const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusCancelled = "cancelled"
	OrderStatusExpired   = "expired"

	TicketStatusValid     = "valid"
	TicketStatusUsed      = "used"
	TicketStatusCancelled = "cancelled"
)

type Order struct {
	ID          string
	UserID      string
	Status      string
	TotalAmount float64
	ExpiresAt   time.Time
	CreatedAt   time.Time
	Items       []OrderItem
	Tickets     []Ticket
}

type OrderItem struct {
	ID           string
	OrderID      string
	TicketTypeID string
	Quantity     int
	UnitPrice    float64
}

type Ticket struct {
	ID          string
	OrderItemID string
	TicketCode  string
	Status      string
	IssuedAt    time.Time
}

// BookingItem is one requested {ticket_type_id, quantity} line from
// POST /bookings' items array.
type BookingItem struct {
	TicketTypeID string
	Quantity     int
}
