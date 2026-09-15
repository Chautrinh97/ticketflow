// Package authclaims defines the internal JWT claim shape shared by every
// service: identity-service signs tokens with it, every other service and
// the API gateway locally verify signature+expiry with it.
// Claim set per docs/04-security/authentication.md: sub, role, jti, exp.
package authclaims

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AccessTokenTTL = 15 * time.Minute

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	Role string `json:"role"`
	JTI  string `json:"jti"`
	jwt.RegisteredClaims
}

func (c *Claims) UserID() string { return c.Subject }

// Sign issues a new HS256 access token for userID/role, with a fresh jti.
func Sign(secret []byte, userID, role, jti string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: role,
		JTI:  jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// Parse verifies signature and expiry only — it does not check revocation
// (blacklist) or live account status, which requires identity-service's
// CheckSession gRPC call (done by the API gateway's auth middleware).
func Parse(secret []byte, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
