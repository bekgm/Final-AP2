# Final-AP2: Messaging Service

This branch contains the implementation of the Messaging Service for the final exam in Golang.

## gRPC Endpoints Implemented

- `SendMessage`: Send a message from one user to another.
- `GetMessages`: Get message history between two users with pagination.
- `GetDialogs`: Get a list of recent dialogs (chats) for a specific user.

## Tech Stack
- Go
- gRPC / Protobuf
- PostgreSQL
- GORM

## Running Locally

1. Set up a PostgreSQL database.
2. Configure `.env` file based on `.env.example` (or set `DATABASE_URL` environment variable).
3. Run the gRPC Messaging Service:
```bash
go run cmd/server/main.go
```
4. Run the API Gateway (in a separate terminal):
```bash
go run cmd/api_gateway/main.go
```

The API Gateway will run on `http://localhost:8080` (or `API_PORT`) and will proxy REST requests to the gRPC service.