package postgres

import (
	"Atlas/internal/models"
	"context"

	"github.com/wb-go/wbf/retry"
)

func (s *CoreStorage) GetItemHistory(ctx context.Context, itemID int64) ([]models.ItemHistory, error) {

	rows, err := s.db.QueryWithRetry(ctx, retry.Strategy(s.config.QueryRetryStrategy), `
        
	SELECT id, item_id, user_id, action, changed_at, old_data, new_data
    FROM item_history
    WHERE item_id = $1
    ORDER BY changed_at DESC`, itemID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.ItemHistory
	for rows.Next() {
		var h models.ItemHistory
		var oldData, newData []byte
		if err = rows.Scan(&h.ID, &h.ItemID, &h.UserID, &h.Action, &h.ChangedAt, &oldData, &newData); err != nil {
			return nil, err
		}
		h.OldData = oldData
		h.NewData = newData
		history = append(history, h)
	}

	return history, rows.Err()

}
