package model

// EventCatalog mirrors docs/03-data/mongodb-schema.md#collection-event_catalog.
// Phase 1 does not validate `attributes` shape per category (explicitly
// deferred to Phase 2 by docs/07-roadmap/phase-1-mvp.md) — it is stored as
// a free-form map.
type EventCatalog struct {
	EventID       string         `bson:"event_id" json:"event_id"`
	Category      string         `bson:"category" json:"category"`
	Attributes    map[string]any `bson:"attributes" json:"attributes"`
	GalleryImages []string       `bson:"gallery_images" json:"gallery_images"`
	FAQ           []FAQItem      `bson:"faq" json:"faq"`
	Tags          []string       `bson:"tags" json:"tags"`
}

type FAQItem struct {
	Question string `bson:"question" json:"question"`
	Answer   string `bson:"answer" json:"answer"`
}
