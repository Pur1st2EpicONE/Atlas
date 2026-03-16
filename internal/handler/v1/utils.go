package v1

import (
	"Atlas/internal/errs"
	"errors"
	"net/http"
	"time"

	"github.com/wb-go/wbf/ginext"
)

func parseTime(timeStr string) (time.Time, error) {

	if timeStr == "" {
		return time.Time{}, errs.ErrMissingDate
	}

	validTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}, errs.ErrInvalidDate
	}

	return validTime.UTC(), nil

}

func respondOK(c *ginext.Context, response any) {
	c.JSON(http.StatusOK, ginext.H{"result": response})
}

func RespondError(c *ginext.Context, err error) {
	if err != nil {
		status, msg := mapErrorToStatus(err)
		c.AbortWithStatusJSON(status, ginext.H{"error": msg})
	}
}

func mapErrorToStatus(err error) (int, string) {

	switch {

	case errors.Is(err, errs.ErrInvalidJSON),
		errors.Is(err, errs.ErrEmptyLogin),
		errors.Is(err, errs.ErrEmptyPassword),
		errors.Is(err, errs.ErrPasswordTooLong),
		errors.Is(err, errs.ErrMissingDate),
		errors.Is(err, errs.ErrInvalidDate),
		errors.Is(err, errs.ErrDateInPast),
		errors.Is(err, errs.ErrDateTooFar),
		errors.Is(err, errs.ErrDateTooSoon),
		errors.Is(err, errs.ErrMissingTitle),
		errors.Is(err, errs.ErrTitleTooShort),
		errors.Is(err, errs.ErrTitleTooLong),
		errors.Is(err, errs.ErrBookingTTLTooLong),
		errors.Is(err, errs.ErrBookingTTLTooShort),
		errors.Is(err, errs.ErrDescriptionTooLong),
		errors.Is(err, errs.ErrInvalidSeatCount),
		errors.Is(err, errs.ErrInvalidUserID),
		errors.Is(err, errs.ErrInvalidBookingTTL),
		errors.Is(err, errs.ErrInvalidEventID),
		errors.Is(err, errs.ErrTooManySeats),
		errors.Is(err, errs.ErrBookingExpired):
		return http.StatusBadRequest, err.Error()

	case errors.Is(err, errs.ErrBookingNotFound),
		errors.Is(err, errs.ErrEventNotFound):
		return http.StatusNotFound, err.Error()

	case errors.Is(err, errs.ErrUserAlreadyExists),
		errors.Is(err, errs.ErrAlreadyConfirmed),
		errors.Is(err, errs.ErrBookingAlreadyExists),
		errors.Is(err, errs.ErrEventFull):
		return http.StatusConflict, err.Error()

	case errors.Is(err, errs.ErrInvalidToken),
		errors.Is(err, errs.ErrEmptyAuthHeader),
		errors.Is(err, errs.ErrInvalidAuthHeader),
		errors.Is(err, errs.ErrInvalidCredentials):
		return http.StatusUnauthorized, err.Error()

	case errors.Is(err, errs.ErrInsufficientPermissions):
		return http.StatusForbidden, err.Error()

	default:
		return http.StatusInternalServerError, errs.ErrInternal.Error()
	}

}
