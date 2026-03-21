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

type AuthStorage interface {
	CreateUser(ctx context.Context, user models.User) (int64, error)
	GetUserByLogin(ctx context.Context, login string) (models.User, error)
}

type CoreStorage interface {
	CreateItem(tx *sql.Tx, ctx context.Context, item models.Item) (models.Item, error)
	DeleteItem(tx *sql.Tx, ctx context.Context, itemID int64) error
	GetItem(ctx context.Context, itemID int64) (models.Item, error)
	GetItems(ctx context.Context) ([]models.Item, error)
	GetItemForUpdate(tx *sql.Tx, ctx context.Context, itemID int64) (models.Item, error)
	UpdateItem(tx *sql.Tx, ctx context.Context, itemID int64, updatedItem models.Item) error
	GetItemHistory(ctx context.Context, itemID int64, filter models.HistoryFilter) ([]models.ItemHistory, error)
	Transaction(ctx context.Context, fn func(tx *sql.Tx, ctx context.Context) error) error
	Close()
}

type Storage struct {
	AuthStorage
	CoreStorage
}

func NewStorage(logger logger.Logger, config config.Storage, db *dbpg.DB) *Storage {
	return &Storage{
		AuthStorage: postgres.NewAuthStorage(logger, config, db),
		CoreStorage: postgres.NewCoreStorage(logger, config, db),
	}
}

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
