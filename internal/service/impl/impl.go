// Package impl provides concrete implementations of the service interfaces.
// It contains AuthService (authentication, JWT, user management) and
// CoreService (item CRUD, history, validation).
package impl

import (
	"Atlas/internal/config"
	"Atlas/internal/logger"
	"Atlas/internal/repository"
)

// AuthService implements the AuthService interface.
type AuthService struct {
	logger  logger.Logger          // logger for structured logging
	config  config.Auth            // configuration for authentication (token TTL, lengths, signing key)
	storage repository.AuthStorage // data storage for user operations
}

// NewAuthService creates a new AuthService with the given dependencies.
func NewAuthService(logger logger.Logger, config config.Auth, storage repository.AuthStorage) *AuthService {
	return &AuthService{logger: logger, config: config, storage: storage}
}

// CoreService implements the CoreService interface.
type CoreService struct {
	logger  logger.Logger          // logger for structured logging
	config  config.Core            // configuration for item constraints (name length, quantity limits, price)
	storage repository.CoreStorage // data storage for item and history operations
}

// NewCoreService creates a new CoreService with the given dependencies.
func NewCoreService(logger logger.Logger, config config.Core, storage repository.CoreStorage) *CoreService {
	return &CoreService{logger: logger, config: config, storage: storage}
}
