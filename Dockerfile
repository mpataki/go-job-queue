# Build stage
FROM golang:1.25-alpine AS builder

# Create workspace directory
WORKDIR /workspace

# Copy the entire repository
COPY . .

# Set up workspace
WORKDIR /workspace/service

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Final stage
FROM alpine:3.20

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /workspace/service/main .

# Run as non-root user
RUN adduser -D appuser
USER appuser

# Command to run the application
CMD ["./main"]
