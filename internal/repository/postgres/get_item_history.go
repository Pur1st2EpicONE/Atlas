package postgres

import (
	"Atlas/internal/models"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/wb-go/wbf/retry"
)

func (s *CoreStorage) GetItemHistory(ctx context.Context, itemID int64, filter models.HistoryFilter) ([]models.ItemHistory, error) {

	where, args := buildWhere(filter)
	args = append([]any{itemID}, args...)

	query := `
        
	SELECT id, item_id, user_id, action, changed_at, old_data, new_data
    FROM item_history
    WHERE item_id = $1` + where + `
    ORDER BY changed_at DESC
    LIMIT $` + strconv.Itoa(len(args)+1)

	args = append(args, filter.Limit)

	rows, err := s.db.QueryWithRetry(ctx, retry.Strategy(s.config.QueryRetryStrategy), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var history []models.ItemHistory

	for rows.Next() {
		var h models.ItemHistory
		var oldData, newData []byte
		if err := rows.Scan(
			&h.ID,
			&h.ItemID,
			&h.UserID,
			&h.Action,
			&h.ChangedAt,
			&oldData,
			&newData,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		h.OldData = oldData
		h.NewData = newData
		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return history, nil

}

func buildWhere(filter models.HistoryFilter) (string, []any) {

	var conditions []string
	args := []any{}
	argIdx := 2

	if !filter.From.IsZero() {
		conditions = append(conditions, fmt.Sprintf("changed_at >= $%d", argIdx))
		args = append(args, filter.From)
		argIdx++
	}

	if !filter.To.IsZero() {
		conditions = append(conditions, fmt.Sprintf("changed_at <= $%d", argIdx))
		args = append(args, filter.To)
		argIdx++
	}

	if filter.UserID > 0 {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, filter.UserID)
		argIdx++
	}

	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIdx))
		args = append(args, filter.Action)
		argIdx++
	}

	if len(conditions) == 0 {
		return "", nil
	}

	return " AND " + strings.Join(conditions, " AND "), args

}
