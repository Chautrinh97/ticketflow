package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"ticketflow/services/booking-service/internal/model"
)

const (
	testCaseSuccess_ToOrderDTO_FullOrder  = "[Success] Map đầy đủ order có items và tickets"
	testCaseSuccess_ToOrderDTO_NilSlices  = "[Success] Items/Tickets nil -> mảng rỗng [], không phải null"
	testCaseSuccess_ToOrderDTO_TimeFormat = "[Success] Thời gian format đúng RFC3339"
)

func TestToOrderDTO(t *testing.T) {
	t.Run(testCaseSuccess_ToOrderDTO_FullOrder, func(t *testing.T) {
		expiresAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		issuedAt := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)
		order := &model.Order{
			ID: "order-1", UserID: "user-1", Status: model.OrderStatusPaid,
			TotalAmount: 250000, ExpiresAt: expiresAt, CreatedAt: createdAt,
			Items: []model.OrderItem{
				{ID: "item-1", OrderID: "order-1", TicketTypeID: "tt-1", Quantity: 2, UnitPrice: 100000},
			},
			Tickets: []model.Ticket{
				{ID: "ticket-1", OrderItemID: "item-1", TicketCode: "ABC123", Status: model.TicketStatusValid, IssuedAt: issuedAt},
			},
		}

		got := toOrderDTO(order)

		assert.Equal(t, "order-1", got.ID)
		assert.Equal(t, "user-1", got.UserID)
		assert.Equal(t, model.OrderStatusPaid, got.Status)
		assert.Equal(t, 250000.0, got.TotalAmount)
		assert.Equal(t, expiresAt.Format(time.RFC3339), got.ExpiresAt)
		assert.Equal(t, createdAt.Format(time.RFC3339), got.CreatedAt)

		if assert.Len(t, got.Items, 1) {
			assert.Equal(t, orderItemDTO{ID: "item-1", TicketTypeID: "tt-1", Quantity: 2, UnitPrice: 100000}, got.Items[0])
		}
		if assert.Len(t, got.Tickets, 1) {
			assert.Equal(t, ticketDTO{
				ID: "ticket-1", OrderItemID: "item-1", TicketCode: "ABC123",
				Status: model.TicketStatusValid, IssuedAt: issuedAt.Format(time.RFC3339),
			}, got.Tickets[0])
		}
	})

	t.Run(testCaseSuccess_ToOrderDTO_NilSlices, func(t *testing.T) {
		order := &model.Order{ID: "order-2", Items: nil, Tickets: nil}

		got := toOrderDTO(order)

		assert.NotNil(t, got.Items, "Items must serialize as [] not null")
		assert.NotNil(t, got.Tickets, "Tickets must serialize as [] not null")
		assert.Empty(t, got.Items)
		assert.Empty(t, got.Tickets)
	})

	t.Run(testCaseSuccess_ToOrderDTO_TimeFormat, func(t *testing.T) {
		loc := time.FixedZone("+07:00", 7*60*60)
		at := time.Date(2026, 6, 15, 12, 30, 0, 0, loc)
		order := &model.Order{ID: "order-3", ExpiresAt: at, CreatedAt: at}

		got := toOrderDTO(order)

		assert.Equal(t, "2026-06-15T12:30:00+07:00", got.ExpiresAt)
		assert.Equal(t, "2026-06-15T12:30:00+07:00", got.CreatedAt)
	})
}
