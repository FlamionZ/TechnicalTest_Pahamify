package inmemrepo_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"jobqueue/entity"
	_interface "jobqueue/interface"
	inmemrepo "jobqueue/repository/inmem"
)

func newRepo() inmemrepo.Initiator {
	return inmemrepo.NewJobRepository().SetInMemConnection(map[string]*entity.Job{})
}

func TestSaveAndFindByID(t *testing.T) {
	repo := newRepo().Build()
	ctx := context.Background()

	job := &entity.Job{ID: "job-1", Task: "task-1", Status: entity.StatusPending, CreatedAt: time.Now()}
	if err := repo.Save(ctx, job); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := repo.FindByID(ctx, "job-1")
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if got.ID != "job-1" || got.Task != "task-1" || got.Status != entity.StatusPending {
		t.Fatalf("unexpected job: %+v", got)
	}
}

func TestFindByIDNotFound(t *testing.T) {
	repo := newRepo().Build()

	_, err := repo.FindByID(context.Background(), "missing")
	if !errors.Is(err, _interface.ErrJobNotFound) {
		t.Fatalf("want ErrJobNotFound, got %v", err)
	}
}

func TestBuildInitializesStorageWhenConnectionIsOmitted(t *testing.T) {
	repo := inmemrepo.NewJobRepository().Build()
	job := &entity.Job{ID: "job-1", Task: "task-1", Status: entity.StatusPending}

	if err := repo.Save(context.Background(), job); err != nil {
		t.Fatalf("save with default storage: %v", err)
	}
	if _, err := repo.FindByID(context.Background(), job.ID); err != nil {
		t.Fatalf("find with default storage: %v", err)
	}
}

// Mutating a job after it is saved (or after it is read) must never change
// the stored data — the repository hands out copies.
func TestRepositoryReturnsCopies(t *testing.T) {
	repo := newRepo().Build()
	ctx := context.Background()

	original := &entity.Job{ID: "job-1", Task: "task-1", Status: entity.StatusPending, CreatedAt: time.Now()}
	if err := repo.Save(ctx, original); err != nil {
		t.Fatalf("save: %v", err)
	}
	original.Status = entity.StatusCompleted

	got, err := repo.FindByID(ctx, "job-1")
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if got.Status != entity.StatusPending {
		t.Fatalf("stored job was mutated through the caller's pointer: %+v", got)
	}

	got.Status = entity.StatusFailed
	again, err := repo.FindByID(ctx, "job-1")
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if again.Status != entity.StatusPending {
		t.Fatalf("stored job was mutated through the returned pointer: %+v", again)
	}
}

func TestFindAll(t *testing.T) {
	repo := newRepo().Build()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		job := &entity.Job{ID: fmt.Sprintf("job-%d", i), Task: "task", Status: entity.StatusPending, CreatedAt: time.Now()}
		if err := repo.Save(ctx, job); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	jobs, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("find all: %v", err)
	}
	if len(jobs) != 3 {
		t.Fatalf("want 3 jobs, got %d", len(jobs))
	}
}

// Concurrent saves, reads and list queries must be race-free and never lose
// a write.
func TestConcurrentAccess(t *testing.T) {
	repo := newRepo().Build()
	ctx := context.Background()

	const workers = 50
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			job := &entity.Job{
				ID:        fmt.Sprintf("job-%d", i),
				Task:      "task",
				Status:    entity.StatusPending,
				CreatedAt: time.Now(),
			}
			if err := repo.Save(ctx, job); err != nil {
				t.Errorf("save: %v", err)
			}
			if _, err := repo.FindByID(ctx, job.ID); err != nil {
				t.Errorf("find by id: %v", err)
			}
			if _, err := repo.FindAll(ctx); err != nil {
				t.Errorf("find all: %v", err)
			}
		}(i)
	}
	wg.Wait()

	jobs, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("find all: %v", err)
	}
	if len(jobs) != workers {
		t.Fatalf("want %d jobs, got %d", workers, len(jobs))
	}
}
