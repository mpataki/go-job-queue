# go-job-queue

A gRPC-based job queue service built with Go and Connect-RPC.

## Project Structure

```
go-job-queue/
├── proto/              # Protocol buffer definitions
│   ├── src/proto/      # Source .proto files
│   ├── gen/            # Generated code (Go, TypeScript)
│   └── README.md       # Proto-specific documentation
├── service/            # Go service implementation
│   ├── main.go         # Server entry point
│   └── servers/        # Service handlers
├── compose.yaml        # Docker Compose configuration
└── Dockerfile          # Container build definition
```

## Prerequisites

- Go 1.25+
- [Buf CLI](https://buf.build/docs/installation) - Protocol buffer tooling
- [grpcurl](https://github.com/fullstorydev/grpcurl) - gRPC testing tool (optional)
- Docker & Docker Compose (optional)

## Getting Started

### Run with Docker Compose

```bash
docker compose up --build
```

The service will be available at `localhost:8080`.

### Run Locally

1. Generate proto code:
```bash
cd proto
buf generate
cd ..
```

2. Run the service:
```bash
cd service
go run main.go
```

## API Documentation

### Service Endpoints

The service implements the `JobService` with the following RPCs:

- `CreateJob` - Create a new job
- `GetJob` - Retrieve a job by ID
- `ListJobs` - List all jobs with pagination
- `UpdateJob` - Update an existing job
- `DeleteJob` - Delete a job by ID

### Job Status

Jobs can have the following statuses:
- `JOB_STATUS_PENDING` - Job is queued
- `JOB_STATUS_RUNNING` - Job is currently executing
- `JOB_STATUS_COMPLETED` - Job finished successfully
- `JOB_STATUS_FAILED` - Job failed

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

### Create a job

```bash
grpcurl -plaintext -d '{
  "name": "example-job",
  "payload": {"key": "value"}
}' localhost:8080 mpataki.jobqueue.v1.JobService/CreateJob
```

### Get a job

```bash
grpcurl -plaintext -d '{"id": "job-id-here"}' \
  localhost:8080 mpataki.jobqueue.v1.JobService/GetJob
```

### List jobs

```bash
grpcurl -plaintext -d '{
  "page_size": 10,
  "page_token": ""
}' localhost:8080 mpataki.jobqueue.v1.JobService/ListJobs
```

### Update a job

```bash
grpcurl -plaintext -d '{
  "id": "job-id-here",
  "name": "updated-name",
  "status": "JOB_STATUS_RUNNING"
}' localhost:8080 mpataki.jobqueue.v1.JobService/UpdateJob
```

### Delete a job

```bash
grpcurl -plaintext -d '{"id": "job-id-here"}' \
  localhost:8080 mpataki.jobqueue.v1.JobService/DeleteJob
```

## Development

### Generate Proto Code

```bash
cd proto
npm run build  # Generates Go and TypeScript/JavaScript
```

See [proto/README.md](proto/README.md) for detailed proto build documentation.

### Go Workspace

This project uses Go workspaces to manage multiple modules:

```bash
go work use ./proto ./service
go work sync
```

### Project Modules

- `proto` - Protocol buffer definitions and generated code
- `service` - Go service implementation

## Architecture

The service uses:
- **Connect-RPC** - Modern, simple gRPC alternative with HTTP/2
- **h2c** - HTTP/2 cleartext for local development (no TLS)
- **gRPC Reflection** - Runtime service discovery for tools like grpcurl
- **Buf** - Modern protobuf tooling for code generation

## License

See LICENSE file for details.
