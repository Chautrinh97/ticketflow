package repository

import (
	"context"

	"gorm.io/gorm"

	"ticketflow/services/identity-service/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, "firebase_uid = ?", firebaseUID).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepository) LinkFirebaseUID(ctx context.Context, id, firebaseUID string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Update("firebase_uid", firebaseUID).Error
}

func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}
