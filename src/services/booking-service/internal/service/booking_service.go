package service

import (
	"context"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/pagination"
	"ticketflow/services/booking-service/internal/model"
)

type BookingService struct {
	repo    Repository
	locker  Locker
	eventCl EventChecker
}

func NewBookingService(repo Repository, locker Locker, eventCl EventChecker) *BookingService {
	return &BookingService{repo: repo, locker: locker, eventCl: eventCl}
}

// CreateOrder orchestrates the full booking flow from docs/02-domains/booking/spec.md:
// acquire the Redis load-shedding lock(s), precheck each owning event is
// published (a correctness gap fixed per the implementation plan — the
// original flow doc didn't call this out), then run the ACID transaction.
func (s *BookingService) CreateOrder(ctx context.Context, userID string, items []model.BookingItem) (*model.Order, error) {
	if len(items) == 0 {
		return nil, apperr.WithMessage(apperr.ErrValidation, "Đơn hàng cần ít nhất 1 loại vé")
	}

	ticketTypeIDs := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, it := range items {
		if it.Quantity < 1 {
			return nil, apperr.WithMessage(apperr.ErrValidation, "Số lượng vé phải lớn hơn 0")
		}
		if _, ok := seen[it.TicketTypeID]; !ok {
			seen[it.TicketTypeID] = struct{}{}
			ticketTypeIDs = append(ticketTypeIDs, it.TicketTypeID)
		}
	}

	unlock, err := s.locker.AcquireMany(ctx, ticketTypeIDs)
	if err != nil {
		return nil, err
	}
	defer unlock()

	eventIDByTicketType, err := s.repo.GetEventIDsForTicketTypes(ctx, ticketTypeIDs)
	if err != nil {
		return nil, err
	}
	if len(eventIDByTicketType) != len(ticketTypeIDs) {
		return nil, apperr.WithMessage(apperr.ErrNotFound, "Một số loại vé không tồn tại")
	}

	checkedEvents := make(map[string]struct{}, len(eventIDByTicketType))
	for _, eventID := range eventIDByTicketType {
		if _, ok := checkedEvents[eventID]; ok {
			continue
		}
		checkedEvents[eventID] = struct{}{}
		ev, err := s.eventCl.GetEvent(ctx, eventID)
		if err != nil {
			return nil, apperr.FromGRPCError(err)
		}
		if ev.Status != "published" {
			return nil, apperr.WithMessage(apperr.ErrConflict, "Sự kiện chưa được xuất bản hoặc đã bị huỷ")
		}
	}

	return s.repo.CreateOrder(ctx, userID, items)
}

func (s *BookingService) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	return s.repo.GetOrderByID(ctx, id)
}

// GetOwnerID backs httpauth.RequireOwnership for GET /bookings/:id.
func (s *BookingService) GetOwnerID(ctx context.Context, orderID string) (string, error) {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return "", err
	}
	return order.UserID, nil
}

func (s *BookingService) ListMyOrders(ctx context.Context, userID, status string, p pagination.Params) ([]*model.Order, int64, error) {
	return s.repo.ListOrdersByUser(ctx, userID, status, p.Limit(), p.Offset())
}

// ConfirmPayment and FailPayment back the booking.proto RPCs consumed by
// payment-service.
func (s *BookingService) ConfirmPayment(ctx context.Context, orderID string) (*model.Order, error) {
	return s.repo.ConfirmPayment(ctx, orderID)
}

func (s *BookingService) FailPayment(ctx context.Context, orderID string) (*model.Order, error) {
	return s.repo.FailPayment(ctx, orderID)
}
