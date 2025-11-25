package jobs

import (
	"context"

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
	jobKey := "job:" + job.ID

	err := s.redisClient.HSet(ctx, jobKey, map[string]any{
		"type": job.Type,
		"payload": job.Payload,
		"created_at": job.CreatedAt,
		"updated_at": job.UpdatedAt,
	}).Err()

	if err != nil {
		return err
	}

	queueKey := "queue:" + job.Type

	err = s.redisClient.ZAdd(ctx, queueKey, redis.Z{
		Score: float64(job.ExecutionTime),
		Member: job.ID,
	}).Err()

	if err != nil {
		return err
	}

	return nil
}

// func (s *Storage) GetJob() (*Job, error) {
// }
//
// func (s *Storage) DeleteJob() error {
// }
