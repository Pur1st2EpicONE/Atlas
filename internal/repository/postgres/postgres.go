package postgres

import (
	"Atlas/internal/config"
	"Atlas/internal/logger"

	"github.com/wb-go/wbf/dbpg"
)

type AuthStorage struct {
	db     *dbpg.DB
	logger logger.Logger
	config config.Storage
}

func NewAuthStorage(logger logger.Logger, config config.Storage, db *dbpg.DB) *AuthStorage {
	return &AuthStorage{db: db, logger: logger, config: config}
}

type CoreStorage struct {
	db     *dbpg.DB
	logger logger.Logger
	config config.Storage
}

func NewCoreStorage(logger logger.Logger, config config.Storage, db *dbpg.DB) *CoreStorage {
	return &CoreStorage{db: db, logger: logger, config: config}
}
