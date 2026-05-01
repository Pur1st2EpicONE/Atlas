// Package server defines the HTTP server abstraction and provides
// a constructor that returns a concrete implementation.
package server

import (
	"Atlas/internal/config"
	"Atlas/internal/logger"
	"Atlas/internal/server/httpserver"

	"context"
	"net/http"
)

// Server defines the contract for an HTTP server that can be started
// and gracefully shut down.
type Server interface {
	Run()      // Run starts the HTTP server and blocks until an error occurs or the server is stopped.
	Shutdown() // Shutdown performs an orderly shutdown of the server, waiting for active connections to finish up to the configured timeout.
}

// NewServer constructs a Server instance using the HTTPServer implementation.
// It accepts a logger, server configuration, an HTTP handler (usually the router),
// and a cancel function to be called if the server fails to start.
func NewServer(logger logger.Logger, config config.Server, handler http.Handler, cancel context.CancelFunc) Server {
	return httpserver.NewServer(logger, config, handler, cancel)
}
