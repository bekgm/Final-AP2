package main

import (
	"log"
	"net"
	"os"

	"github.com/bekgm/Final-AP2/internal/models"
	"github.com/bekgm/Final-AP2/internal/repository"
	"github.com/bekgm/Final-AP2/internal/service"
	pb "github.com/bekgm/Final-AP2/pkg/messaging"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=messaging_db port=5432 sslmode=disable TimeZone=UTC"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.Message{}); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}

	repo := repository.NewMessagingRepository(db)
	svc := service.NewMessagingService(repo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMessagingServiceServer(grpcServer, svc)

	log.Printf("Messaging Service listening on :%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
