package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"

	"ticketflow/services/event-service/internal/model"
)

const (
	testCaseSuccess_CatalogUpsertThenGet    = "[Success] upsert then get by event id round-trip"
	testCaseSuccess_CatalogUpsertOverwrites = "[Success] upsert on an existing event id overwrites the document"
	testCaseError_CatalogGetNotFound        = "[Error] get by event id returns mongo.ErrNoDocuments when missing"
)

// EventCatalogMongoRepoTestSuite shares the package-level testMongoDB set up
// in event_postgres_repo_test.go's TestMain (one Mongo container per
// package, per backend-conventions.md's testing convention — event_catalog
// lives in MongoDB, not Postgres, see docs/03-data/mongodb-schema.md).
type EventCatalogMongoRepoTestSuite struct {
	suite.Suite
	repo *EventCatalogMongoRepo
}

func TestEventCatalogMongoRepoTestSuite(t *testing.T) {
	suite.Run(t, new(EventCatalogMongoRepoTestSuite))
}

func (s *EventCatalogMongoRepoTestSuite) SetupSuite() {
	s.repo = NewEventCatalogMongoRepo(testMongoDB)
}

func (s *EventCatalogMongoRepoTestSuite) TearDownTest() {
	_, err := testMongoDB.Collection("event_catalog").DeleteMany(context.Background(), map[string]any{})
	s.Require().NoError(err)
}

func newTestCatalog(eventID string) *model.EventCatalog {
	return &model.EventCatalog{
		EventID:       eventID,
		Category:      model.CategoryConcert,
		Attributes:    map[string]any{"stage": "main"},
		GalleryImages: []string{"https://example.test/1.jpg"},
		FAQ:           []model.FAQItem{{Question: "Q1", Answer: "A1"}},
		Tags:          []string{"live", "music"},
	}
}

func (s *EventCatalogMongoRepoTestSuite) TestUpsertAndGetByEventID() {
	s.Run(testCaseSuccess_CatalogUpsertThenGet, func() {
		ctx := context.Background()
		doc := newTestCatalog(uuid.NewString())

		s.Require().NoError(s.repo.Upsert(ctx, doc))

		got, err := s.repo.GetByEventID(ctx, doc.EventID)
		s.Require().NoError(err)
		s.Equal(doc.Category, got.Category)
		s.Equal(doc.Tags, got.Tags)
		s.Require().Len(got.FAQ, 1)
		s.Equal("Q1", got.FAQ[0].Question)
	})

	s.Run(testCaseSuccess_CatalogUpsertOverwrites, func() {
		ctx := context.Background()
		eventID := uuid.NewString()
		doc := newTestCatalog(eventID)
		s.Require().NoError(s.repo.Upsert(ctx, doc))

		updated := newTestCatalog(eventID)
		updated.Tags = []string{"updated"}
		s.Require().NoError(s.repo.Upsert(ctx, updated))

		got, err := s.repo.GetByEventID(ctx, eventID)
		s.Require().NoError(err)
		s.Equal([]string{"updated"}, got.Tags)
	})

	s.Run(testCaseError_CatalogGetNotFound, func() {
		_, err := s.repo.GetByEventID(context.Background(), uuid.NewString())

		s.Require().Error(err)
		s.True(errors.Is(err, mongo.ErrNoDocuments))
	})
}

func (s *EventCatalogMongoRepoTestSuite) TestEnsureIndexes() {
	s.Require().NoError(s.repo.EnsureIndexes(context.Background()))
}
