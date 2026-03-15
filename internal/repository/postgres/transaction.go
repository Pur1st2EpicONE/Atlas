package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func (c *CoreStorage) Transaction(ctx context.Context, fn func(tx *sql.Tx, ctx context.Context) error) error {

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := fn(tx, ctx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil

}
