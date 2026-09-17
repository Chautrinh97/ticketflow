package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ticketflow/pkg/apperr"
	"ticketflow/services/identity-service/internal/model"
)

type UserService struct {
	users UserRepository
}

func NewUserService(users UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

type UpdateProfileInput struct {
	FullName  *string
	AvatarURL *string
}

func (s *UserService) UpdateProfile(ctx context.Context, id string, in UpdateProfileInput) (*model.User, error) {
	user, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.FullName != nil {
		user.FullName = in.FullName
	}
	if in.AvatarURL != nil {
		user.AvatarURL = in.AvatarURL
	}
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
