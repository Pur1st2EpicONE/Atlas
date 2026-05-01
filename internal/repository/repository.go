// Package repository provides data access layer interfaces and implementations.
// It defines storage interfaces for authentication and core operations,
// and offers a composite Storage struct that combines them.
package repository

import (
	"Atlas/internal/config"
	"Atlas/internal/logger"
	"Atlas/internal/models"
	"Atlas/internal/repository/postgres"
	"context"
	"database/sql"
	"fmt"

	"github.com/wb-go/wbf/dbpg"
)

// AuthStorage defines the interface for user authentication database operations.
type AuthStorage interface {
	CreateUser(ctx context.Context, user models.User) (int64, error)       // CreateUser inserts a new user and returns the generated ID.
	GetUserByLogin(ctx context.Context, login string) (models.User, error) // GetUserByLogin retrieves a user by their login name.
}

// CoreStorage defines the interface for item and history database operations.
type CoreStorage interface {
	CreateItem(tx *sql.Tx, ctx context.Context, item models.Item) (models.Item, error)                           // CreateItem inserts a new item within a transaction and returns the created item.
	DeleteItem(tx *sql.Tx, ctx context.Context, itemID int64) error                                              // DeleteItem removes an item by ID within a transaction.
	GetItem(ctx context.Context, itemID int64) (models.Item, error)                                              // GetItem retrieves a single item by ID.
	GetItems(ctx context.Context) ([]models.Item, error)                                                         // GetItems returns all items.
	GetItemForUpdate(tx *sql.Tx, ctx context.Context, itemID int64) (models.Item, error)                         // GetItemForUpdate retrieves an item with a row lock for update within a transaction.
	UpdateItem(tx *sql.Tx, ctx context.Context, itemID int64, updatedItem models.Item) error                     // UpdateItem updates an existing item within a transaction.
	GetItemHistory(ctx context.Context, itemID int64, filter models.HistoryFilter) ([]models.ItemHistory, error) // GetItemHistory returns historical change records for an item, filtered by the given filter.
	Transaction(ctx context.Context, fn func(tx *sql.Tx, ctx context.Context) error) error                       // Transaction executes a function within a database transaction.
	Close()                                                                                                      // Close releases the database connection pool.
}

// Storage composes AuthStorage and CoreStorage into a single struct.
// It provides a unified data access layer for the application.
type Storage struct {
	AuthStorage // AuthStorage handles user authentication operations
	CoreStorage // CoreStorage handles item and history operations
}

// NewStorage creates a new Storage instance that combines authentication and core storage
// implementations using PostgreSQL. It takes a logger, storage configuration, and database connection.
func NewStorage(logger logger.Logger, config config.Storage, db *dbpg.DB) *Storage {
	return &Storage{
		AuthStorage: postgres.NewAuthStorage(logger, config, db),
		CoreStorage: postgres.NewCoreStorage(logger, config, db),
	}
}

// ConnectDB establishes a connection to the database using the provided configuration.
// It returns a dbpg.DB instance after pinging the database to ensure connectivity.
func ConnectDB(config config.Storage) (*dbpg.DB, error) {

	options := &dbpg.Options{
		MaxOpenConns:    config.MaxOpenConns,
		MaxIdleConns:    config.MaxIdleConns,
		ConnMaxLifetime: config.ConnMaxLifetime,
	}

	db, err := dbpg.New(fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.DBName, config.SSLMode), nil, options)
	if err != nil {
		return nil, fmt.Errorf("database driver not found or DSN invalid: %w", err)
	}

	if err := db.Master.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return db, nil

}
