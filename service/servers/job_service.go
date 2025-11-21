package servers

import (
	"context"

	"connectrpc.com/connect"
	jobv1 "github.com/mpataki/go-job-queue/proto/gen/go/mpataki/jobqueue/v1"
)

type JobServer struct{}

func NewJobServer() *JobServer {
	return &JobServer{}
}

func (s *JobServer) CreateJob(
	ctx context.Context,
	req *connect.Request[jobv1.CreateJobRequest],
) (*connect.Response[jobv1.CreateJobResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (s *JobServer) GetJob(
	ctx context.Context,
	req *connect.Request[jobv1.GetJobRequest],
) (*connect.Response[jobv1.GetJobResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (s *JobServer) ListJobs(
	ctx context.Context,
	req *connect.Request[jobv1.ListJobsRequest],
) (*connect.Response[jobv1.ListJobsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (s *JobServer) UpdateJob(
	ctx context.Context,
	req *connect.Request[jobv1.UpdateJobRequest],
) (*connect.Response[jobv1.UpdateJobResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (s *JobServer) DeleteJob(
	ctx context.Context,
	req *connect.Request[jobv1.DeleteJobRequest],
) (*connect.Response[jobv1.DeleteJobResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
