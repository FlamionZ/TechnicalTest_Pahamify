package service_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"jobqueue/entity"
	_interface "jobqueue/interface"
	inmemrepo "jobqueue/repository/inmem"
	"jobqueue/service"

	"github.com/sirupsen/logrus"
)

// newTestService builds a service against a fresh in-memory repository with
// short work/retry durations so lifecycle tests stay fast.
func newTestService(t *testing.T, workTime, retryDelay time.Duration) _interface.JobService {
	t.Helper()
	logger := logrus.New()
	logger.Out = io.Discard

	repo := inmemrepo.NewJobRepository().SetInMemConnection(map[string]*entity.Job{}).Build()
	return service.NewJobService().
		SetJobRepository(repo).
		SetLogger(logger).
		SetTiming(workTime, retryDelay).
		Build()
}

// waitForStatus polls the job until it reaches the wanted status or the
// deadline expires.
func waitForStatus(t *testing.T, svc _interface.JobService, id, want string) *entity.Job {
	t.Helper()

	ctx := context.Background()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		job, err := svc.GetJobByID(ctx, id)
		if err != nil {
			t.Fatalf("get job %s: %v", id, err)
		}
		if job.Status == want {
			return job
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("job %s did not reach status %q in time", id, want)
	return nil
}

func TestEnqueueCreatesJobAndProcessesIt(t *testing.T) {
	svc := newTestService(t, 5*time.Millisecond, 5*time.Millisecond)
	ctx := context.Background()

	id, err := svc.Enqueue(ctx, "simple-task")
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if id == "" {
		t.Fatal("enqueue returned an empty id")
	}

	// Right after enqueue the job must exist and not be completed yet.
	job, err := svc.GetJobByID(ctx, id)
	if err != nil {
		t.Fatalf("get job: %v", err)
	}
	if job.Task != "simple-task" {
		t.Fatalf("unexpected task %q", job.Task)
	}

	job = waitForStatus(t, svc, id, entity.StatusCompleted)
	if job.Attempts != 1 {
		t.Fatalf("want 1 attempt, got %d", job.Attempts)
	}
}

// The special unstable-job task must fail twice before it finally succeeds,
// ending completed with three attempts.
func TestUnstableJobFailsTwiceBeforePassing(t *testing.T) {
	svc := newTestService(t, 5*time.Millisecond, 5*time.Millisecond)
	ctx := context.Background()

	id, err := svc.Enqueue(ctx, entity.UnstableJobTask)
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	job := waitForStatus(t, svc, id, entity.StatusCompleted)
	if job.Attempts != 3 {
		t.Fatalf("want 3 attempts for unstable-job, got %d", job.Attempts)
	}
}

// A task that is still pending or running must not be enqueued twice: the
// second Enqueue returns the existing job's ID.
func TestEnqueueIsIdempotentWhileInFlight(t *testing.T) {
	svc := newTestService(t, 200*time.Millisecond, 5*time.Millisecond)
	ctx := context.Background()

	first, err := svc.Enqueue(ctx, "slow-task")
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	second, err := svc.Enqueue(ctx, "slow-task")
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	if first != second {
		t.Fatalf("idempotent enqueue returned different ids: %s vs %s", first, second)
	}

	jobs, err := svc.GetAllJobs(ctx)
	if err != nil {
		t.Fatalf("get all jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("want 1 job after duplicate enqueue, got %d", len(jobs))
	}
}

// A failed attempt is not a terminal state while retries remain. Enqueuing
// the same task during the retry delay must still return the original job.
func TestEnqueueIsIdempotentDuringRetryDelay(t *testing.T) {
	svc := newTestService(t, 10*time.Millisecond, 200*time.Millisecond)
	ctx := context.Background()

	first, err := svc.Enqueue(ctx, entity.UnstableJobTask)
	if err != nil {
		t.Fatalf("enqueue unstable job: %v", err)
	}
	waitForStatus(t, svc, first, entity.StatusFailed)

	second, err := svc.Enqueue(ctx, entity.UnstableJobTask)
	if err != nil {
		t.Fatalf("enqueue unstable job during retry delay: %v", err)
	}
	if first != second {
		t.Fatalf("retry-window enqueue created a duplicate: %s vs %s", first, second)
	}

	jobs, err := svc.GetAllJobs(ctx)
	if err != nil {
		t.Fatalf("get all jobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("want 1 unstable job during retry, got %d", len(jobs))
	}

	job := waitForStatus(t, svc, first, entity.StatusCompleted)
	if job.Attempts != 3 {
		t.Fatalf("want 3 attempts for unstable job, got %d", job.Attempts)
	}
}

// Once a job is completed, enqueuing the same task again starts a new run.
func TestEnqueueAfterCompletionCreatesNewJob(t *testing.T) {
	svc := newTestService(t, 5*time.Millisecond, 5*time.Millisecond)
	ctx := context.Background()

	first, err := svc.Enqueue(ctx, "recurring-task")
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	waitForStatus(t, svc, first, entity.StatusCompleted)

	second, err := svc.Enqueue(ctx, "recurring-task")
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if first == second {
		t.Fatal("expected a new job id after the previous run completed")
	}
}

func TestGetJobStatusCounts(t *testing.T) {
	svc := newTestService(t, 5*time.Millisecond, 5*time.Millisecond)
	ctx := context.Background()

	for _, task := range []string{"a", "b", "c", "d"} {
		if _, err := svc.Enqueue(ctx, task); err != nil {
			t.Fatalf("enqueue %q: %v", task, err)
		}
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		status, err := svc.GetJobStatus(ctx)
		if err != nil {
			t.Fatalf("get job status: %v", err)
		}
		if status.Completed == 4 {
			if status.Pending != 0 || status.Running != 0 || status.Failed != 0 {
				t.Fatalf("unexpected leftover counts: %+v", status)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("jobs did not all complete in time: %+v", status)
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// 50 simultaneous enqueues of only 5 distinct tasks must collapse into
// exactly 5 jobs and all must complete — no duplicates, no lost writes.
func TestConcurrentEnqueueIsSafe(t *testing.T) {
	svc := newTestService(t, 300*time.Millisecond, 5*time.Millisecond)
	ctx := context.Background()

	const goroutines = 50
	const tasks = 5

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			task := fmt.Sprintf("task-%d", i%tasks)
			if _, err := svc.Enqueue(ctx, task); err != nil {
				t.Errorf("enqueue %q: %v", task, err)
			}
		}(i)
	}
	wg.Wait()

	jobs, err := svc.GetAllJobs(ctx)
	if err != nil {
		t.Fatalf("get all jobs: %v", err)
	}
	if len(jobs) != tasks {
		t.Fatalf("want %d jobs after concurrent duplicate enqueues, got %d", tasks, len(jobs))
	}

	for _, job := range jobs {
		waitForStatus(t, svc, job.ID, entity.StatusCompleted)
	}
}

// The evaluation target explicitly calls for 50-100 concurrent jobs. This
// exercises the upper end with unique tasks so no request is deduplicated.
func TestOneHundredConcurrentJobsComplete(t *testing.T) {
	svc := newTestService(t, 10*time.Millisecond, time.Millisecond)
	ctx := context.Background()

	const jobsToCreate = 100
	var wg sync.WaitGroup
	for i := 0; i < jobsToCreate; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := svc.Enqueue(ctx, fmt.Sprintf("unique-task-%d", i)); err != nil {
				t.Errorf("enqueue job %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()

	if err := svc.Shutdown(context.Background()); err != nil {
		t.Fatalf("wait for jobs: %v", err)
	}

	jobs, err := svc.GetAllJobs(ctx)
	if err != nil {
		t.Fatalf("get all jobs: %v", err)
	}
	if len(jobs) != jobsToCreate {
		t.Fatalf("want %d jobs, got %d", jobsToCreate, len(jobs))
	}
	for _, job := range jobs {
		if job.Status != entity.StatusCompleted || job.Attempts != 1 {
			t.Fatalf("job did not complete cleanly: %+v", job)
		}
	}
}

func TestEnqueueRejectsBlankTask(t *testing.T) {
	svc := newTestService(t, time.Millisecond, time.Millisecond)

	if _, err := svc.Enqueue(context.Background(), "  \t\n "); !errors.Is(err, _interface.ErrInvalidTask) {
		t.Fatalf("want ErrInvalidTask, got %v", err)
	}
}

func TestGetJobByIDNotFound(t *testing.T) {
	svc := newTestService(t, 5*time.Millisecond, 5*time.Millisecond)

	_, err := svc.GetJobByID(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected an error for unknown job id")
	}
}

func TestGetAllJobsSortedByCreation(t *testing.T) {
	svc := newTestService(t, 5*time.Millisecond, 5*time.Millisecond)
	ctx := context.Background()

	for _, task := range []string{"first", "second", "third"} {
		if _, err := svc.Enqueue(ctx, task); err != nil {
			t.Fatalf("enqueue %q: %v", task, err)
		}
		time.Sleep(2 * time.Millisecond)
	}

	jobs, err := svc.GetAllJobs(ctx)
	if err != nil {
		t.Fatalf("get all jobs: %v", err)
	}
	if len(jobs) != 3 {
		t.Fatalf("want 3 jobs, got %d", len(jobs))
	}
	for i, want := range []string{"first", "second", "third"} {
		if jobs[i].Task != want {
			t.Fatalf("job %d: want task %q, got %q", i, want, jobs[i].Task)
		}
	}
}

// Shutdown must return once all in-flight workers have finished.
func TestShutdownWaitsForWorkers(t *testing.T) {
	svc := newTestService(t, 50*time.Millisecond, 5*time.Millisecond)
	ctx := context.Background()

	id, err := svc.Enqueue(ctx, "shutdown-task")
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	if err := svc.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	job, err := svc.GetJobByID(ctx, id)
	if err != nil {
		t.Fatalf("get job: %v", err)
	}
	if job.Status != entity.StatusCompleted {
		t.Fatalf("worker did not finish before shutdown returned: %+v", job)
	}
}

func TestShutdownRejectsNewJobs(t *testing.T) {
	svc := newTestService(t, time.Millisecond, time.Millisecond)
	if err := svc.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	_, err := svc.Enqueue(context.Background(), "too-late")
	if !errors.Is(err, _interface.ErrJobServiceShuttingDown) {
		t.Fatalf("want ErrJobServiceShuttingDown, got %v", err)
	}
}

func TestServiceWithoutRepositoryFailsGracefully(t *testing.T) {
	svc := service.NewJobService().Build()

	if _, err := svc.Enqueue(context.Background(), "task"); !errors.Is(err, _interface.ErrJobRepositoryNotConfigured) {
		t.Fatalf("want ErrJobRepositoryNotConfigured, got %v", err)
	}
	if _, err := svc.GetAllJobs(context.Background()); !errors.Is(err, _interface.ErrJobRepositoryNotConfigured) {
		t.Fatalf("want ErrJobRepositoryNotConfigured from GetAllJobs, got %v", err)
	}
}
