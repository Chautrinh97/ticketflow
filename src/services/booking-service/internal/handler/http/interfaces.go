package http

import (
	"context"

	"ticketflow/pkg/pagination"
	"ticketflow/services/booking-service/internal/model"
)

// Bookings is the subset of *service.BookingService's exported methods
// BookingHandler actually calls. Defined here at the consuming package so
// it can be mocked with mockery for unit tests (see backend-conventions.md
// "Unit test có mock") — *service.BookingService still satisfies this
// implicitly, so cmd/main.go's wiring keeps compiling unchanged.
type Bookings interface {
	CreateOrder(ctx context.Context, userID string, items []model.BookingItem) (*model.Order, error)
	GetOrderByID(ctx context.Context, id string) (*model.Order, error)
	ListMyOrders(ctx context.Context, userID, status string, p pagination.Params) ([]*model.Order, int64, error)
	GetOwnerID(ctx context.Context, orderID string) (string, error)
}
