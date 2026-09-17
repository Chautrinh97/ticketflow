package router

import (
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

func noopMiddleware(c *gin.Context) { c.Next() }

// newFakeBackend returns a server that echoes its own name in a response
// header, so a test can tell which of the 4 configured backends actually
// received a proxied request.
func newFakeBackend(t *testing.T, name string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", name)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// newTestGateway wraps the router's *gin.Engine in a real net/http server.
// This (rather than httptest.NewRecorder + Engine.ServeHTTP directly)
// matters because the routes proxy through httputil.ReverseProxy, which
// probes the ResponseWriter for http.CloseNotifier support — gin's
// responseWriter.CloseNotify blindly type-asserts its underlying writer to
// that interface, and a bare httptest.ResponseRecorder doesn't implement
// it, so that path panics under a fake recorder. A real http.Server's
// response writer does implement it.
func newTestGateway(t *testing.T, r *gin.Engine) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func TestNew_RoutesEachPathToItsOwningBackend(t *testing.T) {
	identity := newFakeBackend(t, "identity")
	event := newFakeBackend(t, "event")
	booking := newFakeBackend(t, "booking")
	payment := newFakeBackend(t, "payment")

	r := New(noopMiddleware, noopMiddleware, Targets{
		Identity: identity.URL,
		Event:    event.URL,
		Booking:  booking.URL,
		Payment:  payment.URL,
	})
	gateway := newTestGateway(t, r)

	cases := []struct {
		method string
		path   string
		want   string
	}{
		{http.MethodPost, "/api/v1/auth/login", "identity"},
		{http.MethodPost, "/api/v1/auth/refresh", "identity"},
		{http.MethodPost, "/api/v1/auth/logout", "identity"},
		{http.MethodGet, "/api/v1/users/me", "identity"},
		{http.MethodGet, "/api/v1/users/me/bookings", "booking"},
		{http.MethodGet, "/api/v1/events", "event"},
		{http.MethodGet, "/api/v1/events/some-slug", "event"},
		{http.MethodPost, "/api/v1/organizer/events", "event"},
		{http.MethodGet, "/api/v1/organizer/events/evt-1", "event"},
		{http.MethodPost, "/api/v1/organizer/events/evt-1/publish", "event"},
		{http.MethodPost, "/api/v1/organizer/events/evt-1/ticket-types", "event"},
		{http.MethodPost, "/api/v1/bookings", "booking"},
		{http.MethodGet, "/api/v1/bookings/booking-1", "booking"},
		{http.MethodPost, "/api/v1/payments/order-1/checkout", "payment"},
		{http.MethodPost, "/api/v1/payments/webhook", "payment"},
	}

	for _, tc := range cases {
		t.Run("[Success] "+tc.method+" "+tc.path+" -> "+tc.want, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, gateway.URL+tc.path, nil)
			require.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tc.want, resp.Header.Get("X-Backend"))
		})
	}
}

func TestNew_RouteNotInPhase1TrimReturns404(t *testing.T) {
	r := New(noopMiddleware, noopMiddleware, Targets{
		Identity: "http://127.0.0.1:1",
		Event:    "http://127.0.0.1:1",
		Booking:  "http://127.0.0.1:1",
		Payment:  "http://127.0.0.1:1",
	})
	gateway := newTestGateway(t, r)

	resp, err := http.Get(gateway.URL + "/api/v1/search")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
