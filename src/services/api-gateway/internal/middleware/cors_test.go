package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const allowedOriginForTest = "http://localhost:3000"

func newCORSTestContext(method, origin string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, "/api/v1/events", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	c.Request = req
	return c, w
}

func TestCORS(t *testing.T) {
	t.Run("[Success] matching origin sets CORS headers and calls Next", func(t *testing.T) {
		c, w := newCORSTestContext(http.MethodGet, allowedOriginForTest)

		CORS(allowedOriginForTest)(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, allowedOriginForTest, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Authorization, Content-Type", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "Origin", w.Header().Get("Vary"))
	})

	t.Run("[Success] OPTIONS preflight from matching origin is short-circuited with 204 and CORS headers", func(t *testing.T) {
		c, w := newCORSTestContext(http.MethodOptions, allowedOriginForTest)

		CORS(allowedOriginForTest)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, allowedOriginForTest, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("[Success] OPTIONS preflight from non-matching origin still gets 204 but no CORS headers", func(t *testing.T) {
		c, w := newCORSTestContext(http.MethodOptions, "http://evil.example.com")

		CORS(allowedOriginForTest)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("[Error] disallowed origin gets no CORS headers but request still proceeds", func(t *testing.T) {
		c, w := newCORSTestContext(http.MethodGet, "http://evil.example.com")

		CORS(allowedOriginForTest)(c)

		assert.False(t, c.IsAborted())
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Empty(t, w.Header().Get("Vary"))
	})

	t.Run("[Error] missing Origin header gets no CORS headers", func(t *testing.T) {
		c, w := newCORSTestContext(http.MethodGet, "")

		CORS(allowedOriginForTest)(c)

		assert.False(t, c.IsAborted())
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}
