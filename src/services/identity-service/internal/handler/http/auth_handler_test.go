package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/httpauth"
	"ticketflow/services/identity-service/internal/handler/http/mocks"
	"ticketflow/services/identity-service/internal/model"
	"ticketflow/services/identity-service/internal/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const (
	testCaseError_LoginInvalidBody         = "[Error] invalid JSON body returns validation error"
	testCaseError_LoginServiceFails        = "[Error] service Login error is passed to apperr.Respond"
	testCaseSuccess_LoginSetsCookieAndBody = "[Success] valid login sets refresh cookie and returns access token + user"

	testCaseSuccess_RefreshRotatesCookie = "[Success] refresh reads cookie, rotates it, returns new access token"
	testCaseError_RefreshServiceFails    = "[Error] service Refresh error is passed to apperr.Respond"

	testCaseSuccess_LogoutClearsCookie = "[Success] logout clears refresh cookie and returns 204"
	testCaseError_LogoutServiceFails   = "[Error] service Logout error is passed to apperr.Respond"
)

func newJSONTestContext(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, w
}

type AuthHandlerTestSuite struct {
	suite.Suite
}

func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func (s *AuthHandlerTestSuite) newHandler(cookieSecure bool) (*AuthHandler, *mocks.AuthService) {
	authSvc := mocks.NewAuthService(s.T())
	return NewAuthHandler(authSvc, cookieSecure), authSvc
}

func (s *AuthHandlerTestSuite) TestLogin() {
	s.Run(testCaseError_LoginInvalidBody, func() {
		h, _ := s.newHandler(false)
		c, w := newJSONTestContext(http.MethodPost, "/auth/login", []byte(`not-json`))

		h.Login(c)

		s.Equal(http.StatusBadRequest, w.Code)
	})

	s.Run(testCaseError_LoginServiceFails, func() {
		h, authSvc := s.newHandler(false)
		authSvc.On("Login", mock.Anything, "bad-token").
			Return(nil, apperr.WithMessage(apperr.ErrUnauthorized, "Firebase token không hợp lệ"))

		body, _ := json.Marshal(map[string]string{"firebase_id_token": "bad-token"})
		c, w := newJSONTestContext(http.MethodPost, "/auth/login", body)

		h.Login(c)

		s.Equal(http.StatusUnauthorized, w.Code)
	})

	s.Run(testCaseSuccess_LoginSetsCookieAndBody, func() {
		h, authSvc := s.newHandler(true)
		user := &model.User{ID: "user-1", Email: "a@example.com", Role: model.RoleUser, Status: model.StatusActive}
		authSvc.On("Login", mock.Anything, "good-token").Return(&service.LoginResult{
			AccessToken:  "access-1",
			ExpiresIn:    900,
			RefreshToken: "refresh-1",
			User:         user,
		}, nil)

		body, _ := json.Marshal(map[string]string{"firebase_id_token": "good-token"})
		c, w := newJSONTestContext(http.MethodPost, "/auth/login", body)

		h.Login(c)

		s.Equal(http.StatusOK, w.Code)
		s.Contains(w.Body.String(), "access-1")
		s.Contains(w.Body.String(), "a@example.com")

		cookies := w.Result().Cookies()
		s.Require().Len(cookies, 1)
		s.Equal(refreshCookieName, cookies[0].Name)
		s.Equal("refresh-1", cookies[0].Value)
		s.True(cookies[0].Secure)
		s.True(cookies[0].HttpOnly)
		s.Equal(http.SameSiteStrictMode, cookies[0].SameSite)
	})
}

func (s *AuthHandlerTestSuite) TestRefresh() {
	s.Run(testCaseError_RefreshServiceFails, func() {
		h, authSvc := s.newHandler(false)
		authSvc.On("Refresh", mock.Anything, "old-refresh").
			Return(nil, apperr.WithMessage(apperr.ErrUnauthorized, "Refresh token không hợp lệ hoặc đã hết hạn"))

		c, w := newJSONTestContext(http.MethodPost, "/auth/refresh", nil)
		c.Request.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "old-refresh"})

		h.Refresh(c)

		s.Equal(http.StatusUnauthorized, w.Code)
	})

	s.Run(testCaseSuccess_RefreshRotatesCookie, func() {
		h, authSvc := s.newHandler(false)
		authSvc.On("Refresh", mock.Anything, "old-refresh").Return(&service.LoginResult{
			AccessToken:  "access-2",
			ExpiresIn:    900,
			RefreshToken: "new-refresh",
		}, nil)

		c, w := newJSONTestContext(http.MethodPost, "/auth/refresh", nil)
		c.Request.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "old-refresh"})

		h.Refresh(c)

		s.Equal(http.StatusOK, w.Code)
		s.Contains(w.Body.String(), "access-2")

		cookies := w.Result().Cookies()
		s.Require().Len(cookies, 1)
		s.Equal("new-refresh", cookies[0].Value)
	})
}

func (s *AuthHandlerTestSuite) TestLogout() {
	s.Run(testCaseError_LogoutServiceFails, func() {
		h, authSvc := s.newHandler(false)
		authSvc.On("Logout", mock.Anything, "jti-1", mock.AnythingOfType("time.Time"), "old-refresh").
			Return(apperr.ErrInternal)

		c, w := newJSONTestContext(http.MethodPost, "/auth/logout", nil)
		c.Request.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "old-refresh"})
		c.Set(httpauth.ContextJTI, "jti-1")
		c.Set(httpauth.ContextExpiresAt, time.Now().Add(time.Minute))

		h.Logout(c)

		s.Equal(http.StatusInternalServerError, w.Code)
	})

	s.Run(testCaseSuccess_LogoutClearsCookie, func() {
		h, authSvc := s.newHandler(false)
		authSvc.On("Logout", mock.Anything, "jti-2", mock.AnythingOfType("time.Time"), "old-refresh").
			Return(nil)

		c, w := newJSONTestContext(http.MethodPost, "/auth/logout", nil)
		c.Request.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "old-refresh"})
		c.Set(httpauth.ContextJTI, "jti-2")
		c.Set(httpauth.ContextExpiresAt, time.Now().Add(time.Minute))

		h.Logout(c)
		// Logout's success path only calls c.Status (no body written), so
		// gin's lazy header buffering never auto-flushes to the recorder
		// here — unlike a real request, we call the handler directly and
		// skip the engine's post-middleware-chain WriteHeaderNow(). Force
		// it so w.Code reflects what a real client would actually receive.
		c.Writer.WriteHeaderNow()

		s.Equal(http.StatusNoContent, w.Code)
		cookies := w.Result().Cookies()
		s.Require().Len(cookies, 1)
		s.Equal(refreshCookieName, cookies[0].Name)
		s.Equal("", cookies[0].Value)
		s.Less(cookies[0].MaxAge, 0)
	})
}
