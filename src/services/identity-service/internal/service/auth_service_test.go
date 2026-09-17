package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/authclaims"
	"ticketflow/services/identity-service/internal/firebase"
	"ticketflow/services/identity-service/internal/model"
	"ticketflow/services/identity-service/internal/repository"
	"ticketflow/services/identity-service/internal/service/mocks"
)

var testJWTSecret = []byte("test-secret-key-for-unit-tests")

const (
	// Login
	testCaseSuccess_LoginFoundByFirebaseUID   = "[Success] user found directly by firebase_uid issues token pair"
	testCaseSuccess_LoginFallbackByEmailLinks = "[Success] fallback lookup by email links firebase_uid onto existing dev account"
	testCaseSuccess_LoginCreatesNewUser       = "[Success] no match by firebase_uid or email creates a brand-new user"
	testCaseError_LoginBannedUserRejected     = "[Error] banned user is rejected even though lookup succeeds"
	testCaseError_LoginVerifyFails            = "[Error] firebase verifier failure maps to unauthorized"
	testCaseError_LoginFirebaseUIDLookupFails = "[Error] unexpected (non-not-found) error from GetByFirebaseUID propagates"
	testCaseError_LoginEmailLookupFails       = "[Error] unexpected (non-not-found) error from GetByEmail propagates"
	testCaseError_LoginCreateFails            = "[Error] Create failure while provisioning a new user propagates"
	testCaseError_LoginLinkFirebaseUIDFails   = "[Error] LinkFirebaseUID failure while linking a dev account propagates"

	// Refresh
	testCaseError_RefreshEmptyToken      = "[Error] empty refresh token is rejected before touching any store"
	testCaseError_RefreshConsumeFails    = "[Error] unknown/expired/already-used refresh token is rejected"
	testCaseError_RefreshUserLookupFails = "[Error] user missing after token is consumed is rejected (token still burned)"
	testCaseError_RefreshBannedUser      = "[Error] banned user is rejected on refresh"
	testCaseSuccess_RefreshRotatesToken  = "[Success] valid refresh rotates the token and issues a new pair"

	// Logout
	testCaseSuccess_LogoutBlacklistsAndRevokes     = "[Success] positive ttl blacklists jti and revokes refresh token"
	testCaseSuccess_LogoutSkipsBlacklistOnZeroTTL  = "[Success] already-expired access token skips the blacklist call entirely"
	testCaseSuccess_LogoutSkipsRevokeOnEmptyToken  = "[Success] empty refresh token skips revoke without erroring"
	testCaseError_LogoutBlacklistFailurePropagates = "[Error] blacklist store failure propagates as an error"

	// CheckSession
	testCaseError_CheckSessionBlacklistLookupFails       = "[Error] blacklist lookup failure propagates and short-circuits"
	testCaseSuccess_CheckSessionFailClosedOnMissingUser  = "[Success] fail-closed: missing user reports StatusBanned regardless of blacklist state"
	testCaseSuccess_CheckSessionActiveUserNotBlacklisted = "[Success] active, non-blacklisted user reports its real status"
	testCaseSuccess_CheckSessionBlacklistedActiveUser    = "[Success] blacklisted flag is preserved alongside the user's real status"
)

// mockFirebaseToken builds a MockVerifier-compatible base64(JSON) token.
func mockFirebaseToken(uid, email, name string) string {
	raw, _ := json.Marshal(firebase.VerifiedToken{UID: uid, Email: email, Name: name})
	return base64.StdEncoding.EncodeToString(raw)
}

func appErrCode(t *testing.T, err error) string {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected *apperr.Error, got %T: %v", err, err)
	}
	return appErr.Code
}

type AuthServiceTestSuite struct {
	suite.Suite
	mr       *miniredis.Miniredis
	rdb      *redis.Client
	sessions *repository.SessionRepository
	verifier *firebase.MockVerifier
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

func (s *AuthServiceTestSuite) SetupTest() {
	mr, err := miniredis.Run()
	s.Require().NoError(err)
	s.mr = mr
	s.rdb = redis.NewClient(&redis.Options{Addr: mr.Addr()})
	s.sessions = repository.NewSessionRepository(s.rdb)
	s.verifier = firebase.NewMockVerifier()
}

// SetupSubTest gives every s.Run(...) sub-test a clean Redis state, since
// SetupTest only runs once per Test method (per backend-conventions.md's
// note on sharing a DB/container across sub-tests within one suite method).
func (s *AuthServiceTestSuite) SetupSubTest() {
	s.mr.FlushAll()
}

func (s *AuthServiceTestSuite) TearDownTest() {
	_ = s.rdb.Close()
	s.mr.Close()
}

func (s *AuthServiceTestSuite) newService(users UserRepository) *AuthService {
	return NewAuthService(users, s.sessions, s.verifier, testJWTSecret)
}

// newServiceWithBrokenSessions builds an AuthService backed by its own,
// separate Redis instance that is shut down immediately — simulating a
// Redis connection failure — without touching the suite-wide s.mr/s.rdb
// shared by every other sub-test in the same Test method (SetupTest/
// SetupSubTest cannot recreate those between s.Run calls).
func (s *AuthServiceTestSuite) newServiceWithBrokenSessions(users UserRepository) *AuthService {
	mr, err := miniredis.Run()
	s.Require().NoError(err)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	mr.Close()
	brokenSessions := repository.NewSessionRepository(rdb)
	return NewAuthService(users, brokenSessions, s.verifier, testJWTSecret)
}

func (s *AuthServiceTestSuite) TestLogin() {
	s.Run(testCaseSuccess_LoginFoundByFirebaseUID, func() {
		users := mocks.NewUserRepository(s.T())
		user := &model.User{ID: "user-1", Email: "a@example.com", Role: model.RoleUser, Status: model.StatusActive}
		users.On("GetByFirebaseUID", mock.Anything, "uid-1").Return(user, nil)
		svc := s.newService(users)

		result, err := svc.Login(context.Background(), mockFirebaseToken("uid-1", "a@example.com", ""))

		s.Require().NoError(err)
		s.NotEmpty(result.AccessToken)
		s.NotEmpty(result.RefreshToken)
		s.Equal(int(authclaims.AccessTokenTTL.Seconds()), result.ExpiresIn)
		s.Same(user, result.User)
	})

	s.Run(testCaseSuccess_LoginFallbackByEmailLinks, func() {
		users := mocks.NewUserRepository(s.T())
		existing := &model.User{ID: "user-2", Email: "b@example.com", Role: model.RoleOrganizer, Status: model.StatusActive}
		users.On("GetByFirebaseUID", mock.Anything, "uid-2").Return(nil, gorm.ErrRecordNotFound)
		users.On("GetByEmail", mock.Anything, "b@example.com").Return(existing, nil)
		users.On("LinkFirebaseUID", mock.Anything, "user-2", "uid-2").Return(nil)
		svc := s.newService(users)

		result, err := svc.Login(context.Background(), mockFirebaseToken("uid-2", "b@example.com", ""))

		s.Require().NoError(err)
		s.Equal("user-2", result.User.ID)
		s.Require().NotNil(result.User.FirebaseUID)
		s.Equal("uid-2", *result.User.FirebaseUID)
	})

	s.Run(testCaseSuccess_LoginCreatesNewUser, func() {
		users := mocks.NewUserRepository(s.T())
		users.On("GetByFirebaseUID", mock.Anything, "uid-3").Return(nil, gorm.ErrRecordNotFound)
		users.On("GetByEmail", mock.Anything, "c@example.com").Return(nil, gorm.ErrRecordNotFound)
		users.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
			return u.Email == "c@example.com" && u.Role == model.RoleUser && u.Status == model.StatusActive &&
				u.FirebaseUID != nil && *u.FirebaseUID == "uid-3"
		})).Return(nil)
		svc := s.newService(users)

		result, err := svc.Login(context.Background(), mockFirebaseToken("uid-3", "c@example.com", "Nguyen Van C"))

		s.Require().NoError(err)
		s.Equal("c@example.com", result.User.Email)
		s.Require().NotNil(result.User.FullName)
		s.Equal("Nguyen Van C", *result.User.FullName)
		s.NotEmpty(result.User.ID)
	})

	s.Run(testCaseError_LoginBannedUserRejected, func() {
		users := mocks.NewUserRepository(s.T())
		banned := &model.User{ID: "user-4", Email: "d@example.com", Role: model.RoleUser, Status: model.StatusBanned}
		users.On("GetByFirebaseUID", mock.Anything, "uid-4").Return(banned, nil)
		svc := s.newService(users)

		result, err := svc.Login(context.Background(), mockFirebaseToken("uid-4", "d@example.com", ""))

		s.Nil(result)
		s.Require().Error(err)
		s.Equal(apperr.ErrUnauthorized.Code, appErrCode(s.T(), err))
	})

	s.Run(testCaseError_LoginVerifyFails, func() {
		users := mocks.NewUserRepository(s.T())
		svc := s.newService(users)

		result, err := svc.Login(context.Background(), "not-valid-base64!!!")

		s.Nil(result)
		s.Require().Error(err)
		s.Equal(apperr.ErrUnauthorized.Code, appErrCode(s.T(), err))
	})

	s.Run(testCaseError_LoginFirebaseUIDLookupFails, func() {
		users := mocks.NewUserRepository(s.T())
		dbErr := errors.New("connection reset")
		users.On("GetByFirebaseUID", mock.Anything, "uid-5").Return(nil, dbErr)
		svc := s.newService(users)

		result, err := svc.Login(context.Background(), mockFirebaseToken("uid-5", "e@example.com", ""))

		s.Nil(result)
		s.Require().ErrorIs(err, dbErr)
	})

	s.Run(testCaseError_LoginEmailLookupFails, func() {
		users := mocks.NewUserRepository(s.T())
		dbErr := errors.New("connection reset")
		users.On("GetByFirebaseUID", mock.Anything, "uid-6").Return(nil, gorm.ErrRecordNotFound)
		users.On("GetByEmail", mock.Anything, "f@example.com").Return(nil, dbErr)
		svc := s.newService(users)

		result, err := svc.Login(context.Background(), mockFirebaseToken("uid-6", "f@example.com", ""))

		s.Nil(result)
		s.Require().ErrorIs(err, dbErr)
	})

	s.Run(testCaseError_LoginCreateFails, func() {
		users := mocks.NewUserRepository(s.T())
		dbErr := errors.New("unique violation")
		users.On("GetByFirebaseUID", mock.Anything, "uid-7").Return(nil, gorm.ErrRecordNotFound)
		users.On("GetByEmail", mock.Anything, "g@example.com").Return(nil, gorm.ErrRecordNotFound)
		users.On("Create", mock.Anything, mock.Anything).Return(dbErr)
		svc := s.newService(users)

		result, err := svc.Login(context.Background(), mockFirebaseToken("uid-7", "g@example.com", ""))

		s.Nil(result)
		s.Require().ErrorIs(err, dbErr)
	})

	s.Run(testCaseError_LoginLinkFirebaseUIDFails, func() {
		users := mocks.NewUserRepository(s.T())
		dbErr := errors.New("update failed")
		existing := &model.User{ID: "user-8", Email: "h@example.com", Role: model.RoleUser, Status: model.StatusActive}
		users.On("GetByFirebaseUID", mock.Anything, "uid-8").Return(nil, gorm.ErrRecordNotFound)
		users.On("GetByEmail", mock.Anything, "h@example.com").Return(existing, nil)
		users.On("LinkFirebaseUID", mock.Anything, "user-8", "uid-8").Return(dbErr)
		svc := s.newService(users)

		result, err := svc.Login(context.Background(), mockFirebaseToken("uid-8", "h@example.com", ""))

		s.Nil(result)
		s.Require().ErrorIs(err, dbErr)
	})
}

func (s *AuthServiceTestSuite) TestRefresh() {
	s.Run(testCaseError_RefreshEmptyToken, func() {
		users := mocks.NewUserRepository(s.T()) // no expectations: must not be called
		svc := s.newService(users)

		result, err := svc.Refresh(context.Background(), "")

		s.Nil(result)
		s.Require().Error(err)
		s.Equal(apperr.ErrUnauthorized.Code, appErrCode(s.T(), err))
	})

	s.Run(testCaseError_RefreshConsumeFails, func() {
		users := mocks.NewUserRepository(s.T())
		svc := s.newService(users)

		result, err := svc.Refresh(context.Background(), "never-issued-token")

		s.Nil(result)
		s.Require().Error(err)
		s.Equal(apperr.ErrUnauthorized.Code, appErrCode(s.T(), err))
	})

	s.Run(testCaseError_RefreshUserLookupFails, func() {
		users := mocks.NewUserRepository(s.T())
		users.On("GetByID", mock.Anything, "user-1").Return(nil, gorm.ErrRecordNotFound)
		svc := s.newService(users)
		s.Require().NoError(s.sessions.StoreRefreshToken(context.Background(), "refresh-tok-1", "user-1", RefreshTokenTTL))

		result, err := svc.Refresh(context.Background(), "refresh-tok-1")

		s.Nil(result)
		s.Require().Error(err)
		s.Equal(apperr.ErrUnauthorized.Code, appErrCode(s.T(), err))

		// token must have been consumed (deleted) even though the subsequent
		// user lookup failed — rotation is unconditional on read.
		_, consumeErr := s.sessions.ConsumeRefreshToken(context.Background(), "refresh-tok-1")
		s.Require().Error(consumeErr)
	})

	s.Run(testCaseError_RefreshBannedUser, func() {
		users := mocks.NewUserRepository(s.T())
		users.On("GetByID", mock.Anything, "user-2").Return(&model.User{ID: "user-2", Status: model.StatusBanned}, nil)
		svc := s.newService(users)
		s.Require().NoError(s.sessions.StoreRefreshToken(context.Background(), "refresh-tok-2", "user-2", RefreshTokenTTL))

		result, err := svc.Refresh(context.Background(), "refresh-tok-2")

		s.Nil(result)
		s.Require().Error(err)
		s.Equal(apperr.ErrUnauthorized.Code, appErrCode(s.T(), err))
	})

	s.Run(testCaseSuccess_RefreshRotatesToken, func() {
		users := mocks.NewUserRepository(s.T())
		user := &model.User{ID: "user-3", Role: model.RoleUser, Status: model.StatusActive}
		users.On("GetByID", mock.Anything, "user-3").Return(user, nil)
		svc := s.newService(users)
		s.Require().NoError(s.sessions.StoreRefreshToken(context.Background(), "refresh-tok-3", "user-3", RefreshTokenTTL))

		result, err := svc.Refresh(context.Background(), "refresh-tok-3")

		s.Require().NoError(err)
		s.NotEmpty(result.AccessToken)
		s.NotEqual("refresh-tok-3", result.RefreshToken)

		// old token is single-use: replaying it must now fail.
		_, replayErr := s.sessions.ConsumeRefreshToken(context.Background(), "refresh-tok-3")
		s.Require().Error(replayErr)

		// new token is valid and points at the same user.
		gotUserID, err := s.sessions.ConsumeRefreshToken(context.Background(), result.RefreshToken)
		s.Require().NoError(err)
		s.Equal("user-3", gotUserID)
	})
}

func (s *AuthServiceTestSuite) TestLogout() {
	s.Run(testCaseSuccess_LogoutBlacklistsAndRevokes, func() {
		users := mocks.NewUserRepository(s.T())
		svc := s.newService(users)
		s.Require().NoError(s.sessions.StoreRefreshToken(context.Background(), "refresh-tok-1", "user-1", RefreshTokenTTL))

		err := svc.Logout(context.Background(), "jti-1", time.Now().Add(time.Minute), "refresh-tok-1")

		s.Require().NoError(err)
		blacklisted, err := s.sessions.IsBlacklisted(context.Background(), "jti-1")
		s.Require().NoError(err)
		s.True(blacklisted)
		_, consumeErr := s.sessions.ConsumeRefreshToken(context.Background(), "refresh-tok-1")
		s.Require().Error(consumeErr, "refresh token must have been revoked")
	})

	s.Run(testCaseSuccess_LogoutSkipsBlacklistOnZeroTTL, func() {
		users := mocks.NewUserRepository(s.T())
		svc := s.newService(users)

		err := svc.Logout(context.Background(), "jti-2", time.Now().Add(-time.Minute), "")

		s.Require().NoError(err)
		blacklisted, err := s.sessions.IsBlacklisted(context.Background(), "jti-2")
		s.Require().NoError(err)
		s.False(blacklisted, "already-expired access token must not be blacklisted")
	})

	s.Run(testCaseSuccess_LogoutSkipsRevokeOnEmptyToken, func() {
		users := mocks.NewUserRepository(s.T())
		svc := s.newService(users)

		err := svc.Logout(context.Background(), "jti-3", time.Now().Add(time.Minute), "")

		s.Require().NoError(err)
		blacklisted, err := s.sessions.IsBlacklisted(context.Background(), "jti-3")
		s.Require().NoError(err)
		s.True(blacklisted)
	})

	s.Run(testCaseError_LogoutBlacklistFailurePropagates, func() {
		users := mocks.NewUserRepository(s.T())
		// Force the underlying Redis SET to fail (dedicated, already-closed
		// Redis instance) while Logout attempts to blacklist the
		// (still-unexpired) token.
		svc := s.newServiceWithBrokenSessions(users)

		err := svc.Logout(context.Background(), "jti-4", time.Now().Add(time.Minute), "")

		s.Require().Error(err)
	})
}

func (s *AuthServiceTestSuite) TestCheckSession() {
	s.Run(testCaseError_CheckSessionBlacklistLookupFails, func() {
		users := mocks.NewUserRepository(s.T())
		svc := s.newServiceWithBrokenSessions(users)

		blacklisted, status, err := svc.CheckSession(context.Background(), "jti-1", "user-1")

		s.Require().Error(err)
		s.False(blacklisted)
		s.Empty(status)
	})

	s.Run(testCaseSuccess_CheckSessionFailClosedOnMissingUser, func() {
		users := mocks.NewUserRepository(s.T())
		users.On("GetByID", mock.Anything, "missing-user").Return(nil, gorm.ErrRecordNotFound)
		svc := s.newService(users)

		blacklisted, status, err := svc.CheckSession(context.Background(), "jti-2", "missing-user")

		s.Require().NoError(err)
		s.False(blacklisted)
		s.Equal(model.StatusBanned, status)
	})

	s.Run(testCaseSuccess_CheckSessionActiveUserNotBlacklisted, func() {
		users := mocks.NewUserRepository(s.T())
		users.On("GetByID", mock.Anything, "user-2").Return(&model.User{ID: "user-2", Status: model.StatusActive}, nil)
		svc := s.newService(users)

		blacklisted, status, err := svc.CheckSession(context.Background(), "jti-3", "user-2")

		s.Require().NoError(err)
		s.False(blacklisted)
		s.Equal(model.StatusActive, status)
	})

	s.Run(testCaseSuccess_CheckSessionBlacklistedActiveUser, func() {
		users := mocks.NewUserRepository(s.T())
		users.On("GetByID", mock.Anything, "user-3").Return(&model.User{ID: "user-3", Status: model.StatusActive}, nil)
		svc := s.newService(users)
		s.Require().NoError(s.sessions.BlacklistAccessToken(context.Background(), "jti-4", time.Minute))

		blacklisted, status, err := svc.CheckSession(context.Background(), "jti-4", "user-3")

		s.Require().NoError(err)
		s.True(blacklisted)
		s.Equal(model.StatusActive, status)
	})
}

// issueTokenPair and nonEmptyPtr are pure/near-pure helpers — tested
// directly without the full suite's mocked UserRepository.

func (s *AuthServiceTestSuite) TestIssueTokenPair() {
	svc := NewAuthService(nil, s.sessions, s.verifier, testJWTSecret)
	user := &model.User{ID: "user-1", Role: model.RoleOrganizer, Status: model.StatusActive}

	result, err := svc.issueTokenPair(context.Background(), user)

	s.Require().NoError(err)
	s.Same(user, result.User)
	s.Equal(int(authclaims.AccessTokenTTL.Seconds()), result.ExpiresIn)

	claims, err := authclaims.Parse(testJWTSecret, result.AccessToken)
	s.Require().NoError(err)
	s.Equal("user-1", claims.UserID())
	s.Equal(model.RoleOrganizer, claims.Role)

	gotUserID, err := s.sessions.ConsumeRefreshToken(context.Background(), result.RefreshToken)
	s.Require().NoError(err)
	s.Equal("user-1", gotUserID)
}

func TestNonEmptyPtr(t *testing.T) {
	t.Run("[Success] empty string maps to nil", func(t *testing.T) {
		got := nonEmptyPtr("")
		if got != nil {
			t.Fatalf("nonEmptyPtr(\"\") = %v, want nil", *got)
		}
	})

	t.Run("[Success] non-empty string maps to a pointer to itself", func(t *testing.T) {
		got := nonEmptyPtr("Nguyen Van A")
		if got == nil || *got != "Nguyen Van A" {
			t.Fatalf("nonEmptyPtr(\"Nguyen Van A\") = %v, want pointer to \"Nguyen Van A\"", got)
		}
	})
}
