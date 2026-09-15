package config

import "os"

type Config struct {
	HTTPPort        string
	DatabaseURL     string
	JWTSecret       []byte
	BookingGRPCAddr string
}

// Default ports are the same across every service — see the identical
// comment in identity-service/internal/config/config.go. payment-service
// itself has no gRPC server (nothing calls it internally), only a client
// to booking-service's.
const (
	defaultHTTPPort = "18080"
	defaultGRPCPort = "18443"
)

func Load() Config {
	return Config{
		HTTPPort:        getenv("HTTP_PORT", defaultHTTPPort),
		DatabaseURL:     getenv("DATABASE_URL", "postgres://ticketflow:ticketflow@localhost:5432/ticketflow?sslmode=disable"),
		JWTSecret:       []byte(getenv("JWT_SIGNING_KEY", "dev-secret-change-me")),
		BookingGRPCAddr: getenv("BOOKING_GRPC_ADDR", "localhost:"+defaultGRPCPort),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
