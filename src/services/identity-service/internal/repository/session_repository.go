// SessionRepository implements the two Redis-backed auth state pieces from
// docs/03-data/redis-keys.md: the access-token blacklist (`session:blacklist:{jti}`)
// and — a Phase 1 design decision documented in the implementation plan,
// since no refresh_tokens Postgres table exists — refresh-token rotation
// state (`refresh:{jti}`, opaque token used directly as the Redis key).
package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	blacklistPrefix = "session:blacklist:"
	refreshPrefix   = "refresh:"
)

type SessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) *SessionRepository {
	return &SessionRepository{rdb: rdb}
}

func (r *SessionRepository) BlacklistAccessToken(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return r.rdb.Set(ctx, blacklistPrefix+jti, "1", ttl).Err()
}

func (r *SessionRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := r.rdb.Exists(ctx, blacklistPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *SessionRepository) StoreRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return r.rdb.Set(ctx, refreshPrefix+token, userID, ttl).Err()
}

// ConsumeRefreshToken looks up and immediately deletes the token (rotation):
// a refresh token can only ever be used once. Returns redis.Nil if the
// token is unknown, expired, or was already rotated/revoked.
func (r *SessionRepository) ConsumeRefreshToken(ctx context.Context, token string) (string, error) {
	userID, err := r.rdb.Get(ctx, refreshPrefix+token).Result()
	if err != nil {
		return "", err
	}
	r.rdb.Del(ctx, refreshPrefix+token)
	return userID, nil
}

func (r *SessionRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	return r.rdb.Del(ctx, refreshPrefix+token).Err()
}
