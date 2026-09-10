package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jobqueue/config"
	"jobqueue/delivery/graphql"
	_dataloader "jobqueue/delivery/graphql/dataloader"
	"jobqueue/delivery/graphql/mutation"
	"jobqueue/delivery/graphql/query"
	"jobqueue/delivery/graphql/schema"
	_htmx "jobqueue/delivery/htmx"
	"jobqueue/entity"
	"jobqueue/pkg/handler"
	"jobqueue/pkg/server"
	inmemrepo "jobqueue/repository/inmem"
	"jobqueue/service"

	_graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	setupLogger()
	logger := logrus.New()
	logger.SetReportCaller(true)
	e := server.New(config.Data.Server)
	e.Echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${remote_ip} ${time_rfc3339_nano} \"${method} ${path}\" ${status} ${bytes_out} \"${referer}\" \"${user_agent}\"\n",
	}))
	e.Echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.OPTIONS},
	}))

	//graphql schema
	opts := make([]_graphql.SchemaOpt, 0)
	opts = append(opts, _graphql.SubscribeResolverTimeout(10*time.Second))

	//initialize in mem database
	inMemDb := make(map[string]*entity.Job)

	//set job repository
	jobRepository := inmemrepo.
		NewJobRepository().
		SetInMemConnection(inMemDb).
		Build()
	dataloader := _dataloader.
		New().
		SetJobRepository(jobRepository).
		SetBatchFunction().
		Build()

	//set job service
	jobService := service.NewJobService().
		SetJobRepository(jobRepository).
		SetLogger(logger).
		Build()

	jobMutation := mutation.NewJobMutation(jobService, dataloader)
	jobQuery := query.NewJobQuery(jobService, dataloader)

	rootResolver := graphql.
		New().
		SetJobMutation(jobMutation).
		SetJobQuery(jobQuery).
		Build()

	graphqlSchema := _graphql.MustParseSchema(schema.String(), rootResolver, opts...)
	e.Echo.POST("/graphql",
		handler.GraphQLHandler(&relay.Handler{Schema: graphqlSchema}),
		dataloader.EchoMiddelware,
	)
	e.Echo.GET("/graphql",
		handler.GraphQLHandler(&relay.Handler{Schema: graphqlSchema}),
		dataloader.EchoMiddelware,
	)
	e.Echo.GET("/graphiql", handler.GraphiQLHandler)

	//htmx dashboard
	dashboardHandler, err := _htmx.NewDashboardHandler(jobService)
	if err != nil {
		logger.Fatalf("parse dashboard templates: %v", err)
	}
	e.Echo.GET("/jobqueue/dashboard", dashboardHandler.Page)
	e.Echo.GET("/jobqueue/dashboard/message", dashboardHandler.Message)
	e.Echo.GET("/jobqueue/dashboard/status", dashboardHandler.Status)
	e.Echo.GET("/jobqueue/dashboard/jobs", dashboardHandler.Jobs)
	e.Echo.GET("/jobqueue/dashboard/jobs/search", dashboardHandler.JobDetail)
	e.Echo.GET("/jobqueue/dashboard/jobs/:id", dashboardHandler.JobDetail)
	e.Echo.POST("/jobqueue/dashboard/jobs/create", dashboardHandler.CreateJobs)
	e.Echo.POST("/jobqueue/dashboard/jobs/unstable", dashboardHandler.CreateUnstable)

	//start server and wait for an interrupt to shut down gracefully
	go func() {
		if err := e.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Echo.Logger.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Echo.Shutdown(shutdownCtx); err != nil {
		e.Echo.Logger.Errorf("http server shutdown: %v", err)
	}
	if err := jobService.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("job workers shutdown: %v", err)
	}
	logger.Info("shutdown complete")
}

func setupLogger() {
	configLogger := zap.NewDevelopmentConfig()
	configLogger.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	configLogger.DisableStacktrace = true
	logger, _ := configLogger.Build()
	zap.ReplaceGlobals(logger)
}
