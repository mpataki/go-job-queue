package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mpataki/go-job-queue/service/internal/jobs"
	"github.com/mpataki/go-job-queue/service/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	w, err := worker.NewWorker("print", jobHandler)
	if err != nil {
		log.Fatalf("Failed to initialize Worker: %v", err)
	}

	if err := w.Start(ctx); err != nil {
		log.Fatalf("Worker error: %v", err)
	}

	log.Println("Shutdown complete")
}

func jobHandler(ctx context.Context, job *jobs.Job) error {
	fmt.Printf("Running job '%s'. Sleeping to simulate hard work\n", job.ID)
	time.Sleep(5 * time.Second)

	fmt.Printf("Handling print job: %s\n", job.Payload)

	return nil
}
