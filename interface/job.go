package _interface

import (
	"context"
	"errors"

	"jobqueue/entity"
)

// ErrJobNotFound is returned by repositories when no job matches the
// requested identifier.
var ErrJobNotFound = errors.New("job not found")

var (
	// ErrInvalidTask is returned when an enqueue request does not contain a
	// usable task name.
	ErrInvalidTask = errors.New("task name is required")
	// ErrJobServiceShuttingDown prevents jobs from being accepted after
	// graceful shutdown has started.
	ErrJobServiceShuttingDown = errors.New("job service is shutting down")
	// ErrJobRepositoryNotConfigured makes a bad service configuration fail
	// gracefully instead of panicking on a nil repository.
	ErrJobRepositoryNotConfigured = errors.New("job repository is not configured")
)

type JobService interface {
	// Enqueue registers a new job for the given task and starts processing
	// it in the background. It returns the job ID. If a job with the same
	// task is still active (including between retry attempts), its ID is
	// returned instead — the same logical task is never processed twice
	// concurrently.
	Enqueue(ctx context.Context, taskName string) (string, error)
	GetAllJobs(ctx context.Context) ([]*entity.Job, error)
	GetJobByID(ctx context.Context, id string) (*entity.Job, error)
	GetJobStatus(ctx context.Context) (entity.JobStatus, error)
	// Shutdown waits for all in-flight job workers to finish, or until ctx
	// is done, whichever comes first.
	Shutdown(ctx context.Context) error
}

type JobRepository interface {
	Save(ctx context.Context, job *entity.Job) error
	FindByID(ctx context.Context, id string) (*entity.Job, error)
	FindAll(ctx context.Context) ([]*entity.Job, error)
}
