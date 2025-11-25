package main

import (
	"context"

	"connectrpc.com/connect"
	jobv1 "github.com/mpataki/go-job-queue/proto/gen/go/mpataki/jobqueue/v1"
	"github.com/mpataki/go-job-queue/service/internal/jobs"
)

type JobServer struct {
	service *jobs.Service
}

func NewJobServer(service *jobs.Service) *JobServer {
	return &JobServer{
		service: service,
	}
}

func (s *JobServer) EnqueueJob(
	ctx context.Context,
	req *connect.Request[jobv1.EnqueueJobRequest],
) (*connect.Response[jobv1.EnqueueJobResponse], error) {
	request := jobs.EnqueueJobRequest{
		Type: req.Msg.GetType(),
		Payload: req.Msg.GetPayload(),
		ExecutionTime: req.Msg.ExecutionTimeMs,
	}

	job, err := s.service.EnqueueJob(ctx, &request)

	if err != nil {
		// perhaps we can use more granular codes here as we fill out failure modes
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := &jobv1.EnqueueJobResponse{
		Job: &jobv1.Job{
			Id:              job.ID,
			Type:            job.Type,
			Payload:         job.Payload,
			Status:          domainJobStatusToProto(job.Status),
			ExecutionTimeMs: job.ExecutionTime,
			CreatedAt:       job.CreatedAt,
			UpdatedAt:       job.UpdatedAt,
		},
	}

	return connect.NewResponse(resp), nil
}

func (s *JobServer) GetJob(
	ctx context.Context,
	req *connect.Request[jobv1.GetJobRequest],
) (*connect.Response[jobv1.GetJobResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (s *JobServer) CancelJob(
	ctx context.Context,
	req *connect.Request[jobv1.CancelJobRequest],
) (*connect.Response[jobv1.CancelJobResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func domainJobStatusToProto(status jobs.JobStatus) jobv1.JobStatus {
	switch status {
	case jobs.JobStatusPending:
		return jobv1.JobStatus_JOB_STATUS_PENDING
	case jobs.JobStatusRunning:
		return jobv1.JobStatus_JOB_STATUS_RUNNING
	case jobs.JobStatusFailed:
		return jobv1.JobStatus_JOB_STATUS_FAILED
	case jobs.JobStatusCompleted:
		return jobv1.JobStatus_JOB_STATUS_COMPLETED
	default:
		return jobv1.JobStatus_JOB_STATUS_UNSPECIFIED
	}
}
