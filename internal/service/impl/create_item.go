package impl

import (
	"Atlas/internal/models"
	"context"
	"database/sql"
	"fmt"
)

func (s *CoreService) CreateItem(ctx context.Context, userID int64, item models.Item) (models.Item, error) {

	if err := s.validateItem(item); err != nil {
		return models.Item{}, err
	}

	err := s.storage.Transaction(ctx, func(tx *sql.Tx, ctx context.Context) error {

		if _, err := tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL app.current_user_id = %d", userID)); err != nil {
			return fmt.Errorf("unable to set application-level userID for transaction: %w", err)
		}

		var innerErr error
		item, innerErr = s.storage.CreateItem(tx, ctx, item)
		if innerErr != nil {
			return fmt.Errorf("error at CreateItem: %w", innerErr)
		}
		return nil

	})
	if err != nil {
		s.logger.LogError("service — failed to save item in storage", err, "layer", "service.impl")
		return models.Item{}, err
	}

	return item, nil

}
