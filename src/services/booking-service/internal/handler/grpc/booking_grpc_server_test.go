package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"ticketflow/pkg/apperr"
	"ticketflow/proto/bookingpb"
	"ticketflow/services/booking-service/internal/handler/grpc/mocks"
	"ticketflow/services/booking-service/internal/model"
)

const (
	testCaseSuccess_GetOrder_ReturnsSummary = "[Success] Trả OrderSummary đúng field"
	testCaseError_GetOrder_NotFound         = "[Error] Order không tồn tại -> gRPC NotFound"

	testCaseSuccess_ConfirmOrderPayment_ReturnsTickets = "[Success] Trả order_status + tickets đã sinh"
	testCaseError_ConfirmOrderPayment_Conflict         = "[Error] Sai trạng thái -> gRPC AlreadyExists (conflict)"

	testCaseSuccess_FailOrderPayment_ReturnsSummary = "[Success] Trả OrderSummary với status cancelled"
	testCaseError_FailOrderPayment_NotFound         = "[Error] Order không tồn tại -> gRPC NotFound"
)

type BookingGRPCServerTestSuite struct {
	suite.Suite
}

func TestBookingGRPCServerTestSuite(t *testing.T) {
	suite.Run(t, new(BookingGRPCServerTestSuite))
}

func (s *BookingGRPCServerTestSuite) TestGetOrder() {
	s.Run(testCaseSuccess_GetOrder_ReturnsSummary, func() {
		bookings := mocks.NewBookings(s.T())
		srv := NewBookingGRPCServer(bookings)
		bookings.On("GetOrderByID", mock.Anything, "order-1").
			Return(&model.Order{ID: "order-1", UserID: "user-1", Status: model.OrderStatusPaid, TotalAmount: 150000}, nil)

		resp, err := srv.GetOrder(context.Background(), &bookingpb.GetOrderRequest{OrderId: "order-1"})

		s.Require().NoError(err)
		s.Assert().Equal("order-1", resp.GetId())
		s.Assert().Equal("user-1", resp.GetUserId())
		s.Assert().Equal(model.OrderStatusPaid, resp.GetStatus())
		s.Assert().Equal(150000.0, resp.GetTotalAmount())
	})

	s.Run(testCaseError_GetOrder_NotFound, func() {
		bookings := mocks.NewBookings(s.T())
		srv := NewBookingGRPCServer(bookings)
		bookings.On("GetOrderByID", mock.Anything, "missing").Return(nil, apperr.ErrNotFound)

		resp, err := srv.GetOrder(context.Background(), &bookingpb.GetOrderRequest{OrderId: "missing"})

		s.Require().Nil(resp)
		st, ok := status.FromError(err)
		s.Require().True(ok)
		s.Assert().Equal(codes.NotFound, st.Code())
	})
}

func (s *BookingGRPCServerTestSuite) TestConfirmOrderPayment() {
	s.Run(testCaseSuccess_ConfirmOrderPayment_ReturnsTickets, func() {
		bookings := mocks.NewBookings(s.T())
		srv := NewBookingGRPCServer(bookings)
		bookings.On("ConfirmPayment", mock.Anything, "order-1").Return(&model.Order{
			ID: "order-1", Status: model.OrderStatusPaid,
			Tickets: []model.Ticket{{ID: "ticket-1", TicketCode: "ABC123"}},
		}, nil)

		resp, err := srv.ConfirmOrderPayment(context.Background(), &bookingpb.ConfirmOrderPaymentRequest{OrderId: "order-1"})

		s.Require().NoError(err)
		s.Assert().Equal(model.OrderStatusPaid, resp.GetOrderStatus())
		if s.Assert().Len(resp.GetTickets(), 1) {
			s.Assert().Equal("ticket-1", resp.GetTickets()[0].GetId())
			s.Assert().Equal("ABC123", resp.GetTickets()[0].GetTicketCode())
		}
	})

	s.Run(testCaseError_ConfirmOrderPayment_Conflict, func() {
		bookings := mocks.NewBookings(s.T())
		srv := NewBookingGRPCServer(bookings)
		bookings.On("ConfirmPayment", mock.Anything, "order-2").
			Return(nil, apperr.WithMessage(apperr.ErrConflict, "Đơn hàng không ở trạng thái chờ thanh toán"))

		resp, err := srv.ConfirmOrderPayment(context.Background(), &bookingpb.ConfirmOrderPaymentRequest{OrderId: "order-2"})

		s.Require().Nil(resp)
		st, ok := status.FromError(err)
		s.Require().True(ok)
		s.Assert().Equal(codes.AlreadyExists, st.Code())
	})
}

func (s *BookingGRPCServerTestSuite) TestFailOrderPayment() {
	s.Run(testCaseSuccess_FailOrderPayment_ReturnsSummary, func() {
		bookings := mocks.NewBookings(s.T())
		srv := NewBookingGRPCServer(bookings)
		bookings.On("FailPayment", mock.Anything, "order-1").
			Return(&model.Order{ID: "order-1", Status: model.OrderStatusCancelled}, nil)

		resp, err := srv.FailOrderPayment(context.Background(), &bookingpb.FailOrderPaymentRequest{OrderId: "order-1"})

		s.Require().NoError(err)
		s.Assert().Equal("order-1", resp.GetId())
		s.Assert().Equal(model.OrderStatusCancelled, resp.GetStatus())
	})

	s.Run(testCaseError_FailOrderPayment_NotFound, func() {
		bookings := mocks.NewBookings(s.T())
		srv := NewBookingGRPCServer(bookings)
		bookings.On("FailPayment", mock.Anything, "missing").Return(nil, apperr.ErrNotFound)

		resp, err := srv.FailOrderPayment(context.Background(), &bookingpb.FailOrderPaymentRequest{OrderId: "missing"})

		s.Require().Nil(resp)
		st, ok := status.FromError(err)
		s.Require().True(ok)
		s.Assert().Equal(codes.NotFound, st.Code())
	})
}
