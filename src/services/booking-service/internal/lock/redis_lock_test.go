package lock

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"

	"ticketflow/pkg/apperr"
)

// This package tests AcquireMany against a REAL *redis.Client backed by an
// in-process miniredis server (not redismock) — per backend-conventions.md
// / the go-test task brief, redsync needs real Lua-script / SET-NX
// transport semantics that a call-expectation mock can't reproduce
// faithfully. miniredis runs in-process, so this needs no Docker.

const (
	testCaseSuccess_AcquireMany_SortsIDsBeforeLocking   = "[Success] Sắp xếp ticket_type_id tăng dần trước khi khoá (khớp thứ tự FOR UPDATE ở Postgres)"
	testCaseSuccess_AcquireMany_RoundTripLockThenUnlock = "[Success] Acquire rồi unlock giải phóng khoá, cho phép acquire lại"
	testCaseError_AcquireMany_FailsFastOnContention     = "[Error] Từ chối ngay (không retry) khi 1 id đã bị khoá, và rollback các id đã lock trong cùng lượt gọi"
)

type TicketTypeLockerTestSuite struct {
	suite.Suite

	mr     *miniredis.Miniredis
	client *redis.Client
}

func (s *TicketTypeLockerTestSuite) SetupTest() {
	mr, err := miniredis.Run()
	s.Require().NoError(err)
	s.mr = mr
	s.client = redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func (s *TicketTypeLockerTestSuite) TearDownTest() {
	_ = s.client.Close()
	s.mr.Close()
}

func TestTicketTypeLockerTestSuite(t *testing.T) {
	suite.Run(t, new(TicketTypeLockerTestSuite))
}

// cmdRecorder is a go-redis v9 Hook that records the key argument of every
// "set" command sent over the wire — used to observe the actual order
// AcquireMany issues its Redis SET NX lock commands in, independent of the
// order ticketTypeIDs was given in.
type cmdRecorder struct {
	mu   sync.Mutex
	keys []string
}

func (r *cmdRecorder) DialHook(next redis.DialHook) redis.DialHook { return next }

func (r *cmdRecorder) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		args := cmd.Args()
		if len(args) >= 2 {
			if name, ok := args[0].(string); ok && strings.EqualFold(name, "set") {
				if key, ok := args[1].(string); ok {
					r.mu.Lock()
					r.keys = append(r.keys, key)
					r.mu.Unlock()
				}
			}
		}
		return next(ctx, cmd)
	}
}

func (r *cmdRecorder) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func (r *cmdRecorder) Keys() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.keys...)
}

func (s *TicketTypeLockerTestSuite) TestAcquireMany() {
	s.Run(testCaseSuccess_AcquireMany_SortsIDsBeforeLocking, func() {
		rec := &cmdRecorder{}
		s.client.AddHook(rec)
		locker := NewTicketTypeLocker(s.client)

		unlock, err := locker.AcquireMany(context.Background(), []string{"charlie", "alpha", "bravo"})
		s.Require().NoError(err)
		defer unlock()

		s.Assert().Equal(
			[]string{"lock:ticket_type:alpha", "lock:ticket_type:bravo", "lock:ticket_type:charlie"},
			rec.Keys(),
			"SET NX commands must be issued in ascending sorted id order, not input order",
		)
	})

	s.Run(testCaseSuccess_AcquireMany_RoundTripLockThenUnlock, func() {
		locker := NewTicketTypeLocker(s.client)

		unlock, err := locker.AcquireMany(context.Background(), []string{"tt-1"})
		s.Require().NoError(err)
		unlock()

		// Now free again — a second AcquireMany on the same id must succeed.
		unlock2, err := locker.AcquireMany(context.Background(), []string{"tt-1"})
		s.Require().NoError(err)
		unlock2()
	})

	s.Run(testCaseError_AcquireMany_FailsFastOnContention, func() {
		locker := NewTicketTypeLocker(s.client)

		// Hold "y" externally (simulating another concurrent booking request).
		unlockY, err := locker.AcquireMany(context.Background(), []string{"y"})
		s.Require().NoError(err)
		defer unlockY()

		// Sorted order is [x, y]: "x" acquires fine, "y" then fails ->
		// AcquireMany must roll back "x" before returning, not leak it held.
		_, err = locker.AcquireMany(context.Background(), []string{"x", "y"})
		s.Require().Error(err)
		var appErr *apperr.Error
		s.Require().True(errors.As(err, &appErr))
		s.Assert().Equal(apperr.ErrTooManyRequests.Code, appErr.Code)

		// "x" must have been released by the rollback — acquiring it alone
		// now must succeed immediately (single try, no retry available).
		unlockX, err := locker.AcquireMany(context.Background(), []string{"x"})
		s.Require().NoError(err, "AcquireMany must roll back partially-acquired locks on failure")
		unlockX()
	})
}
