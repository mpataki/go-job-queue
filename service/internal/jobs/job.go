package jobs

type JobStatus string

const (
	JobStatusPending JobStatus = "pending"
	JobStatusRunning JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed JobStatus = "failed"
)

type Job struct {
	ID string
	Type string
	Payload []byte
	ExecutionTime int64
	Status JobStatus
	CreatedAt int64
	UpdatedAt int64
}
