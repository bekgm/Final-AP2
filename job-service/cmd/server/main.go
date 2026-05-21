package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/yourname/freelance-platform/job-service/config"
	grpchandler "github.com/yourname/freelance-platform/job-service/internal/delivery/grpc"
	"github.com/yourname/freelance-platform/job-service/internal/email"
	"github.com/yourname/freelance-platform/job-service/internal/messaging"
	pgRepo "github.com/yourname/freelance-platform/job-service/internal/repository/postgres"
	"github.com/yourname/freelance-platform/job-service/internal/usecase"
	pb "github.com/yourname/freelance-platform/job-service/proto/job"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	var pingErr error
	for attempt := 1; attempt <= cfg.PostgresPingRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		pingErr = db.PingContext(ctx)
		cancel()
		if pingErr == nil {
			break
		}

		if attempt < cfg.PostgresPingRetries {
			log.Printf("waiting for postgres... (%d/%d): %v", attempt, cfg.PostgresPingRetries, pingErr)
			time.Sleep(cfg.PostgresPingInterval)
		}
	}
	if pingErr != nil {
		log.Fatalf("postgres ping failed after %d attempts: %v", cfg.PostgresPingRetries, pingErr)
	}

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to parse redis URL: %v", err)
	}
	_ = redis.NewClient(opt) // cache client — passed to repo layer as needed

	publisher, err := messaging.NewPublisher(cfg.RabbitMQURL, cfg.RabbitMQExchange)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer publisher.Close()

	emailSender := email.NewSMTPSender(
		cfg.SMTPHost, cfg.SMTPPort,
		cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom,
	)

	jobRepo := pgRepo.NewJobRepository(db)
	appRepo := pgRepo.NewApplicationRepository(db)

	jobUC := usecase.NewJobUseCase(jobRepo, appRepo, publisher, emailSender)

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterJobServiceServer(server, grpchandler.NewJobHandler(jobUC))
	reflection.Register(server) // useful for grpcurl during development

	log.Printf("Job Service listening on %s", cfg.GRPCPort)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
