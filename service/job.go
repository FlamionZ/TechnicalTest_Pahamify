package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"jobqueue/entity"
	_interface "jobqueue/interface"

	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

const (
	// maxAttempts is the number of times a job is executed before it is
	// permanently marked as failed.
	maxAttempts = 3
	// unstablePassAttempt is the attempt on which the unstable-job task
	// finally succeeds. Attempts 1 and 2 always fail.
	unstablePassAttempt = 3

	defaultWorkTime   = 2 * time.Second
	defaultRetryDelay = 1 * time.Second
)

type jobService struct {
	jobRepo _interface.JobRepository
	logger  *logrus.Logger

	// mu serializes the check-then-create sequence inside Enqueue so two
	// simultaneous enqueues of the same task can never create duplicates.
	mu sync.Mutex
	// activeTasks is the authoritative in-process idempotency index. A task
	// remains here for its whole lifecycle, including the delay between retry
	// attempts when the persisted status is temporarily failed.
	activeTasks  map[string]string
	shuttingDown bool
	// wg tracks in-flight workers for graceful shutdown.
	wg sync.WaitGroup

	workTime   time.Duration
	retryDelay time.Duration
}

func (s *jobService) Enqueue(ctx context.Context, taskName string) (string, error) {
	taskName = strings.TrimSpace(taskName)
	if taskName == "" {
		return "", _interface.ErrInvalidTask
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if s.jobRepo == nil {
		return "", _interface.ErrJobRepositoryNotConfigured
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shuttingDown {
		return "", _interface.ErrJobServiceShuttingDown
	}

	// Fast-path for jobs created by this service. The entry is removed only
	// after the worker reaches a terminal state. Confirm the persisted state
	// here because a reader can observe the terminal status just before the
	// worker's deferred cleanup acquires s.mu.
	if id, exists := s.activeTasks[taskName]; exists {
		job, err := s.jobRepo.FindByID(ctx, id)
		switch {
		case err == nil && isJobActive(job):
			s.logger.WithField("job_id", id).
				Infof("task %q is already active, returning existing job", taskName)
			return id, nil
		case err == nil || errors.Is(err, _interface.ErrJobNotFound):
			delete(s.activeTasks, taskName)
		default:
			return "", fmt.Errorf("verify active job %s: %w", id, err)
		}
	}

	// Also inspect the repository so idempotency still works if the service is
	// built over a pre-populated store.
	jobs, err := s.jobRepo.FindAll(ctx)
	if err != nil {
		return "", fmt.Errorf("find existing jobs: %w", err)
	}
	for _, job := range jobs {
		if job.Task == taskName && isJobActive(job) {
			s.activeTasks[taskName] = job.ID
			s.logger.WithField("job_id", job.ID).
				Infof("task %q is already active with status %s, returning existing job", taskName, job.Status)
			return job.ID, nil
		}
	}

	job := &entity.Job{
		ID:        uuid.NewV4().String(),
		Task:      taskName,
		Status:    entity.StatusPending,
		CreatedAt: time.Now(),
	}
	if err := s.jobRepo.Save(ctx, job); err != nil {
		return "", fmt.Errorf("save job: %w", err)
	}

	s.activeTasks[taskName] = job.ID
	s.wg.Add(1)
	go s.processJob(*job)

	s.logger.WithField("job_id", job.ID).Infof("enqueued task %q", taskName)
	return job.ID, nil
}

// processJob runs the full lifecycle of one job in a background goroutine:
// each attempt moves the job to running, simulates work, then completes or
// fails. Failed attempts are retried with a delay until maxAttempts is hit.
func (s *jobService) processJob(job entity.Job) {
	defer func() {
		s.releaseTask(job.Task, job.ID)
		s.wg.Done()
	}()
	ctx := context.Background()

	for attempt := int32(1); attempt <= maxAttempts; attempt++ {
		if err := s.transition(ctx, job.ID, entity.StatusRunning, attempt); err != nil {
			s.logger.WithField("job_id", job.ID).
				Errorf("start attempt %d for task %q: %v", attempt, job.Task, err)
			return
		}
		s.logger.WithField("job_id", job.ID).Infof("attempt %d/%d started for task %q", attempt, maxAttempts, job.Task)
		time.Sleep(s.workTime)

		if s.jobFails(job.Task, attempt) {
			if err := s.transition(ctx, job.ID, entity.StatusFailed, attempt); err != nil {
				s.logger.WithField("job_id", job.ID).
					Errorf("record failed attempt %d for task %q: %v", attempt, job.Task, err)
				return
			}
			if attempt < maxAttempts {
				s.logger.WithField("job_id", job.ID).
					Warnf("attempt %d/%d failed for task %q, retrying in %s", attempt, maxAttempts, job.Task, s.retryDelay)
				time.Sleep(s.retryDelay)
				continue
			}
			s.logger.WithField("job_id", job.ID).
				Errorf("task %q permanently failed after %d attempts", job.Task, maxAttempts)
			return
		}

		if err := s.transition(ctx, job.ID, entity.StatusCompleted, attempt); err != nil {
			s.logger.WithField("job_id", job.ID).
				Errorf("complete task %q: %v", job.Task, err)
			return
		}
		s.logger.WithField("job_id", job.ID).
			Infof("task %q completed after %d attempt(s)", job.Task, attempt)
		return
	}
}

// jobFails encodes the retry simulation rule: the unstable-job task is
// expected to fail on its first two attempts and pass on the third.
func (s *jobService) jobFails(task string, attempt int32) bool {
	return task == entity.UnstableJobTask && attempt < unstablePassAttempt
}

// transition persists a status change for the job with the given ID.
func (s *jobService) transition(ctx context.Context, id, status string, attempt int32) error {
	job, err := s.jobRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("load job for status %q: %w", status, err)
	}
	job.Status = status
	job.Attempts = attempt
	if err := s.jobRepo.Save(ctx, job); err != nil {
		return fmt.Errorf("save status %q: %w", status, err)
	}
	return nil
}

func (s *jobService) releaseTask(task, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if activeID, exists := s.activeTasks[task]; exists && activeID == id {
		delete(s.activeTasks, task)
	}
}

func isJobActive(job *entity.Job) bool {
	if job == nil {
		return false
	}
	switch job.Status {
	case entity.StatusPending, entity.StatusRunning:
		return true
	case entity.StatusFailed:
		// Failed is an intermediate status until the final attempt is used.
		return job.Attempts < maxAttempts
	default:
		return false
	}
}

func (s *jobService) GetAllJobs(ctx context.Context) ([]*entity.Job, error) {
	if s.jobRepo == nil {
		return nil, _interface.ErrJobRepositoryNotConfigured
	}
	jobs, err := s.jobRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("find all jobs: %w", err)
	}
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.Before(jobs[j].CreatedAt)
	})
	return jobs, nil
}

func (s *jobService) GetJobByID(ctx context.Context, id string) (*entity.Job, error) {
	if s.jobRepo == nil {
		return nil, _interface.ErrJobRepositoryNotConfigured
	}
	return s.jobRepo.FindByID(ctx, id)
}

func (s *jobService) GetJobStatus(ctx context.Context) (entity.JobStatus, error) {
	if s.jobRepo == nil {
		return entity.JobStatus{}, _interface.ErrJobRepositoryNotConfigured
	}
	jobs, err := s.jobRepo.FindAll(ctx)
	if err != nil {
		return entity.JobStatus{}, fmt.Errorf("find all jobs: %w", err)
	}

	var status entity.JobStatus
	for _, job := range jobs {
		switch job.Status {
		case entity.StatusPending:
			status.Pending++
		case entity.StatusRunning:
			status.Running++
		case entity.StatusFailed:
			status.Failed++
		case entity.StatusCompleted:
			status.Completed++
		}
	}
	return status, nil
}

func (s *jobService) Shutdown(ctx context.Context) error {
	// Serialize shutdown with Enqueue. Once this flag is set no future call
	// can increment the WaitGroup, making Wait safe even under concurrency.
	s.mu.Lock()
	s.shuttingDown = true
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Initiator ...
type Initiator func(s *jobService) *jobService

// NewJobService ...
func NewJobService() Initiator {
	return func(s *jobService) *jobService {
		return s
	}
}

// SetJobRepository ...
func (i Initiator) SetJobRepository(jobRepository _interface.JobRepository) Initiator {
	return func(s *jobService) *jobService {
		i(s).jobRepo = jobRepository
		return s
	}
}

// SetLogger injects the structured logger used for job lifecycle events.
func (i Initiator) SetLogger(logger *logrus.Logger) Initiator {
	return func(s *jobService) *jobService {
		i(s).logger = logger
		return s
	}
}

// SetTiming overrides the simulated work duration and the retry delay; used
// by tests to keep them fast.
func (i Initiator) SetTiming(workTime, retryDelay time.Duration) Initiator {
	return func(s *jobService) *jobService {
		i(s).workTime = workTime
		i(s).retryDelay = retryDelay
		return s
	}
}

// Build ...
func (i Initiator) Build() _interface.JobService {
	s := i(&jobService{})
	s.activeTasks = make(map[string]string)
	if s.workTime <= 0 {
		s.workTime = defaultWorkTime
	}
	if s.retryDelay <= 0 {
		s.retryDelay = defaultRetryDelay
	}
	if s.logger == nil {
		s.logger = logrus.StandardLogger()
	}
	return s
}
