package postgres

import (
	"Atlas/internal/models"
	"context"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

func (s *AuthStorage) GetUserByLogin(ctx context.Context, login string) (models.User, error) {

	row, err := s.db.QueryRowWithRetry(ctx, retry.Strategy(s.config.QueryRetryStrategy), `

    SELECT id, username, password
    FROM users
    WHERE username = $1`,

		login)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to query row %w", err)
	}

	var user models.User
	err = row.Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		return models.User{}, err
	}

	return user, nil

}
