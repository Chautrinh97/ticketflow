package authclaims

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSecret = []byte("test-secret-key-for-unit-tests")

func TestSignAndParse_Success(t *testing.T) {
	tokenStr, err := Sign(testSecret, "user-1", "organizer", "jti-1", time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	claims, err := Parse(testSecret, tokenStr)
	require.NoError(t, err)
	assert.Equal(t, "user-1", claims.UserID())
	assert.Equal(t, "organizer", claims.Role)
	assert.Equal(t, "jti-1", claims.JTI)
	require.NotNil(t, claims.ExpiresAt)
	assert.WithinDuration(t, time.Now().Add(time.Hour), claims.ExpiresAt.Time, 5*time.Second)
}

func TestParse_WrongSecret(t *testing.T) {
	tokenStr, err := Sign(testSecret, "user-1", "user", "jti-1", time.Hour)
	require.NoError(t, err)

	_, err = Parse([]byte("a-completely-different-secret"), tokenStr)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestParse_Expired(t *testing.T) {
	tokenStr, err := Sign(testSecret, "user-1", "user", "jti-1", -time.Minute)
	require.NoError(t, err)

	_, err = Parse(testSecret, tokenStr)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestParse_MalformedToken(t *testing.T) {
	_, err := Parse(testSecret, "not-a-jwt-at-all")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestParse_RejectsNonHMACAlgorithm(t *testing.T) {
	// Alg-confusion guard: Parse's keyfunc type-asserts on *jwt.SigningMethodHMAC,
	// so a token signed with "none" must be rejected even though the JWT
	// library would otherwise happily parse an unsigned token.
	unsigned := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{
		Role: "super_admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "attacker",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	tokenStr, err := unsigned.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = Parse(testSecret, tokenStr)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestClaims_UserID(t *testing.T) {
	c := &Claims{RegisteredClaims: jwt.RegisteredClaims{Subject: "u-42"}}
	assert.Equal(t, "u-42", c.UserID())
}
