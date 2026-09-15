package config

import "os"

type Config struct {
	HTTPPort                string
	GRPCPort                string
	DatabaseURL             string
	RedisAddr               string
	JWTSecret               []byte
	FirebaseMode            string // "mock" (default) | "real"
	FirebaseCredentialsFile string
	CookieSecure            bool
}

// Default ports match every other service's default (see repo-structure.md's
// "unified container port" note) — each service is its own container with
// its own network namespace, so the port number carries no identity; only
// the Docker Compose service name does. Only api-gateway's HOST-side
// mapping needs to be unique (deployments/docker-compose.yaml), since the
// host itself has one shared network namespace.
const (
	defaultHTTPPort = "18080"
	defaultGRPCPort = "18443"
)

func Load() Config {
	return Config{
		HTTPPort:                getenv("HTTP_PORT", defaultHTTPPort),
		GRPCPort:                getenv("GRPC_PORT", defaultGRPCPort),
		DatabaseURL:             getenv("DATABASE_URL", "postgres://ticketflow:ticketflow@localhost:5432/ticketflow?sslmode=disable"),
		RedisAddr:               getenv("REDIS_ADDR", "localhost:6379"),
		JWTSecret:               []byte(getenv("JWT_SIGNING_KEY", "dev-secret-change-me")),
		FirebaseMode:            getenv("AUTH_FIREBASE_MODE", "mock"),
		FirebaseCredentialsFile: os.Getenv("FIREBASE_CREDENTIALS_FILE"),
		CookieSecure:            getenv("COOKIE_SECURE", "false") == "true",
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
