package impl

import (
	"Atlas/internal/models"
	"context"
)

func (s *CoreService) GetItems(ctx context.Context) ([]models.Item, error) {
	items, err := s.storage.GetItems(ctx)
	if err != nil {
		s.logger.LogError("service — failed to get items from storage", err, "layer", "service.impl")
		return nil, err
	}
	return items, nil
}
