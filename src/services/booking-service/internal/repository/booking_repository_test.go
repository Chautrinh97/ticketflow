package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"ticketflow/pkg/apperr"
	"ticketflow/services/booking-service/internal/model"
)

// This file is the NEW-convention integration test for this package (see
// backend-conventions.md "Integration test (cần Postgres thật)") — separate
// from the pre-existing booking_repository_concurrency_test.go, which is
// left untouched per that same doc's explicit grandfathering note. It adds
// its own package-level TestMain, so — as a consequence of Go allowing only
// one TestMain per package — the whole package (including the pre-existing
// concurrency tests) is now gated behind this TestMain's testcontainers
// bring-up. When Docker isn't reachable, TestMain os.Exit(0)s before
// m.Run() is ever called, which also means the pre-existing file's own
// per-test TEST_DATABASE_URL skip logic simply never gets a chance to run
// in that case — this matches the convention doc's own wording that a
// failed container bring-up should "skip toàn bộ package" (skip the WHOLE
// package), not just this file's tests.

const migrationsSourceURL = "file://../../migrations"

// itPool is the shared pgxpool.Pool for every test in this file, backed by
// one testcontainers Postgres container for the whole package run.
var itPool *pgxpool.Pool

func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx, "postgres:16-alpine",
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
		// No Docker daemon reachable (or similar environment issue) — log
		// and skip the whole package, per backend-conventions.md.
		log.Printf("WARN: skipping booking-service repository integration tests: cannot start postgres testcontainer (no Docker daemon reachable?): %v", err)
		return 0
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("WARN: skipping booking-service repository integration tests: cannot resolve container connection string: %v", err)
		_ = pgContainer.Terminate(ctx)
		return 0
	}

	if err := bootstrapSchema(ctx, connStr); err != nil {
		log.Printf("ERROR: bootstrap schema for booking-service repository integration tests: %v", err)
		_ = pgContainer.Terminate(ctx)
		return 1
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Printf("ERROR: connect pgx pool for booking-service repository integration tests: %v", err)
		_ = pgContainer.Terminate(ctx)
		return 1
	}
	itPool = pool

	code := m.Run()

	pool.Close()
	_ = pgContainer.Terminate(ctx)
	return code
}

// bootstrapSchema prepares the fresh container's schema in two steps:
//
//  1. Hand-creates minimal stand-in tables for `users`/`events`/
//     `ticket_types` — these belong to identity-service/event-service, not
//     booking-service, but booking-service's own migration below declares
//     real FOREIGN KEY constraints against them (Phase 1 shares one
//     Postgres cluster across services, see docs/03-data/postgres-schema.md
//     "Ghi chú về ranh giới database-per-service"). Without these,
//     booking-service's real migration file cannot even apply. This is NOT
//     hand-copying booking-service's own schema — it is the minimal
//     external fixture booking's migration needs to exist at all.
//  2. Applies booking-service's REAL migrations/*.up.sql via
//     golang-migrate — the orders/order_items/tickets tables themselves are
//     never hand-copied.
//
// Uses the pgx stdlib driver (part of the pgx/v5 module already in go.mod)
// to get a database/sql.DB for golang-migrate, since lib/pq isn't a
// dependency of this repo.
func bootstrapSchema(ctx context.Context, connStr string) error {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("open pgx sql.DB: %w", err)
	}
	defer db.Close()

	externalFixtureDDL := []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE TABLE IF NOT EXISTS users (id UUID PRIMARY KEY DEFAULT gen_random_uuid())`,
		`CREATE TABLE IF NOT EXISTS events (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organizer_id UUID NOT NULL REFERENCES users(id),
			status VARCHAR(20) NOT NULL DEFAULT 'published'
		)`,
		`CREATE TABLE IF NOT EXISTS ticket_types (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			event_id UUID NOT NULL REFERENCES events(id),
			price NUMERIC(12,2) NOT NULL,
			quota INT NOT NULL,
			sold_count INT NOT NULL DEFAULT 0
		)`,
	}
	for _, stmt := range externalFixtureDDL {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("create external fixture table: %w", err)
		}
	}

	driver, err := migratepg.WithInstance(db, &migratepg.Config{})
	if err != nil {
		return fmt.Errorf("migrate postgres driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(migrationsSourceURL, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate.NewWithDatabaseInstance: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run booking-service migrations: %w", err)
	}
	return nil
}

// --- seed helpers (distinct names from the pre-existing concurrency test's
// own testPool/setupSchema/seedTicketType, which stay untouched) ---

func itSeedUser(ctx context.Context, t *testing.T) string {
	t.Helper()
	var id string
	err := itPool.QueryRow(ctx, `INSERT INTO users (id) VALUES (gen_random_uuid()) RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return id
}

func itSeedEvent(ctx context.Context, t *testing.T, organizerID, status string) string {
	t.Helper()
	var id string
	err := itPool.QueryRow(ctx,
		`INSERT INTO events (id, organizer_id, status) VALUES (gen_random_uuid(), $1, $2) RETURNING id`,
		organizerID, status).Scan(&id)
	if err != nil {
		t.Fatalf("seed event: %v", err)
	}
	return id
}

func itSeedTicketType(ctx context.Context, t *testing.T, eventID string, price float64, quota int) string {
	t.Helper()
	var id string
	err := itPool.QueryRow(ctx,
		`INSERT INTO ticket_types (id, event_id, price, quota, sold_count) VALUES (gen_random_uuid(), $1, $2, $3, 0) RETURNING id`,
		eventID, price, quota).Scan(&id)
	if err != nil {
		t.Fatalf("seed ticket_type: %v", err)
	}
	return id
}

func itReadSoldCount(ctx context.Context, t *testing.T, ticketTypeID string) int {
	t.Helper()
	var sold int
	if err := itPool.QueryRow(ctx, `SELECT sold_count FROM ticket_types WHERE id = $1`, ticketTypeID).Scan(&sold); err != nil {
		t.Fatalf("read sold_count: %v", err)
	}
	return sold
}

func isNotFound(err error) bool {
	var appErr *apperr.Error
	return errors.As(err, &appErr) && appErr.Code == apperr.ErrNotFound.Code
}

// --- suite: GetOrderByID / ListOrdersByUser / ConfirmPayment / FailPayment happy+conflict paths ---

const (
	testCaseSuccess_GetOrderByID_Found  = "[Success] Tìm thấy order kèm items đã hydrate"
	testCaseError_GetOrderByID_NotFound = "[Error] order_id không tồn tại -> ErrNotFound"

	testCaseSuccess_ListOrdersByUser_Pagination   = "[Success] Phân trang đúng limit/offset, sắp xếp created_at giảm dần"
	testCaseSuccess_ListOrdersByUser_StatusFilter = "[Success] Lọc đúng theo status (pending/paid/cancelled)"

	testCaseSuccess_ConfirmPayment_HappyPath = "[Success] pending -> paid, sinh đúng số lượng tickets"
	testCaseError_ConfirmPayment_WrongStatus = "[Error] Xác nhận lần 2 trên order đã paid -> Conflict"

	testCaseSuccess_FailPayment_HappyPath = "[Success] pending -> cancelled, hoàn trả sold_count"
	testCaseError_FailPayment_WrongStatus = "[Error] Huỷ lần 2 trên order đã cancelled -> Conflict"
)

type BookingRepositoryIntegrationTestSuite struct {
	suite.Suite
	repo *BookingRepository
	ctx  context.Context
}

func TestBookingRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(BookingRepositoryIntegrationTestSuite))
}

func (s *BookingRepositoryIntegrationTestSuite) SetupSuite() {
	s.repo = NewBookingRepository(itPool)
	s.ctx = context.Background()
}

// TearDownTest truncates every table this suite touches between sub-tests
// so one scenario's data never leaks into the next, per
// backend-conventions.md "Integration test" step 6.
func (s *BookingRepositoryIntegrationTestSuite) TearDownTest() {
	_, err := itPool.Exec(s.ctx, `TRUNCATE tickets, order_items, orders, ticket_types, events, users RESTART IDENTITY CASCADE`)
	s.Require().NoError(err)
}

// seedOrder creates one pending order (one line item) via the repository's
// own CreateOrder, for a fresh user/event/ticket_type — dogfooding the same
// method the concurrency test exercises directly, used here purely as
// setup for GetOrderByID/ListOrdersByUser/ConfirmPayment/FailPayment tests.
func (s *BookingRepositoryIntegrationTestSuite) seedOrder(quota, qty int) (*model.Order, ticketTypeSeed) {
	organizerID := itSeedUser(s.ctx, s.T())
	userID := itSeedUser(s.ctx, s.T())
	eventID := itSeedEvent(s.ctx, s.T(), organizerID, "published")
	ticketTypeID := itSeedTicketType(s.ctx, s.T(), eventID, 100000, quota)

	order, err := s.repo.CreateOrder(s.ctx, userID, []model.BookingItem{{TicketTypeID: ticketTypeID, Quantity: qty}})
	s.Require().NoError(err)
	return order, ticketTypeSeed{ticketTypeID: ticketTypeID, userID: userID}
}

type ticketTypeSeed struct {
	ticketTypeID string
	userID       string
}

func (s *BookingRepositoryIntegrationTestSuite) TestGetOrderByID() {
	s.Run(testCaseSuccess_GetOrderByID_Found, func() {
		order, _ := s.seedOrder(10, 2)

		got, err := s.repo.GetOrderByID(s.ctx, order.ID)

		s.Require().NoError(err)
		s.Assert().Equal(order.ID, got.ID)
		s.Assert().Equal(order.UserID, got.UserID)
		s.Assert().Equal(model.OrderStatusPending, got.Status)
		if s.Assert().Len(got.Items, 1) {
			s.Assert().Equal(2, got.Items[0].Quantity)
		}
		s.Assert().Empty(got.Tickets, "tickets are only generated after ConfirmPayment")
	})

	s.Run(testCaseError_GetOrderByID_NotFound, func() {
		_, err := s.repo.GetOrderByID(s.ctx, uuid.NewString())

		s.Require().Error(err)
		s.Assert().True(isNotFound(err))
	})
}

func (s *BookingRepositoryIntegrationTestSuite) TestListOrdersByUser() {
	s.Run(testCaseSuccess_ListOrdersByUser_Pagination, func() {
		organizerID := itSeedUser(s.ctx, s.T())
		userID := itSeedUser(s.ctx, s.T())
		eventID := itSeedEvent(s.ctx, s.T(), organizerID, "published")
		ticketTypeID := itSeedTicketType(s.ctx, s.T(), eventID, 50000, 100)

		const total = 5
		var ids []string
		for i := 0; i < total; i++ {
			order, err := s.repo.CreateOrder(s.ctx, userID, []model.BookingItem{{TicketTypeID: ticketTypeID, Quantity: 1}})
			s.Require().NoError(err)
			ids = append(ids, order.ID)
			time.Sleep(time.Millisecond) // keep created_at strictly increasing for a deterministic DESC order
		}

		page1, gotTotal, err := s.repo.ListOrdersByUser(s.ctx, userID, "", 3, 0)
		s.Require().NoError(err)
		s.Assert().EqualValues(total, gotTotal)
		s.Require().Len(page1, 3)

		page2, gotTotal2, err := s.repo.ListOrdersByUser(s.ctx, userID, "", 3, 3)
		s.Require().NoError(err)
		s.Assert().EqualValues(total, gotTotal2)
		s.Require().Len(page2, 2)

		// Newest-first: the very last order created must be page1[0].
		s.Assert().Equal(ids[total-1], page1[0].ID)
		// No overlap between the two pages.
		seen := make(map[string]bool)
		for _, o := range page1 {
			seen[o.ID] = true
		}
		for _, o := range page2 {
			s.Assert().False(seen[o.ID], "page2 order %s must not also appear in page1", o.ID)
		}
	})

	s.Run(testCaseSuccess_ListOrdersByUser_StatusFilter, func() {
		organizerID := itSeedUser(s.ctx, s.T())
		userID := itSeedUser(s.ctx, s.T())
		eventID := itSeedEvent(s.ctx, s.T(), organizerID, "published")
		ticketTypeID := itSeedTicketType(s.ctx, s.T(), eventID, 50000, 100)

		mkOrder := func() *model.Order {
			order, err := s.repo.CreateOrder(s.ctx, userID, []model.BookingItem{{TicketTypeID: ticketTypeID, Quantity: 1}})
			s.Require().NoError(err)
			return order
		}

		pendingOrder := mkOrder()
		toConfirm := mkOrder()
		toFail := mkOrder()

		_, err := s.repo.ConfirmPayment(s.ctx, toConfirm.ID)
		s.Require().NoError(err)
		_, err = s.repo.FailPayment(s.ctx, toFail.ID)
		s.Require().NoError(err)

		pending, pendingTotal, err := s.repo.ListOrdersByUser(s.ctx, userID, model.OrderStatusPending, 10, 0)
		s.Require().NoError(err)
		s.Assert().EqualValues(1, pendingTotal)
		if s.Assert().Len(pending, 1) {
			s.Assert().Equal(pendingOrder.ID, pending[0].ID)
		}

		paid, paidTotal, err := s.repo.ListOrdersByUser(s.ctx, userID, model.OrderStatusPaid, 10, 0)
		s.Require().NoError(err)
		s.Assert().EqualValues(1, paidTotal)
		if s.Assert().Len(paid, 1) {
			s.Assert().Equal(toConfirm.ID, paid[0].ID)
		}

		cancelled, cancelledTotal, err := s.repo.ListOrdersByUser(s.ctx, userID, model.OrderStatusCancelled, 10, 0)
		s.Require().NoError(err)
		s.Assert().EqualValues(1, cancelledTotal)
		if s.Assert().Len(cancelled, 1) {
			s.Assert().Equal(toFail.ID, cancelled[0].ID)
		}

		all, allTotal, err := s.repo.ListOrdersByUser(s.ctx, userID, "", 10, 0)
		s.Require().NoError(err)
		s.Assert().EqualValues(3, allTotal)
		s.Assert().Len(all, 3)
	})
}

func (s *BookingRepositoryIntegrationTestSuite) TestConfirmPayment() {
	s.Run(testCaseSuccess_ConfirmPayment_HappyPath, func() {
		order, _ := s.seedOrder(10, 3)

		got, err := s.repo.ConfirmPayment(s.ctx, order.ID)

		s.Require().NoError(err)
		s.Assert().Equal(model.OrderStatusPaid, got.Status)
		s.Assert().Len(got.Tickets, 3)

		// Re-read to confirm the write actually persisted, not just the
		// in-memory return value.
		reloaded, err := s.repo.GetOrderByID(s.ctx, order.ID)
		s.Require().NoError(err)
		s.Assert().Equal(model.OrderStatusPaid, reloaded.Status)
		s.Assert().Len(reloaded.Tickets, 3)
	})

	s.Run(testCaseError_ConfirmPayment_WrongStatus, func() {
		order, _ := s.seedOrder(10, 1)
		_, err := s.repo.ConfirmPayment(s.ctx, order.ID)
		s.Require().NoError(err)

		_, err = s.repo.ConfirmPayment(s.ctx, order.ID)

		s.Require().Error(err)
		s.Assert().True(isConflict(err))
	})
}

func (s *BookingRepositoryIntegrationTestSuite) TestFailPayment() {
	s.Run(testCaseSuccess_FailPayment_HappyPath, func() {
		order, seed := s.seedOrder(10, 4)
		s.Require().Equal(4, itReadSoldCount(s.ctx, s.T(), seed.ticketTypeID))

		got, err := s.repo.FailPayment(s.ctx, order.ID)

		s.Require().NoError(err)
		s.Assert().Equal(model.OrderStatusCancelled, got.Status)
		s.Assert().Equal(0, itReadSoldCount(s.ctx, s.T(), seed.ticketTypeID), "sold_count must be restored (compensating transaction)")

		reloaded, err := s.repo.GetOrderByID(s.ctx, order.ID)
		s.Require().NoError(err)
		s.Assert().Equal(model.OrderStatusCancelled, reloaded.Status)
	})

	s.Run(testCaseError_FailPayment_WrongStatus, func() {
		order, _ := s.seedOrder(10, 1)
		_, err := s.repo.FailPayment(s.ctx, order.ID)
		s.Require().NoError(err)

		_, err = s.repo.FailPayment(s.ctx, order.ID)

		s.Require().Error(err)
		s.Assert().True(isConflict(err))
	})
}

// --- mandatory race-condition tests (AGENTS.md: booking/payment transactional
// flows must have concurrent-request coverage, not just happy-path) ---
//
// Modeled after booking_repository_concurrency_test.go's goroutines +
// sync/atomic + start-channel + invariant-assertion structure — only the
// pool source changes (the shared testcontainers pool, itPool, instead of
// TEST_DATABASE_URL).

// TestConfirmPayment_ConcurrentDoubleConfirm_OnlyOneWins races N goroutines
// all trying to confirm payment for the SAME pending order at once. The
// `SELECT status ... FOR UPDATE` + status-check inside ConfirmPayment's
// transaction must let exactly one of them transition pending->paid and
// generate tickets; every other concurrent caller must see the
// already-changed status and get a Conflict, never a second set of tickets.
func TestConfirmPayment_ConcurrentDoubleConfirm_OnlyOneWins(t *testing.T) {
	ctx := context.Background()
	repo := NewBookingRepository(itPool)

	organizerID := itSeedUser(ctx, t)
	userID := itSeedUser(ctx, t)
	eventID := itSeedEvent(ctx, t, organizerID, "published")
	ticketTypeID := itSeedTicketType(ctx, t, eventID, 100000, 10)

	const qty = 2
	order, err := repo.CreateOrder(ctx, userID, []model.BookingItem{{TicketTypeID: ticketTypeID, Quantity: qty}})
	if err != nil {
		t.Fatalf("seed order: %v", err)
	}

	const concurrency = 20
	var (
		wg        sync.WaitGroup
		succeeded int64
		conflicts int64
		otherErrs int64
	)
	start := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := repo.ConfirmPayment(context.Background(), order.ID)
			switch {
			case err == nil:
				atomic.AddInt64(&succeeded, 1)
			case isConflict(err):
				atomic.AddInt64(&conflicts, 1)
			default:
				atomic.AddInt64(&otherErrs, 1)
				t.Logf("unexpected error: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if otherErrs != 0 {
		t.Fatalf("got %d unexpected (non-conflict) errors, want 0", otherErrs)
	}
	if succeeded != 1 {
		t.Fatalf("succeeded = %d, want exactly 1 (exactly one caller must win the pending->paid transition)", succeeded)
	}
	if conflicts != concurrency-1 {
		t.Fatalf("conflicts = %d, want exactly %d", conflicts, concurrency-1)
	}

	var status string
	var ticketCount int
	if err := itPool.QueryRow(ctx, `SELECT status FROM orders WHERE id = $1`, order.ID).Scan(&status); err != nil {
		t.Fatalf("read back order status: %v", err)
	}
	if status != model.OrderStatusPaid {
		t.Fatalf("final order status = %q, want %q", status, model.OrderStatusPaid)
	}
	if err := itPool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tickets t JOIN order_items oi ON oi.id = t.order_item_id WHERE oi.order_id = $1`,
		order.ID).Scan(&ticketCount); err != nil {
		t.Fatalf("count tickets: %v", err)
	}
	if ticketCount != qty {
		t.Fatalf("ticket count = %d, want exactly %d (never double-issued by a losing concurrent confirm)", ticketCount, qty)
	}
}

// TestFailPayment_ConcurrentConfirmVsFail_OnlyOneWins races ConfirmPayment
// against FailPayment on the SAME pending order at once (e.g. a payment
// webhook success racing a hold-expiry cron cancellation). Exactly one of
// the two transitions must win; the loser must see the already-changed
// status and get a Conflict — and the resulting sold_count must match
// whichever transition actually won (paid: stays reserved; cancelled:
// restored), never something in between.
func TestFailPayment_ConcurrentConfirmVsFail_OnlyOneWins(t *testing.T) {
	ctx := context.Background()
	repo := NewBookingRepository(itPool)

	organizerID := itSeedUser(ctx, t)
	userID := itSeedUser(ctx, t)
	eventID := itSeedEvent(ctx, t, organizerID, "published")
	ticketTypeID := itSeedTicketType(ctx, t, eventID, 100000, 10)

	const qty = 1
	order, err := repo.CreateOrder(ctx, userID, []model.BookingItem{{TicketTypeID: ticketTypeID, Quantity: qty}})
	if err != nil {
		t.Fatalf("seed order: %v", err)
	}

	const eachSide = 10
	var (
		wg               sync.WaitGroup
		confirmSucceeded int64
		failSucceeded    int64
		conflicts        int64
		otherErrs        int64
	)
	start := make(chan struct{})

	for i := 0; i < eachSide; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := repo.ConfirmPayment(context.Background(), order.ID)
			switch {
			case err == nil:
				atomic.AddInt64(&confirmSucceeded, 1)
			case isConflict(err):
				atomic.AddInt64(&conflicts, 1)
			default:
				atomic.AddInt64(&otherErrs, 1)
				t.Logf("unexpected ConfirmPayment error: %v", err)
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := repo.FailPayment(context.Background(), order.ID)
			switch {
			case err == nil:
				atomic.AddInt64(&failSucceeded, 1)
			case isConflict(err):
				atomic.AddInt64(&conflicts, 1)
			default:
				atomic.AddInt64(&otherErrs, 1)
				t.Logf("unexpected FailPayment error: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if otherErrs != 0 {
		t.Fatalf("got %d unexpected (non-conflict) errors, want 0", otherErrs)
	}
	totalWins := confirmSucceeded + failSucceeded
	if totalWins != 1 {
		t.Fatalf("total successful transitions = %d (confirm=%d, fail=%d), want exactly 1 — Confirm and Fail must be mutually exclusive on the same order",
			totalWins, confirmSucceeded, failSucceeded)
	}
	if conflicts != int64(2*eachSide)-1 {
		t.Fatalf("conflicts = %d, want exactly %d", conflicts, int64(2*eachSide)-1)
	}

	sold := itReadSoldCount(ctx, t, ticketTypeID)
	var status string
	if err := itPool.QueryRow(ctx, `SELECT status FROM orders WHERE id = $1`, order.ID).Scan(&status); err != nil {
		t.Fatalf("read back order status: %v", err)
	}

	switch {
	case confirmSucceeded == 1:
		if status != model.OrderStatusPaid {
			t.Fatalf("Confirm won but final status = %q, want %q", status, model.OrderStatusPaid)
		}
		if sold != qty {
			t.Fatalf("Confirm won but sold_count = %d, want %d (stock stays reserved on a paid order)", sold, qty)
		}
	case failSucceeded == 1:
		if status != model.OrderStatusCancelled {
			t.Fatalf("Fail won but final status = %q, want %q", status, model.OrderStatusCancelled)
		}
		if sold != 0 {
			t.Fatalf("Fail won but sold_count = %d, want 0 (compensating transaction must restore it)", sold)
		}
	}
}
