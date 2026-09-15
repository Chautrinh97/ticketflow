package config

import "os"

type Config struct {
	HTTPPort      string
	GRPCPort      string
	DatabaseURL   string
	RedisAddr     string
	JWTSecret     []byte
	EventGRPCAddr string
}

// Default ports are the same across every service — see the identical
// comment in identity-service/internal/config/config.go.
const (
	defaultHTTPPort = "18080"
	defaultGRPCPort = "18443"
)

func Load() Config {
	return Config{
		HTTPPort:      getenv("HTTP_PORT", defaultHTTPPort),
		GRPCPort:      getenv("GRPC_PORT", defaultGRPCPort),
		DatabaseURL:   getenv("DATABASE_URL", "postgres://ticketflow:ticketflow@localhost:5432/ticketflow?sslmode=disable"),
		RedisAddr:     getenv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:     []byte(getenv("JWT_SIGNING_KEY", "dev-secret-change-me")),
		EventGRPCAddr: getenv("EVENT_GRPC_ADDR", "localhost:"+defaultGRPCPort),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
