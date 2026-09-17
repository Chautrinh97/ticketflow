package http

import (
	"testing"
	"time"

	"ticketflow/services/event-service/internal/model"
	"ticketflow/services/event-service/internal/repository"
	"ticketflow/services/event-service/internal/service"
)

func strPtr(s string) *string { return &s }

// toEventDetailDTO is a pure mapper (no DB/mock deps): computes min_price by
// scanning TicketTypes and formats an optional EndTime — high-value,
// zero-dependency coverage per docs/01-architecture/backend-conventions.md's
// "Test logic thuần" rule.
func TestToEventDetailDTO(t *testing.T) {
	baseEvent := &model.Event{
		ID:          "event-1",
		OrganizerID: "org-1",
		Title:       "Rock Concert",
		Slug:        "rock-concert",
		Category:    model.CategoryConcert,
		City:        strPtr("Hanoi"),
		StartTime:   time.Date(2026, 1, 1, 19, 0, 0, 0, time.UTC),
		Status:      model.StatusPublished,
	}

	t.Run("[Success] min_price là giá thấp nhất trong danh sách ticket_types", func(t *testing.T) {
		detail := &service.EventDetail{
			Event: baseEvent,
			TicketTypes: []model.TicketType{
				{ID: "tt-1", Name: "VIP", Price: 500000},
				{ID: "tt-2", Name: "Standard", Price: 200000},
				{ID: "tt-3", Name: "Early Bird", Price: 350000},
			},
			Catalog: &model.EventCatalog{EventID: "event-1", Category: model.CategoryConcert},
		}

		got := toEventDetailDTO(detail)

		if got.MinPrice == nil || *got.MinPrice != 200000 {
			t.Fatalf("MinPrice = %v, want 200000", got.MinPrice)
		}
		if len(got.TicketTypes) != 3 {
			t.Fatalf("len(TicketTypes) = %d, want 3", len(got.TicketTypes))
		}
	})

	t.Run("[Success] Không có ticket_types nào -> min_price là nil", func(t *testing.T) {
		detail := &service.EventDetail{Event: baseEvent, TicketTypes: nil, Catalog: nil}

		got := toEventDetailDTO(detail)

		if got.MinPrice != nil {
			t.Fatalf("MinPrice = %v, want nil", *got.MinPrice)
		}
		if got.TicketTypes == nil || len(got.TicketTypes) != 0 {
			t.Fatalf("TicketTypes = %v, want an empty (non-nil) slice", got.TicketTypes)
		}
	})

	t.Run("[Success] EndTime có giá trị -> format RFC3339", func(t *testing.T) {
		end := time.Date(2026, 1, 1, 22, 0, 0, 0, time.UTC)
		eventWithEnd := *baseEvent
		eventWithEnd.EndTime = &end
		detail := &service.EventDetail{Event: &eventWithEnd}

		got := toEventDetailDTO(detail)

		if got.EndTime == nil || *got.EndTime != end.Format(time.RFC3339) {
			t.Fatalf("EndTime = %v, want %q", got.EndTime, end.Format(time.RFC3339))
		}
	})

	t.Run("[Success] EndTime là nil -> DTO EndTime cũng là nil", func(t *testing.T) {
		detail := &service.EventDetail{Event: baseEvent}

		got := toEventDetailDTO(detail)

		if got.EndTime != nil {
			t.Fatalf("EndTime = %v, want nil", *got.EndTime)
		}
	})

	t.Run("[Success] Catalog nil -> Attributes/Tags là zero-value", func(t *testing.T) {
		detail := &service.EventDetail{Event: baseEvent, Catalog: nil}

		got := toEventDetailDTO(detail)

		if got.Attributes != nil {
			t.Fatalf("Attributes = %v, want nil", got.Attributes)
		}
		if got.Tags != nil {
			t.Fatalf("Tags = %v, want nil", got.Tags)
		}
	})

	t.Run("[Success] Catalog có giá trị -> Attributes/Tags được copy nguyên vẹn", func(t *testing.T) {
		detail := &service.EventDetail{
			Event: baseEvent,
			Catalog: &model.EventCatalog{
				EventID:    "event-1",
				Category:   model.CategoryConcert,
				Attributes: map[string]any{"artists": []string{"Band A"}},
				Tags:       []string{"rock", "outdoor"},
			},
		}

		got := toEventDetailDTO(detail)

		if got.Attributes["artists"] == nil {
			t.Fatalf("Attributes = %v, want it to carry through the catalog's map", got.Attributes)
		}
		if len(got.Tags) != 2 || got.Tags[0] != "rock" || got.Tags[1] != "outdoor" {
			t.Fatalf("Tags = %v, want [rock outdoor]", got.Tags)
		}
	})
}

func TestToEventSummaryDTO(t *testing.T) {
	t.Run("[Success] Copy đúng field và format start_time theo RFC3339", func(t *testing.T) {
		price := 150000.0
		row := repository.EventSummaryRow{
			ID: "e-1", Title: "Jazz Night", Slug: "jazz-night", Category: model.CategoryConcert,
			City: strPtr("HCMC"), StartTime: time.Date(2026, 3, 1, 20, 0, 0, 0, time.UTC),
			BannerURL: strPtr("https://example.test/banner.png"), MinPrice: &price,
		}

		got := toEventSummaryDTO(row)

		if got.ID != row.ID || got.Title != row.Title || got.Slug != row.Slug || got.Category != row.Category {
			t.Fatalf("toEventSummaryDTO() = %+v, mismatched identity fields against row %+v", got, row)
		}
		if got.StartTime != row.StartTime.Format(time.RFC3339) {
			t.Fatalf("StartTime = %q, want %q", got.StartTime, row.StartTime.Format(time.RFC3339))
		}
		if got.MinPrice == nil || *got.MinPrice != price {
			t.Fatalf("MinPrice = %v, want %v", got.MinPrice, price)
		}
	})
}

func TestToTicketTypeDTO(t *testing.T) {
	t.Run("[Success] Copy đúng toàn bộ field", func(t *testing.T) {
		tt := model.TicketType{ID: "tt-1", Name: "VIP", Price: 500000, Currency: "VND", Quota: 100, SoldCount: 42}

		got := toTicketTypeDTO(tt)

		if got != (ticketTypeDTO{ID: "tt-1", Name: "VIP", Price: 500000, Currency: "VND", Quota: 100, SoldCount: 42}) {
			t.Fatalf("toTicketTypeDTO() = %+v, want a 1:1 field copy of %+v", got, tt)
		}
	})
}
