package htmx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"

	"jobqueue/entity"
	_interface "jobqueue/interface"
	htmxtemplates "jobqueue/web/htmx"

	"github.com/labstack/echo/v4"
)

// pageData is the view model shared by the dashboard page and every
// fragment endpoint; each fragment only reads the fields it needs.
type pageData struct {
	Jobs   []*entity.Job
	Status entity.JobStatus
	Job    *entity.Job
}

// DashboardHandler serves the HTMX job queue dashboard. It only bridges HTTP
// to the service layer; all data comes from _interface.JobService.
type DashboardHandler struct {
	jobService _interface.JobService
	tmpl       *template.Template
}

// NewDashboardHandler creates a DashboardHandler and parses the templates
// once, up front.
func NewDashboardHandler(jobService _interface.JobService) (*DashboardHandler, error) {
	tmpl, err := htmxtemplates.Parse()
	if err != nil {
		return nil, err
	}
	return &DashboardHandler{jobService: jobService, tmpl: tmpl}, nil
}

// Page serves the full HTML shell for the dashboard.
func (h *DashboardHandler) Page(c echo.Context) error {
	jobs, status, err := h.collect(c.Request().Context())
	if err != nil {
		return err
	}
	return h.render(c, "dashboard", pageData{
		Jobs:   jobs,
		Status: status,
	})
}

// Message serves the HTMX HTML fragment swapped into the page on demand.
func (h *DashboardHandler) Message(c echo.Context) error {
	return c.HTML(http.StatusOK, `<p id="message">Hello, World!</p>`)
}

// Status serves the status summary fragment, polled every 2 seconds.
func (h *DashboardHandler) Status(c echo.Context) error {
	status, err := h.jobService.GetJobStatus(c.Request().Context())
	if err != nil {
		return err
	}
	return h.render(c, "status_summary", pageData{Status: status})
}

// Jobs serves the full jobs table fragment, polled every 2 seconds.
func (h *DashboardHandler) Jobs(c echo.Context) error {
	jobs, err := h.jobService.GetAllJobs(c.Request().Context())
	if err != nil {
		return err
	}
	return h.render(c, "jobs_table", pageData{Jobs: jobs})
}

// JobDetail serves the job detail fragment for a single job. The ID comes
// from the path (/jobs/:id) or, for the search form, from the query string
// (/jobs/search?id=...).
func (h *DashboardHandler) JobDetail(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		id = c.QueryParam("id")
	}

	job, err := h.jobService.GetJobByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, _interface.ErrJobNotFound) {
			return h.render(c, "job_not_found", nil)
		}
		return err
	}
	return h.render(c, "job_detail", pageData{Job: job})
}

// CreateJobs enqueues the three tasks from the variables form simultaneously
// and returns the refreshed jobs table fragment.
func (h *DashboardHandler) CreateJobs(c echo.Context) error {
	tasks := []string{
		c.FormValue("Job1"),
		c.FormValue("Job2"),
		c.FormValue("Job3"),
	}
	for _, task := range tasks {
		if strings.TrimSpace(task) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "Job1, Job2, and Job3 are required")
		}
	}

	ctx := c.Request().Context()
	var wg sync.WaitGroup
	errCh := make(chan error, len(tasks))
	for _, task := range tasks {
		wg.Add(1)
		go func(task string) {
			defer wg.Done()
			if _, err := h.jobService.Enqueue(ctx, task); err != nil {
				errCh <- fmt.Errorf("enqueue task %q: %w", task, err)
			}
		}(task)
	}
	wg.Wait()
	close(errCh)
	if err := <-errCh; err != nil {
		return err
	}

	return h.Jobs(c)
}

// CreateUnstable enqueues the special unstable-job task and returns the
// refreshed jobs table fragment.
func (h *DashboardHandler) CreateUnstable(c echo.Context) error {
	if _, err := h.jobService.Enqueue(c.Request().Context(), entity.UnstableJobTask); err != nil {
		return err
	}
	return h.Jobs(c)
}

func (h *DashboardHandler) collect(ctx context.Context) ([]*entity.Job, entity.JobStatus, error) {
	jobs, err := h.jobService.GetAllJobs(ctx)
	if err != nil {
		return nil, entity.JobStatus{}, err
	}
	status, err := h.jobService.GetJobStatus(ctx)
	if err != nil {
		return nil, entity.JobStatus{}, err
	}
	return jobs, status, nil
}

func (h *DashboardHandler) render(c echo.Context, name string, data any) error {
	var output bytes.Buffer
	if err := h.tmpl.ExecuteTemplate(&output, name, data); err != nil {
		return fmt.Errorf("render template %q: %w", name, err)
	}
	return c.HTMLBlob(http.StatusOK, output.Bytes())
}
