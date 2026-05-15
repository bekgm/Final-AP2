package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/freelance-market/user-service/config"
	"github.com/freelance-market/user-service/internal/auth"
	"github.com/freelance-market/user-service/internal/db"
	"github.com/freelance-market/user-service/internal/handler"
	"github.com/freelance-market/user-service/internal/repository"
	"github.com/freelance-market/user-service/internal/service"
	pb "github.com/freelance-market/user-service/proto/user"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()

	dsn := db.DSN(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	pool, err := db.NewPool(dsn)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("connected to database", "host", cfg.Database.Host, "db", cfg.Database.DBName)

	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpirationHours)
	userRepo := repository.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo, jwtManager)
	userHandler := handler.NewUserHandler(userSvc, logger)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(logger),
			recoveryInterceptor(logger),
		),
	)

	pb.RegisterUserServiceServer(grpcServer, userHandler)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("failed to listen", "addr", addr, "error", err)
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("gRPC server started", "addr", addr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("server error", "error", err)
		}
	}()

	<-quit
	logger.Info("shutting down server...")

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		logger.Info("server stopped gracefully")
	case <-time.After(10 * time.Second):
		logger.Warn("forcing server stop after timeout")
		grpcServer.Stop()
	}
}

func loggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		if err != nil {
			logger.Error("gRPC call failed", "method", info.FullMethod, "duration_ms", duration.Milliseconds(), "error", err)
		} else {
			logger.Info("gRPC call succeeded", "method", info.FullMethod, "duration_ms", duration.Milliseconds())
		}
		return resp, err
	}
}

func recoveryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered", "method", info.FullMethod, "panic", r)
				err = fmt.Errorf("internal server error")
			}
		}()
		return handler(ctx, req)
	}
}