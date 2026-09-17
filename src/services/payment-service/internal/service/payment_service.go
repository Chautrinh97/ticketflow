package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ticketflow/pkg/apperr"
	"ticketflow/services/payment-service/internal/bookingclient"
	"ticketflow/services/payment-service/internal/model"
)

// PaymentStore is the narrow slice of *repository.PaymentRepository's
// methods PaymentService actually calls, extracted here at the consuming
// package per docs/01-architecture/backend-conventions.md's "Unit test có
// mock" convention — mockery can only generate a mock from an interface,
// not a concrete struct, and *repository.PaymentRepository already
// implements this implicitly, so no repository-side change is needed.
type PaymentStore interface {
	Create(ctx context.Context, p *model.Payment) error
	UpdateStatus(ctx context.Context, id, status string, providerTxnID *string) error
	GetLatestByOrderID(ctx context.Context, orderID string) (*model.Payment, error)
}

// BookingClient is the narrow slice of *bookingclient.Client's methods
// PaymentService actually calls, extracted for the same reason as
// PaymentStore above.
type BookingClient interface {
	GetOrder(ctx context.Context, orderID string) (*bookingclient.Order, error)
	ConfirmOrderPayment(ctx context.Context, orderID, paymentID string) (*bookingclient.ConfirmResult, error)
	FailOrderPayment(ctx context.Context, orderID, paymentID, reason string) (*bookingclient.ConfirmResult, error)
}

type PaymentService struct {
	payments  PaymentStore
	bookingCl BookingClient
}

func NewPaymentService(payments PaymentStore, bookingCl BookingClient) *PaymentService {
	return &PaymentService{payments: payments, bookingCl: bookingCl}
}

type CheckoutResult struct {
	PaymentID   string
	CheckoutURL string
}

// Checkout implements the Phase 1 mock payment flow (docs/02-domains/payment/spec.md:
// "Mock ở Phase 1... miễn luồng orders.status: pending -> paid hoạt động
// đúng"): it validates the order via booking-service's GetOrder RPC,
// records a `payments` row, and — per the implementation plan's resolved
// "payment mock flow" decision — immediately marks it successful and calls
// ConfirmOrderPayment in the same request, rather than waiting for a
// separate webhook call. The webhook endpoint still exists (HandleWebhook)
// for contract stability into Phase 2 and manual failure-path testing.
func (s *PaymentService) Checkout(ctx context.Context, orderID, provider string) (*CheckoutResult, error) {
	order, err := s.bookingCl.GetOrder(ctx, orderID)
	if err != nil {
		return nil, apperr.FromGRPCError(err)
	}
	if order.Status != "pending" {
		return nil, apperr.WithMessage(apperr.ErrConflict, "Đơn hàng không ở trạng thái chờ thanh toán")
	}
	if provider == "" {
		provider = "mock"
	}

	payment := &model.Payment{
		ID:       uuid.NewString(),
		OrderID:  orderID,
		Provider: &provider,
		Amount:   order.TotalAmount,
		Status:   model.StatusInitiated,
	}
	if err := s.payments.Create(ctx, payment); err != nil {
		return nil, err
	}
	if err := s.markSuccess(ctx, payment); err != nil {
		return nil, err
	}

	return &CheckoutResult{PaymentID: payment.ID, CheckoutURL: mockCheckoutURL(orderID)}, nil
}

func (s *PaymentService) markSuccess(ctx context.Context, payment *model.Payment) error {
	txnID := "mock-" + uuid.NewString()
	if err := s.payments.UpdateStatus(ctx, payment.ID, model.StatusSuccess, &txnID); err != nil {
		return err
	}
	if _, err := s.bookingCl.ConfirmOrderPayment(ctx, payment.OrderID, payment.ID); err != nil {
		return apperr.FromGRPCError(err)
	}
	return nil
}

func (s *PaymentService) markFailed(ctx context.Context, payment *model.Payment, reason string) error {
	if err := s.payments.UpdateStatus(ctx, payment.ID, model.StatusFailed, payment.ProviderTxnID); err != nil {
		return err
	}
	if _, err := s.bookingCl.FailOrderPayment(ctx, payment.OrderID, payment.ID, reason); err != nil {
		return apperr.FromGRPCError(err)
	}
	return nil
}

// HandleWebhook is a thin Phase 1 pass-through — no HMAC signature
// verification or provider_txn_id idempotency yet, both explicitly Phase 2
// per docs/02-domains/payment/spec.md's phase table.
func (s *PaymentService) HandleWebhook(ctx context.Context, orderID, status string) error {
	payment, err := s.payments.GetLatestByOrderID(ctx, orderID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperr.ErrNotFound
	}
	if err != nil {
		return err
	}
	if status == model.StatusSuccess {
		return s.markSuccess(ctx, payment)
	}
	return s.markFailed(ctx, payment, "payment_failed_webhook")
}

func mockCheckoutURL(orderID string) string {
	return "/orders/" + orderID + "/confirmation"
}
