package _dataloader

import (
	"context"

	_interface "jobqueue/interface"
	"jobqueue/pkg/constant"

	"github.com/graph-gophers/dataloader/v6"
	"github.com/labstack/echo/v4"
)

// GeneralDataloader ...
type GeneralDataloader struct {
	JobLoader *dataloader.Loader
	jobRepo   _interface.JobRepository
}

func (g GeneralDataloader) EchoMiddelware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// add general dataloader into echo context
		oriReq := c.Request()
		req := oriReq.WithContext(context.WithValue(oriReq.Context(), constant.DataloaderContextKey, g))
		c.SetRequest(req)
		return next(c)
	}
}

// Initiator ...
type Initiator func(r *GeneralDataloader) *GeneralDataloader

// New ...
func New() Initiator {
	return func(r *GeneralDataloader) *GeneralDataloader {
		return r
	}
}

// SetJobRepository ...
func (i Initiator) SetJobRepository(repo _interface.JobRepository) Initiator {
	return func(s *GeneralDataloader) *GeneralDataloader {
		i(s).jobRepo = repo
		return s
	}
}

// SetBatchFunction ...
func (i Initiator) SetBatchFunction() Initiator {
	return func(s *GeneralDataloader) *GeneralDataloader {
		i(s).JobLoader = dataloader.NewBatchedLoader(s.JobBatchFunc, dataloader.WithCache(dataloader.NewCache()))
		return s
	}
}

// Build ...
func (i Initiator) Build() *GeneralDataloader {
	return i(&GeneralDataloader{})
}
