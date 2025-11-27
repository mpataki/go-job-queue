// Package jobs provides job queue management
package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Service is the top level API that all runners and entrypoints use.
//
//	All system functionality is exposed from this this struct
type Service struct {
	config  *Config
	storage *Storage
}

func NewService(config *Config, storage *Storage) (*Service, error) {
	s := Service{
		config:  config,
		storage: storage,
	}

	return &s, nil
}

type EnqueueJobRequest struct {
	Type          string
	Payload       []byte
	ExecutionTime *int64
}

func (s *Service) EnqueueJob(ctx context.Context, request *EnqueueJobRequest) (*Job, error) {
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
	}

	job, err := s.storage.PutJob(ctx, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *Service) GetJob(ctx context.Context, id string) (*Job, error) {
	return s.storage.GetJob(ctx, id)
}

func (s *Service) DeleteJob(ctx context.Context, id string) error {
	return s.storage.DeleteJob(ctx, id)
}

func (s *Service) GetExecutableJob(ctx context.Context, jobType string) (*Job, error) {
	return s.storage.GetExecutableJob(ctx, jobType)
}

func (s *Service) MarkJobAsRunning(ctx context.Context, id string) error {
	return s.storage.SetJobStatus(ctx, id, JobStatusRunning)
}

func (s *Service) MarkJobAsFailed(ctx context.Context, id string) error {
	err := s.storage.DequeueJob(ctx, id)
	if err != nil {
		return err
	}

	err = s.storage.SetJobStatus(ctx, id, JobStatusFailed)
	if err != nil {
		return err
	}

	return s.storage.SetExpiry(ctx, id, 5*time.Minute)
}

func (s *Service) MarkJobComplete(ctx context.Context, id string) error {
	err := s.storage.DequeueJob(ctx, id)
	if err != nil {
		return err
	}

	err = s.storage.SetJobStatus(ctx, id, JobStatusCompleted)
	if err != nil {
		return err
	}

	return s.storage.SetExpiry(ctx, id, 5*time.Minute)
}
