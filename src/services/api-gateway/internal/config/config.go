package config

import "os"

type Config struct {
	HTTPPort         string
	JWTSecret        []byte
	IdentityGRPCAddr string
	IdentityHTTPURL  string
	EventHTTPURL     string
	BookingHTTPURL   string
	PaymentHTTPURL   string
	AllowedOrigin    string
}

// Default ports are the same across every service, api-gateway's own
// listen port included — see the identical comment in
// identity-service/internal/config/config.go. Only api-gateway's HOST-side
// port mapping (deployments/docker-compose.yaml) needs to stay a
// well-known, stable value (8080) — this internal port does not.
const (
	defaultHTTPPort = "18080"
	defaultGRPCPort = "18443"
)

func Load() Config {
	return Config{
		HTTPPort:         getenv("HTTP_PORT", defaultHTTPPort),
		JWTSecret:        []byte(getenv("JWT_SIGNING_KEY", "dev-secret-change-me")),
		IdentityGRPCAddr: getenv("IDENTITY_GRPC_ADDR", "localhost:"+defaultGRPCPort),
		IdentityHTTPURL:  getenv("IDENTITY_HTTP_URL", "http://localhost:"+defaultHTTPPort),
		EventHTTPURL:     getenv("EVENT_HTTP_URL", "http://localhost:"+defaultHTTPPort),
		BookingHTTPURL:   getenv("BOOKING_HTTP_URL", "http://localhost:"+defaultHTTPPort),
		PaymentHTTPURL:   getenv("PAYMENT_HTTP_URL", "http://localhost:"+defaultHTTPPort),
		AllowedOrigin:    getenv("ALLOWED_ORIGIN", "http://localhost:3000"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
