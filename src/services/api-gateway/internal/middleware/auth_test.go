package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"ticketflow/pkg/authclaims"
	"ticketflow/services/api-gateway/internal/middleware/mocks"
)

const (
	testCaseSuccess_NoBearerPrefix_PassesThrough = "[Success] no Bearer prefix passes through without calling CheckSession"
	testCaseError_InvalidJWT_Returns401          = "[Error] invalid/expired JWT returns 401 without calling CheckSession"
	testCaseError_CheckSessionError_Returns503   = "[Error] CheckSession transport error returns 503"
	testCaseError_Blacklisted_Returns401         = "[Error] blacklisted jti returns 401"
	testCaseError_Banned_Returns401              = "[Error] banned account status returns 401"
	testCaseSuccess_ValidActiveSession_CallsNext = "[Success] valid, non-blacklisted, active session calls Next"
)

var authTestSecret = []byte("test-secret-key-for-unit-tests")

type OptionalSessionCheckTestSuite struct {
	suite.Suite
}

func TestOptionalSessionCheckTestSuite(t *testing.T) {
	suite.Run(t, new(OptionalSessionCheckTestSuite))
}

// newAuthTestContext builds a gin.Context for a GET request, optionally with
// the given raw Authorization header value (pass "" to omit the header).
func (s *OptionalSessionCheckTestSuite) newAuthTestContext(authHeader string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	c.Request = req
	return c, w
}

func (s *OptionalSessionCheckTestSuite) signToken(userID, role, jti string, ttl time.Duration) string {
	tok, err := authclaims.Sign(authTestSecret, userID, role, jti, ttl)
	s.Require().NoError(err)
	return tok
}

func (s *OptionalSessionCheckTestSuite) TestOptionalSessionCheck() {
	s.Run(testCaseSuccess_NoBearerPrefix_PassesThrough, func() {
		mockCl := mocks.NewSessionChecker(s.T())
		c, w := s.newAuthTestContext("")

		OptionalSessionCheck(authTestSecret, mockCl)(c)

		s.Assert().False(c.IsAborted())
		s.Assert().NotEqual(http.StatusUnauthorized, w.Code)
		mockCl.AssertNotCalled(s.T(), "CheckSession", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_InvalidJWT_Returns401, func() {
		mockCl := mocks.NewSessionChecker(s.T())
		c, w := s.newAuthTestContext("Bearer not-a-real-jwt")

		OptionalSessionCheck(authTestSecret, mockCl)(c)

		s.Assert().True(c.IsAborted())
		s.Assert().Equal(http.StatusUnauthorized, w.Code)
		mockCl.AssertNotCalled(s.T(), "CheckSession", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_InvalidJWT_Returns401+" (expired)", func() {
		mockCl := mocks.NewSessionChecker(s.T())
		expiredTok := s.signToken("user-1", "user", "jti-1", -time.Minute)
		c, w := s.newAuthTestContext("Bearer " + expiredTok)

		OptionalSessionCheck(authTestSecret, mockCl)(c)

		s.Assert().True(c.IsAborted())
		s.Assert().Equal(http.StatusUnauthorized, w.Code)
		mockCl.AssertNotCalled(s.T(), "CheckSession", mock.Anything, mock.Anything, mock.Anything)
	})

	s.Run(testCaseError_CheckSessionError_Returns503, func() {
		tok := s.signToken("user-1", "user", "jti-1", time.Hour)
		mockCl := mocks.NewSessionChecker(s.T())
		mockCl.On("CheckSession", mock.Anything, "jti-1", "user-1").Return(false, "", errors.New("grpc dial failed"))
		c, w := s.newAuthTestContext("Bearer " + tok)

		OptionalSessionCheck(authTestSecret, mockCl)(c)

		s.Assert().True(c.IsAborted())
		s.Assert().Equal(http.StatusServiceUnavailable, w.Code)
	})

	s.Run(testCaseError_Blacklisted_Returns401, func() {
		tok := s.signToken("user-1", "user", "jti-1", time.Hour)
		mockCl := mocks.NewSessionChecker(s.T())
		mockCl.On("CheckSession", mock.Anything, "jti-1", "user-1").Return(true, "active", nil)
		c, w := s.newAuthTestContext("Bearer " + tok)

		OptionalSessionCheck(authTestSecret, mockCl)(c)

		s.Assert().True(c.IsAborted())
		s.Assert().Equal(http.StatusUnauthorized, w.Code)
	})

	s.Run(testCaseError_Banned_Returns401, func() {
		tok := s.signToken("user-1", "user", "jti-1", time.Hour)
		mockCl := mocks.NewSessionChecker(s.T())
		mockCl.On("CheckSession", mock.Anything, "jti-1", "user-1").Return(false, "banned", nil)
		c, w := s.newAuthTestContext("Bearer " + tok)

		OptionalSessionCheck(authTestSecret, mockCl)(c)

		s.Assert().True(c.IsAborted())
		s.Assert().Equal(http.StatusUnauthorized, w.Code)
	})

	s.Run(testCaseSuccess_ValidActiveSession_CallsNext, func() {
		tok := s.signToken("user-1", "user", "jti-1", time.Hour)
		mockCl := mocks.NewSessionChecker(s.T())
		mockCl.On("CheckSession", mock.Anything, "jti-1", "user-1").Return(false, "active", nil)
		c, w := s.newAuthTestContext("Bearer " + tok)

		OptionalSessionCheck(authTestSecret, mockCl)(c)

		s.Assert().False(c.IsAborted())
		s.Assert().NotEqual(http.StatusUnauthorized, w.Code)
		s.Assert().NotEqual(http.StatusServiceUnavailable, w.Code)
	})
}
