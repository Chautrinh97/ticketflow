package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ticketflow/pkg/apperr"
	"ticketflow/pkg/authclaims"
	"ticketflow/services/identity-service/internal/firebase"
	"ticketflow/services/identity-service/internal/model"
	"ticketflow/services/identity-service/internal/repository"
)

// RefreshTokenTTL per docs/04-security/authentication.md ("refresh_token
// hạn 7 ngày").
const RefreshTokenTTL = 7 * 24 * time.Hour

type AuthService struct {
	users     UserRepository
	sessions  *repository.SessionRepository
	verifier  firebase.Verifier
	jwtSecret []byte
}

func NewAuthService(users UserRepository, sessions *repository.SessionRepository, verifier firebase.Verifier, jwtSecret []byte) *AuthService {
	return &AuthService{users: users, sessions: sessions, verifier: verifier, jwtSecret: jwtSecret}
}

type LoginResult struct {
	AccessToken  string
	ExpiresIn    int
	RefreshToken string
	User         *model.User
}

// Login exchanges a Firebase ID token for an internal token pair, per the
// numbered flow in docs/04-security/authentication.md. A user record is
// looked up by firebase_uid first, then by email (to link a seeded
// organizer/super_admin dev account on its first real login), then created
// fresh with role='user' if neither exists.
func (s *AuthService) Login(ctx context.Context, firebaseIDToken string) (*LoginResult, error) {
	verified, err := s.verifier.Verify(ctx, firebaseIDToken)
	if err != nil {
		return nil, apperr.WithMessage(apperr.ErrUnauthorized, "Firebase token không hợp lệ")
	}

	user, err := s.users.GetByFirebaseUID(ctx, verified.UID)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		user, err = s.users.GetByEmail(ctx, verified.Email)
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			user = &model.User{
				ID:          uuid.NewString(),
				FirebaseUID: &verified.UID,
				Email:       verified.Email,
				FullName:    nonEmptyPtr(verified.Name),
				Role:        model.RoleUser,
				Status:      model.StatusActive,
			}
			if err := s.users.Create(ctx, user); err != nil {
				return nil, err
			}
		case err != nil:
			return nil, err
		default:
			if err := s.users.LinkFirebaseUID(ctx, user.ID, verified.UID); err != nil {
				return nil, err
			}
			user.FirebaseUID = &verified.UID
		}
	case err != nil:
		return nil, err
	}

	if user.Status == model.StatusBanned {
		return nil, apperr.WithMessage(apperr.ErrUnauthorized, "Tài khoản đã bị khoá")
	}

	return s.issueTokenPair(ctx, user)
}

func (s *AuthService) issueTokenPair(ctx context.Context, user *model.User) (*LoginResult, error) {
	jti := uuid.NewString()
	accessToken, err := authclaims.Sign(s.jwtSecret, user.ID, user.Role, jti, authclaims.AccessTokenTTL)
	if err != nil {
		return nil, err
	}
	refreshToken := uuid.NewString() + uuid.NewString()
	if err := s.sessions.StoreRefreshToken(ctx, refreshToken, user.ID, RefreshTokenTTL); err != nil {
		return nil, err
	}
	return &LoginResult{
		AccessToken:  accessToken,
		ExpiresIn:    int(authclaims.AccessTokenTTL.Seconds()),
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// Refresh rotates the refresh token (old one is consumed/deleted by
// ConsumeRefreshToken regardless of outcome) and issues a fresh access
// token, per the mandatory rotation rule in authentication.md.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*LoginResult, error) {
	if refreshToken == "" {
		return nil, apperr.WithMessage(apperr.ErrUnauthorized, "Thiếu refresh token")
	}
	userID, err := s.sessions.ConsumeRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, apperr.WithMessage(apperr.ErrUnauthorized, "Refresh token không hợp lệ hoặc đã hết hạn")
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, apperr.WithMessage(apperr.ErrUnauthorized, "Tài khoản không tồn tại")
	}
	if user.Status == model.StatusBanned {
		return nil, apperr.WithMessage(apperr.ErrUnauthorized, "Tài khoản đã bị khoá")
	}
	return s.issueTokenPair(ctx, user)
}

// Logout blacklists the current access token's jti until its natural
// expiry and revokes the refresh token, per authentication.md.
func (s *AuthService) Logout(ctx context.Context, jti string, accessTokenExpiresAt time.Time, refreshToken string) error {
	if ttl := time.Until(accessTokenExpiresAt); ttl > 0 {
		if err := s.sessions.BlacklistAccessToken(ctx, jti, ttl); err != nil {
			return err
		}
	}
	if refreshToken != "" {
		_ = s.sessions.RevokeRefreshToken(ctx, refreshToken)
	}
	return nil
}

// CheckSession backs the identity.proto CheckSession RPC used by the API
// Gateway's auth middleware on every authenticated request: it combines
// the blacklist check with a live users.status read (fail closed —
// missing user is treated as banned/invalid) since only identity-service
// owns both stores.
func (s *AuthService) CheckSession(ctx context.Context, jti, userID string) (blacklisted bool, status string, err error) {
	blacklisted, err = s.sessions.IsBlacklisted(ctx, jti)
	if err != nil {
		return false, "", err
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return blacklisted, model.StatusBanned, nil
	}
	return blacklisted, user.Status, nil
}

func nonEmptyPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
