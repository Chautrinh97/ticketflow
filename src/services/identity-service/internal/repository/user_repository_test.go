package repository

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ticketflow/services/identity-service/internal/model"
)

// Integration test per docs/01-architecture/backend-conventions.md's
// "Integration test (cần Postgres thật)" convention: one testcontainers
// Postgres for the whole package, schema applied via the real
// migrations/*.up.sql through golang-migrate — no hand-copied CREATE TABLE.
// No reachable Docker daemon -> log a warning and os.Exit(0) in TestMain,
// gracefully skipping the whole package instead of failing CI/sandboxes
// that have no Docker.

var testDB *gorm.DB

// migrationsDir resolves the service's migrations/ directory relative to
// this test file's own location, so it works regardless of the working
// directory `go test` is invoked from.
func migrationsDir() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("cannot determine caller to resolve migrations dir")
	}
	// this file: internal/repository/user_repository_test.go
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
		log.Printf("WARN: no Docker daemon available, skipping identity-service repository integration tests: %v", err)
		os.Exit(0)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("WARN: skipping identity-service repository integration tests: cannot get postgres connection string: %v", err)
		os.Exit(0)
	}

	if err := applyMigrations(connStr); err != nil {
		log.Printf("ERROR: failed to apply identity-service migrations to test container: %v", err)
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

	code := m.Run()

	_ = pgContainer.Terminate(ctx)
	os.Exit(code)
}

const (
	testCaseSuccess_CreateAndGetByID       = "[Success] create then get by id round-trip"
	testCaseError_GetByIDNotFound          = "[Error] get by id returns gorm.ErrRecordNotFound when missing"
	testCaseSuccess_GetByFirebaseUID       = "[Success] get by firebase_uid finds the linked user"
	testCaseError_GetByFirebaseUIDNotFound = "[Error] get by firebase_uid returns gorm.ErrRecordNotFound when missing"
	testCaseSuccess_GetByEmail             = "[Success] get by email finds the user"
	testCaseSuccess_LinkFirebaseUID        = "[Success] link firebase_uid updates the existing user"
	testCaseSuccess_Update                 = "[Success] update persists changed fields"
)

func newTestUser() *model.User {
	return &model.User{
		ID:     uuid.NewString(),
		Email:  uuid.NewString() + "@example.test",
		Role:   model.RoleUser,
		Status: model.StatusActive,
	}
}

type UserRepositoryTestSuite struct {
	suite.Suite
	repo *UserRepository
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	s.repo = NewUserRepository(testDB)
}

// TearDownTest truncates between tests sharing the one package-level
// container/DB, per backend-conventions.md's guidance for suites sharing a
// DB across sub-tests.
func (s *UserRepositoryTestSuite) TearDownTest() {
	s.Require().NoError(testDB.Exec("TRUNCATE TABLE users CASCADE").Error)
}

func (s *UserRepositoryTestSuite) TestGetByID() {
	s.Run(testCaseSuccess_CreateAndGetByID, func() {
		ctx := context.Background()
		u := newTestUser()
		s.Require().NoError(s.repo.Create(ctx, u))

		got, err := s.repo.GetByID(ctx, u.ID)

		s.Require().NoError(err)
		s.Equal(u.Email, got.Email)
		s.Equal(model.RoleUser, got.Role)
		s.Equal(model.StatusActive, got.Status)
	})

	s.Run(testCaseError_GetByIDNotFound, func() {
		_, err := s.repo.GetByID(context.Background(), uuid.NewString())

		s.Require().Error(err)
		s.True(errors.Is(err, gorm.ErrRecordNotFound))
	})
}

func (s *UserRepositoryTestSuite) TestGetByFirebaseUID() {
	s.Run(testCaseSuccess_GetByFirebaseUID, func() {
		ctx := context.Background()
		fbUID := uuid.NewString()
		u := newTestUser()
		u.FirebaseUID = &fbUID
		s.Require().NoError(s.repo.Create(ctx, u))

		got, err := s.repo.GetByFirebaseUID(ctx, fbUID)

		s.Require().NoError(err)
		s.Equal(u.ID, got.ID)
	})

	s.Run(testCaseError_GetByFirebaseUIDNotFound, func() {
		_, err := s.repo.GetByFirebaseUID(context.Background(), uuid.NewString())

		s.Require().Error(err)
		s.True(errors.Is(err, gorm.ErrRecordNotFound))
	})
}

func (s *UserRepositoryTestSuite) TestGetByEmail() {
	s.Run(testCaseSuccess_GetByEmail, func() {
		ctx := context.Background()
		u := newTestUser()
		s.Require().NoError(s.repo.Create(ctx, u))

		got, err := s.repo.GetByEmail(ctx, u.Email)

		s.Require().NoError(err)
		s.Equal(u.ID, got.ID)
	})
}

func (s *UserRepositoryTestSuite) TestLinkFirebaseUID() {
	s.Run(testCaseSuccess_LinkFirebaseUID, func() {
		ctx := context.Background()
		u := newTestUser()
		s.Require().NoError(s.repo.Create(ctx, u))

		fbUID := uuid.NewString()
		s.Require().NoError(s.repo.LinkFirebaseUID(ctx, u.ID, fbUID))

		got, err := s.repo.GetByFirebaseUID(ctx, fbUID)
		s.Require().NoError(err)
		s.Equal(u.ID, got.ID)
	})
}

func (s *UserRepositoryTestSuite) TestUpdate() {
	s.Run(testCaseSuccess_Update, func() {
		ctx := context.Background()
		u := newTestUser()
		oldName := "Old Name"
		u.FullName = &oldName
		s.Require().NoError(s.repo.Create(ctx, u))

		newName := "New Name"
		u.FullName = &newName
		s.Require().NoError(s.repo.Update(ctx, u))

		got, err := s.repo.GetByID(ctx, u.ID)
		s.Require().NoError(err)
		s.Require().NotNil(got.FullName)
		s.Equal(newName, *got.FullName)
	})
}
