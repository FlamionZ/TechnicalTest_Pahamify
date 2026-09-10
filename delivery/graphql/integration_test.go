package graphql_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	jobgraphql "jobqueue/delivery/graphql"
	_dataloader "jobqueue/delivery/graphql/dataloader"
	"jobqueue/delivery/graphql/mutation"
	"jobqueue/delivery/graphql/query"
	"jobqueue/delivery/graphql/schema"
	"jobqueue/entity"
	_interface "jobqueue/interface"
	inmemrepo "jobqueue/repository/inmem"
	"jobqueue/service"

	_graphql "github.com/graph-gophers/graphql-go"
)

func newGraphQLSchema(t *testing.T) (*_graphql.Schema, _interface.JobService) {
	t.Helper()
	repo := inmemrepo.NewJobRepository().SetInMemConnection(map[string]*entity.Job{}).Build()
	loader := _dataloader.New().SetJobRepository(repo).SetBatchFunction().Build()
	jobService := service.NewJobService().
		SetJobRepository(repo).
		SetTiming(5*time.Millisecond, 5*time.Millisecond).
		Build()
	root := jobgraphql.New().
		SetJobMutation(mutation.NewJobMutation(jobService, loader)).
		SetJobQuery(query.NewJobQuery(jobService, loader)).
		Build()

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := jobService.Shutdown(ctx); err != nil {
			t.Errorf("shutdown job service: %v", err)
		}
	})

	return _graphql.MustParseSchema(schema.String(), root), jobService
}

func executeGraphQL(t *testing.T, schema *_graphql.Schema, document, operation string, variables map[string]any) map[string]any {
	t.Helper()
	response := schema.Exec(context.Background(), document, operation, variables)
	if len(response.Errors) != 0 {
		t.Fatalf("GraphQL errors: %+v", response.Errors)
	}

	var data map[string]any
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatalf("decode GraphQL response %s: %v", response.Data, err)
	}
	return data
}

func waitForCompletedJob(t *testing.T, service _interface.JobService, id string) *entity.Job {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		job, err := service.GetJobByID(context.Background(), id)
		if err != nil {
			t.Fatalf("get job %s: %v", id, err)
		}
		if job.Status == entity.StatusCompleted {
			return job
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("job %s did not complete", id)
	return nil
}

func TestRequiredGraphQLOperations(t *testing.T) {
	gqlSchema, jobService := newGraphQLSchema(t)

	created := executeGraphQL(t, gqlSchema, `
		mutation SimultaneousCreateJob($Job1: String!, $Job2: String!, $Job3: String!) {
			job1: Enqueue(task: $Job1) { id task status attempts }
			job2: Enqueue(task: $Job2) { id task status attempts }
			job3: Enqueue(task: $Job3) { id task status attempts }
		}`,
		"SimultaneousCreateJob",
		map[string]any{"Job1": "JobTest1", "Job2": "JobTest2", "Job3": "JobTest3"},
	)

	ids := make(map[string]bool)
	for _, field := range []string{"job1", "job2", "job3"} {
		job, ok := created[field].(map[string]any)
		if !ok {
			t.Fatalf("%s is not a job: %#v", field, created[field])
		}
		id, _ := job["id"].(string)
		if id == "" || ids[id] {
			t.Fatalf("%s returned missing or duplicate ID %q", field, id)
		}
		ids[id] = true
		waitForCompletedJob(t, jobService, id)
	}

	allJobs := executeGraphQL(t, gqlSchema, `
		query GetAllJobs { Jobs { id task status attempts } }
	`, "GetAllJobs", nil)
	jobs, ok := allJobs["Jobs"].([]any)
	if !ok || len(jobs) != 3 {
		t.Fatalf("want 3 jobs, got %#v", allJobs["Jobs"])
	}

	var knownID string
	for id := range ids {
		knownID = id
		break
	}
	byID := executeGraphQL(t, gqlSchema, `
		query GetJobById($id: String!) { Job(id: $id) { id task status attempts } }
	`, "GetJobById", map[string]any{"id": knownID})
	job, ok := byID["Job"].(map[string]any)
	if !ok || job["id"] != knownID || job["status"] != entity.StatusCompleted {
		t.Fatalf("unexpected GetJobById result: %#v", byID["Job"])
	}

	missing := executeGraphQL(t, gqlSchema, `
		query GetMissingJob { Job(id: "missing") { id } }
	`, "GetMissingJob", nil)
	if missing["Job"] != nil {
		t.Fatalf("missing job must resolve to null, got %#v", missing["Job"])
	}

	status := executeGraphQL(t, gqlSchema, `
		query GetAllJobStatus { JobStatus { pending running failed completed } }
	`, "GetAllJobStatus", nil)
	counts, ok := status["JobStatus"].(map[string]any)
	if !ok || counts["completed"] != float64(3) || counts["pending"] != float64(0) || counts["running"] != float64(0) || counts["failed"] != float64(0) {
		t.Fatalf("unexpected status counts: %#v", status["JobStatus"])
	}
}

func TestSimulateUnstableJobThroughGraphQL(t *testing.T) {
	gqlSchema, jobService := newGraphQLSchema(t)

	created := executeGraphQL(t, gqlSchema, `
		mutation SimulateUnstableJob {
			Enqueue(task: "unstable-job") { id }
		}`,
		"SimulateUnstableJob",
		nil,
	)
	jobData, ok := created["Enqueue"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected enqueue result: %#v", created["Enqueue"])
	}
	id, _ := jobData["id"].(string)
	job := waitForCompletedJob(t, jobService, id)
	if job.Attempts != 3 {
		t.Fatalf("unstable job must complete on attempt 3, got %+v", job)
	}
}
