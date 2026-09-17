package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoad is a pure-unit test (no DB/gRPC/Redis) per
// docs/01-architecture/backend-conventions.md's "Test logic thuần" carve-out
// — plain t.Run sub-tests, no suite/mock needed. t.Setenv auto-restores the
// previous value after each sub-test, so no manual save/restore is needed.
func TestLoad(t *testing.T) {
	t.Run("[Success] all env vars empty uses documented defaults", func(t *testing.T) {
		t.Setenv("HTTP_PORT", "")
		t.Setenv("DATABASE_URL", "")
		t.Setenv("JWT_SIGNING_KEY", "")
		t.Setenv("BOOKING_GRPC_ADDR", "")

		cfg := Load()

		assert.Equal(t, defaultHTTPPort, cfg.HTTPPort)
		assert.Equal(t, "postgres://ticketflow:ticketflow@localhost:5432/ticketflow?sslmode=disable", cfg.DatabaseURL)
		assert.Equal(t, "dev-secret-change-me", string(cfg.JWTSecret))
		assert.Equal(t, "localhost:"+defaultGRPCPort, cfg.BookingGRPCAddr)
	})

	t.Run("[Success] every env var set overrides its default", func(t *testing.T) {
		t.Setenv("HTTP_PORT", "9999")
		t.Setenv("DATABASE_URL", "postgres://custom:custom@db:5432/custom?sslmode=disable")
		t.Setenv("JWT_SIGNING_KEY", "super-secret")
		t.Setenv("BOOKING_GRPC_ADDR", "booking-service:9443")

		cfg := Load()

		assert.Equal(t, "9999", cfg.HTTPPort)
		assert.Equal(t, "postgres://custom:custom@db:5432/custom?sslmode=disable", cfg.DatabaseURL)
		assert.Equal(t, "super-secret", string(cfg.JWTSecret))
		assert.Equal(t, "booking-service:9443", cfg.BookingGRPCAddr)
	})

	t.Run("[Error] env var explicitly set to empty string still falls back to default", func(t *testing.T) {
		t.Setenv("HTTP_PORT", "")

		cfg := Load()

		assert.Equal(t, defaultHTTPPort, cfg.HTTPPort, "getenv treats empty string same as unset")
	})
}

func TestGetenv(t *testing.T) {
	t.Run("[Success] unset/empty key returns default", func(t *testing.T) {
		t.Setenv("TICKETFLOW_TEST_ONLY_KEY", "")
		assert.Equal(t, "fallback", getenv("TICKETFLOW_TEST_ONLY_KEY", "fallback"))
	})

	t.Run("[Success] set key returns its value", func(t *testing.T) {
		t.Setenv("TICKETFLOW_TEST_ONLY_KEY", "actual-value")
		assert.Equal(t, "actual-value", getenv("TICKETFLOW_TEST_ONLY_KEY", "fallback"))
	})
}
