package postgres

import (
	"Atlas/internal/models"
	"context"
	"database/sql"
	"fmt"
)

func (s *CoreStorage) UpdateItem(tx *sql.Tx, ctx context.Context, itemID int64, updatedItem models.Item) error {

	_, err := tx.ExecContext(ctx, `

    UPDATE items
    SET name = $1, description = $2, quantity = $3, price = $4, updated_at = NOW()
    WHERE id = $5`,

		updatedItem.Name, updatedItem.Description, updatedItem.Quantity, updatedItem.Price, itemID)

	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return err

}
