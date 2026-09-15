// Package bookingclient is payment-service's gRPC client to booking-service
// (see src/proto/booking.proto) — payment-service never touches the
// orders/order_items/tickets tables directly, only through this RPC surface.
package bookingclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ticketflow/pkg/grpcinterceptor"
	"ticketflow/proto/bookingpb"
)

type Client struct {
	conn *grpc.ClientConn
	api  bookingpb.BookingServiceClient
}

func Dial(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(grpcinterceptor.UnaryClientLogging("payment-service")),
	)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, api: bookingpb.NewBookingServiceClient(conn)}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

type Order struct {
	ID          string
	UserID      string
	Status      string
	TotalAmount float64
}

func (c *Client) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	resp, err := c.api.GetOrder(ctx, &bookingpb.GetOrderRequest{OrderId: orderID})
	if err != nil {
		return nil, err
	}
	return &Order{ID: resp.GetId(), UserID: resp.GetUserId(), Status: resp.GetStatus(), TotalAmount: resp.GetTotalAmount()}, nil
}

type ConfirmResult struct {
	OrderStatus string
}

func (c *Client) ConfirmOrderPayment(ctx context.Context, orderID, paymentID string) (*ConfirmResult, error) {
	resp, err := c.api.ConfirmOrderPayment(ctx, &bookingpb.ConfirmOrderPaymentRequest{OrderId: orderID, PaymentId: paymentID})
	if err != nil {
		return nil, err
	}
	return &ConfirmResult{OrderStatus: resp.GetOrderStatus()}, nil
}

func (c *Client) FailOrderPayment(ctx context.Context, orderID, paymentID, reason string) (*ConfirmResult, error) {
	resp, err := c.api.FailOrderPayment(ctx, &bookingpb.FailOrderPaymentRequest{OrderId: orderID, PaymentId: paymentID, Reason: reason})
	if err != nil {
		return nil, err
	}
	return &ConfirmResult{OrderStatus: resp.GetStatus()}, nil
}
