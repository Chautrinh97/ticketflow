package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Pure-unit test: Load()/getenv only touch process env vars — plain
// testify/assert, no suite/mock needed, per backend-conventions.md's
// "Test logic thuần" carve-out.

const (
	testCaseSuccess_DefaultsWhenUnset      = "[Success] returns documented defaults when no env vars are set"
	testCaseSuccess_OverridesWhenSet       = "[Success] returns env var values when they are set"
	testCaseSuccess_CookieSecureTrueString = "[Success] COOKIE_SECURE=true parses to true"
	testCaseSuccess_CookieSecureOtherValue = "[Success] any non-'true' COOKIE_SECURE value parses to false"
)

// allConfigEnvVars lists every env var Load() reads, so tests can unset all
// of them deterministically regardless of what the host environment has set.
var allConfigEnvVars = []string{
	"HTTP_PORT", "GRPC_PORT", "DATABASE_URL", "REDIS_ADDR",
	"JWT_SIGNING_KEY", "AUTH_FIREBASE_MODE", "FIREBASE_CREDENTIALS_FILE", "COOKIE_SECURE",
}

// unsetEnv removes key for the duration of the test, restoring whatever
// value (or absence) it had before, via t.Cleanup.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	old, existed := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if existed {
			require.NoError(t, os.Setenv(key, old))
		}
	})
}

func unsetAllConfigEnvVars(t *testing.T) {
	t.Helper()
	for _, key := range allConfigEnvVars {
		unsetEnv(t, key)
	}
}

func TestLoad(t *testing.T) {
	t.Run(testCaseSuccess_DefaultsWhenUnset, func(t *testing.T) {
		unsetAllConfigEnvVars(t)

		cfg := Load()

		assert.Equal(t, defaultHTTPPort, cfg.HTTPPort)
		assert.Equal(t, defaultGRPCPort, cfg.GRPCPort)
		assert.Equal(t, "postgres://ticketflow:ticketflow@localhost:5432/ticketflow?sslmode=disable", cfg.DatabaseURL)
		assert.Equal(t, "localhost:6379", cfg.RedisAddr)
		assert.Equal(t, []byte("dev-secret-change-me"), cfg.JWTSecret)
		assert.Equal(t, "mock", cfg.FirebaseMode)
		assert.Equal(t, "", cfg.FirebaseCredentialsFile)
		assert.False(t, cfg.CookieSecure)
	})

	t.Run(testCaseSuccess_OverridesWhenSet, func(t *testing.T) {
		t.Setenv("HTTP_PORT", "8080")
		t.Setenv("GRPC_PORT", "8443")
		t.Setenv("DATABASE_URL", "postgres://custom/db")
		t.Setenv("REDIS_ADDR", "redis-host:6380")
		t.Setenv("JWT_SIGNING_KEY", "super-secret")
		t.Setenv("AUTH_FIREBASE_MODE", "real")
		t.Setenv("FIREBASE_CREDENTIALS_FILE", "/etc/firebase/creds.json")
		t.Setenv("COOKIE_SECURE", "true")

		cfg := Load()

		assert.Equal(t, "8080", cfg.HTTPPort)
		assert.Equal(t, "8443", cfg.GRPCPort)
		assert.Equal(t, "postgres://custom/db", cfg.DatabaseURL)
		assert.Equal(t, "redis-host:6380", cfg.RedisAddr)
		assert.Equal(t, []byte("super-secret"), cfg.JWTSecret)
		assert.Equal(t, "real", cfg.FirebaseMode)
		assert.Equal(t, "/etc/firebase/creds.json", cfg.FirebaseCredentialsFile)
		assert.True(t, cfg.CookieSecure)
	})

	t.Run(testCaseSuccess_CookieSecureTrueString, func(t *testing.T) {
		t.Setenv("COOKIE_SECURE", "true")
		assert.True(t, Load().CookieSecure)
	})

	t.Run(testCaseSuccess_CookieSecureOtherValue, func(t *testing.T) {
		unsetEnv(t, "COOKIE_SECURE")
		for _, v := range []string{"false", "1", "yes", "TRUE", ""} {
			if v == "" {
				require.NoError(t, os.Unsetenv("COOKIE_SECURE"))
			} else {
				require.NoError(t, os.Setenv("COOKIE_SECURE", v))
			}
			assert.False(t, Load().CookieSecure, "value %q should parse to false", v)
		}
	})
}

func TestGetenv(t *testing.T) {
	t.Run("[Success] returns env value when set", func(t *testing.T) {
		t.Setenv("TICKETFLOW_TEST_GETENV_KEY", "custom-value")
		assert.Equal(t, "custom-value", getenv("TICKETFLOW_TEST_GETENV_KEY", "default-value"))
	})

	t.Run("[Success] returns default when unset", func(t *testing.T) {
		unsetEnv(t, "TICKETFLOW_TEST_GETENV_KEY_UNSET")
		assert.Equal(t, "default-value", getenv("TICKETFLOW_TEST_GETENV_KEY_UNSET", "default-value"))
	})

	t.Run("[Success] returns default when set to empty string", func(t *testing.T) {
		t.Setenv("TICKETFLOW_TEST_GETENV_KEY_EMPTY", "")
		assert.Equal(t, "default-value", getenv("TICKETFLOW_TEST_GETENV_KEY_EMPTY", "default-value"))
	})
}
