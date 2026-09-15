// Package grpc implements bookingpb.BookingServiceServer — the internal RPC
// surface in src/proto/booking.proto, consumed by payment-service.
package grpc

import (
	"context"

	"ticketflow/pkg/apperr"
	"ticketflow/proto/bookingpb"
	"ticketflow/services/booking-service/internal/model"
	"ticketflow/services/booking-service/internal/service"
)

type BookingGRPCServer struct {
	bookingpb.UnimplementedBookingServiceServer
	bookings *service.BookingService
}

func NewBookingGRPCServer(bookings *service.BookingService) *BookingGRPCServer {
	return &BookingGRPCServer{bookings: bookings}
}

func (s *BookingGRPCServer) GetOrder(ctx context.Context, req *bookingpb.GetOrderRequest) (*bookingpb.OrderSummary, error) {
	order, err := s.bookings.GetOrderByID(ctx, req.GetOrderId())
	if err != nil {
		return nil, apperr.AsGRPCStatus(err)
	}
	return &bookingpb.OrderSummary{Id: order.ID, UserId: order.UserID, Status: order.Status, TotalAmount: order.TotalAmount}, nil
}

func (s *BookingGRPCServer) ConfirmOrderPayment(ctx context.Context, req *bookingpb.ConfirmOrderPaymentRequest) (*bookingpb.ConfirmOrderPaymentResponse, error) {
	order, err := s.bookings.ConfirmPayment(ctx, req.GetOrderId())
	if err != nil {
		return nil, apperr.AsGRPCStatus(err)
	}
	return &bookingpb.ConfirmOrderPaymentResponse{
		OrderStatus: order.Status,
		Tickets:     toTicketSummaries(order.Tickets),
	}, nil
}

func (s *BookingGRPCServer) FailOrderPayment(ctx context.Context, req *bookingpb.FailOrderPaymentRequest) (*bookingpb.OrderSummary, error) {
	order, err := s.bookings.FailPayment(ctx, req.GetOrderId())
	if err != nil {
		return nil, apperr.AsGRPCStatus(err)
	}
	return &bookingpb.OrderSummary{Id: order.ID, Status: order.Status}, nil
}

func toTicketSummaries(tickets []model.Ticket) []*bookingpb.TicketSummary {
	out := make([]*bookingpb.TicketSummary, 0, len(tickets))
	for _, t := range tickets {
		out = append(out, &bookingpb.TicketSummary{Id: t.ID, TicketCode: t.TicketCode})
	}
	return out
}
