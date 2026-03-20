package impl

import (
	"Atlas/internal/models"
	"context"
)

func (s *CoreService) GetItemHistory(ctx context.Context, itemID int64) ([]models.ItemHistory, error) {
	return s.storage.GetItemHistory(ctx, itemID)
}
