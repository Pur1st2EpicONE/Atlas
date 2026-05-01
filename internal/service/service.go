// Package service defines the business logic interfaces and composes them into a Service struct.
// It provides authentication (AuthService) and core item management (CoreService) abstractions.
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

// AuthService defines the interface for authentication operations.
type AuthService interface {
	CreateUser(ctx context.Context, user models.User) (int64, error)    // CreateUser registers a new user and returns the user ID.
	CreateToken(user models.User) (string, error)                       // CreateToken generates a JWT token for the authenticated user.
	GetUser(ctx context.Context, user models.User) (models.User, error) // GetUser validates credentials and returns the user.
	ParseToken(tokenString string) (int64, error)                       // ParseToken extracts the user ID from a JWT token.
	KeyFunc(token *jwt.Token) (any, error)                              // KeyFunc returns the signing key for JWT validation.
}

// CoreService defines the interface for item and history management.
type CoreService interface {
	CreateItem(ctx context.Context, userID int64, item models.Item) (models.Item, error)                         // CreateItem adds a new item and returns it.
	UpdateItem(ctx context.Context, userID int64, itemID int64, update models.Update) error                      // UpdateItem partially updates an existing item.
	DeleteItem(ctx context.Context, userID int64, itemID int64) error                                            // DeleteItem removes an item by ID.
	GetItem(ctx context.Context, itemID int64) (models.Item, error)                                              // GetItem retrieves a single item by ID.
	GetItems(ctx context.Context) ([]models.Item, error)                                                         // GetItems returns all items.
	GetItemHistory(ctx context.Context, itemID int64, filter models.HistoryFilter) ([]models.ItemHistory, error) // GetItemHistory returns historical changes for an item.
}

// Service composes AuthService and CoreService into a single struct.
// It serves as the application's service layer entry point.
type Service struct {
	AuthService // AuthService handles user authentication and token management.
	CoreService // CoreService handles item CRUD and history operations.
}

// NewService creates a new Service instance with the given logger, configuration,
// and storage. It wires the concrete implementations (impl.AuthService and impl.CoreService).
func NewService(logger logger.Logger, config config.Service, storage *repository.Storage) *Service {
	return &Service{
		AuthService: impl.NewAuthService(logger, config.Auth, storage.AuthStorage),
		CoreService: impl.NewCoreService(logger, config.Core, storage.CoreStorage),
	}
}
