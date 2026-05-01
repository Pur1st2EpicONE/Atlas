package postgres

import (
	"Atlas/internal/models"
	"context"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

// CreateUser inserts a new user record into the database.
// It returns the auto-generated user ID.
func (s *AuthStorage) CreateUser(ctx context.Context, user models.User) (int64, error) {

	var userID int64
	row, err := s.db.QueryRowWithRetry(ctx, retry.Strategy(s.config.QueryRetryStrategy), `

    INSERT INTO users (username, password, role)
    VALUES ($1, $2, $3)
    RETURNING id;`,

		user.Login, user.Password, user.Role)
	if err != nil {
		return 0, err
	}

	err = row.Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to scan row: %w", err)
	}

	return userID, nil

}
