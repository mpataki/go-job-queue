package main

import (
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/grpcreflect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/mpataki/go-job-queue/proto/gen/go/mpataki/jobqueue/v1/jobqueuev1connect"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Initialize Redis (ready for when we implement the service)
	_ = redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
	})

	mux := http.NewServeMux()

	// Initialize the job server
	jobServer := NewJobServer()

	// Register the job service
	path, handler := jobqueuev1connect.NewJobServiceHandler(jobServer)
	mux.Handle(path, handler)

	// Register reflection service
	reflector := grpcreflect.NewStaticReflector(
		jobqueuev1connect.JobServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Start the server with h2c support
	port := "8080"
	fmt.Printf("gRPC server listening on port %s...\n", port)
	if err := http.ListenAndServe(
		":"+port,
		h2c.NewHandler(mux, &http2.Server{}),
	); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)

	if !exists {
		return defaultValue
	}

	return value
}
