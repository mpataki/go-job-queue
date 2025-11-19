package main

import (
	"fmt"
	"net/http"

	"connectrpc.com/grpcreflect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/mpataki/go-job-queue/proto/gen/go/mpataki/go_job_queue/proto/job/v1/jobv1connect"
	"github.com/mpataki/go-job-queue/service/servers"
)

func main() {
	mux := http.NewServeMux()

	// Initialize the job server
	jobServer := servers.NewJobServer()

	// Register the job service
	path, handler := jobv1connect.NewJobServiceHandler(jobServer)
	mux.Handle(path, handler)

	// Register reflection service
	reflector := grpcreflect.NewStaticReflector(
		jobv1connect.JobServiceName,
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
