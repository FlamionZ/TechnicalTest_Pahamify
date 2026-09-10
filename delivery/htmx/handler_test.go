package htmx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"jobqueue/entity"
	_interface "jobqueue/interface"

	"github.com/labstack/echo/v4"
)

type fakeJobService struct {
	mu            sync.Mutex
	jobs          []*entity.Job
	enqueueErrFor map[string]error
	activeCalls   int
	maxConcurrent int
}

func (s *fakeJobService) Enqueue(_ context.Context, task string) (string, error) {
	s.mu.Lock()
	s.activeCalls++
	if s.activeCalls > s.maxConcurrent {
		s.maxConcurrent = s.activeCalls
	}
	s.mu.Unlock()

	// Keep calls overlapped long enough to verify that CreateJobs launches the
	// three enqueue operations concurrently.
	time.Sleep(10 * time.Millisecond)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeCalls--
	if err := s.enqueueErrFor[task]; err != nil {
		return "", err
	}
	id := fmt.Sprintf("id-%s", task)
	s.jobs = append(s.jobs, &entity.Job{ID: id, Task: task, Status: entity.StatusPending})
	return id, nil
}

func (s *fakeJobService) GetAllJobs(context.Context) ([]*entity.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	jobs := make([]*entity.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		cp := *job
		jobs = append(jobs, &cp)
	}
	return jobs, nil
}

func (s *fakeJobService) GetJobByID(_ context.Context, id string) (*entity.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, job := range s.jobs {
		if job.ID == id {
			cp := *job
			return &cp, nil
		}
	}
	return nil, _interface.ErrJobNotFound
}

func (s *fakeJobService) GetJobStatus(context.Context) (entity.JobStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	status := entity.JobStatus{}
	for _, job := range s.jobs {
		switch job.Status {
		case entity.StatusPending:
			status.Pending++
		case entity.StatusRunning:
			status.Running++
		case entity.StatusFailed:
			status.Failed++
		case entity.StatusCompleted:
			status.Completed++
		}
	}
	return status, nil
}

func (*fakeJobService) Shutdown(context.Context) error { return nil }

func newDashboardHandler(t *testing.T, service _interface.JobService) *DashboardHandler {
	t.Helper()
	handler, err := NewDashboardHandler(service)
	if err != nil {
		t.Fatalf("create dashboard handler: %v", err)
	}
	return handler
}

func newFormContext(values url.Values) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/jobqueue/dashboard/jobs/create", strings.NewReader(values.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	recorder := httptest.NewRecorder()
	return echo.New().NewContext(req, recorder), recorder
}

func TestCreateJobsEnqueuesAllFieldsConcurrently(t *testing.T) {
	service := &fakeJobService{enqueueErrFor: make(map[string]error)}
	handler := newDashboardHandler(t, service)
	ctx, recorder := newFormContext(url.Values{
		"Job1": {"JobTest1"},
		"Job2": {"JobTest2"},
		"Job3": {"JobTest3"},
	})

	if err := handler.CreateJobs(ctx); err != nil {
		t.Fatalf("create jobs: %v", err)
	}
	if service.maxConcurrent != 3 {
		t.Fatalf("want 3 concurrent enqueue calls, got %d", service.maxConcurrent)
	}
	for _, task := range []string{"JobTest1", "JobTest2", "JobTest3"} {
		if !strings.Contains(recorder.Body.String(), task) {
			t.Fatalf("response does not contain task %q: %s", task, recorder.Body.String())
		}
	}
}

func TestCreateJobsRejectsMissingField(t *testing.T) {
	service := &fakeJobService{enqueueErrFor: make(map[string]error)}
	handler := newDashboardHandler(t, service)
	ctx, _ := newFormContext(url.Values{
		"Job1": {"JobTest1"},
		"Job2": {"   "},
		"Job3": {"JobTest3"},
	})

	err := handler.CreateJobs(ctx)
	var httpErr *echo.HTTPError
	if !errors.As(err, &httpErr) || httpErr.Code != http.StatusBadRequest {
		t.Fatalf("want HTTP 400, got %v", err)
	}
	if len(service.jobs) != 0 {
		t.Fatalf("validation failure must not partially enqueue jobs: %+v", service.jobs)
	}
}

func TestCreateJobsPropagatesEnqueueError(t *testing.T) {
	wantErr := errors.New("repository unavailable")
	service := &fakeJobService{enqueueErrFor: map[string]error{"JobTest2": wantErr}}
	handler := newDashboardHandler(t, service)
	ctx, _ := newFormContext(url.Values{
		"Job1": {"JobTest1"},
		"Job2": {"JobTest2"},
		"Job3": {"JobTest3"},
	})

	if err := handler.CreateJobs(ctx); !errors.Is(err, wantErr) {
		t.Fatalf("want enqueue error, got %v", err)
	}
}

func TestDashboardAndFragmentsRender(t *testing.T) {
	service := &fakeJobService{
		jobs: []*entity.Job{{
			ID:       "job-1",
			Task:     "task-1",
			Status:   entity.StatusCompleted,
			Attempts: 1,
		}},
		enqueueErrFor: make(map[string]error),
	}
	handler := newDashboardHandler(t, service)

	tests := []struct {
		name     string
		path     string
		paramID  string
		handle   func(echo.Context) error
		contains string
	}{
		{name: "page", path: "/jobqueue/dashboard", handle: handler.Page, contains: "Job Queue Dashboard"},
		{name: "status", path: "/jobqueue/dashboard/status", handle: handler.Status, contains: "status-summary"},
		{name: "jobs", path: "/jobqueue/dashboard/jobs", handle: handler.Jobs, contains: "task-1"},
		{name: "detail", path: "/jobqueue/dashboard/jobs/job-1", paramID: "job-1", handle: handler.JobDetail, contains: "job-1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.path, nil)
			recorder := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, recorder)
			if test.paramID != "" {
				ctx.SetPath("/jobqueue/dashboard/jobs/:id")
				ctx.SetParamNames("id")
				ctx.SetParamValues(test.paramID)
			}
			if err := test.handle(ctx); err != nil {
				t.Fatalf("render %s: %v", test.name, err)
			}
			if !strings.Contains(recorder.Body.String(), test.contains) {
				t.Fatalf("response does not contain %q: %s", test.contains, recorder.Body.String())
			}
		})
	}
}

func TestJobDetailRendersNotFoundFragment(t *testing.T) {
	service := &fakeJobService{enqueueErrFor: make(map[string]error)}
	handler := newDashboardHandler(t, service)
	req := httptest.NewRequest(http.MethodGet, "/jobqueue/dashboard/jobs/missing", nil)
	recorder := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, recorder)
	ctx.SetPath("/jobqueue/dashboard/jobs/:id")
	ctx.SetParamNames("id")
	ctx.SetParamValues("missing")

	if err := handler.JobDetail(ctx); err != nil {
		t.Fatalf("render missing job: %v", err)
	}
	if !strings.Contains(recorder.Body.String(), "Job not found") {
		t.Fatalf("unexpected not-found fragment: %s", recorder.Body.String())
	}
}
