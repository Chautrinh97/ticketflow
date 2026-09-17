package grpc

import (
	"context"

	"ticketflow/services/booking-service/internal/model"
)

// Bookings is the subset of *service.BookingService's exported methods
// BookingGRPCServer actually calls. Defined here at the consuming package
// so it can be mocked with mockery for unit tests (see
// backend-conventions.md "Unit test có mock") — *service.BookingService
// still satisfies this implicitly, so cmd/main.go's wiring keeps
// compiling unchanged.
type Bookings interface {
	GetOrderByID(ctx context.Context, id string) (*model.Order, error)
	ConfirmPayment(ctx context.Context, orderID string) (*model.Order, error)
	FailPayment(ctx context.Context, orderID string) (*model.Order, error)
}
