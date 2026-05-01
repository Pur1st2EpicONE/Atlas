// Package v1 implements the version 1 HTTP handlers for the Atlas API.
// It provides endpoints for item management with role-based access.
package v1

import (
	"Atlas/internal/config"
	"Atlas/internal/service"
)

// Handler holds the dependencies for v1 API endpoints.
// It contains the server configuration and the core service layer.
type Handler struct {
	config  config.Server   // config holds HTTP server settings (timeouts, ports, etc.)
	service service.Service // service provides business logic operations
}

// NewHandler creates a new v1 Handler with the given configuration and service.
func NewHandler(config config.Server, service service.Service) *Handler {
	return &Handler{config: config, service: service}
}
