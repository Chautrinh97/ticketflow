package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/httpauth"
	"ticketflow/services/identity-service/internal/handler/http/mocks"
	"ticketflow/services/identity-service/internal/model"
	"ticketflow/services/identity-service/internal/service"
)

const (
	testCaseSuccess_GetMeReturnsUser = "[Success] returns current user's DTO"
	testCaseError_GetMeNotFound      = "[Error] not-found error passed to apperr.Respond"

	testCaseError_UpdateMeInvalidBody     = "[Error] invalid JSON body returns validation error"
	testCaseSuccess_UpdateMeUpdatesFields = "[Success] valid body updates profile and returns DTO"
	testCaseError_UpdateMeServiceFails    = "[Error] service error passed to apperr.Respond"
)

func newAuthedTestContext(method, path string, body []byte, userID string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set(httpauth.ContextUserID, userID)
	return c, w
}

type UserHandlerTestSuite struct {
	suite.Suite
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

func (s *UserHandlerTestSuite) newHandler() (*UserHandler, *mocks.UserService) {
	users := mocks.NewUserService(s.T())
	return NewUserHandler(users), users
}

func (s *UserHandlerTestSuite) TestGetMe() {
	s.Run(testCaseSuccess_GetMeReturnsUser, func() {
		h, users := s.newHandler()
		user := &model.User{ID: "user-1", Email: "a@example.com", Role: model.RoleUser, Status: model.StatusActive}
		users.On("GetByID", mock.Anything, "user-1").Return(user, nil)

		c, w := newAuthedTestContext(http.MethodGet, "/users/me", nil, "user-1")
		h.GetMe(c)

		s.Equal(http.StatusOK, w.Code)
		s.Contains(w.Body.String(), "a@example.com")
	})

	s.Run(testCaseError_GetMeNotFound, func() {
		h, users := s.newHandler()
		users.On("GetByID", mock.Anything, "missing").Return(nil, apperr.ErrNotFound)

		c, w := newAuthedTestContext(http.MethodGet, "/users/me", nil, "missing")
		h.GetMe(c)

		s.Equal(http.StatusNotFound, w.Code)
	})
}

func (s *UserHandlerTestSuite) TestUpdateMe() {
	s.Run(testCaseError_UpdateMeInvalidBody, func() {
		h, _ := s.newHandler()
		c, w := newAuthedTestContext(http.MethodPatch, "/users/me", []byte(`not-json`), "user-1")

		h.UpdateMe(c)

		s.Equal(http.StatusBadRequest, w.Code)
	})

	s.Run(testCaseSuccess_UpdateMeUpdatesFields, func() {
		h, users := s.newHandler()
		fullName := "New Name"
		updated := &model.User{ID: "user-1", Email: "a@example.com", FullName: &fullName, Role: model.RoleUser, Status: model.StatusActive}
		users.On("UpdateProfile", mock.Anything, "user-1", mock.MatchedBy(func(in service.UpdateProfileInput) bool {
			return in.FullName != nil && *in.FullName == fullName && in.AvatarURL == nil
		})).Return(updated, nil)

		body, _ := json.Marshal(map[string]string{"full_name": fullName})
		c, w := newAuthedTestContext(http.MethodPatch, "/users/me", body, "user-1")

		h.UpdateMe(c)

		s.Equal(http.StatusOK, w.Code)
		s.Contains(w.Body.String(), "New Name")
	})

	s.Run(testCaseError_UpdateMeServiceFails, func() {
		h, users := s.newHandler()
		users.On("UpdateProfile", mock.Anything, "user-1", mock.Anything).Return(nil, apperr.ErrInternal)

		body, _ := json.Marshal(map[string]string{"full_name": "X"})
		c, w := newAuthedTestContext(http.MethodPatch, "/users/me", body, "user-1")

		h.UpdateMe(c)

		s.Equal(http.StatusInternalServerError, w.Code)
	})
}
