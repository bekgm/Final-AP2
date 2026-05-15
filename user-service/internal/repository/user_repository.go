package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/freelance-market/user-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, role, bio, skills, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Password,
		user.Name,
		string(user.Role),
		user.Bio,
		user.Skills,
		user.AvatarURL,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrEmailAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, name, role, bio, skills, avatar_url, created_at, updated_at
		FROM users WHERE id = $1
	`
	user, err := r.scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, name, role, bio, skills, avatar_url, created_at, updated_at
		FROM users WHERE email = $1
	`
	user, err := r.scanUser(r.pool.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET name = $1, bio = $2, skills = $3, avatar_url = $4, updated_at = $5
		WHERE id = $6
	`
	result, err := r.pool.Exec(ctx, query,
		user.Name,
		user.Bio,
		user.Skills,
		user.AvatarURL,
		time.Now(),
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) scanUser(row pgx.Row) (*model.User, error) {
	var u model.User
	var role string
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.Password,
		&u.Name,
		&role,
		&u.Bio,
		&u.Skills,
		&u.AvatarURL,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.Role = model.Role(role)
	return &u, nil
}

func isDuplicateKeyError(err error) bool {
	return err != nil && (containsString(err.Error(), "duplicate key") ||
		containsString(err.Error(), "unique constraint") ||
		containsString(err.Error(), "23505"))
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && findSubstring(s, substr))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
