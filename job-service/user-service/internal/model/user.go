package model

import (
	"time"

	pb "github.com/freelance-market/user-service/proto/user"
)

type Role string

const (
	RoleClient     Role = "client"
	RoleFreelancer Role = "freelancer"
)

type User struct {
	ID        string    `db:"id"`
	Email     string    `db:"email"`
	Password  string    `db:"password_hash"`
	Name      string    `db:"name"`
	Role      Role      `db:"role"`
	Bio       string    `db:"bio"`
	Skills    []string  `db:"skills"`
	AvatarURL string    `db:"avatar_url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (u *User) ToProto() *pb.User {
	role := pb.Role_ROLE_CLIENT
	if u.Role == RoleFreelancer {
		role = pb.Role_ROLE_FREELANCER
	}

	return &pb.User{
		Id:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      role,
		Bio:       u.Bio,
		Skills:    u.Skills,
		AvatarUrl: u.AvatarURL,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}

func RoleFromProto(r pb.Role) Role {
	if r == pb.Role_ROLE_FREELANCER {
		return RoleFreelancer
	}
	return RoleClient
}
