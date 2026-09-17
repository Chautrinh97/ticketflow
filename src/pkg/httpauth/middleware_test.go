package httpauth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ticketflow/pkg/authclaims"
)

func init() {
	gin.SetMode(gin.TestMode)
}

var testSecret = []byte("test-secret-key-for-unit-tests")

func newTestContext(method, path string, header http.Header) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, path, nil)
	if header != nil {
		req.Header = header
	}
	c.Request = req
	return c, w
}

func TestRequireAuth(t *testing.T) {
	t.Run("[Error] missing Authorization header", func(t *testing.T) {
		c, w := newTestContext(http.MethodGet, "/", nil)
		RequireAuth(testSecret)(c)
		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("[Error] header without Bearer prefix", func(t *testing.T) {
		h := http.Header{}
		h.Set("Authorization", "Basic abc123")
		c, w := newTestContext(http.MethodGet, "/", h)
		RequireAuth(testSecret)(c)
		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("[Error] invalid/expired token", func(t *testing.T) {
		h := http.Header{}
		h.Set("Authorization", "Bearer not-a-real-token")
		c, w := newTestContext(http.MethodGet, "/", h)
		RequireAuth(testSecret)(c)
		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("[Success] valid token sets context values and calls Next", func(t *testing.T) {
		tokenStr, err := authclaims.Sign(testSecret, "user-1", "organizer", "jti-1", time.Hour)
		require.NoError(t, err)
		h := http.Header{}
		h.Set("Authorization", "Bearer "+tokenStr)
		c, w := newTestContext(http.MethodGet, "/", h)

		RequireAuth(testSecret)(c)

		assert.False(t, c.IsAborted())
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
		assert.Equal(t, "user-1", UserID(c))
		assert.Equal(t, "organizer", Role(c))
		assert.Equal(t, "jti-1", JTI(c))
		assert.WithinDuration(t, time.Now().Add(time.Hour), ExpiresAt(c), 5*time.Second)
	})
}

func TestContextGetters_AbsentOrWrongType(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	assert.Equal(t, "", UserID(c))
	assert.Equal(t, "", Role(c))
	assert.Equal(t, "", JTI(c))
	assert.True(t, ExpiresAt(c).IsZero())

	c.Set(ContextUserID, 12345) // wrong type on purpose
	c.Set(ContextExpiresAt, "not-a-time")
	assert.Equal(t, "", UserID(c))
	assert.True(t, ExpiresAt(c).IsZero())
}

func TestRequireRole(t *testing.T) {
	t.Run("[Success] role in allowed set calls Next", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(ContextRole, "organizer")

		RequireRole("organizer", "super_admin")(c)
		assert.False(t, c.IsAborted())
	})

	t.Run("[Error] role not in allowed set aborts 403", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(ContextRole, "user")

		RequireRole("organizer", "super_admin")(c)
		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("[Error] empty/missing role aborts 403", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		RequireRole("organizer")(c)
		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRequireOwnership(t *testing.T) {
	t.Run("[Success] super_admin bypasses lookup entirely", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(ContextRole, "super_admin")
		lookupCalled := false

		RequireOwnership(func(c *gin.Context) (string, error) {
			lookupCalled = true
			return "", nil
		})(c)

		assert.False(t, c.IsAborted())
		assert.False(t, lookupCalled, "resourceLookup must not be called for super_admin")
	})

	t.Run("[Error] lookup failure returns 404, not 403, to avoid leaking existence", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(ContextRole, "user")
		c.Set(ContextUserID, "user-1")

		RequireOwnership(func(c *gin.Context) (string, error) {
			return "", errors.New("not found in db")
		})(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("[Error] owner mismatch returns 403", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(ContextRole, "user")
		c.Set(ContextUserID, "user-1")

		RequireOwnership(func(c *gin.Context) (string, error) {
			return "someone-else", nil
		})(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("[Success] owner match calls Next", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(ContextRole, "user")
		c.Set(ContextUserID, "user-1")

		RequireOwnership(func(c *gin.Context) (string, error) {
			return "user-1", nil
		})(c)

		assert.False(t, c.IsAborted())
	})
}
