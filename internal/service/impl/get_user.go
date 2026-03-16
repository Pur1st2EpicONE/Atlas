package impl

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func (a *AuthService) GetUser(ctx context.Context, allegedUser models.User) (models.User, error) {

	if err := validateUser(allegedUser); err != nil {
		return models.User{}, err
	}

	realUser, err := a.storage.GetUserByLogin(ctx, allegedUser.Login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errs.ErrInvalidCredentials
		}
		a.logger.LogError("service — failed to get userID by login", err, "layer", "service.impl")
		return models.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(realUser.Password), []byte(allegedUser.Password)); err != nil {
		return models.User{}, errs.ErrInvalidCredentials
	}

	return realUser, nil

}
