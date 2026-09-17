package service

import (
	"context"

	"ticketflow/services/identity-service/internal/model"
)

// UserRepository is a narrow interface over *repository.UserRepository
// covering exactly the methods called from this package (by AuthService and
// UserService combined), extracted so unit tests can mock it via mockery
// instead of hitting a real Postgres — per docs/01-architecture/backend-conventions.md's
// "Unit test có mock" convention. *repository.UserRepository already
// implements this interface structurally, so cmd/main.go's wiring keeps
// compiling unchanged.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, u *model.User) error
	LinkFirebaseUID(ctx context.Context, id, firebaseUID string) error
	Update(ctx context.Context, u *model.User) error
}
