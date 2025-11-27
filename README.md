# go-job-queue

A distributed job queue system built with Go, Redis, and gRPC. Supports scheduled job execution with **at-least-once delivery** semantics.

## Quick Start

```bash
# Start everything (server, worker, Redis)
docker compose up --build

# In another terminal, enqueue a job
grpcurl -plaintext -d '{
  "type": "print",
  "payload": "SGVsbG8gV29ybGQh"
}' localhost:8080 mpataki.jobqueue.v1.JobService/EnqueueJob
```

The example worker will process the job and log the payload.

## Architecture

**Components:**
- **Server** - gRPC API for job submission and querying
- **Worker SDK** - Library for building custom job processors
- **Redis** - Job storage and scheduling (sorted sets + hashes)

**Job Flow:**
```
Client → gRPC → Server → Redis (job stored)
Worker → Redis (poll) → Execute Handler → Update Status
```

**Key Properties:**
- At-least-once delivery (jobs may execute multiple times on worker crashes)
- Scheduled execution via Unix timestamps
- Completed/failed jobs expire after 24 hours
- Single-threaded worker (v1)

## API Examples

### Enqueue a Job

Execute immediately:
```bash
grpcurl -plaintext -d '{
  "type": "send_email",
  "payload": "eyJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20ifQ=="
}' localhost:8080 mpataki.jobqueue.v1.JobService/EnqueueJob
```

Schedule for later (Unix milliseconds):
```bash
grpcurl -plaintext -d '{
  "type": "send_email",
  "payload": "eyJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20ifQ==",
  "execution_time_ms": 1735689600000
}' localhost:8080 mpataki.jobqueue.v1.JobService/EnqueueJob
```

### Get Job Status

```bash
grpcurl -plaintext -d '{"id": "JOB_ID_HERE"}' \
  localhost:8080 mpataki.jobqueue.v1.JobService/GetJob
```

### Cancel a Job

```bash
grpcurl -plaintext -d '{"id": "JOB_ID_HERE"}' \
  localhost:8080 mpataki.jobqueue.v1.JobService/CancelJob
```

### List Services

```bash
grpcurl -plaintext localhost:8080 list
# Output: mpataki.jobqueue.v1.JobService
```

## Building a Custom Worker

See `service/cmd/worker/main.go` for a complete example.

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/mpataki/go-job-queue/service/worker"
    "github.com/mpataki/go-job-queue/service/internal/jobs"
)

func main() {
    // Define handler for job type "email"
    handler := func(ctx context.Context, job *jobs.Job) error {
        log.Printf("Sending email: %s", string(job.Payload))
        // Send email here
        return nil
    }

    w, err := worker.NewWorker("email", handler)
    if err != nil {
        log.Fatal(err)
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    if err := w.Start(ctx); err != nil {
        log.Fatalf("Worker error: %v", err)
    }

    log.Println("Shutdown complete")
}
```

Run with: `go run ./service/cmd/worker`

## Job Status

- `JOB_STATUS_PENDING` - Waiting for execution time
- `JOB_STATUS_RUNNING` - Currently executing
- `JOB_STATUS_COMPLETED` - Finished successfully (expires in 24h)
- `JOB_STATUS_FAILED` - Execution failed (expires in 24h)

## Development

### Project Structure

```
go-job-queue/
├── proto/              # Protocol buffer definitions
│   └── gen/           # Generated gRPC code
├── service/
│   ├── cmd/
│   │   ├── server/    # gRPC server
│   │   └── worker/    # Example worker
│   ├── internal/jobs/ # Job domain logic & Redis storage
│   └── worker/        # PUBLIC - Worker SDK
├── compose.yaml       # Docker setup
└── Dockerfile         # Server image
```

### Run Tests

```bash
go test ./...  # Requires Docker (uses testcontainers)
```

### Run Without Docker

```bash
# Terminal 1: Redis
redis-server

# Terminal 2: Server
cd service
go run ./cmd/server

# Terminal 3: Worker
cd service
go run ./cmd/worker
```

### Environment Variables

- `REDIS_ADDR` - Redis address (default: `localhost:6379`)

## Technology

- **Go 1.25+** - Server and SDK implementation
- **Connect-RPC** - gRPC over HTTP/2
- **Redis** - Job storage and scheduling
- **Protocol Buffers** - API definitions
- **Docker** - Containerization
