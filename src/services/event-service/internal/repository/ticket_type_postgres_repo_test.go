package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"ticketflow/services/event-service/internal/model"
)

const (
	testCaseSuccess_TicketTypeCreate            = "[Success] create a ticket type for an event"
	testCaseSuccess_ListByEventIDOrderedByPrice = "[Success] list by event id orders by price ascending"
	testCaseSuccess_ListByEventIDEmpty          = "[Success] list by event id returns empty slice for an event with none"
)

// TicketTypePostgresRepoTestSuite shares the package-level testDB/TestMain
// defined in event_postgres_repo_test.go (one Postgres container per
// package, per backend-conventions.md's testing convention).
type TicketTypePostgresRepoTestSuite struct {
	suite.Suite
	repo      *TicketTypePostgresRepo
	eventRepo *EventPostgresRepo
}

func TestTicketTypePostgresRepoTestSuite(t *testing.T) {
	suite.Run(t, new(TicketTypePostgresRepoTestSuite))
}

func (s *TicketTypePostgresRepoTestSuite) SetupSuite() {
	s.repo = NewTicketTypePostgresRepo(testDB)
	s.eventRepo = NewEventPostgresRepo(testDB)
}

func (s *TicketTypePostgresRepoTestSuite) TearDownTest() {
	s.Require().NoError(testDB.Exec("TRUNCATE TABLE ticket_types, events CASCADE").Error)
}

func (s *TicketTypePostgresRepoTestSuite) seedEvent() *model.Event {
	e := newTestEvent(seedOrganizerUser(testDB), model.StatusPublished)
	s.Require().NoError(s.eventRepo.CreateTx(testDB, e))
	return e
}

func (s *TicketTypePostgresRepoTestSuite) TestCreate() {
	s.Run(testCaseSuccess_TicketTypeCreate, func() {
		ctx := context.Background()
		e := s.seedEvent()
		tt := &model.TicketType{ID: uuid.NewString(), EventID: e.ID, Name: "VIP", Price: 500, Currency: "VND", Quota: 20}

		err := s.repo.Create(ctx, tt)

		s.Require().NoError(err)
		rows, err := s.repo.ListByEventID(ctx, e.ID)
		s.Require().NoError(err)
		s.Require().Len(rows, 1)
		s.Equal("VIP", rows[0].Name)
	})
}

func (s *TicketTypePostgresRepoTestSuite) TestListByEventID() {
	s.Run(testCaseSuccess_ListByEventIDOrderedByPrice, func() {
		ctx := context.Background()
		e := s.seedEvent()
		s.Require().NoError(s.repo.Create(ctx, &model.TicketType{ID: uuid.NewString(), EventID: e.ID, Name: "VIP", Price: 500, Currency: "VND", Quota: 10}))
		s.Require().NoError(s.repo.Create(ctx, &model.TicketType{ID: uuid.NewString(), EventID: e.ID, Name: "GA", Price: 100, Currency: "VND", Quota: 50}))

		rows, err := s.repo.ListByEventID(ctx, e.ID)

		s.Require().NoError(err)
		s.Require().Len(rows, 2)
		s.Equal("GA", rows[0].Name)
		s.Equal("VIP", rows[1].Name)
	})

	s.Run(testCaseSuccess_ListByEventIDEmpty, func() {
		e := s.seedEvent()

		rows, err := s.repo.ListByEventID(context.Background(), e.ID)

		s.Require().NoError(err)
		s.Empty(rows)
	})
}
