package grpc

import (
	"context"
	"time"

	"ticketflow/pkg/apperr"
	"ticketflow/proto/eventpb"
	"ticketflow/services/event-service/internal/service"
)

type EventGRPCServer struct {
	eventpb.UnimplementedEventServiceServer
	events *service.EventService
}

func NewEventGRPCServer(events *service.EventService) *EventGRPCServer {
	return &EventGRPCServer{events: events}
}

func (s *EventGRPCServer) GetEvent(ctx context.Context, req *eventpb.GetEventRequest) (*eventpb.GetEventResponse, error) {
	detail, err := s.events.GetDetailByID(ctx, req.GetEventId())
	if err != nil {
		return nil, apperr.AsGRPCStatus(err)
	}
	return &eventpb.GetEventResponse{
		Id:          detail.Event.ID,
		OrganizerId: detail.Event.OrganizerID,
		Status:      detail.Event.Status,
		Title:       detail.Event.Title,
		StartTime:   detail.Event.StartTime.Format(time.RFC3339),
	}, nil
}
