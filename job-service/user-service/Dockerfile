# ---------- build stage ----------
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install protoc dependencies (needed if generating proto inside container)
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o user-service ./cmd/server

# ---------- runtime stage ----------
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/user-service .

EXPOSE 50051

ENTRYPOINT ["./user-service"]
