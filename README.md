# go-job-queue

A distributed job queue system built with Go, Redis, and gRPC. Inspired by Sidekiq, this queue allows users to define and register custom job handlers while the system manages scheduling, persistence, and execution.

## Project Structure

```
go-job-queue/
├── proto/                      # Protocol buffer definitions (Module 1)
│   ├── src/proto/             # Source .proto files
│   ├── gen/                   # Generated code (Go, TypeScript)
│   ├── package.json           # npm package for TypeScript clients
│   └── go.mod
├── server/                     # Go service implementation (Module 2)
│   ├── cmd/
│   │   ├── server/            # gRPC API server binary
│   │   └── example-worker/    # Example worker implementation
│   ├── internal/              # Private application code
│   │   └── jobs/              # Job domain logic and Redis storage
│   ├── worker/                # PUBLIC - Worker SDK for users
│   ├── client/                # PUBLIC - Client SDK for job submission
│   └── go.mod
├── compose.yaml               # Docker Compose (Redis + Server)
└── go.work                    # Go workspace configuration
```

## Prerequisites

- Go 1.25+
- Redis 7+ (or use Docker Compose)
- [Buf CLI](https://buf.build/docs/installation) - Protocol buffer tooling
- [grpcurl](https://github.com/fullstorydev/grpcurl) - gRPC testing tool (optional)
- Docker & Docker Compose (for containerized deployment)

## Getting Started

### Run with Docker Compose

```bash
docker compose up --build
```

The service will be available at `localhost:8080`.

### Run Locally

1. Start Redis:
```bash
redis-server
```

2. Generate proto code (if modified):
```bash
cd proto
buf generate
cd ..
```

3. Run the server:
```bash
go run ./server/cmd/server
```

The service will be available at `localhost:8080`.

## How It Works

### Architecture

The job queue follows a **registration pattern** similar to Sidekiq:

1. **Server** - Central gRPC API for job submission and querying (you deploy this)
2. **Client** - SDK for submitting jobs via gRPC
3. **Worker** - SDK for processing jobs (users import and register handlers)
4. **Redis** - Job storage and queue (workers poll Redis directly, not via gRPC)

### Job Flow

```
Client → gRPC → Server → Redis (job stored)
Worker → Redis (poll) → Execute Handler → Update Status
Client → gRPC → Server → Redis (query status)
```

## API Documentation

### gRPC Service Endpoints

The `JobService` provides the following RPCs:

- `EnqueueJob` - Submit a new job to the queue
- `GetJob` - Retrieve job status by ID
- `CancelJob` - Cancel a pending job

### Job Status

- `JOB_STATUS_PENDING` - Job is queued, waiting to execute
- `JOB_STATUS_RUNNING` - Job is currently executing
- `JOB_STATUS_COMPLETED` - Job finished successfully
- `JOB_STATUS_FAILED` - Job execution failed

## Testing with grpcurl

The service includes gRPC reflection for easy testing.

### List available services

```bash
grpcurl -plaintext localhost:8080 list
```

Output:
```
mpataki.jobqueue.v1.JobService
```

### Describe the JobService

```bash
grpcurl -plaintext localhost:8080 describe mpataki.jobqueue.v1.JobService
```

### List all methods

```bash
grpcurl -plaintext localhost:8080 list mpataki.jobqueue.v1.JobService
```

### Enqueue a job

```bash
grpcurl -plaintext -d '{
  "type": "send_email",
  "payload": "eyJ0byI6InRlc3RAZXhhbXBsZS5jb20ifQ==",
  "execution_time_ms": 1640000000000
}' localhost:8080 mpataki.jobqueue.v1.JobService/EnqueueJob
```

**With immediate execution** (omit execution_time_ms):
```bash
grpcurl -plaintext -d '{
  "type": "send_email",
  "payload": "eyJ0byI6InRlc3RAZXhhbXBsZS5jb20ifQ=="
}' localhost:8080 mpataki.jobqueue.v1.JobService/EnqueueJob
```

### Get job status

```bash
grpcurl -plaintext -d '{"id": "job-id-here"}' \
  localhost:8080 mpataki.jobqueue.v1.JobService/GetJob
```

### Cancel a job

```bash
grpcurl -plaintext -d '{"id": "job-id-here"}' \
  localhost:8080 mpataki.jobqueue.v1.JobService/CancelJob
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./server/internal/jobs
```

**Note:** Tests use testcontainers and require Docker to be running.

### Generate Proto Code

```bash
cd proto
npm run build  # Generates Go and TypeScript/JavaScript
```

See [proto/README.md](proto/README.md) for detailed proto build documentation.

### Go Workspace

This project uses Go workspaces to manage multiple modules:

```bash
go work use ./proto ./server
go work sync
```

### Project Modules

- `proto` - Protocol buffer definitions and generated code (multi-language)
- `server` - Go service implementation, worker SDK, and client SDK

## Architecture & Technology

### Stack
- **Connect-RPC** - Modern gRPC with HTTP/2 (h2c for local development)
- **Redis** - Job storage and queue (sorted sets for scheduling, hashes for job data)
- **Go** - Service implementation with clean architecture
- **Protocol Buffers** - Type-safe API definitions
- **Buf** - Modern protobuf tooling for code generation

### Design Patterns
- **Registration Pattern** - Users register job handlers (like Sidekiq)
- **Domain-Driven Design** - Internal domain logic separated from transport/storage
- **Package by Feature** - Code organized by job queue domain, not layers
- **Consumer-Defined Interfaces** - Storage interfaces defined by domain, not provider

## License

See LICENSE file for details.
