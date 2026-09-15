// Package lock wraps redsync (Redlock) around the `lock:ticket_type:{id}`
// key pattern from docs/03-data/redis-keys.md: a load-shedding gate in
// front of the Postgres SELECT...FOR UPDATE transaction that is the real
// correctness guarantee (see docs/02-domains/booking/spec.md) — losing this
// lock (Redis restart, expiry) must never risk overselling.
package lock

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/go-redsync/redsync/v4"
	redsyncgoredis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"

	"ticketflow/pkg/apperr"
)

const lockTTL = 8 * time.Second

type TicketTypeLocker struct {
	rs *redsync.Redsync
}

func NewTicketTypeLocker(client *redis.Client) *TicketTypeLocker {
	return &TicketTypeLocker{rs: redsync.New(redsyncgoredis.NewPool(client))}
}

type Unlock func()

// AcquireMany locks every distinct ticket_type_id needed for one booking
// request in one call. IDs are sorted before locking so concurrent
// multi-item bookings that share a ticket type always acquire Redis locks
// in the same order — mirrors the Postgres-side `ORDER BY id FOR UPDATE`
// deadlock-safety ordering. Fails fast (single try, no blocking/retry) on
// contention, matching the "load-shedding" intent: reject fast under a
// flash-sale spike rather than queueing requests up.
func (l *TicketTypeLocker) AcquireMany(ctx context.Context, ticketTypeIDs []string) (Unlock, error) {
	sorted := append([]string(nil), ticketTypeIDs...)
	sort.Strings(sorted)

	acquired := make([]*redsync.Mutex, 0, len(sorted))
	for _, id := range sorted {
		m := l.rs.NewMutex(fmt.Sprintf("lock:ticket_type:%s", id),
			redsync.WithExpiry(lockTTL),
			redsync.WithTries(1),
		)
		if err := m.LockContext(ctx); err != nil {
			for _, done := range acquired {
				_, _ = done.UnlockContext(ctx)
			}
			return nil, apperr.WithMessage(apperr.ErrTooManyRequests,
				"Hệ thống đang xử lý nhiều yêu cầu cho loại vé này, vui lòng thử lại sau vài giây")
		}
		acquired = append(acquired, m)
	}

	return func() {
		for _, m := range acquired {
			_, _ = m.UnlockContext(ctx)
		}
	}, nil
}
