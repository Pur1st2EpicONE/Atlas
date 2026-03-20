package impl

import (
	"Atlas/internal/config"
	"Atlas/internal/logger"
	"Atlas/internal/repository"
)

type AuthService struct {
	logger  logger.Logger
	config  config.Auth
	storage repository.AuthStorage
}

func NewAuthService(logger logger.Logger, config config.Auth, storage repository.AuthStorage) *AuthService {
	return &AuthService{logger: logger, config: config, storage: storage}
}

type CoreService struct {
	logger  logger.Logger
	config  config.Core
	storage repository.CoreStorage
}

func NewCoreService(logger logger.Logger, config config.Core, storage repository.CoreStorage) *CoreService {
	return &CoreService{logger: logger, config: config, storage: storage}
}
