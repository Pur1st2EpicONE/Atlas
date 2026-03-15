package impl

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (a *AuthService) GetUserId(ctx context.Context, allegedUser models.User) (int64, error) {

	if err := validateUser(allegedUser); err != nil {
		return 0, err
	}

	realUser, err := a.storage.GetUserByLogin(ctx, allegedUser.Login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errs.ErrInvalidCredentials
		}
		a.logger.LogError("service — failed to get userID by login", err, "layer", "service.impl")
		return 0, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(realUser.Password), []byte(allegedUser.Password)); err != nil {
		return 0, errs.ErrInvalidCredentials
	}

	return int64(realUser.ID), nil

}
