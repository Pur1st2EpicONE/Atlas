package postgres

import (
	"Atlas/internal/models"
	"context"

	"github.com/wb-go/wbf/retry"
)

func (s *CoreStorage) GetItems(ctx context.Context) ([]models.Item, error) {

	rows, err := s.db.QueryWithRetry(ctx, retry.Strategy(s.config.QueryRetryStrategy), `

    SELECT id, name, description, quantity, price, created_at, updated_at
    FROM items ORDER BY id`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var i models.Item
		if err = rows.Scan(&i.ID, &i.Name, &i.Description, &i.Quantity, &i.Price, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	return items, rows.Err()

}
