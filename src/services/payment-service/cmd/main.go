package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ticketflow/services/payment-service/internal/bookingclient"
	"ticketflow/services/payment-service/internal/config"
	httpserver "ticketflow/services/payment-service/internal/handler/http"
	"ticketflow/services/payment-service/internal/repository"
	"ticketflow/services/payment-service/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}

	bookingCl, err := bookingclient.Dial(cfg.BookingGRPCAddr)
	if err != nil {
		log.Fatalf("dial booking-service grpc: %v", err)
	}
	defer bookingCl.Close()

	paymentRepo := repository.NewPaymentRepository(db)
	paymentService := service.NewPaymentService(paymentRepo, bookingCl)

	paymentHandler := httpserver.NewPaymentHandler(paymentService)
	router := httpserver.NewRouter(cfg.JWTSecret, paymentHandler)

	httpSrv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: router}
	go func() {
		log.Printf("payment-service http listening on :%s", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http serve: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	log.Println("shutting down payment-service")
	_ = httpSrv.Shutdown(context.Background())
}
