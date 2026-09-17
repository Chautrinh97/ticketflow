package config

import "testing"

// Load reads plain env vars with hardcoded fallbacks — no DB/network/mocks
// needed, so a plain table-driven test is enough per
// docs/01-architecture/backend-conventions.md's "Test logic thuần" rule.
// t.Setenv("KEY", "") is used to force the "unset" branch deterministically
// (getenv treats an empty string exactly like an absent var) regardless of
// whatever the real process environment happens to have, and t.Setenv
// auto-restores the previous value after each test.
func TestLoad(t *testing.T) {
	t.Run("[Success] Dùng giá trị mặc định khi không có biến môi trường nào được set", func(t *testing.T) {
		t.Setenv("HTTP_PORT", "")
		t.Setenv("GRPC_PORT", "")
		t.Setenv("DATABASE_URL", "")
		t.Setenv("MONGO_URI", "")
		t.Setenv("MONGO_DB_NAME", "")
		t.Setenv("JWT_SIGNING_KEY", "")

		cfg := Load()

		if cfg.HTTPPort != defaultHTTPPort {
			t.Errorf("HTTPPort = %q, want default %q", cfg.HTTPPort, defaultHTTPPort)
		}
		if cfg.GRPCPort != defaultGRPCPort {
			t.Errorf("GRPCPort = %q, want default %q", cfg.GRPCPort, defaultGRPCPort)
		}
		if cfg.DatabaseURL != "postgres://ticketflow:ticketflow@localhost:5432/ticketflow?sslmode=disable" {
			t.Errorf("DatabaseURL = %q, want the hardcoded local default", cfg.DatabaseURL)
		}
		if cfg.MongoURI != "mongodb://localhost:27017" {
			t.Errorf("MongoURI = %q, want the hardcoded local default", cfg.MongoURI)
		}
		if cfg.MongoDBName != "ticketflow" {
			t.Errorf("MongoDBName = %q, want \"ticketflow\"", cfg.MongoDBName)
		}
		if string(cfg.JWTSecret) != "dev-secret-change-me" {
			t.Errorf("JWTSecret = %q, want \"dev-secret-change-me\"", string(cfg.JWTSecret))
		}
	})

	t.Run("[Success] Ghi đè toàn bộ giá trị khi biến môi trường được set", func(t *testing.T) {
		t.Setenv("HTTP_PORT", "9090")
		t.Setenv("GRPC_PORT", "9443")
		t.Setenv("DATABASE_URL", "postgres://custom:pw@db:5432/custom?sslmode=disable")
		t.Setenv("MONGO_URI", "mongodb://mongo-host:27017")
		t.Setenv("MONGO_DB_NAME", "custom_db")
		t.Setenv("JWT_SIGNING_KEY", "super-secret")

		cfg := Load()

		if cfg.HTTPPort != "9090" {
			t.Errorf("HTTPPort = %q, want %q", cfg.HTTPPort, "9090")
		}
		if cfg.GRPCPort != "9443" {
			t.Errorf("GRPCPort = %q, want %q", cfg.GRPCPort, "9443")
		}
		if cfg.DatabaseURL != "postgres://custom:pw@db:5432/custom?sslmode=disable" {
			t.Errorf("DatabaseURL = %q, want the overridden value", cfg.DatabaseURL)
		}
		if cfg.MongoURI != "mongodb://mongo-host:27017" {
			t.Errorf("MongoURI = %q, want the overridden value", cfg.MongoURI)
		}
		if cfg.MongoDBName != "custom_db" {
			t.Errorf("MongoDBName = %q, want %q", cfg.MongoDBName, "custom_db")
		}
		if string(cfg.JWTSecret) != "super-secret" {
			t.Errorf("JWTSecret = %q, want %q", string(cfg.JWTSecret), "super-secret")
		}
	})

	t.Run("[Success] Chỉ ghi đè một biến, các biến còn lại vẫn dùng mặc định", func(t *testing.T) {
		t.Setenv("HTTP_PORT", "7070")
		t.Setenv("GRPC_PORT", "")
		t.Setenv("DATABASE_URL", "")
		t.Setenv("MONGO_URI", "")
		t.Setenv("MONGO_DB_NAME", "")
		t.Setenv("JWT_SIGNING_KEY", "")

		cfg := Load()

		if cfg.HTTPPort != "7070" {
			t.Errorf("HTTPPort = %q, want overridden %q", cfg.HTTPPort, "7070")
		}
		if cfg.GRPCPort != defaultGRPCPort {
			t.Errorf("GRPCPort = %q, want default %q", cfg.GRPCPort, defaultGRPCPort)
		}
	})
}

func TestGetenv(t *testing.T) {
	t.Run("[Success] Trả về giá trị mặc định khi biến không tồn tại/rỗng", func(t *testing.T) {
		t.Setenv("TICKETFLOW_TEST_UNSET_KEY", "")
		if got := getenv("TICKETFLOW_TEST_UNSET_KEY", "fallback"); got != "fallback" {
			t.Errorf("getenv() = %q, want %q", got, "fallback")
		}
	})

	t.Run("[Success] Trả về giá trị đã set khi biến tồn tại", func(t *testing.T) {
		t.Setenv("TICKETFLOW_TEST_SET_KEY", "actual-value")
		if got := getenv("TICKETFLOW_TEST_SET_KEY", "fallback"); got != "actual-value" {
			t.Errorf("getenv() = %q, want %q", got, "actual-value")
		}
	})
}
