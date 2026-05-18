FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o job-service ./cmd/server
RUN go build -o gateway ./cmd/gateway

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/job-service .
COPY --from=builder /app/gateway .
COPY --from=builder /app/migrations ./migrations

EXPOSE 50052 8080
CMD ["./job-service"]
