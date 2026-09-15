package repository

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"ticketflow/pkg/apperr"
	"ticketflow/services/booking-service/internal/model"
)

// This is the mandatory race-condition test required by
// src/services/booking-service/README.md before the booking flow counts as
// done: N goroutines race to book the same ticket_type concurrently, and
// the ACID transaction in CreateOrder (SELECT ... FOR UPDATE, see
// booking_repository.go) must ensure sold_count never exceeds quota no
// matter how many requests arrive at once — the Redis lock in
// internal/lock is a load-shedding gate, not tested here, since this test
// specifically verifies the Postgres-level correctness guarantee holds
// even without it (per docs/02-domains/booking/spec.md: "Không được bỏ
// SELECT ... FOR UPDATE với lý do đã có Redis lock rồi").
//
// Requires a reachable Postgres (TEST_DATABASE_URL, defaults to the
// docker-compose instance) — skips instead of failing when none is
// available, so `go build`/`go vet` and CI steps without Postgres still pass.

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://ticketflow:ticketflow@localhost:5432/ticketflow?sslmode=disable"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("skipping concurrency test: cannot create postgres pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skipping concurrency test: postgres not reachable at %s: %v", dsn, err)
	}
	return pool
}

// setupSchema creates the minimal subset of the real schema this test
// touches (mirroring the migrations in identity-service/event-service/
// booking-service) so the test is self-contained and doesn't depend on a
// migration runner having been executed first.
func setupSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	stmts := []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto`,
		`CREATE TABLE IF NOT EXISTS users (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), email VARCHAR(255) UNIQUE NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS events (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), organizer_id UUID NOT NULL REFERENCES users(id), status VARCHAR(20) NOT NULL DEFAULT 'published')`,
		`CREATE TABLE IF NOT EXISTS ticket_types (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), event_id UUID NOT NULL REFERENCES events(id), price NUMERIC(12,2) NOT NULL, quota INT NOT NULL, sold_count INT NOT NULL DEFAULT 0)`,
		`CREATE TABLE IF NOT EXISTS orders (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID NOT NULL REFERENCES users(id), status VARCHAR(20) NOT NULL DEFAULT 'pending', total_amount NUMERIC(12,2) NOT NULL, expires_at TIMESTAMPTZ NOT NULL, created_at TIMESTAMPTZ DEFAULT now())`,
		`CREATE TABLE IF NOT EXISTS order_items (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), order_id UUID NOT NULL REFERENCES orders(id), ticket_type_id UUID NOT NULL REFERENCES ticket_types(id), quantity INT NOT NULL, unit_price NUMERIC(12,2) NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS tickets (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), order_item_id UUID NOT NULL REFERENCES order_items(id), ticket_code VARCHAR(64) UNIQUE NOT NULL, status VARCHAR(20) NOT NULL DEFAULT 'valid', issued_at TIMESTAMPTZ DEFAULT now())`,
	}
	for _, stmt := range stmts {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			t.Fatalf("setup schema: %v", err)
		}
	}
}

// seedTicketType inserts one fresh user+event+ticket_type row with the
// given quota and returns the ticket_type id and a user id to book as.
func seedTicketType(t *testing.T, pool *pgxpool.Pool, quota int) (ticketTypeID, userID string) {
	t.Helper()
	ctx := context.Background()

	if err := pool.QueryRow(ctx,
		`INSERT INTO users (id, email) VALUES ($1, $2) RETURNING id`,
		uuid.NewString(), uuid.NewString()+"@example.test").Scan(&userID); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	var organizerID string
	if err := pool.QueryRow(ctx,
		`INSERT INTO users (id, email) VALUES ($1, $2) RETURNING id`,
		uuid.NewString(), uuid.NewString()+"@example.test").Scan(&organizerID); err != nil {
		t.Fatalf("seed organizer: %v", err)
	}
	var eventID string
	if err := pool.QueryRow(ctx,
		`INSERT INTO events (id, organizer_id, status) VALUES ($1, $2, 'published') RETURNING id`,
		uuid.NewString(), organizerID).Scan(&eventID); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	if err := pool.QueryRow(ctx,
		`INSERT INTO ticket_types (id, event_id, price, quota, sold_count) VALUES ($1, $2, 100000, $3, 0) RETURNING id`,
		uuid.NewString(), eventID, quota).Scan(&ticketTypeID); err != nil {
		t.Fatalf("seed ticket_type: %v", err)
	}
	return ticketTypeID, userID
}

func TestCreateOrder_ConcurrentBookings_NeverOversells(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	setupSchema(t, pool)

	const quota = 10
	const concurrency = 50

	ticketTypeID, userID := seedTicketType(t, pool, quota)
	repo := NewBookingRepository(pool)

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
			<-start // release every goroutine as close to simultaneously as possible
			_, err := repo.CreateOrder(context.Background(), userID, []model.BookingItem{
				{TicketTypeID: ticketTypeID, Quantity: 1},
			})
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
	if succeeded != quota {
		t.Fatalf("succeeded = %d, want exactly quota = %d", succeeded, quota)
	}
	if conflicts != concurrency-quota {
		t.Fatalf("conflicts = %d, want exactly %d", conflicts, concurrency-quota)
	}

	var soldCount int
	if err := pool.QueryRow(context.Background(), `SELECT sold_count FROM ticket_types WHERE id = $1`, ticketTypeID).Scan(&soldCount); err != nil {
		t.Fatalf("read back sold_count: %v", err)
	}
	if soldCount != quota {
		t.Fatalf("final sold_count = %d, want exactly quota = %d (never oversold, never undercounted)", soldCount, quota)
	}

	var totalQuantity int
	if err := pool.QueryRow(context.Background(),
		`SELECT COALESCE(SUM(quantity), 0) FROM order_items WHERE ticket_type_id = $1`, ticketTypeID).Scan(&totalQuantity); err != nil {
		t.Fatalf("sum order_items.quantity: %v", err)
	}
	if totalQuantity != quota {
		t.Fatalf("sum(order_items.quantity) = %d, want exactly quota = %d", totalQuantity, quota)
	}
}

// TestCreateOrder_ConcurrentBookings_PartialFitRejected exercises the case
// where a request wants more units than the single remaining slot can
// cover: with quota=10 and quantity=3 per request, only 3 requests can
// fully succeed (consuming 9) — a 4th succeeding would overshoot the
// remaining 1 unit and must be rejected outright, not partially fulfilled.
func TestCreateOrder_ConcurrentBookings_PartialFitRejected(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	setupSchema(t, pool)

	const quota = 10
	const quantityPerRequest = 3
	const concurrency = 20

	ticketTypeID, userID := seedTicketType(t, pool, quota)
	repo := NewBookingRepository(pool)

	var (
		wg        sync.WaitGroup
		succeeded int64
	)
	start := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := repo.CreateOrder(context.Background(), userID, []model.BookingItem{
				{TicketTypeID: ticketTypeID, Quantity: quantityPerRequest},
			})
			if err == nil {
				atomic.AddInt64(&succeeded, 1)
			}
		}()
	}
	close(start)
	wg.Wait()

	wantSucceeded := int64(quota / quantityPerRequest) // 3
	if succeeded != wantSucceeded {
		t.Fatalf("succeeded = %d, want exactly %d (floor(quota/quantity))", succeeded, wantSucceeded)
	}

	var soldCount int
	if err := pool.QueryRow(context.Background(), `SELECT sold_count FROM ticket_types WHERE id = $1`, ticketTypeID).Scan(&soldCount); err != nil {
		t.Fatalf("read back sold_count: %v", err)
	}
	if soldCount != int(wantSucceeded)*quantityPerRequest {
		t.Fatalf("final sold_count = %d, want %d (never overshoots quota)", soldCount, int(wantSucceeded)*quantityPerRequest)
	}
	if soldCount > quota {
		t.Fatalf("final sold_count = %d exceeds quota = %d — oversold!", soldCount, quota)
	}
}

func isConflict(err error) bool {
	var appErr *apperr.Error
	return errors.As(err, &appErr) && appErr.Code == apperr.ErrConflict.Code
}
