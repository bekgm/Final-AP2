# Freelance Platform — Go Microservices

A backend for a freelance marketplace built with Go microservices. Clients can post jobs, freelancers can apply, and both sides communicate through a messaging system. The project demonstrates microservices architecture with service separation, gRPC inter-service communication, and event-driven behavior via RabbitMQ.

---

## Project Goal

The goal of this project is to build a simple freelance platform where clients and freelancers can interact. Clients can create job postings, freelancers can apply for them, and both sides can communicate through a messaging system.

The main objective is to demonstrate how a microservices architecture can be implemented using the Go programming language, focusing on service separation, inter-service communication, and basic scalability principles.

---

## System Overview

The system consists of three independent microservices, each responsible for a specific domain of the application:

- **User Service** — manages users, authentication, and profiles
- **Job Service** — handles job listings and applications
- **Messaging Service** — provides communication between users

Services communicate using:
- **gRPC** for synchronous communication between services
- **RabbitMQ** for asynchronous event-driven communication

All external HTTP requests are handled through an **API Gateway**, which serves as the single entry point to the system.

---

## Architecture

```
                        ┌─────────────────────────────┐
                        │     job-gateway :8080        │  ← Main entry point (HTTP)
                        │  (proxies all 3 services)    │
                        └──────────┬──────────────┬────┘
                                   │              │
              ┌────────────────────┤              ├────────────────────┐
              ▼                    ▼              ▼                    ▼
     user-service:50051   job-service:50052   messaging-service:50053
          │                      │  │               │
     postgres-users          postgres-jobs      postgres-messaging
          :5432                  :5433              :5434
                               redis:6379
                             rabbitmq:5672

     messaging-gateway:8081  ← Direct HTTP gateway for messaging only
```

### Services

| Service | Protocol | Port | Description |
|---|---|---|---|
| `job-gateway` | HTTP | **8080** | Main API gateway — proxies all 3 services |
| `messaging-gateway` | HTTP | **8081** | Dedicated messaging HTTP gateway (Gin) |
| `user-service` | gRPC | 50051 | User registration, login, JWT auth |
| `job-service` | gRPC | 50052 | Job listings, applications, acceptance |
| `messaging-service` | gRPC | 50053 | Real-time messaging between users |

### Infrastructure

| Service | Port | Purpose |
|---|---|---|
| `postgres-users` | 5432 | Users database |
| `postgres-jobs` | 5433 | Jobs database |
| `postgres-messaging` | 5434 | Messages database |
| `redis` | 6379 | Caching (job-service) |
| `rabbitmq` | 5672 / 15672 | Message queue (job notifications) |

---

## Services

### User Service

Responsible for user-related functionality including registration, authentication, and profile management. Authentication is implemented using JWT tokens; passwords are stored hashed.

**Features:** registration, login, profile management, role support (`client` / `freelancer`)

**gRPC methods:** `Register`, `Login`, `GetUser`, `UpdateUser`

**Database:** PostgreSQL — `users` table (id, email, password hash, role, name, created_at)

---

### Job Service

Manages job postings and applications. Clients create jobs, freelancers apply. When a freelancer is accepted, the service **publishes an event to RabbitMQ** (`job.accepted`) so other services can react without direct coupling.

**Features:** create/manage jobs, view listings, apply to jobs, accept a freelancer

**gRPC methods:** `CreateJob`, `GetJob`, `ListJobs`, `ApplyToJob`, `AcceptFreelancer`

**Database:** PostgreSQL — `jobs` table + `applications` table

---

### Messaging Service

Handles communication between clients and freelancers. Stores messages and allows users to view their conversations. Can also consume RabbitMQ events — for example, when a freelancer is accepted, a conversation can be automatically initiated.

**Features:** send messages, retrieve message history, list user dialogs

**gRPC methods:** `SendMessage`, `GetMessages`, `GetDialogs`

**Database:** PostgreSQL — `messages` table (sender_id, receiver_id, content, created_at)

---

## Communication

### gRPC (Synchronous)

All services expose their functionality through gRPC. The API Gateway converts incoming HTTP requests into gRPC calls and forwards them to the appropriate service. This provides efficient, strongly typed, low-latency communication.

### RabbitMQ (Asynchronous / Event-Driven)

RabbitMQ is used for asynchronous communication between services. When an important event occurs — such as a freelancer being accepted for a job — the Job Service publishes a message to the `jobs` exchange. Other services consume these messages and react accordingly without direct dependencies on each other.

```
Job Service  ──publish──▶  RabbitMQ (jobs exchange)  ──consume──▶  Messaging Service
                                                                   (auto-create dialog)
```

---

## API Gateway

The API Gateway (`job-gateway :8080`) is the single entry point for all client HTTP requests. It:
- Routes requests to the correct gRPC service
- Validates JWT tokens
- Converts HTTP → gRPC calls
- Contains no business logic

---

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) with Linux containers enabled
- [Postman](https://www.postman.com/downloads/) for API testing

---

## Running the Project

```bash
# Clone and start everything
git clone <repo-url>
cd job-service
docker compose up --build
```

All services will start in the correct order. First boot takes ~1–2 minutes for images to build.

**Verify everything is running:**
```bash
docker compose ps
```

Expected output — all containers should show `Running` or `Healthy`:
```
job-service-job-gateway-1        Running   0.0.0.0:8080->8080/tcp
job-service-messaging-gateway-1  Running   0.0.0.0:8081->8081/tcp
job-service-user-service-1       Running   0.0.0.0:50051->50051/tcp
job-service-job-service-1        Running   0.0.0.0:50052->50052/tcp
job-service-messaging-service-1  Running   0.0.0.0:50053->50053/tcp
job-service-postgres-users-1     Healthy   0.0.0.0:5432->5432/tcp
job-service-postgres-jobs-1      Healthy   0.0.0.0:5433->5432/tcp
job-service-postgres-messaging-1 Healthy   0.0.0.0:5434->5432/tcp
job-service-redis-1              Running   0.0.0.0:6379->6379/tcp
job-service-rabbitmq-1           Healthy   0.0.0.0:5672->5672/tcp
```

---

## API Reference

Base URL: `http://localhost:8080`

### Health

```
GET /healthz
```
Returns `200 OK` if the gateway is up.

---

### Users

#### Register
```
POST /users/register
```
```json
{
  "email": "user@example.com",
  "password": "secret123",
  "name": "John Doe",
  "role": "client"
}
```
> `role` can be `"client"` or `"freelancer"`

**Response `201`:**
```json
{
  "user_id": "uuid",
  "token": "jwt-token"
}
```

#### Login
```
POST /users/login
```
```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```
**Response `200`:**
```json
{
  "token": "jwt-token",
  "user_id": "uuid"
}
```

#### Get User
```
GET /users/{user_id}
Authorization: Bearer <token>
```

#### Update User
```
PUT /users/{user_id}
Authorization: Bearer <token>
```
```json
{
  "name": "New Name"
}
```

---

### Jobs

#### Create Job
```
POST /jobs
Authorization: Bearer <token>
```
```json
{
  "title": "Go Developer",
  "description": "Need experienced Go developer for 3-month project",
  "budget": 1500,
  "client_id": "<user_id>"
}
```

#### List Jobs
```
GET /jobs?page=1&page_size=10
```

#### Get Job
```
GET /jobs/{job_id}
```

#### Apply to Job
```
POST /jobs/{job_id}/apply
Authorization: Bearer <token>
```
```json
{
  "freelancer_id": "<user_id>",
  "cover_letter": "I am a perfect fit for this role."
}
```

#### Accept Freelancer
```
POST /jobs/{job_id}/accept
Authorization: Bearer <token>
```
```json
{
  "freelancer_id": "<freelancer_user_id>"
}
```

---

### Messaging

Base URL: `http://localhost:8081`

All messaging endpoints require the `X-User-ID` header.

#### Send Message
```
POST /api/messages
X-User-ID: <your_user_id>
```
```json
{
  "receiver_id": "<other_user_id>",
  "content": "Hello!"
}
```

#### Get Messages (conversation)
```
GET /api/messages?other_user_id=<other_user_id>
X-User-ID: <your_user_id>
```

#### Get Dialogs (all conversations)
```
GET /api/dialogs
X-User-ID: <your_user_id>
```

---

## Testing with Postman

1. Open Postman → click **Import**
2. Import the file `postman_collection.json` (recreate if missing — see below)
3. Run requests in this order:
   - `Health → Healthz` — verify gateway is live
   - `Users → Register` — token and user_id are saved automatically
   - `Users → Login` — refreshes token
   - `Jobs → Create Job` — job_id is saved automatically
   - `Jobs → List Jobs`, `Get Job`
   - `Messaging → Send Message` (replace `receiver-user-id` with a real user ID)

Collection variables are set automatically via Postman test scripts after Register/Login/Create Job.

---

## Viewing Logs

```bash
# All services
docker compose logs -f

# Individual service
docker compose logs job-gateway --tail=50
docker compose logs user-service --tail=50
docker compose logs job-service --tail=50
docker compose logs messaging-service --tail=50
docker compose logs messaging-gateway --tail=50
```

---

## Stopping

```bash
# Stop all containers
docker compose down

# Stop and remove all data (volumes)
docker compose down -v
```

---

## Project Structure

```
job-service/                     ← repo root
├── docker-compose.yml           ← unified orchestration for all services
├── README.md
├── user-service/                ← gRPC user service (Go 1.22)
│   ├── cmd/server/main.go
│   ├── proto/user/
│   ├── migrations/
│   └── Dockerfile
├── job-service/                 ← gRPC job service + HTTP gateway (Go 1.22)
│   ├── cmd/server/main.go
│   ├── cmd/gateway/main.go      ← main API gateway (proxies all 3 services)
│   ├── proto/job/
│   ├── proto/user/              ← copied proto stubs from user-service
│   ├── proto/messaging/         ← copied proto stubs from messaging_service
│   ├── migrations/
│   └── Dockerfile
└── messaging_service/           ← gRPC messaging service + Gin HTTP gateway (Go 1.24)
    ├── cmd/server/main.go
    ├── cmd/api_gateway/main.go
    ├── pkg/messaging/
    ├── proto/
    └── Dockerfile
```

---

## Environment Variables

### user-service
| Variable | Default | Description |
|---|---|---|
| `GRPC_PORT` | `50051` | gRPC listen port |
| `DB_HOST` | `postgres-users` | Postgres host |
| `DB_NAME` | `userservice` | Database name |
| `JWT_SECRET` | — | **Change in production** |
| `JWT_EXPIRATION_HOURS` | `72` | Token lifetime |

### job-service
| Variable | Default | Description |
|---|---|---|
| `GRPC_PORT` | `:50052` | gRPC listen port |
| `DATABASE_URL` | — | Postgres connection string |
| `REDIS_URL` | `redis://redis:6379` | Redis connection |
| `RABBITMQ_URL` | `amqp://guest:guest@rabbitmq:5672/` | RabbitMQ connection |
| `SMTP_HOST` / `SMTP_PORT` | — | Email notifications (optional) |

### job-gateway
| Variable | Default | Description |
|---|---|---|
| `GATEWAY_HTTP_ADDR` | `:8080` | HTTP listen address |
| `JOB_SERVICE_GRPC_ADDR` | `job-service:50052` | Job service gRPC address |
| `USER_SERVICE_GRPC_ADDR` | `user-service:50051` | User service gRPC address |
| `MESSAGING_SERVICE_GRPC_ADDR` | `messaging-service:50053` | Messaging service gRPC address |

### messaging-service
| Variable | Default | Description |
|---|---|---|
| `PORT` | `50053` | gRPC listen port |
| `DATABASE_URL` | — | Postgres GORM connection string |

### messaging-gateway
| Variable | Default | Description |
|---|---|---|
| `API_PORT` | `8081` | HTTP listen port |
| `GRPC_TARGET` | `messaging-service:50053` | Messaging service gRPC address |

---

## Tech Stack

| Technology | Purpose |
|---|---|
| **Go 1.22 / 1.24** | All services |
| **gRPC + Protobuf** | Inter-service communication |
| **PostgreSQL 16** | Persistent storage (separate DB per service) |
| **Redis 7** | Caching (job-service) |
| **RabbitMQ 3** | Async event-driven communication |
| **Gin** | HTTP framework for messaging gateway |
| **GORM** | ORM for messaging service |
| **JWT** | Authentication (user-service) |
| **Docker / Docker Compose** | Containerization and local orchestration |

---

## Conclusion

This project demonstrates how a microservices architecture can be implemented in practice using Go. It shows how to separate responsibilities between services, use gRPC for synchronous communication, and implement event-driven behavior with RabbitMQ.

Each service has its own database, its own Dockerfile, and can be developed and scaled independently. The API Gateway acts as a single entry point and keeps the client interface simple regardless of how many services exist internally.

The system is intentionally simple but structured in a way that allows it to be extended with additional features — such as notifications, search, payments, or a frontend — without restructuring the existing services.
