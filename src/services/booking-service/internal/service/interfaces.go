package service

import (
	"context"

	"ticketflow/services/booking-service/internal/eventclient"
	"ticketflow/services/booking-service/internal/lock"
	"ticketflow/services/booking-service/internal/model"
)

// Repository is the subset of *repository.BookingRepository's exported
// methods BookingService actually calls. Defined here at the consuming
// package (idiom "accept interfaces, return structs") purely so it can be
// mocked with mockery for unit tests — see backend-conventions.md
// "Unit test có mock". *repository.BookingRepository still satisfies this
// implicitly, so cmd/main.go's wiring keeps compiling unchanged.
type Repository interface {
	GetEventIDsForTicketTypes(ctx context.Context, ticketTypeIDs []string) (map[string]string, error)
	CreateOrder(ctx context.Context, userID string, items []model.BookingItem) (*model.Order, error)
	GetOrderByID(ctx context.Context, id string) (*model.Order, error)
	ListOrdersByUser(ctx context.Context, userID, status string, limit, offset int) ([]*model.Order, int64, error)
	ConfirmPayment(ctx context.Context, orderID string) (*model.Order, error)
	FailPayment(ctx context.Context, orderID string) (*model.Order, error)
}

// Locker matches *lock.TicketTypeLocker's AcquireMany.
type Locker interface {
	AcquireMany(ctx context.Context, ticketTypeIDs []string) (lock.Unlock, error)
}

// EventChecker matches *eventclient.Client's GetEvent.
type EventChecker interface {
	GetEvent(ctx context.Context, eventID string) (*eventclient.Event, error)
}
