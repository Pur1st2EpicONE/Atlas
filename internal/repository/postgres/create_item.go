package postgres

import (
	"Atlas/internal/models"
	"context"
	"database/sql"
	"fmt"
)

func (s *CoreStorage) CreateItem(tx *sql.Tx, ctx context.Context, item models.Item) (models.Item, error) {

	var created models.Item

	err := tx.QueryRowContext(ctx, `

    INSERT INTO items (name, description, quantity, price)
    VALUES ($1, $2, $3, $4)
    RETURNING id, name, description, quantity, price, created_at, updated_at`,

		item.Name, item.Description, item.Quantity, item.Price).
		Scan(&created.ID, &created.Name, &created.Description,
			&created.Quantity, &created.Price, &created.CreatedAt, &created.UpdatedAt)

	if err != nil {
		return models.Item{}, fmt.Errorf("failed to scan row: %w", err)
	}
	return created, nil

}
