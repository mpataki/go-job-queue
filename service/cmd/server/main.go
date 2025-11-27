package main

import (
	"log"
	"net/http"

	"connectrpc.com/grpcreflect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/mpataki/go-job-queue/proto/gen/go/mpataki/jobqueue/v1/jobqueuev1connect"
	"github.com/mpataki/go-job-queue/service/internal/jobs"
)

func main() {
	config, err := jobs.NewConfig()

	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	storage, err := jobs.NewStorage(config)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	service, err := jobs.NewService(config, storage)
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	jobServer := NewJobServer(service)

	mux := http.NewServeMux()

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
	log.Printf("gRPC server listening on port %s...", port)
	if err := http.ListenAndServe(
		":"+port,
		h2c.NewHandler(mux, &http2.Server{}),
	); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
