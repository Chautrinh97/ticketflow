package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"ticketflow/services/api-gateway/internal/config"
	"ticketflow/services/api-gateway/internal/identityclient"
	"ticketflow/services/api-gateway/internal/middleware"
	"ticketflow/services/api-gateway/internal/router"
)

func main() {
	cfg := config.Load()

	identityCl, err := identityclient.Dial(cfg.IdentityGRPCAddr)
	if err != nil {
		log.Fatalf("dial identity-service grpc: %v", err)
	}
	defer identityCl.Close()

	r := router.New(
		middleware.CORS(cfg.AllowedOrigin),
		middleware.OptionalSessionCheck(cfg.JWTSecret, identityCl),
		router.Targets{
			Identity: cfg.IdentityHTTPURL,
			Event:    cfg.EventHTTPURL,
			Booking:  cfg.BookingHTTPURL,
			Payment:  cfg.PaymentHTTPURL,
		},
	)

	srv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: r}
	go func() {
		log.Printf("api-gateway http listening on :%s", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http serve: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	log.Println("shutting down api-gateway")
	_ = srv.Shutdown(context.Background())
}
