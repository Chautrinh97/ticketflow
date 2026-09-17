package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"ticketflow/pkg/apperr"
	"ticketflow/services/payment-service/internal/bookingclient"
	"ticketflow/services/payment-service/internal/model"
	"ticketflow/services/payment-service/internal/service/mocks"
)

const (
	testCaseSuccess_Checkout_HappyPath              = "[Success] Checkout thành công với provider chỉ định"
	testCaseSuccess_Checkout_DefaultsProviderToMock = "[Success] Checkout để trống provider mặc định thành 'mock'"
	testCaseError_Checkout_OrderNotPending          = "[Error] Đơn hàng không ở trạng thái pending bị từ chối"
	testCaseError_Checkout_GetOrderGRPCError        = "[Error] GetOrder lỗi gRPC được map qua apperr.FromGRPCError"
	testCaseError_Checkout_CreateRepoError          = "[Error] Create payment lỗi trả nguyên lỗi repo"
	testCaseError_Checkout_MarkSuccessPropagates    = "[Error] Lỗi ở bước markSuccess (ConfirmOrderPayment) được trả về Checkout"

	testCaseSuccess_MarkSuccess_HappyPath                  = "[Success] Cập nhật status=success và gọi ConfirmOrderPayment"
	testCaseError_MarkSuccess_UpdateStatusError            = "[Error] UpdateStatus lỗi thì không gọi ConfirmOrderPayment"
	testCaseError_MarkSuccess_ConfirmOrderPaymentGRPCError = "[Error] ConfirmOrderPayment lỗi gRPC được map qua apperr"

	testCaseSuccess_MarkFailed_HappyPath               = "[Success] Cập nhật status=failed và gọi FailOrderPayment với đúng reason"
	testCaseError_MarkFailed_UpdateStatusError         = "[Error] UpdateStatus lỗi thì không gọi FailOrderPayment"
	testCaseError_MarkFailed_FailOrderPaymentGRPCError = "[Error] FailOrderPayment lỗi gRPC được map qua apperr"

	testCaseSuccess_HandleWebhook_RoutesToMarkSuccess = "[Success] status=success điều hướng sang markSuccess"
	testCaseSuccess_HandleWebhook_RoutesToMarkFailed  = "[Success] status khác success (vd 'failed') điều hướng sang markFailed"
	testCaseError_HandleWebhook_NotFound              = "[Error] Không tìm thấy payment (gorm.ErrRecordNotFound) trả về apperr.ErrNotFound"
	testCaseError_HandleWebhook_RepoErrorPropagated   = "[Error] Lỗi repo khác (không phải NotFound) được trả nguyên"
)

type PaymentServiceTestSuite struct {
	suite.Suite
	store   *mocks.PaymentStore
	booking *mocks.BookingClient
	svc     *PaymentService
}

func TestPaymentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentServiceTestSuite))
}

// SetupTest/SetupSubTest both reinitialize fresh mocks bound to the
// currently-running (sub)test's *testing.T, so mockery's generated
// AssertExpectations cleanup fires per case, not once for the whole suite.
func (s *PaymentServiceTestSuite) SetupTest()    { s.initMocks() }
func (s *PaymentServiceTestSuite) SetupSubTest() { s.initMocks() }

func (s *PaymentServiceTestSuite) initMocks() {
	s.store = mocks.NewPaymentStore(s.T())
	s.booking = mocks.NewBookingClient(s.T())
	s.svc = NewPaymentService(s.store, s.booking)
}

// requireAppErrCode asserts err is an *apperr.Error with the given Code —
// per backend-conventions.md's "Unit test có mock" note, *apperr.Error has
// no Is(target) method, so errors.Is(err, apperr.ErrX) would never match;
// compare .Code via errors.As instead.
func requireAppErrCode(t require.TestingT, err error, wantCode string) {
	var appErr *apperr.Error
	ok := errors.As(err, &appErr)
	require.True(t, ok, "expected *apperr.Error, got %#v", err)
	require.Equal(t, wantCode, appErr.Code)
}

func (s *PaymentServiceTestSuite) TestCheckout() {
	s.Run(testCaseSuccess_Checkout_HappyPath, func() {
		const orderID = "order-1"
		s.booking.On("GetOrder", mock.Anything, orderID).
			Return(&bookingclient.Order{ID: orderID, Status: "pending", TotalAmount: 150000}, nil)
		s.store.On("Create", mock.Anything, mock.MatchedBy(func(p *model.Payment) bool {
			return p.OrderID == orderID && p.Status == model.StatusInitiated &&
				p.Provider != nil && *p.Provider == "vnpay" && p.Amount == 150000 && p.ID != ""
		})).Return(nil)
		s.store.On("UpdateStatus", mock.Anything, mock.AnythingOfType("string"), model.StatusSuccess, mock.AnythingOfType("*string")).
			Return(nil)
		s.booking.On("ConfirmOrderPayment", mock.Anything, orderID, mock.AnythingOfType("string")).
			Return(&bookingclient.ConfirmResult{OrderStatus: "paid"}, nil)

		result, err := s.svc.Checkout(context.Background(), orderID, "vnpay")

		s.Require().NoError(err)
		s.Require().NotEmpty(result.PaymentID)
		s.Assert().Equal("/orders/"+orderID+"/confirmation", result.CheckoutURL)
	})

	s.Run(testCaseSuccess_Checkout_DefaultsProviderToMock, func() {
		const orderID = "order-2"
		s.booking.On("GetOrder", mock.Anything, orderID).
			Return(&bookingclient.Order{ID: orderID, Status: "pending", TotalAmount: 50000}, nil)
		s.store.On("Create", mock.Anything, mock.MatchedBy(func(p *model.Payment) bool {
			return p.Provider != nil && *p.Provider == "mock"
		})).Return(nil)
		s.store.On("UpdateStatus", mock.Anything, mock.Anything, model.StatusSuccess, mock.Anything).Return(nil)
		s.booking.On("ConfirmOrderPayment", mock.Anything, orderID, mock.Anything).
			Return(&bookingclient.ConfirmResult{OrderStatus: "paid"}, nil)

		_, err := s.svc.Checkout(context.Background(), orderID, "")

		s.Require().NoError(err)
	})

	s.Run(testCaseError_Checkout_OrderNotPending, func() {
		const orderID = "order-3"
		s.booking.On("GetOrder", mock.Anything, orderID).
			Return(&bookingclient.Order{ID: orderID, Status: "paid", TotalAmount: 50000}, nil)

		result, err := s.svc.Checkout(context.Background(), orderID, "mock")

		s.Require().Nil(result)
		requireAppErrCode(s.T(), err, apperr.ErrConflict.Code)
		s.store.AssertNotCalled(s.T(), "Create", mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_Checkout_GetOrderGRPCError, func() {
		const orderID = "order-4"
		s.booking.On("GetOrder", mock.Anything, orderID).
			Return(nil, status.Error(codes.NotFound, "order not found"))

		result, err := s.svc.Checkout(context.Background(), orderID, "mock")

		s.Require().Nil(result)
		requireAppErrCode(s.T(), err, apperr.ErrNotFound.Code)
	})

	s.Run(testCaseError_Checkout_CreateRepoError, func() {
		const orderID = "order-5"
		repoErr := errors.New("insert failed: connection reset")
		s.booking.On("GetOrder", mock.Anything, orderID).
			Return(&bookingclient.Order{ID: orderID, Status: "pending", TotalAmount: 10000}, nil)
		s.store.On("Create", mock.Anything, mock.Anything).Return(repoErr)

		result, err := s.svc.Checkout(context.Background(), orderID, "mock")

		s.Require().Nil(result)
		s.Assert().ErrorIs(err, repoErr)
		s.booking.AssertNotCalled(s.T(), "ConfirmOrderPayment", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_Checkout_MarkSuccessPropagates, func() {
		const orderID = "order-6"
		s.booking.On("GetOrder", mock.Anything, orderID).
			Return(&bookingclient.Order{ID: orderID, Status: "pending", TotalAmount: 10000}, nil)
		s.store.On("Create", mock.Anything, mock.Anything).Return(nil)
		s.store.On("UpdateStatus", mock.Anything, mock.Anything, model.StatusSuccess, mock.Anything).Return(nil)
		s.booking.On("ConfirmOrderPayment", mock.Anything, orderID, mock.Anything).
			Return(nil, status.Error(codes.Unauthenticated, "booking rejected caller"))

		result, err := s.svc.Checkout(context.Background(), orderID, "mock")

		s.Require().Nil(result)
		requireAppErrCode(s.T(), err, apperr.ErrUnauthorized.Code)
	})
}

func (s *PaymentServiceTestSuite) TestMarkSuccess() {
	s.Run(testCaseSuccess_MarkSuccess_HappyPath, func() {
		payment := &model.Payment{ID: "pay-1", OrderID: "order-1", Status: model.StatusInitiated}
		s.store.On("UpdateStatus", mock.Anything, "pay-1", model.StatusSuccess, mock.AnythingOfType("*string")).
			Return(nil)
		s.booking.On("ConfirmOrderPayment", mock.Anything, "order-1", "pay-1").
			Return(&bookingclient.ConfirmResult{OrderStatus: "paid"}, nil)

		err := s.svc.markSuccess(context.Background(), payment)

		s.Require().NoError(err)
	})

	s.Run(testCaseError_MarkSuccess_UpdateStatusError, func() {
		payment := &model.Payment{ID: "pay-2", OrderID: "order-2"}
		repoErr := errors.New("update failed")
		s.store.On("UpdateStatus", mock.Anything, "pay-2", model.StatusSuccess, mock.Anything).Return(repoErr)

		err := s.svc.markSuccess(context.Background(), payment)

		s.Assert().ErrorIs(err, repoErr)
		s.booking.AssertNotCalled(s.T(), "ConfirmOrderPayment", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_MarkSuccess_ConfirmOrderPaymentGRPCError, func() {
		payment := &model.Payment{ID: "pay-3", OrderID: "order-3"}
		s.store.On("UpdateStatus", mock.Anything, "pay-3", model.StatusSuccess, mock.Anything).Return(nil)
		s.booking.On("ConfirmOrderPayment", mock.Anything, "order-3", "pay-3").
			Return(nil, status.Error(codes.AlreadyExists, "order already confirmed"))

		err := s.svc.markSuccess(context.Background(), payment)

		requireAppErrCode(s.T(), err, apperr.ErrConflict.Code)
	})
}

func (s *PaymentServiceTestSuite) TestMarkFailed() {
	s.Run(testCaseSuccess_MarkFailed_HappyPath, func() {
		payment := &model.Payment{ID: "pay-4", OrderID: "order-4"}
		s.store.On("UpdateStatus", mock.Anything, "pay-4", model.StatusFailed, payment.ProviderTxnID).Return(nil)
		s.booking.On("FailOrderPayment", mock.Anything, "order-4", "pay-4", "gateway_declined").
			Return(&bookingclient.ConfirmResult{OrderStatus: "payment_failed"}, nil)

		err := s.svc.markFailed(context.Background(), payment, "gateway_declined")

		s.Require().NoError(err)
	})

	s.Run(testCaseError_MarkFailed_UpdateStatusError, func() {
		payment := &model.Payment{ID: "pay-5", OrderID: "order-5"}
		repoErr := errors.New("update failed")
		s.store.On("UpdateStatus", mock.Anything, "pay-5", model.StatusFailed, mock.Anything).Return(repoErr)

		err := s.svc.markFailed(context.Background(), payment, "reason")

		s.Assert().ErrorIs(err, repoErr)
		s.booking.AssertNotCalled(s.T(), "FailOrderPayment", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_MarkFailed_FailOrderPaymentGRPCError, func() {
		payment := &model.Payment{ID: "pay-6", OrderID: "order-6"}
		s.store.On("UpdateStatus", mock.Anything, "pay-6", model.StatusFailed, mock.Anything).Return(nil)
		s.booking.On("FailOrderPayment", mock.Anything, "order-6", "pay-6", "reason").
			Return(nil, status.Error(codes.InvalidArgument, "bad payment id"))

		err := s.svc.markFailed(context.Background(), payment, "reason")

		requireAppErrCode(s.T(), err, apperr.ErrValidation.Code)
	})
}

func (s *PaymentServiceTestSuite) TestHandleWebhook() {
	s.Run(testCaseSuccess_HandleWebhook_RoutesToMarkSuccess, func() {
		const orderID = "order-7"
		payment := &model.Payment{ID: "pay-7", OrderID: orderID, Status: model.StatusInitiated}
		s.store.On("GetLatestByOrderID", mock.Anything, orderID).Return(payment, nil)
		s.store.On("UpdateStatus", mock.Anything, "pay-7", model.StatusSuccess, mock.Anything).Return(nil)
		s.booking.On("ConfirmOrderPayment", mock.Anything, orderID, "pay-7").
			Return(&bookingclient.ConfirmResult{OrderStatus: "paid"}, nil)

		err := s.svc.HandleWebhook(context.Background(), orderID, model.StatusSuccess)

		s.Require().NoError(err)
		s.booking.AssertNotCalled(s.T(), "FailOrderPayment", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseSuccess_HandleWebhook_RoutesToMarkFailed, func() {
		const orderID = "order-8"
		payment := &model.Payment{ID: "pay-8", OrderID: orderID, Status: model.StatusInitiated}
		s.store.On("GetLatestByOrderID", mock.Anything, orderID).Return(payment, nil)
		s.store.On("UpdateStatus", mock.Anything, "pay-8", model.StatusFailed, mock.Anything).Return(nil)
		s.booking.On("FailOrderPayment", mock.Anything, orderID, "pay-8", "payment_failed_webhook").
			Return(&bookingclient.ConfirmResult{OrderStatus: "payment_failed"}, nil)

		err := s.svc.HandleWebhook(context.Background(), orderID, model.StatusFailed)

		s.Require().NoError(err)
		s.booking.AssertNotCalled(s.T(), "ConfirmOrderPayment", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_HandleWebhook_NotFound, func() {
		const orderID = "order-9"
		s.store.On("GetLatestByOrderID", mock.Anything, orderID).Return(nil, gorm.ErrRecordNotFound)

		err := s.svc.HandleWebhook(context.Background(), orderID, model.StatusSuccess)

		requireAppErrCode(s.T(), err, apperr.ErrNotFound.Code)
	})

	s.Run(testCaseError_HandleWebhook_RepoErrorPropagated, func() {
		const orderID = "order-10"
		repoErr := errors.New("db connection lost")
		s.store.On("GetLatestByOrderID", mock.Anything, orderID).Return(nil, repoErr)

		err := s.svc.HandleWebhook(context.Background(), orderID, model.StatusSuccess)

		s.Assert().ErrorIs(err, repoErr)
	})
}
