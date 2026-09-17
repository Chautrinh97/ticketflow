package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// envVars lists every environment variable Load reads, used to snapshot and
// clear the environment around each test so tests never depend on (or leak
// into) whatever happens to be set in the process environment.
var envVars = []string{
	"HTTP_PORT",
	"JWT_SIGNING_KEY",
	"IDENTITY_GRPC_ADDR",
	"IDENTITY_HTTP_URL",
	"EVENT_HTTP_URL",
	"BOOKING_HTTP_URL",
	"PAYMENT_HTTP_URL",
	"ALLOWED_ORIGIN",
}

// withCleanEnv unsets every relevant env var, restoring the original values
// (or absence thereof) after the test via t.Cleanup.
func withCleanEnv(t *testing.T) {
	t.Helper()
	original := make(map[string]string, len(envVars))
	present := make(map[string]bool, len(envVars))
	for _, k := range envVars {
		v, ok := os.LookupEnv(k)
		present[k] = ok
		original[k] = v
		require.NoError(t, os.Unsetenv(k))
	}
	t.Cleanup(func() {
		for _, k := range envVars {
			if present[k] {
				_ = os.Setenv(k, original[k])
			} else {
				_ = os.Unsetenv(k)
			}
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("[Success] no env vars set falls back to defaults", func(t *testing.T) {
		withCleanEnv(t)

		cfg := Load()

		assert.Equal(t, defaultHTTPPort, cfg.HTTPPort)
		assert.Equal(t, []byte("dev-secret-change-me"), cfg.JWTSecret)
		assert.Equal(t, "localhost:"+defaultGRPCPort, cfg.IdentityGRPCAddr)
		assert.Equal(t, "http://localhost:"+defaultHTTPPort, cfg.IdentityHTTPURL)
		assert.Equal(t, "http://localhost:"+defaultHTTPPort, cfg.EventHTTPURL)
		assert.Equal(t, "http://localhost:"+defaultHTTPPort, cfg.BookingHTTPURL)
		assert.Equal(t, "http://localhost:"+defaultHTTPPort, cfg.PaymentHTTPURL)
		assert.Equal(t, "http://localhost:3000", cfg.AllowedOrigin)
	})

	t.Run("[Success] every env var set overrides its matching default", func(t *testing.T) {
		withCleanEnv(t)
		t.Setenv("HTTP_PORT", "9090")
		t.Setenv("JWT_SIGNING_KEY", "super-secret")
		t.Setenv("IDENTITY_GRPC_ADDR", "identity:9443")
		t.Setenv("IDENTITY_HTTP_URL", "http://identity:9000")
		t.Setenv("EVENT_HTTP_URL", "http://event:9001")
		t.Setenv("BOOKING_HTTP_URL", "http://booking:9002")
		t.Setenv("PAYMENT_HTTP_URL", "http://payment:9003")
		t.Setenv("ALLOWED_ORIGIN", "https://ticketflow.example.com")

		cfg := Load()

		assert.Equal(t, "9090", cfg.HTTPPort)
		assert.Equal(t, []byte("super-secret"), cfg.JWTSecret)
		assert.Equal(t, "identity:9443", cfg.IdentityGRPCAddr)
		assert.Equal(t, "http://identity:9000", cfg.IdentityHTTPURL)
		assert.Equal(t, "http://event:9001", cfg.EventHTTPURL)
		assert.Equal(t, "http://booking:9002", cfg.BookingHTTPURL)
		assert.Equal(t, "http://payment:9003", cfg.PaymentHTTPURL)
		assert.Equal(t, "https://ticketflow.example.com", cfg.AllowedOrigin)
	})

	t.Run("[Success] empty string env var is treated as unset and falls back to default", func(t *testing.T) {
		withCleanEnv(t)
		t.Setenv("HTTP_PORT", "")

		cfg := Load()

		assert.Equal(t, defaultHTTPPort, cfg.HTTPPort)
	})
}

func TestGetenv(t *testing.T) {
	t.Run("[Success] returns env value when set", func(t *testing.T) {
		t.Setenv("TICKETFLOW_TEST_GETENV_KEY", "custom-value")
		assert.Equal(t, "custom-value", getenv("TICKETFLOW_TEST_GETENV_KEY", "default-value"))
	})

	t.Run("[Success] returns default when unset", func(t *testing.T) {
		require.NoError(t, os.Unsetenv("TICKETFLOW_TEST_GETENV_KEY_UNSET"))
		assert.Equal(t, "default-value", getenv("TICKETFLOW_TEST_GETENV_KEY_UNSET", "default-value"))
	})

	t.Run("[Success] returns default when set to empty string", func(t *testing.T) {
		t.Setenv("TICKETFLOW_TEST_GETENV_KEY_EMPTY", "")
		assert.Equal(t, "default-value", getenv("TICKETFLOW_TEST_GETENV_KEY_EMPTY", "default-value"))
	})
}
