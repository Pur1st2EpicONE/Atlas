package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *CoreStorage) DeleteItem(tx *sql.Tx, ctx context.Context, itemID int64) error {

	result, err := tx.ExecContext(ctx, `
	
	DELETE FROM items 
	WHERE id = $1`, itemID)

	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get number of affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return err

}
