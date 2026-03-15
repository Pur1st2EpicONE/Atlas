package service

import (
	"Atlas/internal/config"
	"Atlas/internal/logger"
	"Atlas/internal/models"
	"Atlas/internal/repository"
	"Atlas/internal/service/impl"
	"context"
)

type AuthService interface {
	CreateUser(ctx context.Context, user models.User) (int64, error)
	CreateToken(userID int64) (string, error)
	GetUserId(ctx context.Context, user models.User) (int64, error)
	ParseToken(tokenString string) (int64, error)
}

type CoreService interface {
	CreateEvent(ctx context.Context, event *models.Event) (string, error)
}

type Service struct {
	AuthService
	CoreService
}

func NewService(logger logger.Logger, config config.Service, storage *repository.Storage) *Service {
	return &Service{
		AuthService: impl.NewAuthService(logger, config, storage.AuthStorage),
		CoreService: impl.NewCoreService(logger, storage.CoreStorage),
	}
}
