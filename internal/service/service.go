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
	CreateItem(ctx context.Context, userID int64, item models.Item) (models.Item, error)
	UpdateItem(ctx context.Context, userID int64, itemID int64, update models.Update) error
	DeleteItem(ctx context.Context, userID int64, itemID int64) error
	GetItem(ctx context.Context, itemID int64) (models.Item, error)
	GetItems(ctx context.Context) ([]models.Item, error)
	GetItemHistory(ctx context.Context, itemID int64, filter models.HistoryFilter) ([]models.ItemHistory, error)
}

type Service struct {
	AuthService
	CoreService
}

func NewService(logger logger.Logger, config config.Service, storage *repository.Storage) *Service {
	return &Service{
		AuthService: impl.NewAuthService(logger, config.Auth, storage.AuthStorage),
		CoreService: impl.NewCoreService(logger, config.Core, storage.CoreStorage),
	}
}
