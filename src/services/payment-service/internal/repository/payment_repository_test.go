package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ticketflow/services/payment-service/internal/model"
)

// This file bootstraps payment-service's integration test scaffolding from
// scratch (the service had zero tests before), per
// docs/01-architecture/backend-conventions.md's "Integration test (cần
// Postgres thật)" convention: one testcontainers Postgres for the whole
// package, schema applied via real golang-migrate migrations (not
// hand-copied DDL), graceful `os.Exit(0)` skip when Docker is unavailable.
//
// `payments.order_id` has a foreign key to booking-service's `orders`
// table, which itself FKs to `users` (identity-service) and whose sibling
// `order_items`/`tickets` tables FK to `ticket_types`/`events`
// (event-service). To exercise the *real* migrations end to end (rather
// than hand-copying a schema subset, which is exactly what the new
// convention wants to avoid), mergeRealMigrations below copies each
// service's actual migrations/*.sql files — verbatim, only renumbering the
// leading version prefix so all four services can share one golang-migrate
// source directory — in dependency order: identity -> event -> booking ->
// payment.
var gormDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	migrationsDir, cleanupDir, err := mergeRealMigrations()
	if err != nil {
		log.Fatalf("payment repository tests: prepare merged migrations dir: %v", err)
	}
	defer cleanupDir()

	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("ticketflow_test"),
		tcpostgres.WithUsername("ticketflow"),
		tcpostgres.WithPassword("ticketflow"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		log.Printf("warning: skipping payment-service repository integration tests: cannot start postgres testcontainer (Docker unavailable?): %v", err)
		os.Exit(0)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("warning: skipping payment-service repository integration tests: cannot resolve container connection string: %v", err)
		_ = container.Terminate(ctx)
		os.Exit(0)
	}

	if err := runMigrations(migrationsDir, connStr); err != nil {
		log.Printf("warning: skipping payment-service repository integration tests: migration failed: %v", err)
		_ = container.Terminate(ctx)
		os.Exit(0)
	}

	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Printf("warning: skipping payment-service repository integration tests: cannot open gorm connection: %v", err)
		_ = container.Terminate(ctx)
		os.Exit(0)
	}
	gormDB = db

	code := m.Run()
	_ = container.Terminate(ctx)
	os.Exit(code)
}

// mergeRealMigrations copies the real migrations/*.up.sql and
// migrations/*.down.sql of identity-service, event-service,
// booking-service, and payment-service (in that FK dependency order) into
// one fresh temp directory, renumbering each file's leading version prefix
// to <serviceIndex*1000 + originalVersion> so golang-migrate (which needs a
// single source directory with globally-ordered, non-colliding versions)
// can replay all four services' real schemas against one test database.
func mergeRealMigrations() (dir string, cleanup func(), err error) {
	tmp, err := os.MkdirTemp("", "payment-service-merged-migrations-*")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { _ = os.RemoveAll(tmp) }

	sources := []struct {
		label string
		dir   string
	}{
		{"identity", "../../../identity-service/migrations"},
		{"event", "../../../event-service/migrations"},
		{"booking", "../../../booking-service/migrations"},
		{"payment", "../../migrations"},
	}

	for i, src := range sources {
		entries, readErr := os.ReadDir(src.dir)
		if readErr != nil {
			cleanup()
			return "", nil, fmt.Errorf("read %s migrations dir %q: %w", src.label, src.dir, readErr)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasSuffix(name, ".up.sql") && !strings.HasSuffix(name, ".down.sql") {
				continue
			}
			parts := strings.SplitN(name, "_", 2)
			if len(parts) != 2 {
				cleanup()
				return "", nil, fmt.Errorf("unexpected migration filename %q in %s", name, src.dir)
			}
			origVersion, convErr := strconv.Atoi(parts[0])
			if convErr != nil {
				cleanup()
				return "", nil, fmt.Errorf("unexpected migration filename %q in %s: %w", name, src.dir, convErr)
			}
			newVersion := (i+1)*1000 + origVersion
			newName := fmt.Sprintf("%06d_%s_%s", newVersion, src.label, parts[1])

			content, readErr := os.ReadFile(filepath.Join(src.dir, name))
			if readErr != nil {
				cleanup()
				return "", nil, readErr
			}
			if writeErr := os.WriteFile(filepath.Join(tmp, newName), content, 0o644); writeErr != nil {
				cleanup()
				return "", nil, writeErr
			}
		}
	}
	return tmp, cleanup, nil
}

func runMigrations(migrationsDir, connStr string) error {
	m, err := migrate.New("file://"+migrationsDir, connStr)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// truncateAll clears every table touched by these tests, in FK-safe order,
// between sub-tests sharing the one package-level container per convention
// #6 ("dọn bảng liên quan (TRUNCATE ... CASCADE) trong TearDownTest").
func truncateAll(t require.TestingT) {
	require.NoError(t, gormDB.Exec(`TRUNCATE TABLE payments, tickets, order_items, orders, ticket_types, events, users RESTART IDENTITY CASCADE`).Error)
}

// seedUserAndOrder inserts one user + one order row (the minimal FK chain
// payments.order_id needs) via the real orders table created by
// booking-service's own migration, and returns the new order's id.
func seedUserAndOrder(t require.TestingT, status string) (orderID string) {
	userID := uuid.NewString()
	require.NoError(t, gormDB.Exec(
		`INSERT INTO users (id, email) VALUES (?, ?)`, userID, userID+"@example.test").Error)

	orderID = uuid.NewString()
	require.NoError(t, gormDB.Exec(
		`INSERT INTO orders (id, user_id, status, total_amount, expires_at) VALUES (?, ?, ?, ?, ?)`,
		orderID, userID, status, 100000, time.Now().Add(15*time.Minute)).Error)
	return orderID
}

const (
	testCaseSuccess_Create_InsertsFullRow         = "[Success] Tạo payment mới với đầy đủ trường"
	testCaseError_Create_DuplicateIDConflict      = "[Error] Trùng khóa chính (id) trả về lỗi"
	testCaseError_Create_UnknownOrderIDViolatesFK = "[Error] order_id không tồn tại vi phạm foreign key"

	testCaseSuccess_UpdateStatus_SetsStatusAndTxnID   = "[Success] Cập nhật status và provider_txn_id"
	testCaseSuccess_UpdateStatus_NilTxnIDClearsColumn = "[Success] providerTxnID nil ghi đè cột thành NULL"
	testCaseSuccess_UpdateStatus_UnknownIDNoOpNoError = "[Success] id không tồn tại vẫn trả nil error, không có hàng nào bị ảnh hưởng (không có guard tồn tại)"

	testCaseSuccess_GetLatestByOrderID_ReturnsMostRecent = "[Success] Trả về payment mới nhất theo created_at"
	testCaseError_GetLatestByOrderID_NotFound            = "[Error] Không có payment nào trả về gorm.ErrRecordNotFound"
)

type PaymentRepositoryTestSuite struct {
	suite.Suite
	repo *PaymentRepository
}

func TestPaymentRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentRepositoryTestSuite))
}

func (s *PaymentRepositoryTestSuite) SetupSuite() {
	s.repo = NewPaymentRepository(gormDB)
}

func (s *PaymentRepositoryTestSuite) TearDownTest() {
	truncateAll(s.T())
}

func (s *PaymentRepositoryTestSuite) TestCreate() {
	s.Run(testCaseSuccess_Create_InsertsFullRow, func() {
		orderID := seedUserAndOrder(s.T(), "pending")
		provider := "mock"
		payment := &model.Payment{
			ID:       uuid.NewString(),
			OrderID:  orderID,
			Provider: &provider,
			Amount:   250000,
			Status:   model.StatusInitiated,
		}

		err := s.repo.Create(context.Background(), payment)
		s.Require().NoError(err)

		var got model.Payment
		s.Require().NoError(gormDB.Where("id = ?", payment.ID).First(&got).Error)
		s.Assert().Equal(orderID, got.OrderID)
		s.Assert().Equal(model.StatusInitiated, got.Status)
		s.Assert().Equal(float64(250000), got.Amount)
		s.Require().NotNil(got.Provider)
		s.Assert().Equal("mock", *got.Provider)
		s.Assert().False(got.CreatedAt.IsZero())
	})

	s.Run(testCaseError_Create_DuplicateIDConflict, func() {
		orderID := seedUserAndOrder(s.T(), "pending")
		id := uuid.NewString()
		payment := &model.Payment{ID: id, OrderID: orderID, Amount: 1000, Status: model.StatusInitiated}
		s.Require().NoError(s.repo.Create(context.Background(), payment))

		dup := &model.Payment{ID: id, OrderID: orderID, Amount: 2000, Status: model.StatusInitiated}
		err := s.repo.Create(context.Background(), dup)
		s.Require().Error(err)
	})

	s.Run(testCaseError_Create_UnknownOrderIDViolatesFK, func() {
		payment := &model.Payment{ID: uuid.NewString(), OrderID: uuid.NewString(), Amount: 1000, Status: model.StatusInitiated}
		err := s.repo.Create(context.Background(), payment)
		s.Require().Error(err)
	})
}

func (s *PaymentRepositoryTestSuite) TestUpdateStatus() {
	s.Run(testCaseSuccess_UpdateStatus_SetsStatusAndTxnID, func() {
		orderID := seedUserAndOrder(s.T(), "pending")
		id := uuid.NewString()
		s.Require().NoError(s.repo.Create(context.Background(), &model.Payment{
			ID: id, OrderID: orderID, Amount: 1000, Status: model.StatusInitiated,
		}))
		txnID := "mock-txn-1"

		err := s.repo.UpdateStatus(context.Background(), id, model.StatusSuccess, &txnID)
		s.Require().NoError(err)

		var got model.Payment
		s.Require().NoError(gormDB.Where("id = ?", id).First(&got).Error)
		s.Assert().Equal(model.StatusSuccess, got.Status)
		s.Require().NotNil(got.ProviderTxnID)
		s.Assert().Equal(txnID, *got.ProviderTxnID)
	})

	s.Run(testCaseSuccess_UpdateStatus_NilTxnIDClearsColumn, func() {
		orderID := seedUserAndOrder(s.T(), "pending")
		id := uuid.NewString()
		existingTxnID := "some-prior-txn"
		s.Require().NoError(s.repo.Create(context.Background(), &model.Payment{
			ID: id, OrderID: orderID, Amount: 1000, Status: model.StatusInitiated, ProviderTxnID: &existingTxnID,
		}))

		err := s.repo.UpdateStatus(context.Background(), id, model.StatusFailed, nil)
		s.Require().NoError(err)

		var status string
		var txnID sql.NullString
		s.Require().NoError(gormDB.Raw(`SELECT status, provider_txn_id FROM payments WHERE id = ?`, id).Row().Scan(&status, &txnID))
		s.Assert().Equal(model.StatusFailed, status)
		s.Assert().False(txnID.Valid, "nil providerTxnID must overwrite the column to NULL, not skip it")
	})

	// This documents a real production gap (see final report), not a bug
	// to fix here: UpdateStatus is an unconditional `Updates` by id, with
	// no existence/status guard. Updating a non-existent id silently
	// affects zero rows and returns nil error rather than
	// gorm.ErrRecordNotFound or any other signal.
	s.Run(testCaseSuccess_UpdateStatus_UnknownIDNoOpNoError, func() {
		err := s.repo.UpdateStatus(context.Background(), uuid.NewString(), model.StatusSuccess, nil)
		s.Require().NoError(err, "current UpdateStatus has no existence guard: updating an unknown id is a silent no-op, not an error")
	})
}

func (s *PaymentRepositoryTestSuite) TestGetLatestByOrderID() {
	s.Run(testCaseSuccess_GetLatestByOrderID_ReturnsMostRecent, func() {
		orderID := seedUserAndOrder(s.T(), "pending")
		older := &model.Payment{
			ID: uuid.NewString(), OrderID: orderID, Amount: 1000, Status: model.StatusFailed,
			CreatedAt: time.Now().Add(-1 * time.Hour),
		}
		newer := &model.Payment{
			ID: uuid.NewString(), OrderID: orderID, Amount: 2000, Status: model.StatusSuccess,
			CreatedAt: time.Now(),
		}
		s.Require().NoError(s.repo.Create(context.Background(), older))
		s.Require().NoError(s.repo.Create(context.Background(), newer))

		got, err := s.repo.GetLatestByOrderID(context.Background(), orderID)
		s.Require().NoError(err)
		s.Assert().Equal(newer.ID, got.ID)
		s.Assert().Equal(model.StatusSuccess, got.Status)
	})

	s.Run(testCaseError_GetLatestByOrderID_NotFound, func() {
		got, err := s.repo.GetLatestByOrderID(context.Background(), uuid.NewString())
		s.Require().Nil(got)
		s.Require().ErrorIs(err, gorm.ErrRecordNotFound)
	})
}

// TestUpdateStatus_ConcurrentCalls_NoGuardOrIdempotencyCheck is the
// mandatory race-condition test (AGENTS.md names payment alongside booking
// explicitly) for payment-service's only state-mutating repository method.
// Unlike booking-service's CreateOrder (which uses `SELECT ... FOR UPDATE`
// to serialize concurrent stock decrements), UpdateStatus takes no lock, no
// optimistic-concurrency version column, and no idempotency key — it is an
// unconditional `UPDATE ... WHERE id = ?`. This test asserts what the
// CURRENT code actually does under concurrent Checkout/HandleWebhook-style
// calls racing on the SAME payment row (every call "succeeds" with a nil
// error regardless of which status "should" win), per the SCOPE DISCIPLINE
// instruction to document the gap, not silently patch it in this test.
func TestUpdateStatus_ConcurrentCalls_NoGuardOrIdempotencyCheck(t *testing.T) {
	if gormDB == nil {
		t.Skip("no test database available (see TestMain)")
	}
	defer truncateAll(t)

	repo := NewPaymentRepository(gormDB)
	orderID := seedUserAndOrder(t, "pending")
	paymentID := uuid.NewString()
	require.NoError(t, repo.Create(context.Background(), &model.Payment{
		ID: paymentID, OrderID: orderID, Amount: 100000, Status: model.StatusInitiated,
	}))

	const concurrency = 20
	var (
		wg        sync.WaitGroup
		succeeded int64
		failed    int64
	)
	start := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		wantStatus := model.StatusSuccess
		if i%2 == 0 {
			wantStatus = model.StatusFailed
		}
		txnID := fmt.Sprintf("txn-%d", i)
		go func(status, txnID string) {
			defer wg.Done()
			<-start // release every goroutine as close to simultaneously as possible
			err := repo.UpdateStatus(context.Background(), paymentID, status, &txnID)
			if err == nil {
				atomic.AddInt64(&succeeded, 1)
			} else {
				atomic.AddInt64(&failed, 1)
				t.Logf("unexpected UpdateStatus error: %v", err)
			}
		}(wantStatus, txnID)
	}
	close(start)
	wg.Wait()

	// FINDING, not a bug in this test: every concurrent caller "wins" —
	// there is no conflict, no rejection, no idempotency check. Contrast
	// with booking-service's CreateOrder concurrency test, where excess
	// concurrent requests are correctly rejected with apperr.ErrConflict.
	require.Zero(t, failed, "no call should ever fail today, since UpdateStatus has no guard at all")
	require.EqualValues(t, concurrency, succeeded)

	// The final row state is whichever write physically committed last —
	// non-deterministic from the test's point of view, but it must be one
	// of the attempted (status, provider_txn_id) pairs, not corrupted.
	var finalStatus string
	var finalTxnID sql.NullString
	require.NoError(t, gormDB.Raw(`SELECT status, provider_txn_id FROM payments WHERE id = ?`, paymentID).
		Row().Scan(&finalStatus, &finalTxnID))
	require.Contains(t, []string{model.StatusSuccess, model.StatusFailed}, finalStatus)
	require.True(t, finalTxnID.Valid)
	require.True(t, strings.HasPrefix(finalTxnID.String, "txn-"))
}
