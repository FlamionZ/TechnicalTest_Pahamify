# Technical Test Result
Muhammad Rakha Abimanyu
## Summary

I completed the job queue backend and the HTMX dashboard from the task. The application uses an in-memory repository, processes jobs in the background, and provides GraphQL and HTTP endpoints to create and check jobs.

## What I completed

Backend:

- Create three jobs from the `SimultaneousCreateJob` mutation.
- Process jobs in the background using goroutines.
- Return all jobs, one job by ID, and the current status counts.
- Retry `unstable-job` two times before completing it on the third attempt.
- Prevent the same task from running twice while it is still active or waiting for a retry.
- Protect the in-memory map with `sync.RWMutex`.
- Wait for running jobs during graceful shutdown.

Dashboard:

- Create three jobs using the Job1, Job2, and Job3 form.
- Create an unstable job.
- Refresh the job table and status summary every two seconds.
- Show job ID, task, status, and attempts.
- Open job details using the View button or the ID search field.

## Job flow

A new job starts as `pending` and is then processed as `running`. A normal job completes on its first attempt.

The `unstable-job` flow is:

```text
attempt 1 -> failed
attempt 2 -> failed
attempt 3 -> completed
```

If the same task is sent again while the job is still active, the service returns the existing job ID. After the previous job is finished, the task can be run again as a new job.

## How to run

From the project folder:

```bash
go test ./...
go run .
```

The application can be opened at:

- GraphiQL: `http://localhost:58579/graphiql`
- HTMX dashboard: `http://localhost:58579/jobqueue/dashboard`
- GraphQL endpoint: `http://localhost:58579/graphql`

Stop the server with `Ctrl+C`.

## Testing

I added tests for the repository, service, GraphQL operations, and HTMX handlers. The tests cover concurrent access, duplicate jobs, retry behavior, status updates, missing jobs, invalid input, and graceful shutdown. I also tested the service with 100 jobs running concurrently.

Commands used for the final check:

```bash
go test ./... -shuffle=on -count=3
go build ./...
go vet ./...
```

## Notes

The job data is stored in memory, so it will be reset when the server restarts. I kept it this way because the task asks for an in-memory job queue simulation. For a production system, I would use persistent storage and a message broker so jobs can survive application restarts and be processed by multiple worker instances.
