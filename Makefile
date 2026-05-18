.PHONY: proto build run test docker-up docker-down

# Generate Go code from .proto files
proto:
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/user/user.proto

# Build binary
build:
	go build -o bin/user-service ./cmd/server

# Run locally (needs .env and running postgres)
run:
	go run ./cmd/server

# Run all tests
test:
	go test ./... -v

# Start postgres + service with docker compose
docker-up:
	docker compose up --build

# Stop everything
docker-down:
	docker compose down -v

# Apply migrations manually (if not using docker init scripts)
migrate:
	psql "$$DATABASE_URL" -f migrations/001_create_users.sql

# Format code
fmt:
	gofmt -w .

# Lint (needs golangci-lint installed)
lint:
	golangci-lint run ./...
