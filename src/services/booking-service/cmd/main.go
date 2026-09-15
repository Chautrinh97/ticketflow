package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	"ticketflow/pkg/grpcinterceptor"
	"ticketflow/proto/bookingpb"
	"ticketflow/services/booking-service/internal/config"
	"ticketflow/services/booking-service/internal/eventclient"
	grpcserver "ticketflow/services/booking-service/internal/handler/grpc"
	httpserver "ticketflow/services/booking-service/internal/handler/http"
	"ticketflow/services/booking-service/internal/lock"
	"ticketflow/services/booking-service/internal/repository"
	"ticketflow/services/booking-service/internal/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	locker := lock.NewTicketTypeLocker(rdb)

	eventCl, err := eventclient.Dial(cfg.EventGRPCAddr)
	if err != nil {
		log.Fatalf("dial event-service grpc: %v", err)
	}
	defer eventCl.Close()

	repo := repository.NewBookingRepository(pool)
	bookingService := service.NewBookingService(repo, locker, eventCl)

	bookingHandler := httpserver.NewBookingHandler(bookingService)
	router := httpserver.NewRouter(cfg.JWTSecret, bookingHandler)

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcinterceptor.UnaryServerRecovery(),
			grpcinterceptor.UnaryServerLogging("booking-service"),
		),
	)
	bookingpb.RegisterBookingServiceServer(grpcSrv, grpcserver.NewBookingGRPCServer(bookingService))

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("grpc listen: %v", err)
		}
		log.Printf("booking-service grpc listening on :%s", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	httpSrv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: router}
	go func() {
		log.Printf("booking-service http listening on :%s", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http serve: %v", err)
		}
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-stopCtx.Done()
	log.Println("shutting down booking-service")
	_ = httpSrv.Shutdown(context.Background())
	grpcSrv.GracefulStop()
}
