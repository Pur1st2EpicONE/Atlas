package postgres

import (
	"Atlas/internal/models"
	"context"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

func (s *CoreStorage) GetItem(ctx context.Context, id int64) (models.Item, error) {

	var item models.Item

	row, err := s.db.QueryRowWithRetry(ctx, retry.Strategy(s.config.QueryRetryStrategy), `
        
	SELECT id, name, description, quantity, price, created_at, updated_at
	FROM items 
	WHERE id = $1`, id)

	if err != nil {
		return item, fmt.Errorf("failed to query row: %w", err)
	}
	err = row.Scan(&item.ID, &item.Name, &item.Description, &item.Quantity, &item.Price, &item.CreatedAt, &item.UpdatedAt)

	return item, err

}
