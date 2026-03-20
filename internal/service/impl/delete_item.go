package impl

import (
	"Atlas/internal/errs"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (s *CoreService) DeleteItem(ctx context.Context, userID int64, itemID int64) error {

	err := s.storage.Transaction(ctx, func(tx *sql.Tx, ctx context.Context) error {

		if _, err := tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL app.current_user_id = %d", userID)); err != nil {
			return fmt.Errorf("unable to set application-level userID for transaction: %w", err)
		}
		if err := s.storage.DeleteItem(tx, ctx, itemID); err != nil {
			return fmt.Errorf("error at DeleteItem: %w", err)
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrItemNotFound
		}
		s.logger.LogError("service — failed to delete item from storage", err, "itemID", itemID, "layer", "service.impl")
	}

	return err

}
