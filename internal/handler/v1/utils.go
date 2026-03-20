package v1

import (
	"Atlas/internal/errs"
	"errors"
	"fmt"
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

func getUserID(c *ginext.Context) (int64, error) {
	val, found := c.Get("userID")
	if !found {
		return 0, errs.ErrInvalidToken
	}
	return val.(int64), nil
}

func getRole(c *ginext.Context) (string, error) {
	val, found := c.Get("role")
	if !found {
		return "", errs.ErrInvalidToken
	}
	return val.(string), nil
}

func respondOK(c *ginext.Context, response any) {
	c.JSON(http.StatusOK, ginext.H{"result": response})
}

func RespondError(c *ginext.Context, err error) {
	if err != nil {
		fmt.Println(err)
		status, msg := mapErrorToStatus(err)
		c.AbortWithStatusJSON(status, ginext.H{"error": msg})
	}
}

func mapErrorToStatus(err error) (int, string) {

	switch {

	case errors.Is(err, errs.ErrInvalidJSON),
		errors.Is(err, errs.ErrEmptyLogin),
		errors.Is(err, errs.ErrEmptyPassword),
		errors.Is(err, errs.ErrEmptyRole),
		errors.Is(err, errs.ErrInvalidRole),
		errors.Is(err, errs.ErrPasswordTooLong),
		errors.Is(err, errs.ErrPasswordTooShort),
		errors.Is(err, errs.ErrLoginTooShort),
		errors.Is(err, errs.ErrLoginTooLong),
		errors.Is(err, errs.ErrLoginInvalidFormat),
		errors.Is(err, errs.ErrMissingItemName),
		errors.Is(err, errs.ErrItemNameTooShort),
		errors.Is(err, errs.ErrItemNameTooLong),
		errors.Is(err, errs.ErrItemDescriptionTooLong),
		errors.Is(err, errs.ErrItemQuantityTooLow),
		errors.Is(err, errs.ErrItemQuantityTooHigh),
		errors.Is(err, errs.ErrNegativeItemPrice),
		errors.Is(err, errs.ErrItemZeroPrice),
		errors.Is(err, errs.ErrItemPriceTooLarge),
		errors.Is(err, errs.ErrItemPriceInvalidPrecision),
		errors.Is(err, errs.ErrMissingRequiredField),
		errors.Is(err, errs.ErrInvalidFieldFormat),
		errors.Is(err, errs.ErrInvalidUserID),
		errors.Is(err, errs.ErrInvalidItemID):
		return http.StatusBadRequest, err.Error()

	case errors.Is(err, errs.ErrEmptyAuthHeader),
		errors.Is(err, errs.ErrInvalidAuthHeader),
		errors.Is(err, errs.ErrInvalidToken),
		errors.Is(err, errs.ErrInvalidCredentials):
		return http.StatusUnauthorized, err.Error()

	case errors.Is(err, errs.ErrInsufficientPermissions),
		errors.Is(err, errs.ErrActionNotAllowedForRole):
		return http.StatusForbidden, err.Error()

	case errors.Is(err, errs.ErrUserNotFound),
		errors.Is(err, errs.ErrItemNotFound),
		errors.Is(err, errs.ErrResourceNotFound):
		return http.StatusNotFound, err.Error()

	case errors.Is(err, errs.ErrUserAlreadyExists),
		errors.Is(err, errs.ErrItemAlreadyExists),
		errors.Is(err, errs.ErrCannotDeleteActiveItem):
		return http.StatusConflict, err.Error()

	default:
		return http.StatusInternalServerError, errs.ErrInternal.Error()
	}

}
