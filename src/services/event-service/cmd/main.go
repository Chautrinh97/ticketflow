package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ticketflow/pkg/grpcinterceptor"
	"ticketflow/proto/eventpb"
	"ticketflow/services/event-service/internal/config"
	grpcserver "ticketflow/services/event-service/internal/handler/grpc"
	httpserver "ticketflow/services/event-service/internal/handler/http"
	"ticketflow/services/event-service/internal/repository"
	"ticketflow/services/event-service/internal/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}

	mongoClient, err := mongo.Connect(ctx, mongooptions.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("connect mongo: %v", err)
	}
	mongoDB := mongoClient.Database(cfg.MongoDBName)

	eventRepo := repository.NewEventPostgresRepo(db)
	ticketTypeRepo := repository.NewTicketTypePostgresRepo(db)
	catalogRepo := repository.NewEventCatalogMongoRepo(mongoDB)
	if err := catalogRepo.EnsureIndexes(ctx); err != nil {
		log.Fatalf("ensure mongo indexes: %v", err)
	}

	eventService := service.NewEventService(eventRepo, ticketTypeRepo, catalogRepo)
	ticketTypeService := service.NewTicketTypeService(eventRepo, ticketTypeRepo)

	publicHandler := httpserver.NewPublicEventHandler(eventService)
	organizerHandler := httpserver.NewOrganizerEventHandler(eventService)
	ticketTypeHandler := httpserver.NewTicketTypeHandler(ticketTypeService)
	router := httpserver.NewRouter(cfg.JWTSecret, publicHandler, organizerHandler, ticketTypeHandler)

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcinterceptor.UnaryServerRecovery(),
			grpcinterceptor.UnaryServerLogging("event-service"),
		),
	)
	eventpb.RegisterEventServiceServer(grpcSrv, grpcserver.NewEventGRPCServer(eventService))

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("grpc listen: %v", err)
		}
		log.Printf("event-service grpc listening on :%s", cfg.GRPCPort)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	httpSrv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: router}
	go func() {
		log.Printf("event-service http listening on :%s", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http serve: %v", err)
		}
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-stopCtx.Done()
	log.Println("shutting down event-service")
	_ = httpSrv.Shutdown(context.Background())
	grpcSrv.GracefulStop()
	_ = mongoClient.Disconnect(context.Background())
}
