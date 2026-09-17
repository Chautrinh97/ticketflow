package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"ticketflow/pkg/apperr"
	"ticketflow/services/identity-service/internal/model"
	"ticketflow/services/identity-service/internal/service/mocks"
)

const (
	testCaseSuccess_GetByIDFound          = "[Success] existing user is returned as-is"
	testCaseError_GetByIDNotFoundMapped   = "[Error] gorm.ErrRecordNotFound maps to apperr.ErrNotFound"
	testCaseError_GetByIDGenericPropagate = "[Error] a non-not-found repository error propagates unmapped"

	testCaseSuccess_UpdateProfileBothNilNoChange  = "[Success] both fields nil leaves existing values untouched"
	testCaseSuccess_UpdateProfileOnlyFullNameSet  = "[Success] only FullName set merges FullName, keeps existing AvatarURL"
	testCaseSuccess_UpdateProfileOnlyAvatarURLSet = "[Success] only AvatarURL set merges AvatarURL, keeps existing FullName"
	testCaseSuccess_UpdateProfileBothFieldsSet    = "[Success] both fields set overwrite both existing values"
	testCaseError_UpdateProfileGetByIDNotFound    = "[Error] not-found from the initial GetByID propagates, Update is never attempted"
	testCaseError_UpdateProfileUpdateFails        = "[Error] repository Update failure propagates"
)

func strPtr(s string) *string { return &s }

type UserServiceTestSuite struct {
	suite.Suite
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) newService() (*UserService, *mocks.UserRepository) {
	users := mocks.NewUserRepository(s.T())
	return NewUserService(users), users
}

func (s *UserServiceTestSuite) TestGetByID() {
	s.Run(testCaseSuccess_GetByIDFound, func() {
		svc, users := s.newService()
		user := &model.User{ID: "user-1", Email: "a@example.com"}
		users.On("GetByID", mock.Anything, "user-1").Return(user, nil)

		got, err := svc.GetByID(context.Background(), "user-1")

		s.Require().NoError(err)
		s.Same(user, got)
	})

	s.Run(testCaseError_GetByIDNotFoundMapped, func() {
		svc, users := s.newService()
		users.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound)

		got, err := svc.GetByID(context.Background(), "missing")

		s.Nil(got)
		s.Require().Error(err)
		var appErr *apperr.Error
		s.Require().True(errors.As(err, &appErr))
		s.Equal(apperr.ErrNotFound.Code, appErr.Code)
	})

	s.Run(testCaseError_GetByIDGenericPropagate, func() {
		svc, users := s.newService()
		dbErr := errors.New("connection reset")
		users.On("GetByID", mock.Anything, "id-1").Return(nil, dbErr)

		got, err := svc.GetByID(context.Background(), "id-1")

		s.Nil(got)
		s.Require().ErrorIs(err, dbErr)
	})
}

func (s *UserServiceTestSuite) TestUpdateProfile() {
	s.Run(testCaseSuccess_UpdateProfileBothNilNoChange, func() {
		svc, users := s.newService()
		user := &model.User{ID: "user-1", FullName: strPtr("Old Name"), AvatarURL: strPtr("old.png")}
		users.On("GetByID", mock.Anything, "user-1").Return(user, nil)
		users.On("Update", mock.Anything, user).Return(nil)

		got, err := svc.UpdateProfile(context.Background(), "user-1", UpdateProfileInput{})

		s.Require().NoError(err)
		s.Require().NotNil(got.FullName)
		s.Equal("Old Name", *got.FullName)
		s.Require().NotNil(got.AvatarURL)
		s.Equal("old.png", *got.AvatarURL)
	})

	s.Run(testCaseSuccess_UpdateProfileOnlyFullNameSet, func() {
		svc, users := s.newService()
		user := &model.User{ID: "user-2", FullName: strPtr("Old Name"), AvatarURL: strPtr("old.png")}
		users.On("GetByID", mock.Anything, "user-2").Return(user, nil)
		users.On("Update", mock.Anything, user).Return(nil)

		got, err := svc.UpdateProfile(context.Background(), "user-2", UpdateProfileInput{FullName: strPtr("New Name")})

		s.Require().NoError(err)
		s.Equal("New Name", *got.FullName)
		s.Equal("old.png", *got.AvatarURL)
	})

	s.Run(testCaseSuccess_UpdateProfileOnlyAvatarURLSet, func() {
		svc, users := s.newService()
		user := &model.User{ID: "user-3", FullName: strPtr("Old Name"), AvatarURL: strPtr("old.png")}
		users.On("GetByID", mock.Anything, "user-3").Return(user, nil)
		users.On("Update", mock.Anything, user).Return(nil)

		got, err := svc.UpdateProfile(context.Background(), "user-3", UpdateProfileInput{AvatarURL: strPtr("new.png")})

		s.Require().NoError(err)
		s.Equal("Old Name", *got.FullName)
		s.Equal("new.png", *got.AvatarURL)
	})

	s.Run(testCaseSuccess_UpdateProfileBothFieldsSet, func() {
		svc, users := s.newService()
		user := &model.User{ID: "user-4", FullName: strPtr("Old Name"), AvatarURL: strPtr("old.png")}
		users.On("GetByID", mock.Anything, "user-4").Return(user, nil)
		users.On("Update", mock.Anything, user).Return(nil)

		got, err := svc.UpdateProfile(context.Background(), "user-4", UpdateProfileInput{
			FullName:  strPtr("New Name"),
			AvatarURL: strPtr("new.png"),
		})

		s.Require().NoError(err)
		s.Equal("New Name", *got.FullName)
		s.Equal("new.png", *got.AvatarURL)
	})

	s.Run(testCaseError_UpdateProfileGetByIDNotFound, func() {
		svc, users := s.newService()
		// No "Update" expectation registered: if UpdateProfile called it
		// anyway despite the not-found short-circuit, the mock would panic
		// and fail this sub-test.
		users.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound)

		got, err := svc.UpdateProfile(context.Background(), "missing", UpdateProfileInput{FullName: strPtr("X")})

		s.Nil(got)
		s.Require().Error(err)
		var appErr *apperr.Error
		s.Require().True(errors.As(err, &appErr))
		s.Equal(apperr.ErrNotFound.Code, appErr.Code)
	})

	s.Run(testCaseError_UpdateProfileUpdateFails, func() {
		svc, users := s.newService()
		user := &model.User{ID: "user-5", FullName: strPtr("Old Name")}
		dbErr := errors.New("update failed")
		users.On("GetByID", mock.Anything, "user-5").Return(user, nil)
		users.On("Update", mock.Anything, user).Return(dbErr)

		got, err := svc.UpdateProfile(context.Background(), "user-5", UpdateProfileInput{FullName: strPtr("New Name")})

		s.Nil(got)
		s.Require().ErrorIs(err, dbErr)
	})
}
