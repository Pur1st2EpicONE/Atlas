package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

// Transaction executes a function within a database transaction.
// It automatically rolls back on error and commits on success.
func (c *CoreStorage) Transaction(ctx context.Context, fn func(tx *sql.Tx, ctx context.Context) error) error {

	tx, err := c.db.BeginTxWithRetry(ctx, retry.Strategy(c.config.TxRetryStrategy), nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := fn(tx, ctx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil

}
