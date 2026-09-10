package entity

import "time"

// Job lifecycle statuses. A job moves pending -> running -> completed, with
// failed as an intermediate state between retry attempts and as the terminal
// state when every attempt is exhausted.
const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusFailed    = "failed"
	StatusCompleted = "completed"
)

// UnstableJobTask is the special task that is expected to fail twice before
// it finally succeeds.
const UnstableJobTask = "unstable-job"

type Job struct {
	ID        string    `json:"id"`
	Task      string    `json:"task"`
	Status    string    `json:"status"`
	Attempts  int32     `json:"attempts"`
	CreatedAt time.Time `json:"created_at"`
}

type JobStatus struct {
	Pending   int32 `json:"pending"`
	Running   int32 `json:"running"`
	Failed    int32 `json:"failed"`
	Completed int32 `json:"completed"`
}
