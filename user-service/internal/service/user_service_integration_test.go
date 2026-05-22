package service_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/freelance-market/user-service/internal/auth"
	"github.com/freelance-market/user-service/internal/model"
	"github.com/freelance-market/user-service/internal/repository"
	"github.com/freelance-market/user-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Integration tests for user-service
// Requires: docker compose up postgres-users -d
// Test DB: postgres://postgres:password@localhost:5432/userservice?sslmode=disable

func getTestUserPool(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("TEST_USER_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/userservice?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Cannot connect to test database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Test database not available: %v", err)
	}

	return pool
}

func cleanupUserTestData(t *testing.T, pool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = pool.Exec(ctx, "DELETE FROM users WHERE email LIKE 'integration-test-%'")
}

func TestIntegrationUserRegister(t *testing.T) {
	pool := getTestUserPool(t)
	defer pool.Close()
	defer cleanupUserTestData(t, pool)

	repo := repository.NewUserRepository(pool)
	jwt := auth.NewJWTManager("test-secret", 24)
	svc := service.NewUserService(repo, jwt)

	ctx := context.Background()
	email := fmt.Sprintf("integration-test-%d@example.com", time.Now().UnixNano())

	result, err := svc.Register(ctx, service.RegisterInput{
		Email:    email,
		Password: "securepassword123",
		Name:     "Integration Test User",
		Role:     model.RoleClient,
	})

	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	if result.Token == "" {
		t.Error("Expected JWT token")
	}

	if result.User.ID == "" {
		t.Error("Expected user ID")
	}

	if result.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, result.User.Email)
	}

	t.Logf("User registered successfully with ID: %s", result.User.ID)
}

func TestIntegrationUserLogin(t *testing.T) {
	pool := getTestUserPool(t)
	defer pool.Close()
	defer cleanupUserTestData(t, pool)

	repo := repository.NewUserRepository(pool)
	jwt := auth.NewJWTManager("test-secret", 24)
	svc := service.NewUserService(repo, jwt)

	ctx := context.Background()
	email := fmt.Sprintf("integration-test-%d@example.com", time.Now().UnixNano())
	password := "securepassword123"

	// Register first
	_, err := svc.Register(ctx, service.RegisterInput{
		Email:    email,
		Password: password,
		Name:     "Login Test User",
		Role:     model.RoleFreelancer,
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Login
	result, err := svc.Login(ctx, email, password)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result.Token == "" {
		t.Error("Expected JWT token on login")
	}

	if result.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, result.User.Email)
	}

	t.Log("Login successful")
}

func TestIntegrationUserProfileUpdate(t *testing.T) {
	pool := getTestUserPool(t)
	defer pool.Close()
	defer cleanupUserTestData(t, pool)

	repo := repository.NewUserRepository(pool)
	jwt := auth.NewJWTManager("test-secret", 24)
	svc := service.NewUserService(repo, jwt)

	ctx := context.Background()
	email := fmt.Sprintf("integration-test-%d@example.com", time.Now().UnixNano())

	// Register
	regResult, err := svc.Register(ctx, service.RegisterInput{
		Email:    email,
		Password: "password123",
		Name:     "Original Name",
		Role:     model.RoleClient,
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	userID := regResult.User.ID

	// Get user profile
	user, err := svc.GetUser(ctx, userID)
	if err != nil {
		t.Fatalf("Get user failed: %v", err)
	}

	if user.Name != "Original Name" {
		t.Errorf("Expected 'Original Name', got '%s'", user.Name)
	}

	// Update profile
	newName := "Updated Name"
	updated, err := svc.UpdateUser(ctx, service.UpdateInput{
		UserID: userID,
		Name:   &newName,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Expected 'Updated Name', got '%s'", updated.Name)
	}

	// Verify update persisted
	userAfterUpdate, err := svc.GetUser(ctx, userID)
	if err != nil {
		t.Fatalf("Get user after update failed: %v", err)
	}

	if userAfterUpdate.Name != "Updated Name" {
		t.Errorf("Update did not persist: expected 'Updated Name', got '%s'", userAfterUpdate.Name)
	}

	t.Log("Profile update test passed")
}

func TestIntegrationDuplicateEmail(t *testing.T) {
	pool := getTestUserPool(t)
	defer pool.Close()
	defer cleanupUserTestData(t, pool)

	repo := repository.NewUserRepository(pool)
	jwt := auth.NewJWTManager("test-secret", 24)
	svc := service.NewUserService(repo, jwt)

	ctx := context.Background()
	email := fmt.Sprintf("integration-test-%d@example.com", time.Now().UnixNano())

	// First registration
	_, err := svc.Register(ctx, service.RegisterInput{
		Email:    email,
		Password: "password123",
		Name:     "First User",
		Role:     model.RoleClient,
	})
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Second registration with same email - should fail
	_, err = svc.Register(ctx, service.RegisterInput{
		Email:    email,
		Password: "password456",
		Name:     "Second User",
		Role:     model.RoleFreelancer,
	})

	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}

	t.Logf("Duplicate email correctly rejected: %v", err)
}
