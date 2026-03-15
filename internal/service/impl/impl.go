package impl

import (
	"Atlas/internal/config"
	"Atlas/internal/logger"
	"Atlas/internal/repository"
)

type AuthService struct {
	logger  logger.Logger
	config  config.Service
	storage repository.AuthStorage
}

func NewAuthService(logger logger.Logger, config config.Service, storage repository.AuthStorage) *AuthService {
	return &AuthService{logger: logger, config: config, storage: storage}
}

type CoreService struct {
	logger  logger.Logger
	storage repository.CoreStorage
}

func NewCoreService(logger logger.Logger, storage repository.CoreStorage) *CoreService {
	return &CoreService{logger: logger, storage: storage}
}
