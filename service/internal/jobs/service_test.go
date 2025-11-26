package jobs

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	testredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

var service *Service

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

	config := &Config{redisAddr: addr}
	storage, err := NewStorage(config)
	if err != nil {
		panic(err)
	}

	service, err = NewService(config, storage)
	if err != nil {
		panic(err)
	}

	exitCode := m.Run()

	redisContainer.Terminate(ctx)

	os.Exit(exitCode)
}

func setupTest(t *testing.T) {
	ctx := context.Background()
	service.storage.redisClient.FlushDB(ctx)
	t.Cleanup(func() {
		service.storage.redisClient.FlushDB(ctx)
	})
}

func TestEnqueueJob(t *testing.T) {
	setupTest(t)

	ctx := context.Background()
	now := time.Now().UnixMilli()

	request := &EnqueueJobRequest{
		Type:          "test",
		Payload:       []byte("test-payload"),
		ExecutionTime: &now,
	}

	job, err := service.EnqueueJob(ctx, request)
	if err != nil {
		t.Fatalf("service.EnqueueJob failed: %v", err)
	}

	if len(job.ID) <= 0 {
		t.Fatal("service.EnqueueJob return a job with no ID")
	}

	if job.Type != "test" {
		t.Fatalf("expected Type %v, got %v", "test", job.Type)
	}

	if job.ExecutionTime != now {
		t.Fatalf("expected ExecutionTime %v, got %v", now, job.ExecutionTime)
	}

	if job.Status != JobStatusPending {
		t.Fatalf("expected Status %v, got %v", JobStatusPending, job.Status)
	}

	if job.CreatedAt < now {
		t.Fatalf("expected CreatedAt >= %v, got %v", now, job.CreatedAt)
	}

	if job.UpdatedAt < now {
		t.Fatalf("expected UpdatedAt >= %v, got %v", now, job.UpdatedAt)
	}
}

func TestGetJob(t *testing.T) {
	setupTest(t)

	ctx := context.Background()
	now := time.Now().UnixMilli()

	request := &EnqueueJobRequest{
		Type:          "test",
		Payload:       []byte("test-payload"),
		ExecutionTime: &now,
	}

	job, err := service.EnqueueJob(ctx, request)
	if err != nil {
		t.Fatalf("service.EnqueueJob failed: %v", err)
	}

	savedJob, err := service.GetJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("service.GetJob failed: %v", err)
	}

	if diff := cmp.Diff(job, savedJob); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestDeleteJob(t *testing.T) {
	setupTest(t)

	ctx := context.Background()
	now := time.Now().UnixMilli()

	request := &EnqueueJobRequest{
		Type:          "test",
		Payload:       []byte("test-payload"),
		ExecutionTime: &now,
	}

	job, err := service.EnqueueJob(ctx, request)
	if err != nil {
		t.Fatalf("service.EnqueueJob failed: %v", err)
	}

	err = service.DeleteJob(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}

	savedJob, err := service.GetJob(ctx, job.ID)
	if err != ErrJobNotFound {
		t.Fatal("expected savedJob to return a ErrJobNotFound")
	}

	if savedJob != nil {
		t.Fatal("expected savedJob to not exist")
	}
}
