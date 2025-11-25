// Package jobs provides job queue management
package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Service is the top level API that all runners and entrypoints use.
//   All system functionality is exposed from this this struct
type Service struct {
	config *Config
	storage *Storage
}

func NewService(config *Config, storage *Storage) (*Service, error) {
	service := Service{
		config: config,
		storage: storage,
	}

	return &service, nil
}

type EnqueueJobRequest struct {
	Type string
	Payload []byte
	ExecutionTime *int64
}

func (s *Service) EnqueueJob(ctx context.Context,  request *EnqueueJobRequest) (*Job, error) {
	executionTime := time.Now().UnixMilli()

	if request.ExecutionTime != nil {
		executionTime = *request.ExecutionTime
	}

	job := &Job{
		ID:            uuid.NewString(),
		Type:          request.Type,
		Payload:       request.Payload,
		ExecutionTime: executionTime,
		Status:        JobStatusPending,
		CreatedAt:     time.Now().UnixMilli(),
		UpdatedAt:     time.Now().UnixMilli(),
	}

	err := s.storage.PutJob(ctx, job)

	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *Service) GetJob(ctx context.Context, id string) (*Job, error) {
	return s.storage.GetJob(ctx, id)
}
