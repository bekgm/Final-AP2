# User Service — Freelance Market

gRPC microservice for user authentication and profile management.

## Endpoints

| Method | Description |
|---|---|
| `Register` | Create account, returns JWT + user |
| `Login` | Authenticate, returns JWT + user |
| `GetUser` | Fetch user profile by ID |
| `UpdateUser` | Partial update of profile fields |

## Stack

- **Language:** Go 1.22
- **Transport:** gRPC (protobuf)
- **Database:** PostgreSQL 16
- **Auth:** JWT (HS256, 72h expiry)
- **Password hashing:** bcrypt

## Project Structure

```
user-service/
├── cmd/server/main.go          # Entry point, gRPC server setup
├── config/config.go            # Environment-based config
├── internal/
│   ├── auth/jwt.go             # JWT generate & validate
│   ├── db/db.go                # pgxpool connection
│   ├── handler/user_handler.go # gRPC handler (4 endpoints)
│   ├── model/user.go           # User struct + proto converter
│   ├── repository/             # DB queries
│   └── service/user_service.go # Business logic
├── migrations/001_create_users.sql
├── proto/user/user.proto       # gRPC contract
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## Quick Start

### Option 1: Docker Compose (Recommended)

```bash
docker compose up --build
```

Service starts on `localhost:50051`, PostgreSQL on `localhost:5432`.

### Option 2: Local Run

1. Start PostgreSQL and create a database:
   ```bash
   createdb userservice
   psql userservice -f migrations/001_create_users.sql
   ```

2. Copy and edit env:
   ```bash
   cp .env.example .env
   # edit .env with your DB password and JWT secret
   ```

3. Run:
   ```bash
   make run
   ```

## Testing with grpcurl

```bash
# Register
grpcurl -plaintext -d '{
  "email": "alice@example.com",
  "password": "secret123",
  "name": "Alice",
  "role": "ROLE_CLIENT"
}' localhost:50051 user.UserService/Register

# Login
grpcurl -plaintext -d '{
  "email": "alice@example.com",
  "password": "secret123"
}' localhost:50051 user.UserService/Login

# Get user
grpcurl -plaintext -d '{"user_id": "<id from above>"}' \
  localhost:50051 user.UserService/GetUser

# Update profile
grpcurl -plaintext -d '{
  "user_id": "<id>",
  "bio": "Senior Go developer",
  "skills": ["Go", "gRPC", "PostgreSQL"]
}' localhost:50051 user.UserService/UpdateUser
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `GRPC_PORT` | `50051` | gRPC server port |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | DB username |
| `DB_PASSWORD` | `password` | DB password |
| `DB_NAME` | `userservice` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `JWT_SECRET` | (change this!) | HS256 signing key |
| `JWT_EXPIRATION_HOURS` | `72` | Token lifetime |

## Integration with Other Services

Other services (Job Service, Messaging Service) call this service via gRPC to:
- Validate JWT tokens via `Login` or independently using the shared `JWT_SECRET`
- Look up user info via `GetUser(user_id)`

Import the proto file into other services:
```
proto/user/user.proto
```
