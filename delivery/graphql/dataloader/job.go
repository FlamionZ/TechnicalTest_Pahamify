package _dataloader

import (
	"context"

	"jobqueue/entity"

	"github.com/graph-gophers/dataloader/v6"
)

// JobBatchFunc resolves a batch of job IDs in a single repository pass and
// keeps result order aligned with the requested keys.
func (s GeneralDataloader) JobBatchFunc(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	results := make([]*dataloader.Result, len(keys))

	jobs, err := s.jobRepo.FindAll(ctx)
	if err != nil {
		for i := range results {
			results[i] = &dataloader.Result{Error: err}
		}
		return results
	}

	byID := make(map[string]*entity.Job, len(jobs))
	for _, job := range jobs {
		byID[job.ID] = job
	}

	for i, key := range keys {
		results[i] = &dataloader.Result{Data: byID[key.String()]}
	}
	return results
}
