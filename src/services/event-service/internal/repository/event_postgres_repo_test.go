package repository

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/suite"
	tcmongodb "github.com/testcontainers/testcontainers-go/modules/mongodb"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ticketflow/pkg/pagination"
	"ticketflow/services/event-service/internal/model"
)

// seedUsersStubTable creates the minimal part of the `users` table (owned by
// identity-service) that events.organizer_id's FK needs — event-service
// does not run identity-service's migrations here, same approach as
// booking-service/internal/repository/booking_repository_concurrency_test.go
// and event_service_test.go use for the same cross-service FK issue.
func seedUsersStubTable(connStr string) error {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto; CREATE TABLE IF NOT EXISTS users (id UUID PRIMARY KEY DEFAULT gen_random_uuid())`)
	return err
}

// Integration test per docs/01-architecture/backend-conventions.md's
// "Integration test (cần Postgres thật)" convention: one testcontainers
// Postgres for the whole package, schema applied via the real
// migrations/*.up.sql through golang-migrate. No reachable Docker daemon ->
// log a warning and os.Exit(0) in TestMain, gracefully skipping the whole
// package instead of failing CI/sandboxes that have no Docker.

var testDB *gorm.DB
var testMongoDB *mongo.Database

func migrationsDir() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("cannot determine caller to resolve migrations dir")
	}
	// this file: internal/repository/event_postgres_repo_test.go
	dir := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations")
	return filepath.Abs(dir)
}

func applyMigrations(connStr string) error {
	dir, err := migrationsDir()
	if err != nil {
		return err
	}
	mig, err := migrate.New("file://"+dir, connStr)
	if err != nil {
		return err
	}
	defer mig.Close()
	if err := mig.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("ticketflow_test"),
		tcpostgres.WithUsername("ticketflow"),
		tcpostgres.WithPassword("ticketflow"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Printf("WARN: no Docker daemon available, skipping event-service repository integration tests: %v", err)
		os.Exit(0)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("WARN: skipping event-service repository integration tests: cannot get postgres connection string: %v", err)
		os.Exit(0)
	}

	if err := seedUsersStubTable(connStr); err != nil {
		log.Printf("ERROR: failed to seed users stub table in test container: %v", err)
		_ = pgContainer.Terminate(ctx)
		os.Exit(1)
	}

	if err := applyMigrations(connStr); err != nil {
		log.Printf("ERROR: failed to apply event-service migrations to test container: %v", err)
		_ = pgContainer.Terminate(ctx)
		os.Exit(1)
	}

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Printf("ERROR: failed to open gorm connection to test container: %v", err)
		_ = pgContainer.Terminate(ctx)
		os.Exit(1)
	}
	testDB = db

	// EventCatalogMongoRepo needs its own container — event_catalog is a
	// MongoDB collection, not a Postgres table (see
	// docs/03-data/mongodb-schema.md#collection-event_catalog). Shared with
	// the Postgres container's lifetime for the whole package per
	// backend-conventions.md.
	mongoContainer, err := tcmongodb.Run(ctx, "mongo:7")
	if err != nil {
		log.Printf("WARN: no Docker daemon available for Mongo, skipping event-service Mongo repository integration tests: %v", err)
		_ = pgContainer.Terminate(ctx)
		os.Exit(0)
	}

	mongoURI, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("WARN: skipping event-service Mongo repository integration tests: cannot get mongo connection string: %v", err)
		_ = pgContainer.Terminate(ctx)
		_ = mongoContainer.Terminate(ctx)
		os.Exit(0)
	}
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Printf("ERROR: failed to connect to Mongo test container: %v", err)
		_ = pgContainer.Terminate(ctx)
		_ = mongoContainer.Terminate(ctx)
		os.Exit(1)
	}
	testMongoDB = mongoClient.Database("ticketflow_test")

	code := m.Run()

	_ = mongoClient.Disconnect(ctx)
	_ = mongoContainer.Terminate(ctx)
	_ = pgContainer.Terminate(ctx)
	os.Exit(code)
}

const (
	testCaseSuccess_CreateTxAndGetByID    = "[Success] create then get by id round-trip"
	testCaseError_GetByIDNotFound         = "[Error] get by id returns gorm.ErrRecordNotFound when missing"
	testCaseSuccess_GetPublishedBySlug    = "[Success] get published event by slug"
	testCaseError_GetPublishedBySlugDraft = "[Error] draft event not found by slug lookup (published-only)"
	testCaseSuccess_Update                = "[Success] update persists changed fields"
	testCaseSuccess_SlugExistsTrue        = "[Success] slug_exists true for an existing slug"
	testCaseSuccess_SlugExistsFalse       = "[Success] slug_exists false for an unused slug"
	testCaseSuccess_ListPublished         = "[Success] list published filters by category/city and computes min_price"
	testCaseSuccess_ListByOrganizer       = "[Success] list by organizer filters by status"
)

// seedOrganizerUser inserts a row into the users stub table (see
// seedUsersStubTable) and returns its id, so events.organizer_id's FK is
// satisfied by a real row rather than a dangling random UUID.
func seedOrganizerUser(db *gorm.DB) string {
	id := uuid.NewString()
	if err := db.Exec("INSERT INTO users (id) VALUES (?)", id).Error; err != nil {
		panic(err)
	}
	return id
}

func newTestEvent(organizerID, status string) *model.Event {
	return &model.Event{
		ID:          uuid.NewString(),
		OrganizerID: organizerID,
		Title:       "Test Event " + uuid.NewString(),
		Slug:        "test-event-" + uuid.NewString(),
		Category:    model.CategoryConcert,
		Status:      status,
		StartTime:   time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second),
		CreatedAt:   time.Now().UTC().Truncate(time.Second),
	}
}

type EventPostgresRepoTestSuite struct {
	suite.Suite
	repo *EventPostgresRepo
}

func TestEventPostgresRepoTestSuite(t *testing.T) {
	suite.Run(t, new(EventPostgresRepoTestSuite))
}

func (s *EventPostgresRepoTestSuite) SetupSuite() {
	s.repo = NewEventPostgresRepo(testDB)
}

// TearDownTest truncates between tests sharing the one package-level
// container/DB, per backend-conventions.md's guidance for suites sharing a
// DB across sub-tests.
func (s *EventPostgresRepoTestSuite) TearDownTest() {
	s.Require().NoError(testDB.Exec("TRUNCATE TABLE ticket_types, events CASCADE").Error)
}

func (s *EventPostgresRepoTestSuite) TestGetByID() {
	s.Run(testCaseSuccess_CreateTxAndGetByID, func() {
		ctx := context.Background()
		e := newTestEvent(seedOrganizerUser(testDB), model.StatusDraft)
		s.Require().NoError(s.repo.CreateTx(testDB, e))

		got, err := s.repo.GetByID(ctx, e.ID)

		s.Require().NoError(err)
		s.Equal(e.Title, got.Title)
		s.Equal(e.Slug, got.Slug)
		s.Equal(model.StatusDraft, got.Status)
	})

	s.Run(testCaseError_GetByIDNotFound, func() {
		_, err := s.repo.GetByID(context.Background(), uuid.NewString())

		s.Require().Error(err)
		s.True(errors.Is(err, gorm.ErrRecordNotFound))
	})
}

func (s *EventPostgresRepoTestSuite) TestGetPublishedBySlug() {
	s.Run(testCaseSuccess_GetPublishedBySlug, func() {
		ctx := context.Background()
		e := newTestEvent(seedOrganizerUser(testDB), model.StatusPublished)
		s.Require().NoError(s.repo.CreateTx(testDB, e))

		got, err := s.repo.GetPublishedBySlug(ctx, e.Slug)

		s.Require().NoError(err)
		s.Equal(e.ID, got.ID)
	})

	s.Run(testCaseError_GetPublishedBySlugDraft, func() {
		ctx := context.Background()
		e := newTestEvent(seedOrganizerUser(testDB), model.StatusDraft)
		s.Require().NoError(s.repo.CreateTx(testDB, e))

		_, err := s.repo.GetPublishedBySlug(ctx, e.Slug)

		s.Require().Error(err)
		s.True(errors.Is(err, gorm.ErrRecordNotFound))
	})
}

func (s *EventPostgresRepoTestSuite) TestUpdate() {
	s.Run(testCaseSuccess_Update, func() {
		ctx := context.Background()
		e := newTestEvent(seedOrganizerUser(testDB), model.StatusDraft)
		s.Require().NoError(s.repo.CreateTx(testDB, e))

		e.Title = "Updated Title"
		e.Status = model.StatusPublished
		s.Require().NoError(s.repo.Update(ctx, e))

		got, err := s.repo.GetByID(ctx, e.ID)
		s.Require().NoError(err)
		s.Equal("Updated Title", got.Title)
		s.Equal(model.StatusPublished, got.Status)
	})
}

func (s *EventPostgresRepoTestSuite) TestSlugExists() {
	s.Run(testCaseSuccess_SlugExistsTrue, func() {
		ctx := context.Background()
		e := newTestEvent(seedOrganizerUser(testDB), model.StatusDraft)
		s.Require().NoError(s.repo.CreateTx(testDB, e))

		exists, err := s.repo.SlugExists(ctx, e.Slug)

		s.Require().NoError(err)
		s.True(exists)
	})

	s.Run(testCaseSuccess_SlugExistsFalse, func() {
		exists, err := s.repo.SlugExists(context.Background(), "no-such-slug-"+uuid.NewString())

		s.Require().NoError(err)
		s.False(exists)
	})
}

func (s *EventPostgresRepoTestSuite) TestListPublished() {
	s.Run(testCaseSuccess_ListPublished, func() {
		ctx := context.Background()
		city := "Hanoi"

		match := newTestEvent(seedOrganizerUser(testDB), model.StatusPublished)
		match.Category = model.CategoryConcert
		match.City = &city
		s.Require().NoError(s.repo.CreateTx(testDB, match))

		tt := &model.TicketType{ID: uuid.NewString(), EventID: match.ID, Name: "GA", Price: 100, Currency: "VND", Quota: 10}
		s.Require().NoError(testDB.Create(tt).Error)

		otherCity := "HCMC"
		nonMatch := newTestEvent(seedOrganizerUser(testDB), model.StatusPublished)
		nonMatch.Category = model.CategoryConcert
		nonMatch.City = &otherCity
		s.Require().NoError(s.repo.CreateTx(testDB, nonMatch))

		draft := newTestEvent(seedOrganizerUser(testDB), model.StatusDraft)
		draft.Category = model.CategoryConcert
		draft.City = &city
		s.Require().NoError(s.repo.CreateTx(testDB, draft))

		rows, total, err := s.repo.ListPublished(ctx, EventFilter{Category: model.CategoryConcert, City: city}, pagination.Params{Page: 1, PageSize: 20})

		s.Require().NoError(err)
		s.Equal(int64(1), total)
		s.Require().Len(rows, 1)
		s.Equal(match.ID, rows[0].ID)
		s.Require().NotNil(rows[0].MinPrice)
		s.Equal(100.0, *rows[0].MinPrice)
	})
}

func (s *EventPostgresRepoTestSuite) TestListByOrganizer() {
	s.Run(testCaseSuccess_ListByOrganizer, func() {
		ctx := context.Background()
		organizerID := seedOrganizerUser(testDB)

		published := newTestEvent(organizerID, model.StatusPublished)
		s.Require().NoError(s.repo.CreateTx(testDB, published))
		draft := newTestEvent(organizerID, model.StatusDraft)
		s.Require().NoError(s.repo.CreateTx(testDB, draft))
		otherOrganizer := newTestEvent(seedOrganizerUser(testDB), model.StatusPublished)
		s.Require().NoError(s.repo.CreateTx(testDB, otherOrganizer))

		rows, total, err := s.repo.ListByOrganizer(ctx, organizerID, model.StatusDraft, pagination.Params{Page: 1, PageSize: 20})

		s.Require().NoError(err)
		s.Equal(int64(1), total)
		s.Require().Len(rows, 1)
		s.Equal(draft.ID, rows[0].ID)
	})
}
