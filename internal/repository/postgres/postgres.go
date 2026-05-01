// Package postgres provides PostgreSQL-specific implementations of the repository interfaces.
// It includes storage for authentication (users) and core business entities (items, history).
package postgres

import (
	"Atlas/internal/config"
	"Atlas/internal/logger"

	"github.com/wb-go/wbf/dbpg"
)

// AuthStorage implements the AuthStorage interface using PostgreSQL.
type AuthStorage struct {
	db     *dbpg.DB       // db is the database connection pool
	logger logger.Logger  // logger for structured logging
	config config.Storage // config holds database connection and retry settings
}

// NewAuthStorage creates a new AuthStorage instance with the given logger, config, and DB connection.
func NewAuthStorage(logger logger.Logger, config config.Storage, db *dbpg.DB) *AuthStorage {
	return &AuthStorage{db: db, logger: logger, config: config}
}

// CoreStorage implements the CoreStorage interface using PostgreSQL.
type CoreStorage struct {
	db     *dbpg.DB       // db is the database connection pool
	logger logger.Logger  // logger for structured logging
	config config.Storage // config holds database connection and retry settings
}

// NewCoreStorage creates a new CoreStorage instance with the given logger, config, and DB connection.
func NewCoreStorage(logger logger.Logger, config config.Storage, db *dbpg.DB) *CoreStorage {
	return &CoreStorage{db: db, logger: logger, config: config}
}
