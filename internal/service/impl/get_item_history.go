package impl

import (
	"Atlas/internal/models"
	"context"
)

func (s *CoreService) GetItemHistory(ctx context.Context, itemID int64, filter models.HistoryFilter) ([]models.ItemHistory, error) {

	if err := s.validateFilter(filter); err != nil {
		return nil, err
	}

	history, err := s.storage.GetItemHistory(ctx, itemID, filter)
	if err != nil {
		s.logger.LogError("service — failed to get item history from storage", err, "layer", "service.impl")
		return nil, err
	}

	return history, nil

}
