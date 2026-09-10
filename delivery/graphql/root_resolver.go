package graphql

import (
	"jobqueue/delivery/graphql/mutation"
	"jobqueue/delivery/graphql/query"
)

type rootResolver struct {
	mutation.JobMutation
	query.JobQuery
}

// Initiator ...
type Initiator func(r *rootResolver) *rootResolver

// New ...
func New() Initiator {
	return func(r *rootResolver) *rootResolver {
		return r
	}
}

// SetJobMutation ...
func (i Initiator) SetJobMutation(jobMutation mutation.JobMutation) Initiator {
	return func(r *rootResolver) *rootResolver {
		i(r).JobMutation = jobMutation
		return r
	}
}

// SetJobQuery ...
func (i Initiator) SetJobQuery(jobQuery query.JobQuery) Initiator {
	return func(r *rootResolver) *rootResolver {
		i(r).JobQuery = jobQuery
		return r
	}
}

// Build ...
func (i Initiator) Build() *rootResolver {
	return i(&rootResolver{})
}
