package postgres

import (
	"Atlas/internal/models"
	"context"

	"github.com/wb-go/wbf/retry"
)

func (s *AuthStorage) CreateUser(ctx context.Context, user models.User) (int64, error) {

	var userID int64
	row, err := s.db.QueryRowWithRetry(ctx, retry.Strategy(s.config.QueryRetryStrategy), `

    INSERT INTO users (username, password)
    VALUES ($1, $2)
    RETURNING id;`,

		user.Login, user.Password)
	if err != nil {
		return 0, err
	}

	err = row.Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil

}
