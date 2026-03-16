package service

import (
	"Atlas/internal/config"
	"Atlas/internal/logger"
	"Atlas/internal/models"
	"Atlas/internal/repository"
	"Atlas/internal/service/impl"
	"context"

	"github.com/golang-jwt/jwt"
)

type AuthService interface {
	CreateUser(ctx context.Context, user models.User) (int64, error)
	CreateToken(user models.User) (string, error)
	GetUser(ctx context.Context, user models.User) (models.User, error)
	ParseToken(tokenString string) (int64, error)
	KeyFunc(token *jwt.Token) (any, error)
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
