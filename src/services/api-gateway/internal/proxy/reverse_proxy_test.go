package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNew(t *testing.T) {
	t.Run("[Error] invalid target URL panics instead of returning an error", func(t *testing.T) {
		// "%zz" is not a valid percent-encoding escape, which makes
		// url.Parse fail — New has no error return, so it panics.
		assert.Panics(t, func() {
			New("http://%zz-invalid-escape")
		})
	})

	t.Run("[Success] forwards method, path, headers and body unchanged to the target", func(t *testing.T) {
		var gotMethod, gotPath, gotAuth string
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			gotAuth = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))
		defer backend.Close()

		handler := New(backend.URL)

		// Drive the handler through a real net/http server rather than a
		// bare httptest.ResponseRecorder: httputil.ReverseProxy checks for
		// http.CloseNotifier support, and gin's responseWriter.CloseNotify
		// blindly type-asserts its underlying writer to that interface —
		// a bare ResponseRecorder doesn't implement it and the assertion
		// panics. A real http.Server's response writer does implement it.
		r := gin.New()
		r.Any("/*proxyPath", handler)
		gateway := httptest.NewServer(r)
		defer gateway.Close()

		req, err := http.NewRequest(http.MethodPost, gateway.URL+"/api/v1/bookings", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer test-token")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.JSONEq(t, `{"ok":true}`, string(body))

		assert.Equal(t, http.MethodPost, gotMethod)
		assert.Equal(t, "/api/v1/bookings", gotPath)
		assert.Equal(t, "Bearer test-token", gotAuth)
	})
}
