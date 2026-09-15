package http

import (
	"time"

	"ticketflow/services/event-service/internal/model"
	"ticketflow/services/event-service/internal/repository"
	"ticketflow/services/event-service/internal/service"
)

// DTOs mirror api-docs/openapi/event-service.yaml's TicketType/EventSummary/EventDetail schemas.

type ticketTypeDTO struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	Quota     int     `json:"quota"`
	SoldCount int     `json:"sold_count"`
}

func toTicketTypeDTO(tt model.TicketType) ticketTypeDTO {
	return ticketTypeDTO{ID: tt.ID, Name: tt.Name, Price: tt.Price, Currency: tt.Currency, Quota: tt.Quota, SoldCount: tt.SoldCount}
}

type eventSummaryDTO struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Slug      string   `json:"slug"`
	Category  string   `json:"category"`
	City      *string  `json:"city"`
	StartTime string   `json:"start_time"`
	BannerURL *string  `json:"banner_url"`
	MinPrice  *float64 `json:"min_price"`
}

func toEventSummaryDTO(r repository.EventSummaryRow) eventSummaryDTO {
	return eventSummaryDTO{
		ID: r.ID, Title: r.Title, Slug: r.Slug, Category: r.Category, City: r.City,
		StartTime: r.StartTime.Format(time.RFC3339), BannerURL: r.BannerURL, MinPrice: r.MinPrice,
	}
}

type eventDetailDTO struct {
	eventSummaryDTO
	OrganizerID string          `json:"organizer_id"`
	VenueName   *string         `json:"venue_name"`
	Address     *string         `json:"address"`
	EndTime     *string         `json:"end_time"`
	Description *string         `json:"description"`
	Status      string          `json:"status"`
	TicketTypes []ticketTypeDTO `json:"ticket_types"`
	Attributes  map[string]any  `json:"attributes"`
	Tags        []string        `json:"tags"`
}

func toEventDetailDTO(d *service.EventDetail) eventDetailDTO {
	var minPrice *float64
	for _, tt := range d.TicketTypes {
		p := tt.Price
		if minPrice == nil || p < *minPrice {
			minPrice = &p
		}
	}
	var endTime *string
	if d.Event.EndTime != nil {
		s := d.Event.EndTime.Format(time.RFC3339)
		endTime = &s
	}
	ticketTypes := make([]ticketTypeDTO, 0, len(d.TicketTypes))
	for _, tt := range d.TicketTypes {
		ticketTypes = append(ticketTypes, toTicketTypeDTO(tt))
	}
	var attributes map[string]any
	var tags []string
	if d.Catalog != nil {
		attributes = d.Catalog.Attributes
		tags = d.Catalog.Tags
	}
	return eventDetailDTO{
		eventSummaryDTO: eventSummaryDTO{
			ID: d.Event.ID, Title: d.Event.Title, Slug: d.Event.Slug, Category: d.Event.Category,
			City: d.Event.City, StartTime: d.Event.StartTime.Format(time.RFC3339), BannerURL: d.Event.BannerURL,
			MinPrice: minPrice,
		},
		OrganizerID: d.Event.OrganizerID,
		VenueName:   d.Event.VenueName,
		Address:     d.Event.Address,
		EndTime:     endTime,
		Description: d.Event.Description,
		Status:      d.Event.Status,
		TicketTypes: ticketTypes,
		Attributes:  attributes,
		Tags:        tags,
	}
}
