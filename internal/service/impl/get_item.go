package impl

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"context"
	"database/sql"
	"errors"
)

func (s *CoreService) GetItem(ctx context.Context, itemID int64) (models.Item, error) {
	item, err := s.storage.GetItemByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Item{}, errs.ErrItemNotFound
		}
		s.logger.LogError("service — failed to get item from storage", err, "itemID", itemID, "layer", "service.impl")
		return models.Item{}, err
	}
	return item, nil
}
