package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

// SessionRepository has no Postgres dependency at all (it's a thin Redis
// wrapper — see the package doc comment in session_repository.go): these
// cases exercise real Redis wire-protocol behavior (TTL expiry, delete-on-
// read rotation) against an embedded miniredis server, not the package's
// shared testcontainers Postgres. miniredis itself needs no Docker.
//
// Note: this file still shares user_repository_test.go's package-level
// TestMain, which — per backend-conventions.md's documented convention —
// gates the *entire* package behind a reachable Docker daemon and calls
// os.Exit(0) before m.Run() when none is found, "coi như skip toàn bộ
// package". So in a Docker-less environment these cases would be skipped
// too, even though nothing here actually touches Postgres — they are not
// reachable independently of that gate given the one-TestMain-per-package
// convention this repo uses for integration tests. Where Docker *is*
// available (as in this sandbox), they run for real every time, regardless
// of whether the Postgres container in TestMain came up.

const (
	testCaseSuccess_BlacklistThenIsBlacklistedTrue = "[Success] blacklisted jti reports true"
	testCaseSuccess_NotBlacklistedReportsFalse     = "[Success] unknown jti reports false"
	testCaseSuccess_TTLZeroNoOpSkipsBlacklist      = "[Success] ttl<=0 is a no-op — key is never set"
	testCaseSuccess_BlacklistExpiresAfterTTL       = "[Success] blacklist entry expires after its ttl"

	testCaseSuccess_StoreThenConsumeReturnsUserID = "[Success] consume returns the stored user id"
	testCaseError_ConsumeTwiceFailsSecondTime     = "[Error] refresh token is single-use (delete-on-read)"
	testCaseError_ConsumeUnknownTokenFails        = "[Error] consuming an unknown/never-stored token fails"

	testCaseSuccess_RevokeRemovesToken = "[Success] revoke deletes the refresh token so a later consume fails"
)

type SessionRepositoryTestSuite struct {
	suite.Suite
	mr   *miniredis.Miniredis
	rdb  *redis.Client
	repo *SessionRepository
}

func TestSessionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(SessionRepositoryTestSuite))
}

func (s *SessionRepositoryTestSuite) SetupTest() {
	mr, err := miniredis.Run()
	s.Require().NoError(err)
	s.mr = mr
	s.rdb = redis.NewClient(&redis.Options{Addr: mr.Addr()})
	s.repo = NewSessionRepository(s.rdb)
}

func (s *SessionRepositoryTestSuite) TearDownTest() {
	_ = s.rdb.Close()
	s.mr.Close()
}

func (s *SessionRepositoryTestSuite) TestBlacklistAccessToken() {
	s.Run(testCaseSuccess_BlacklistThenIsBlacklistedTrue, func() {
		ctx := context.Background()
		s.Require().NoError(s.repo.BlacklistAccessToken(ctx, "jti-1", time.Minute))

		got, err := s.repo.IsBlacklisted(ctx, "jti-1")

		s.Require().NoError(err)
		s.True(got)
	})

	s.Run(testCaseSuccess_TTLZeroNoOpSkipsBlacklist, func() {
		ctx := context.Background()
		s.Require().NoError(s.repo.BlacklistAccessToken(ctx, "jti-2", 0))
		got, err := s.repo.IsBlacklisted(ctx, "jti-2")
		s.Require().NoError(err)
		s.False(got)

		s.Require().NoError(s.repo.BlacklistAccessToken(ctx, "jti-3", -time.Second))
		got, err = s.repo.IsBlacklisted(ctx, "jti-3")
		s.Require().NoError(err)
		s.False(got)
	})

	s.Run(testCaseSuccess_BlacklistExpiresAfterTTL, func() {
		ctx := context.Background()
		s.Require().NoError(s.repo.BlacklistAccessToken(ctx, "jti-4", time.Second))
		s.mr.FastForward(2 * time.Second)

		got, err := s.repo.IsBlacklisted(ctx, "jti-4")

		s.Require().NoError(err)
		s.False(got)
	})
}

func (s *SessionRepositoryTestSuite) TestIsBlacklisted() {
	s.Run(testCaseSuccess_NotBlacklistedReportsFalse, func() {
		got, err := s.repo.IsBlacklisted(context.Background(), "never-blacklisted")

		s.Require().NoError(err)
		s.False(got)
	})
}

func (s *SessionRepositoryTestSuite) TestRefreshTokenLifecycle() {
	s.Run(testCaseSuccess_StoreThenConsumeReturnsUserID, func() {
		ctx := context.Background()
		s.Require().NoError(s.repo.StoreRefreshToken(ctx, "token-1", "user-1", time.Hour))

		userID, err := s.repo.ConsumeRefreshToken(ctx, "token-1")

		s.Require().NoError(err)
		s.Equal("user-1", userID)
	})

	s.Run(testCaseError_ConsumeTwiceFailsSecondTime, func() {
		ctx := context.Background()
		s.Require().NoError(s.repo.StoreRefreshToken(ctx, "token-2", "user-2", time.Hour))
		_, err := s.repo.ConsumeRefreshToken(ctx, "token-2")
		s.Require().NoError(err)

		_, err = s.repo.ConsumeRefreshToken(ctx, "token-2")

		s.Require().Error(err)
	})

	s.Run(testCaseError_ConsumeUnknownTokenFails, func() {
		_, err := s.repo.ConsumeRefreshToken(context.Background(), "never-stored")

		s.Require().Error(err)
	})
}

func (s *SessionRepositoryTestSuite) TestRevokeRefreshToken() {
	s.Run(testCaseSuccess_RevokeRemovesToken, func() {
		ctx := context.Background()
		s.Require().NoError(s.repo.StoreRefreshToken(ctx, "token-3", "user-3", time.Hour))

		s.Require().NoError(s.repo.RevokeRefreshToken(ctx, "token-3"))

		_, err := s.repo.ConsumeRefreshToken(ctx, "token-3")
		s.Require().Error(err)
	})
}
