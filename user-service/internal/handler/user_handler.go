package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/freelance-market/user-service/internal/model"
	"github.com/freelance-market/user-service/internal/service"
	pb "github.com/freelance-market/user-service/proto/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	svc    *service.UserService
	logger *slog.Logger
}

func NewUserHandler(svc *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, logger: logger}
}

// Register creates a new user account and returns a JWT token.
func (h *UserHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Validate input
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if len(req.Password) < 6 {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 6 characters")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Role == pb.Role_ROLE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "role must be client or freelancer")
	}

	result, err := h.svc.Register(ctx, service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Role:     model.RoleFromProto(req.Role),
	})
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			return nil, status.Error(codes.AlreadyExists, "email already registered")
		}
		h.logger.Error("register failed", "error", err)
		return nil, status.Error(codes.Internal, "registration failed")
	}

	h.logger.Info("user registered", "user_id", result.User.ID, "email", result.User.Email)
	return &pb.RegisterResponse{
		Token: result.Token,
		User:  result.User.ToProto(),
	}, nil
}

// Login authenticates a user and returns a JWT token.
func (h *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	result, err := h.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		h.logger.Error("login failed", "error", err, "email", req.Email)
		return nil, status.Error(codes.Internal, "login failed")
	}

	h.logger.Info("user logged in", "user_id", result.User.ID)
	return &pb.LoginResponse{
		Token: result.Token,
		User:  result.User.ToProto(),
	}, nil
}

// GetUser retrieves user profile by ID.
func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := h.svc.GetUser(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		h.logger.Error("get user failed", "error", err, "user_id", req.UserId)
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &pb.GetUserResponse{User: user.ToProto()}, nil
}

// UpdateUser updates user profile fields (partial update supported).
func (h *UserHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	input := service.UpdateInput{
		UserID: req.UserId,
		Skills: req.Skills,
	}
	if req.Name != nil {
		name := req.GetName()
		input.Name = &name
	}
	if req.Bio != nil {
		bio := req.GetBio()
		input.Bio = &bio
	}
	if req.AvatarUrl != nil {
		url := req.GetAvatarUrl()
		input.AvatarURL = &url
	}

	user, err := h.svc.UpdateUser(ctx, input)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		h.logger.Error("update user failed", "error", err, "user_id", req.UserId)
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	h.logger.Info("user updated", "user_id", user.ID)
	return &pb.UpdateUserResponse{User: user.ToProto()}, nil
}
