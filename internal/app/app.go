// Package app wires application components, provides lifecycle management,
// and exposes the entry point for booting and running the service.
package app

import (
	"Atlas/internal/config"
	"Atlas/internal/handler"
	"Atlas/internal/logger"
	"Atlas/internal/repository"
	"Atlas/internal/server"
	"Atlas/internal/service"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"context"

	"github.com/pressly/goose/v3"
	"github.com/wb-go/wbf/dbpg"
)

// App represents the application's composition root.
// It holds long-lived resources (logger, database storage, server) and
// the context/cancel function used for graceful shutdown.
type App struct {
	logger  logger.Logger       // structured logger used across layers
	logFile *os.File            // file handle where logs are written
	server  server.Server       // HTTP server instance
	ctx     context.Context     // root context for shutdown coordination
	cancel  context.CancelFunc  // cancels ctx when a shutdown signal is received
	storage *repository.Storage // data storage abstraction backed by the database
}

// Boot loads configuration, initializes logger, connects to the database,
// applies migrations, wires all components, and returns a fully constructed *App.
func Boot() *App {

	config, err := config.Load()
	if err != nil {
		log.Fatalf("app — failed to load configs: %v", err)
	}

	logger, logFile := logger.NewLogger(config.Logger)

	db, err := bootstrapDB(logger, config.Storage)
	if err != nil {
		logger.LogFatal("app — failed to connect to database", err, "layer", "app")
	}

	return wireApp(db, logger, logFile, config)

}

// bootstrapDB establishes a database connection using repository.ConnectDB,
// runs pending Goose migrations, and returns the DB handle.
// It logs a successful connection and migration application.
func bootstrapDB(logger logger.Logger, config config.Storage) (*dbpg.DB, error) {

	db, err := repository.ConnectDB(config)
	if err != nil {
		return nil, err
	}

	logger.LogInfo("app — connected to database", "layer", "app")

	if err := goose.SetDialect(config.Dialect); err != nil {
		return nil, fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db.Master, config.MigrationsDir); err != nil {
		return nil, fmt.Errorf("failed to apply goose migrations: %w", err)
	}

	logger.Debug("app — migrations applied", "layer", "app")

	return db, nil

}

// wireApp constructs all application components (storage, service, handler, server),
// creates a cancellable context, and returns the assembled *App.
func wireApp(db *dbpg.DB, logger logger.Logger, logFile *os.File, config config.Config) *App {

	ctx, cancel := newContext(logger)
	storage := repository.NewStorage(logger, config.Storage, db)
	service := service.NewService(logger, config.Service, storage)
	server := server.NewServer(logger, config.Server, handler.NewHandler(config.Server, service), cancel)

	return &App{
		logger:  logger,
		logFile: logFile,
		server:  server,
		ctx:     ctx,
		cancel:  cancel,
		storage: storage,
	}

}

// newContext creates a context that is cancelled when the process receives
// SIGINT or SIGTERM. It logs the received signal and triggers graceful shutdown
// by calling the returned cancel function.
func newContext(logger logger.Logger) (context.Context, context.CancelFunc) {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := <-sigCh
		sigString := sig.String()
		if sig == syscall.SIGTERM {
			sigString = "terminate" // sig.String() returns the SIGTERM string in past tense for some reason
		}
		logger.LogInfo("app — received signal "+sigString+", initiating graceful shutdown", "layer", "app")
		cancel()
	}()

	return ctx, cancel

}

// Run starts the HTTP server in a background goroutine,
// then blocks until the application's context is cancelled.
// After cancellation it calls Stop to perform orderly shutdown.
func (a *App) Run() {

	go a.server.Run()
	<-a.ctx.Done()

	a.Stop()

}

// Stop performs an orderly shutdown of application components:
// it shuts down the HTTP server, closes the database storage,
// and closes the log file if it is not os.Stdout.
func (a *App) Stop() {

	a.server.Shutdown()
	a.storage.Close()

	if a.logFile != nil && a.logFile != os.Stdout {
		_ = a.logFile.Close()
	}

}
