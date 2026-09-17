package http

import (
	"context"
	"time"

	"ticketflow/services/identity-service/internal/model"
	"ticketflow/services/identity-service/internal/service"
)

// AuthService is a narrow interface over *service.AuthService, extracted so
// handler tests can mock the service layer via mockery instead of exercising
// the real Firebase/Postgres/Redis-backed implementation — per
// docs/01-architecture/backend-conventions.md's "Unit test có mock"
// convention. *service.AuthService already implements this interface
// structurally, so cmd/main.go's wiring keeps compiling unchanged.
type AuthService interface {
	Login(ctx context.Context, firebaseIDToken string) (*service.LoginResult, error)
	Refresh(ctx context.Context, refreshToken string) (*service.LoginResult, error)
	Logout(ctx context.Context, jti string, accessTokenExpiresAt time.Time, refreshToken string) error
}

// UserService is a narrow interface over *service.UserService for the same
// reason as AuthService above.
type UserService interface {
	GetByID(ctx context.Context, id string) (*model.User, error)
	UpdateProfile(ctx context.Context, id string, in service.UpdateProfileInput) (*model.User, error)
}
