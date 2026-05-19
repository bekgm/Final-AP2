package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/freelance-market/user-service/internal/auth"
	"github.com/freelance-market/user-service/internal/model"
	"github.com/freelance-market/user-service/internal/service"
)

// ─────────────────────────────────────────────
// Mock repository
// ─────────────────────────────────────────────

type mockUserRepo struct {
	users  map[string]*model.User
	byEmail map[string]*model.User
}

func newMockRepo() *mockUserRepo {
	return &mockUserRepo{
		users:   make(map[string]*model.User),
		byEmail: make(map[string]*model.User),
	}
}

func (m *mockUserRepo) Create(_ context.Context, user *model.User) error {
	if _, exists := m.byEmail[user.Email]; exists {
		return service.ErrEmailAlreadyExists
	}
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (*model.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	if u, ok := m.byEmail[email]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepo) Update(_ context.Context, user *model.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return errors.New("user not found")
	}
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

func newService() (*service.UserService, *mockUserRepo) {
	repo := newMockRepo()
	jwt := auth.NewJWTManager("test-secret", 24)
	svc := service.NewUserService(repo, jwt)
	return svc, repo
}

// ─────────────────────────────────────────────
// Register tests
// ─────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	svc, _ := newService()
	result, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		Role:     model.RoleClient,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Error("expected JWT token to be non-empty")
	}
	if result.User.ID == "" {
		t.Error("expected user ID to be set")
	}
	if result.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", result.User.Email)
	}
	if result.User.Role != model.RoleClient {
		t.Errorf("expected role client, got %s", result.User.Role)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _ := newService()
	input := service.RegisterInput{
		Email:    "dup@example.com",
		Password: "password123",
		Name:     "User One",
		Role:     model.RoleClient,
	}
	if _, err := svc.Register(context.Background(), input); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	_, err := svc.Register(context.Background(), input)
	if !errors.Is(err, service.ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestRegister_PasswordIsHashed(t *testing.T) {
	svc, repo := newService()
	svc.Register(context.Background(), service.RegisterInput{
		Email:    "hash@example.com",
		Password: "plaintext",
		Name:     "Hash Test",
		Role:     model.RoleFreelancer,
	})

	stored := repo.byEmail["hash@example.com"]
	if stored.Password == "plaintext" {
		t.Error("password must not be stored as plaintext")
	}
	if len(stored.Password) < 30 {
		t.Error("password hash looks too short")
	}
}

// ─────────────────────────────────────────────
// Login tests
// ─────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	svc, _ := newService()
	svc.Register(context.Background(), service.RegisterInput{
		Email:    "login@example.com",
		Password: "mypassword",
		Name:     "Login User",
		Role:     model.RoleFreelancer,
	})

	result, err := svc.Login(context.Background(), "login@example.com", "mypassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, _ := newService()
	svc.Register(context.Background(), service.RegisterInput{
		Email:    "wp@example.com",
		Password: "correct",
		Name:     "WP User",
		Role:     model.RoleClient,
	})

	_, err := svc.Login(context.Background(), "wp@example.com", "wrong")
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc, _ := newService()
	_, err := svc.Login(context.Background(), "nobody@example.com", "pass")
	if !errors.Is(err, service.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

// ─────────────────────────────────────────────
// GetUser tests
// ─────────────────────────────────────────────

func TestGetUser_Success(t *testing.T) {
	svc, _ := newService()
	result, _ := svc.Register(context.Background(), service.RegisterInput{
		Email:    "get@example.com",
		Password: "pass123",
		Name:     "Get User",
		Role:     model.RoleClient,
	})

	user, err := svc.GetUser(context.Background(), result.User.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "get@example.com" {
		t.Errorf("expected get@example.com, got %s", user.Email)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	svc, _ := newService()
	_, err := svc.GetUser(context.Background(), "nonexistent-id")
	if !errors.Is(err, service.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

// ─────────────────────────────────────────────
// UpdateUser tests
// ─────────────────────────────────────────────

func TestUpdateUser_PartialUpdate(t *testing.T) {
	svc, _ := newService()
	result, _ := svc.Register(context.Background(), service.RegisterInput{
		Email:    "upd@example.com",
		Password: "pass123",
		Name:     "Old Name",
		Role:     model.RoleFreelancer,
	})

	newName := "New Name"
	updated, err := svc.UpdateUser(context.Background(), service.UpdateInput{
		UserID: result.User.ID,
		Name:   &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected 'New Name', got %s", updated.Name)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	svc, _ := newService()
	name := "Ghost"
	_, err := svc.UpdateUser(context.Background(), service.UpdateInput{
		UserID: "ghost-id",
		Name:   &name,
	})
	if !errors.Is(err, service.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}
