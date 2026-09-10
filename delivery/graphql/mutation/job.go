package mutation

import (
	"context"

	_dataloader "jobqueue/delivery/graphql/dataloader"
	"jobqueue/delivery/graphql/resolver"
	_interface "jobqueue/interface"
)

type JobMutation struct {
	jobService _interface.JobService
	dataloader *_dataloader.GeneralDataloader
}

func (m JobMutation) Enqueue(ctx context.Context, args struct {
	Task string
}) (*resolver.JobResolver, error) {
	id, err := m.jobService.Enqueue(ctx, args.Task)
	if err != nil {
		return nil, err
	}

	job, err := m.jobService.GetJobByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &resolver.JobResolver{
		Data:       *job,
		JobService: m.jobService,
		Dataloader: m.dataloader,
	}, nil
}

// NewJobMutation to create new instance
func NewJobMutation(jobService _interface.JobService, dataloader *_dataloader.GeneralDataloader) JobMutation {
	return JobMutation{
		jobService: jobService,
		dataloader: dataloader,
	}
}
