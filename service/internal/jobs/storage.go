package jobs

import (
	"context"
	"fmt"
	"strconv"
	"time"

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

func (s *Storage) PutJob(ctx context.Context, job *Job) (*Job, error) {
	jobKey := jobKey(job.ID)
	now := time.Now().UnixMilli()
	createdAt := now

	c, err := s.redisClient.HGet(ctx, jobKey, "created_at").Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("storage.PutJob failed to HGet the job: %w", err)
	}
	if len(c) > 0 {
		createdAt, err = strconv.ParseInt(c, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("storage.PutJob failed parse the createdAt time: %w", err)
		}
	}

	toReturn := Job{
		ID:            job.ID,
		Type:          job.Type,
		Payload:       job.Payload,
		ExecutionTime: job.ExecutionTime,
		Status:        job.Status,
		CreatedAt:     createdAt,
		UpdatedAt:     now,
	}

	err = s.redisClient.HSet(ctx, jobKey, map[string]any{
		"type":       job.Type,
		"payload":    job.Payload,
		"status":     string(job.Status),
		"created_at": strconv.FormatInt(createdAt, 10),
		"updated_at": strconv.FormatInt(now, 10),
	}).Err()
	if err != nil {
		return nil, fmt.Errorf("storage.PutJob failed to HSet the job: %w", err)
	}

	err = s.redisClient.ZAdd(ctx, queueKey(job.Type), redis.Z{
		Score:  float64(job.ExecutionTime),
		Member: job.ID,
	}).Err()
	if err != nil {
		return nil, fmt.Errorf("storage.PutJob failed to ZAdd the job: %w", err)
	}

	return &toReturn, nil
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
	err := s.DequeueJob(ctx, id)
	if err != nil {
		return err
	}

	err = s.redisClient.Del(ctx, jobKey(id)).Err()
	if err != nil {
		return fmt.Errorf("storage.DeleteJob failed to Del: %w", err)
	}

	return nil
}

func (s *Storage) GetExecutableJob(ctx context.Context, jobType string) (*Job, error) {
	queueKey := queueKey(jobType)
	now := time.Now().UnixMilli()

	results, err := s.redisClient.ZRangeByScore(ctx, queueKey, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    strconv.FormatInt(now, 10),
		Offset: 0,
		Count:  1,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ZRangeByScore: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	jobID := results[0]

	job, err := s.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// Note that this isn't automic, but could use a lua script in the future
func (s *Storage) SetJobStatus(ctx context.Context, jobID string, status JobStatus) error {
	jobKey := jobKey(jobID)

	exists, err := s.redisClient.Exists(ctx, jobKey).Result()
	if err != nil {
		return err
	}

	if exists == 0 {
		return ErrJobNotFound
	}

	return s.redisClient.HSet(ctx, jobKey, "status", string(status)).Err()
}

func (s *Storage) DequeueJob(ctx context.Context, id string) error {
	jobKey := jobKey(id)

	m, err := s.redisClient.HGetAll(ctx, jobKey).Result()
	if err != nil {
		return fmt.Errorf("storage.DeleteJob failed to HGetAll: %w", err)
	}

	err = s.redisClient.ZRem(ctx, queueKey(m["type"]), id).Err()
	if err != nil {
		return fmt.Errorf("storage.DeleteJob failed to ZRem: %w", err)
	}

	return nil
}

func (s *Storage) SetExpiry(ctx context.Context, id string, duration time.Duration) error {
	return s.redisClient.Expire(ctx, jobKey(id), duration).Err()
}

func (s *Storage) FlushDB(ctx context.Context) error {
	return s.redisClient.FlushDB(ctx).Err()
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
