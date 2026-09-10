package resolver

import (
	_dataloader "jobqueue/delivery/graphql/dataloader"
	"jobqueue/entity"
	_interface "jobqueue/interface"
)

type JobResolver struct {
	Data       entity.Job
	JobService _interface.JobService
	Dataloader *_dataloader.GeneralDataloader
}

type JobStatusResolver struct {
	Data       entity.JobStatus
	JobService _interface.JobService
	Dataloader *_dataloader.GeneralDataloader
}

// ID ....
func (q JobResolver) ID() string {
	return q.Data.ID
}

// Task ....
func (q JobResolver) Task() string {
	return q.Data.Task
}

// Status ....
func (q JobResolver) Status() string {
	return q.Data.Status
}

// Attempts ....
func (q JobResolver) Attempts() int32 {
	return q.Data.Attempts
}

// Pending ...
func (t JobStatusResolver) Pending() int32 {
	return t.Data.Pending
}

// Running ...
func (t JobStatusResolver) Running() int32 {
	return t.Data.Running
}

// Failed ...
func (t JobStatusResolver) Failed() int32 {
	return t.Data.Failed
}

// Completed ...
func (t JobStatusResolver) Completed() int32 {
	return t.Data.Completed
}
