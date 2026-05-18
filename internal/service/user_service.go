package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/freelance-market/user-service/internal/auth"
	"github.com/freelance-market/user-service/internal/model"
	"github.com/freelance-market/user-service/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already taken")
)

type UserService struct {
	repo       *repository.UserRepository
	jwtManager *auth.JWTManager
}

func NewUserService(repo *repository.UserRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

type RegisterInput struct {
	Email    string
	Password string
	Name     string
	Role     model.Role
}

type AuthResult struct {
	Token string
	User  *model.User
}

func (s *UserService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &model.User{
		ID:        uuid.New().String(),
		Email:     input.Email,
		Password:  string(hash),
		Name:      input.Name,
		Role:      input.Role,
		Bio:       "",
		Skills:    []string{},
		AvatarURL: "",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := s.jwtManager.Generate(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResult{Token: token, User: user}, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwtManager.Generate(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResult{Token: token, User: user}, nil
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

type UpdateInput struct {
	UserID    string
	Name      *string
	Bio       *string
	Skills    []string
	AvatarURL *string
}

func (s *UserService) UpdateUser(ctx context.Context, input UpdateInput) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	// Apply partial updates
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Bio != nil {
		user.Bio = *input.Bio
	}
	if input.Skills != nil {
		user.Skills = input.Skills
	}
	if input.AvatarURL != nil {
		user.AvatarURL = *input.AvatarURL
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	// Fetch fresh copy
	updated, err := s.repo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get updated user: %w", err)
	}
	return updated, nil
}
