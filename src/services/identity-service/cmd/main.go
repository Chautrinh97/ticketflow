package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ticketflow/pkg/grpcinterceptor"
	"ticketflow/proto/identitypb"
	"ticketflow/services/identity-service/internal/config"
	firebaseverify "ticketflow/services/identity-service/internal/firebase"
	grpcserver "ticketflow/services/identity-service/internal/handler/grpc"
	httpserver "ticketflow/services/identity-service/internal/handler/http"
	"ticketflow/services/identity-service/internal/repository"
	"ticketflow/services/identity-service/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})

	var verifier firebaseverify.Verifier
	if cfg.FirebaseMode == "real" {
		v, err := firebaseverify.NewRealVerifier(context.Background(), cfg.FirebaseCredentialsFile)
		if err != nil {
			log.Fatalf("init firebase verifier: %v", err)
		}
		verifier = v
	} else {
		log.Println("AUTH_FIREBASE_MODE=mock — using mock Firebase verifier (dev only, do not use in production)")
		verifier = firebaseverify.NewMockVerifier()
	}

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(rdb)
	authService := service.NewAuthService(userRepo, sessionRepo, verifier, cfg.JWTSecret)
	userService := service.NewUserService(userRepo)

	authHandler := httpserver.NewAuthHandler(authService, cfg.CookieSecure)
	userHandler := httpserver.NewUserHandler(userService)
	router := httpserver.NewRouter(cfg.JWTSecret, authHandler, userHandler)

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcinterceptor.UnaryServerRecovery(),
			grpcinterceptor.UnaryServerLogging("identity-service"),
		),
	)
	identitypb.RegisterIdentityServiceServer(grpcSrv, grpcserver.NewIdentityGRPCServer(authService))

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("grpc listen: %v", err)
		}
		log.Printf("identity-service grpc listening on :%s", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	httpSrv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: router}
	go func() {
		log.Printf("identity-service http listening on :%s", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http serve: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	log.Println("shutting down identity-service")
	_ = httpSrv.Shutdown(context.Background())
	grpcSrv.GracefulStop()
}
