package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/mpataki/go-job-queue/service/internal/jobs"
	testredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

var testService *jobs.Service
var testStorage *jobs.Storage

func TestMain(m *testing.M) {
	ctx := context.Background()

	redisContainer, err := testredis.Run(ctx, "redis:8-alpine")
	if err != nil {
		panic(err)
	}

	addr, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		panic(err)
	}

	// Set environment variable for worker to use
	os.Setenv("REDIS_ADDR", addr)

	config, err := jobs.NewConfig()
	if err != nil {
		panic(err)
	}

	storage, err := jobs.NewStorage(config)
	if err != nil {
		panic(err)
	}

	testService, err = jobs.NewService(config, storage)
	if err != nil {
		panic(err)
	}

	testStorage = storage

	exitCode := m.Run()

	redisContainer.Terminate(ctx)

	os.Exit(exitCode)
}

func setupTest(t *testing.T, jobType string) *Worker {
	ctx := context.Background()
	testStorage.FlushDB(ctx)
	t.Cleanup(func() {
		testStorage.FlushDB(ctx)
	})

	handler := func(ctx context.Context, j *jobs.Job) error {
		return nil
	}

	logger := log.New(os.Stderr, fmt.Sprintf("[Worker:%s]", jobType), log.LstdFlags)

	return &Worker{
		service: testService,
		jobType: jobType,
		handler: handler,
		logger:  logger,
	}
}

func TestWorkerProcessesJob(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UnixMilli()
	jobType := "test-process"

	w := setupTest(t, jobType)

	// Enqueue a job
	request := &jobs.EnqueueJobRequest{
		Type:          jobType,
		Payload:       []byte("test-payload"),
		ExecutionTime: &now,
	}
	job, err := testService.EnqueueJob(ctx, request)
	if err != nil {
		t.Fatalf("failed to enqueue job: %v", err)
	}

	// Track handler execution
	handlerCalled := false
	w.handler = func(ctx context.Context, j *jobs.Job) error {
		handlerCalled = true
		if j.ID != job.ID {
			t.Errorf("expected job ID %s, got %s", job.ID, j.ID)
		}
		return nil
	}

	// Poll once
	err = w.poll(ctx)
	if err != nil {
		t.Fatalf("poll failed: %v", err)
	}

	if !handlerCalled {
		t.Fatal("handler was not called")
	}

	// Verify job is marked complete
	savedJob, err := testService.GetJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}

	if savedJob.Status != jobs.JobStatusCompleted {
		t.Errorf("expected status %v, got %v", jobs.JobStatusCompleted, savedJob.Status)
	}
}

func TestWorkerMarksJobAsFailedOnHandlerError(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UnixMilli()
	jobType := "test-failure"

	w := setupTest(t, jobType)

	// Enqueue a job
	request := &jobs.EnqueueJobRequest{
		Type:          jobType,
		Payload:       []byte("test-payload"),
		ExecutionTime: &now,
	}
	job, err := testService.EnqueueJob(ctx, request)
	if err != nil {
		t.Fatalf("failed to enqueue job: %v", err)
	}

	// Handler that returns an error
	expectedErr := errors.New("handler failed")
	w.handler = func(ctx context.Context, j *jobs.Job) error {
		return expectedErr
	}

	// Poll once
	err = w.poll(ctx)
	if err != expectedErr {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	// Verify job is marked failed
	savedJob, err := testService.GetJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}

	if savedJob.Status != jobs.JobStatusFailed {
		t.Errorf("expected status %v, got %v", jobs.JobStatusFailed, savedJob.Status)
	}
}

func TestWorkerIgnoresJobsForOtherTypes(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UnixMilli()

	w := setupTest(t, "worker-type")

	// Enqueue a job with different type
	request := &jobs.EnqueueJobRequest{
		Type:          "other-type",
		Payload:       []byte("test-payload"),
		ExecutionTime: &now,
	}
	_, err := testService.EnqueueJob(ctx, request)
	if err != nil {
		t.Fatalf("failed to enqueue job: %v", err)
	}

	// Track handler execution
	handlerCalled := false
	w.handler = func(ctx context.Context, j *jobs.Job) error {
		handlerCalled = true
		return nil
	}

	// Poll once
	err = w.poll(ctx)
	if err != nil {
		t.Fatalf("poll failed: %v", err)
	}

	if handlerCalled {
		t.Fatal("handler should not be called for different job type")
	}
}

func TestWorkerShutdownOnContextCancellation(t *testing.T) {
	w := setupTest(t, "test-shutdown")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := w.Start(ctx)
	if err != nil {
		t.Fatalf("expected no error on shutdown, got %v", err)
	}
}
