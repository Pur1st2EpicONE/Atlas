// Package httpserver implements an HTTP server using the standard library's
// http.Server. It provides graceful shutdown capabilities and integrates
// with the application's logger and cancellation mechanism.
package httpserver

import (
	"Atlas/internal/config"
	"Atlas/internal/logger"
	"context"
	"errors"
	"net/http"
	"time"
)

// HttpServer is the concrete implementation of the Server interface.
// It wraps an http.Server and adds logging and shutdown coordination.
type HttpServer struct {
	shutdownTimeout time.Duration      // maximum time allowed for graceful shutdown
	logger          logger.Logger      // structured logger
	cancel          context.CancelFunc // function to cancel the application root context on fatal error
	instance        *http.Server       // underlying standard HTTP server
}

// NewServer creates and initialises an HttpServer with the given dependencies.
// It does not start the server; that is done by calling Run.
func NewServer(logger logger.Logger, config config.Server, handler http.Handler, cancel context.CancelFunc) *HttpServer {

	return &HttpServer{
		shutdownTimeout: config.ShutdownTimeout,
		logger:          logger,
		cancel:          cancel,
		instance: &http.Server{
			Addr:           ":" + config.Port,
			Handler:        handler,
			ReadTimeout:    config.ReadTimeout,
			WriteTimeout:   config.WriteTimeout,
			MaxHeaderBytes: config.MaxHeaderBytes},
	}

}

// Run starts the HTTP server and listens for incoming requests.
// It logs a startup message, then calls ListenAndServe. If the server
// stops with an error that is not http.ErrServerClosed, it logs a fatal
// error and triggers the application's cancel function to initiate shutdown.
func (s *HttpServer) Run() {
	s.logger.LogInfo("server — receiving requests", "layer", "server.httpserver")
	if err := s.instance.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.LogError("server — fatal at ListenAndServe, initiating emergency shutdown", err, "layer", "server.httpserver")
		s.cancel()
	}
}

// Shutdown gracefully stops the HTTP server, allowing active connections
// to complete up to the shutdownTimeout. If the shutdown succeeds, it logs
// a completion message; otherwise, it logs an error.
func (s *HttpServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()
	if err := s.instance.Shutdown(ctx); err != nil {
		s.logger.LogError("server — failed to shutdown gracefully", err, "layer", "server.httpserver")
	} else {
		s.logger.LogInfo("server — shutdown complete", "layer", "server.httpserver")
	}
}
