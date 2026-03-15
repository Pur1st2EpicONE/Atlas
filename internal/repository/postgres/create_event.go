package postgres

import (
	"Atlas/internal/models"
	"context"
	"fmt"

	"github.com/wb-go/wbf/retry"
)

func (c *CoreStorage) CreateEvent(ctx context.Context, event *models.Event) error {

	_, err := c.db.ExecWithRetry(ctx, retry.Strategy(c.config.QueryRetryStrategy), `

	INSERT INTO events (uuid, userID, title, description, event_date, available_seats, booking_ttl)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`,

		event.ID, event.UserID, event.Title, event.Description, event.Date, event.Seats, int(event.BookingTTL.Seconds()))
	if err != nil {
		return fmt.Errorf("failed to execute query %w", err)
	}

	return nil

}
