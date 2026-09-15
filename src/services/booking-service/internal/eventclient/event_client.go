// Package eventclient is booking-service's gRPC client to event-service,
// used only for the published-event precheck (see src/proto/event.proto).
package eventclient

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"ticketflow/pkg/grpcinterceptor"
	"ticketflow/proto/eventpb"
)

type Client struct {
	conn *grpc.ClientConn
	api  eventpb.EventServiceClient
}

func Dial(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(grpcinterceptor.UnaryClientLogging("booking-service")),
	)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, api: eventpb.NewEventServiceClient(conn)}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

type Event struct {
	ID          string
	OrganizerID string
	Status      string
	Title       string
}

func (c *Client) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	resp, err := c.api.GetEvent(ctx, &eventpb.GetEventRequest{EventId: eventID})
	if err != nil {
		return nil, err
	}
	return &Event{ID: resp.GetId(), OrganizerID: resp.GetOrganizerId(), Status: resp.GetStatus(), Title: resp.GetTitle()}, nil
}
