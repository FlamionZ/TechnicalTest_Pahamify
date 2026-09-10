package inmemrepo

import (
	"context"

	"jobqueue/entity"
	_interface "jobqueue/interface"
	"sync"
)

type jobRepository struct {
	mu      sync.RWMutex
	inMemDb map[string]*entity.Job
}

// Save Job. A copy is stored so later mutations of the caller's struct never
// leak into the repository.
func (t *jobRepository) Save(ctx context.Context, job *entity.Job) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	cp := *job
	t.inMemDb[job.ID] = &cp
	return nil
}

// Find Job By ID. The returned job is a copy, safe to read and mutate outside
// the repository lock.
func (t *jobRepository) FindByID(ctx context.Context, id string) (*entity.Job, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	job, exists := t.inMemDb[id]
	if !exists {
		return nil, _interface.ErrJobNotFound
	}
	cp := *job
	return &cp, nil
}

// FindAll Job. Returns copies of every stored job.
func (t *jobRepository) FindAll(ctx context.Context) ([]*entity.Job, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	jobs := make([]*entity.Job, 0, len(t.inMemDb))
	for _, job := range t.inMemDb {
		cp := *job
		jobs = append(jobs, &cp)
	}
	return jobs, nil
}

// Initiator ...
type Initiator func(s *jobRepository) *jobRepository

// NewJobRepository ...
func NewJobRepository() Initiator {
	return func(q *jobRepository) *jobRepository {
		return q
	}
}

// SetInMemConnection set database client connection
func (i Initiator) SetInMemConnection(inMemDb map[string]*entity.Job) Initiator {
	return func(s *jobRepository) *jobRepository {
		i(s).inMemDb = inMemDb
		return s
	}
}

// Build ...
func (i Initiator) Build() _interface.JobRepository {
	repo := i(&jobRepository{})
	if repo.inMemDb == nil {
		repo.inMemDb = make(map[string]*entity.Job)
	}
	return repo
}
