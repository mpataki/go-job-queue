// Package worker provides a simple SDK for processing jobs from the job queue.
//
// This worker implementation makes the following assumptions:
//   - Single-threaded execution: processes one job at a time sequentially
//   - At-least-once delivery: jobs may be executed multiple times in failure scenarios
//   - Job handlers should be idempotent to handle duplicate execution gracefully
//   - Single worker instance per job type (distributed workers to be implemented in future versions)
//
// Example usage:
//
//	handler := func(ctx context.Context, job *jobs.Job) error {
//	    log.Printf("Processing job: %s", string(job.Payload))
//	    // Perform work here
//	    return nil
//	}
//
//	w, err := worker.NewWorker("email", handler)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
//	defer cancel()
//
//	if err := w.Start(ctx); err != nil {
//	    log.Fatal(err)
//	}
package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mpataki/go-job-queue/service/internal/jobs"
)

type HandlerFunc func(ctx context.Context, job *jobs.Job) error

type Worker struct {
	service *jobs.Service
	jobType string
	handler HandlerFunc
	logger  *log.Logger
}

func NewWorker(jobType string, handler HandlerFunc) (*Worker, error) {
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

	logger := log.New(os.Stderr, fmt.Sprintf("[Worker:%s]", jobType), log.LstdFlags)

	return &Worker{
		service: service,
		jobType: jobType,
		handler: handler,
		logger:  logger,
	}, nil
}

func (w *Worker) Start(ctx context.Context) error {
	w.logger.Println("Starting job worker")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.poll(ctx)
		case <-ctx.Done():
			w.logger.Println("Worker shutting down")
			return nil
		}
	}
}

func (w *Worker) poll(ctx context.Context) error {
	job, err := w.service.GetExecutableJob(ctx, w.jobType)
	if err != nil {
		return err
	}

	if job == nil {
		return nil
	}

	w.service.MarkJobAsRunning(ctx, job.ID)

	err = w.handler(ctx, job)
	if err != nil {
		// We could start a retry here, for example
		w.service.MarkJobAsFailed(ctx, job.ID)
		return err
	}

	w.service.MarkJobComplete(ctx, job.ID)

	return nil
}
