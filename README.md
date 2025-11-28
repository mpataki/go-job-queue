# go-job-queue

A distributed job queue system built with Go, Redis, and Connect. Supports scheduled job execution with **at-least-once delivery** semantics.

## Quick Start

```bash
# Start everything (server, worker, Redis)
docker compose up --build

# In another terminal, enqueue a job using the CLI
go run ./service/cmd/cli submit --type print --payload "Hello World"

# Or use grpcurl directly
grpcurl -plaintext -d '{
  "type": "print",
  "payload": "SGVsbG8gV29ybGQh"
}' localhost:8080 mpataki.jobqueue.v1.JobService/EnqueueJob
```

The example worker will process the job and log the payload.

## Architecture

**Components:**
- **Server** - Connect API for job submission and querying (supports Connect, gRPC, and gRPC-Web protocols)
- **Worker SDK** - Library for building custom job processors
- **Redis** - Job storage and scheduling (sorted sets + hashes)

**Job Flow:**
```
Client → Connect/gRPC → Server → Redis (job stored)
Worker → Redis (poll) → Execute Handler → Update Status
```

**Key Properties:**
- At-least-once delivery (jobs may execute multiple times on worker crashes)
- Scheduled execution via Unix timestamps
- Completed/failed jobs expire after 24 hours
- Single-threaded worker (v1)

## CLI Tool

A command-line client for submitting and managing jobs:

```bash
# Build the CLI
cd service
go build -o job ./cmd/cli

# Submit a job (executes immediately)
./job submit --type print --payload "Hello World"

# Schedule a job for later
./job submit --type email --payload "test@example.com" --at 1735689600000

# Get job status
./job get --id <job-id>

# Cancel a job
./job cancel --id <job-id>

# Connect to different server
./job submit --type print --payload "hello" --server http://prod:8080

# Get help
./job --help
./job submit --help
```

### CLI Commands

**Global Flags:**
- `--server, -s` - Server address (default: `http://localhost:8080`)

**Commands:**

- `submit` - Enqueue a job with type and payload
  - `--type` (required) - Job type
  - `--payload` (required) - Job payload as string
  - `--at` (optional) - Execution time in Unix milliseconds (default: now)

- `get` - Get job status and details
  - `--id` (required) - Job ID

- `cancel` - Cancel a pending or running job
  - `--id` (required) - Job ID

## API Examples (grpcurl)

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
│   │   ├── server/    # Connect server
│   │   ├── worker/    # Example worker
│   │   └── cli/       # CLI tool
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

# Terminal 4: Submit a job using CLI
go run ./service/cmd/cli submit --type print --payload "Hello World"
```

### Environment Variables

- `REDIS_ADDR` - Redis address (default: `localhost:6379`)

## Technology

- **Go 1.25+** - Server and SDK implementation
- **Connect-RPC** - Protocol-agnostic RPC framework
  - Supports Connect, gRPC, and gRPC-Web protocols
  - Works over HTTP/1.1 and HTTP/2
  - Browser-compatible without proxy (no Envoy needed)
- **Redis** - Job storage and scheduling
- **Protocol Buffers** - API definitions
- **Docker** - Containerization
