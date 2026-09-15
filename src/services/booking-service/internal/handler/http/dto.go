package http

import (
	"time"

	"ticketflow/services/booking-service/internal/model"
)

// DTOs mirror api-docs/openapi/booking-service.yaml's Order/OrderItem/Ticket
// schemas (the request body is extended to items[] per the implementation
// plan's multi-item decision — update the OpenAPI file alongside this code).

type orderItemDTO struct {
	ID           string  `json:"id"`
	TicketTypeID string  `json:"ticket_type_id"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
}

type ticketDTO struct {
	ID          string `json:"id"`
	OrderItemID string `json:"order_item_id"`
	TicketCode  string `json:"ticket_code"`
	Status      string `json:"status"`
	IssuedAt    string `json:"issued_at"`
}

type orderDTO struct {
	ID          string         `json:"id"`
	UserID      string         `json:"user_id"`
	Status      string         `json:"status"`
	TotalAmount float64        `json:"total_amount"`
	ExpiresAt   string         `json:"expires_at"`
	CreatedAt   string         `json:"created_at"`
	Items       []orderItemDTO `json:"items"`
	Tickets     []ticketDTO    `json:"tickets"`
}

func toOrderDTO(o *model.Order) orderDTO {
	items := make([]orderItemDTO, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, orderItemDTO{ID: it.ID, TicketTypeID: it.TicketTypeID, Quantity: it.Quantity, UnitPrice: it.UnitPrice})
	}
	tickets := make([]ticketDTO, 0, len(o.Tickets))
	for _, t := range o.Tickets {
		tickets = append(tickets, ticketDTO{ID: t.ID, OrderItemID: t.OrderItemID, TicketCode: t.TicketCode, Status: t.Status, IssuedAt: t.IssuedAt.Format(time.RFC3339)})
	}
	return orderDTO{
		ID: o.ID, UserID: o.UserID, Status: o.Status, TotalAmount: o.TotalAmount,
		ExpiresAt: o.ExpiresAt.Format(time.RFC3339), CreatedAt: o.CreatedAt.Format(time.RFC3339),
		Items: items, Tickets: tickets,
	}
}
