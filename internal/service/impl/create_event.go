package impl

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"context"

	"github.com/lib/pq"
	"github.com/wb-go/wbf/helpers"
)

func (c *CoreService) CreateEvent(ctx context.Context, event *models.Event) (string, error) {

	if err := validateEvent(event); err != nil {
		return "", err
	}

	initialize(event)

	if err := c.storage.CreateEvent(ctx, event); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23503" {
				return "", errs.ErrInvalidUserID
			}
		}
		c.logger.LogError("service — failed to create event", err, "layer", "service.impl")
		return "", err
	}

	return event.ID, nil

}

func initialize(event *models.Event) {
	event.ID = helpers.CreateUUID()
}
