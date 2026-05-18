FROM golang:1.22-alpine AS builder

WORKDIR /app
RUN apk add --no-cache protobuf protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.1
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
ENV PATH="/go/bin:${PATH}"

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN protoc -I proto -I /usr/include --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/job/job.proto
RUN go build -o job-service ./cmd/server
RUN go build -o gateway ./cmd/gateway

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/job-service .
COPY --from=builder /app/gateway .
COPY migrations ./migrations

EXPOSE 50052 8080
CMD ["./job-service"]
