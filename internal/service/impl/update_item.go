package impl

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (s *CoreService) UpdateItem(ctx context.Context, userID int64, itemID int64, update models.Update) error {

	err := s.storage.Transaction(ctx, func(tx *sql.Tx, ctx context.Context) error {

		if _, err := tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL app.current_user_id = %d", userID)); err != nil {
			return fmt.Errorf("failed to set application-level userID for transaction: %w", err)
		}

		item, err := s.storage.GetItemForUpdate(tx, ctx, itemID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errs.ErrItemNotFound
			}
			return fmt.Errorf("unable to get item for update: %w", err)
		}

		updateFields(&item, update)

		if err := s.validateItem(item); err != nil {
			return err
		}

		if err := s.storage.UpdateItem(tx, ctx, itemID, item); err != nil {
			return fmt.Errorf("error at UpdateItem: %w", err)
		}
		return nil

	})
	if unexpectedError(err) {
		s.logger.LogError("service — failed to update item in storage", err, "layer", "service.impl")
		return err
	}

	return err

}

func updateFields(old *models.Item, new models.Update) {
	if new.Name != nil {
		old.Name = *new.Name
	}
	if new.Description != nil {
		old.Description = *new.Description
	}
	if new.Quantity != nil {
		old.Quantity = *new.Quantity
	}
	if new.Price != nil {
		old.Price = *new.Price
	}
}

func unexpectedError(err error) bool {

	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, errs.ErrMissingItemName),
		errors.Is(err, errs.ErrItemNameTooShort),
		errors.Is(err, errs.ErrItemNameTooLong),
		errors.Is(err, errs.ErrItemDescriptionTooLong),
		errors.Is(err, errs.ErrItemQuantityTooLow),
		errors.Is(err, errs.ErrItemQuantityTooHigh),
		errors.Is(err, errs.ErrNegativeItemPrice),
		errors.Is(err, errs.ErrItemZeroPrice),
		errors.Is(err, errs.ErrItemPriceTooLarge),
		errors.Is(err, errs.ErrItemNotFound),
		errors.Is(err, errs.ErrItemPriceInvalidPrecision):
		return false
	default:
		return true
	}

}
