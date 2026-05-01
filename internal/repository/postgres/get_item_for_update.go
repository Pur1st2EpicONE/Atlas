package postgres

import (
	"Atlas/internal/models"
	"context"
	"database/sql"
)

// GetItemForUpdate retrieves an item by ID with a row lock for update within a transaction.
// It is used to prevent concurrent modifications.
func (s *CoreStorage) GetItemForUpdate(tx *sql.Tx, ctx context.Context, itemID int64) (models.Item, error) {

	var item models.Item
	row := tx.QueryRowContext(ctx, `

    SELECT id, name, description, quantity, price, created_at, updated_at
    FROM items WHERE id = $1`, itemID)

	if err := row.Scan(&item.ID, &item.Name, &item.Description,
		&item.Quantity, &item.Price, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return models.Item{}, err
	}

	return item, nil

}
