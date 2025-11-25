package jobs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Storage struct {
	redisClient *redis.Client
}

func NewStorage(config *Config) (*Storage, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.redisAddr,
	})

	storage := Storage{
		redisClient: redisClient,
	}

	return &storage, nil
}

func (s *Storage) PutJob(ctx context.Context, job *Job) error {
	err := s.redisClient.HSet(ctx, jobKey(job.ID), map[string]any{
		"type": job.Type,
		"payload": job.Payload,
		"status": string(job.Status),
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
	}).Err()

	if err != nil {
		return fmt.Errorf("storage.PutJob failed to HSet the job: %w", err)
	}

	err = s.redisClient.ZAdd(ctx, queueKey(job.Type), redis.Z{
		Score: float64(job.ExecutionTime),
		Member: job.ID,
	}).Err()

	if err != nil {
		return fmt.Errorf("storage.PutJob failed to ZAdd the job: %w", err)
	}

	return nil
}

func (s *Storage) GetJob(ctx context.Context, id string) (*Job, error) {
	m, err := s.redisClient.HGetAll(ctx, jobKey(id)).Result()
	if err != nil {
		return nil, fmt.Errorf("storage.GetJob failed to HGetAll: %w", err)
	}

	if len(m) == 0 {
		return nil, ErrJobNotFound
	}

	createdAt, err := strconv.ParseInt(m["created_at"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("storage.GetJob failed to ParseInt on the created_at field: %w", err)
	}

	updatedAt, err := strconv.ParseInt(m["updated_at"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("storage.GetJob failed to ParseInt on the updated_at field: %w", err)
	}

	jobType := m["type"]

	score, err := s.redisClient.ZScore(ctx, queueKey(jobType), id).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("storage.GetJob failed to get the ZScore: %w", err)
	}

	job := Job{
		ID:            id,
		Type:          jobType,
		Payload:       []byte(m["payload"]),
		ExecutionTime: int64(score),
		Status:        jobStatusForString(m["status"]),
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}

	return &job, nil
}

func (s *Storage) DeleteJob(ctx context.Context, id string) error {
	jobKey := jobKey(id)

	m, err := s.redisClient.HGetAll(ctx, jobKey).Result()
	if err != nil {
		return fmt.Errorf("storage.DeleteJob failed to HGetAll: %w", err)
	}

	err = s.redisClient.ZRem(ctx, queueKey(m["type"]), id).Err()
	if err != nil {
		return fmt.Errorf("storage.DeleteJob failed to ZRem: %w", err)
	}

	err = s.redisClient.Del(ctx, jobKey).Err()
	if err != nil {
		return fmt.Errorf("storage.DeleteJob failed to Del: %w", err)
	}

	return nil
}

func jobKey(id string) string {
	return "job:" + id
}

func queueKey(jobType string) string {
	return "queue:" + jobType
}

func jobStatusForString(s string) JobStatus {
	switch s {
	case "pending":
		return JobStatusPending
	case "running":
		return JobStatusRunning
	case "completed":
		return JobStatusCompleted
	case "failed":
		return JobStatusFailed
	default:
		return JobStatusUnspecified
	}
}
