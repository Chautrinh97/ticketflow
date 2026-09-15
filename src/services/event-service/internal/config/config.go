package config

import "os"

type Config struct {
	HTTPPort    string
	GRPCPort    string
	DatabaseURL string
	MongoURI    string
	MongoDBName string
	JWTSecret   []byte
}

// Default ports are the same across every service — see the identical
// comment in identity-service/internal/config/config.go.
const (
	defaultHTTPPort = "18080"
	defaultGRPCPort = "18443"
)

func Load() Config {
	return Config{
		HTTPPort:    getenv("HTTP_PORT", defaultHTTPPort),
		GRPCPort:    getenv("GRPC_PORT", defaultGRPCPort),
		DatabaseURL: getenv("DATABASE_URL", "postgres://ticketflow:ticketflow@localhost:5432/ticketflow?sslmode=disable"),
		MongoURI:    getenv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getenv("MONGO_DB_NAME", "ticketflow"),
		JWTSecret:   []byte(getenv("JWT_SIGNING_KEY", "dev-secret-change-me")),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
