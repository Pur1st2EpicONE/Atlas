package v1

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/wb-go/wbf/ginext"
)

const defaultLimit = 100

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

func parseQuery(c *ginext.Context) (models.HistoryFilter, error) {

	filter := models.HistoryFilter{Limit: defaultLimit}

	if fromStr := c.Query("from"); fromStr != "" {
		from, err := parseTime(fromStr)
		if err != nil {
			return models.HistoryFilter{}, err
		}
		filter.From = from
	}

	if toStr := c.Query("to"); toStr != "" {
		to, err := parseTime(toStr)
		if err != nil {
			return models.HistoryFilter{}, err
		}
		filter.To = to
	}

	if userStr := c.Query("user"); userStr != "" {
		userID, err := strconv.ParseInt(userStr, 10, 64)
		if err != nil || userID <= 0 {
			return models.HistoryFilter{}, errs.ErrInvalidUserID
		}
		filter.UserID = userID
	}

	if action := c.Query("action"); action != "" {
		action = strings.ToUpper(strings.TrimSpace(action))
		switch action {
		case "INSERT", "UPDATE", "DELETE":
			filter.Action = action
		default:
			return models.HistoryFilter{}, errs.ErrInvalidAction
		}
	}

	if limStr := c.Query("limit"); limStr != "" {
		limit, err := strconv.Atoi(limStr)
		if err != nil {
			return models.HistoryFilter{}, errs.ErrInvalidLimit
		}
		if limit < 1 {
			limit = 1
		}
		if limit > 2000 {
			limit = 2000
		}
		filter.Limit = limit
	}

	return filter, nil

}

func parseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, errs.ErrMissingDate
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z07:00",
	}

	for _, layout := range layouts {
		to, err := time.Parse(layout, timeStr)
		if err == nil {
			return to.UTC(), nil
		}
	}

	return time.Time{}, errs.ErrInvalidDate

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
		errors.Is(err, errs.ErrMissingDate),
		errors.Is(err, errs.ErrInvalidDate),
		errors.Is(err, errs.ErrInvalidDateRange),
		errors.Is(err, errs.ErrInvalidLimit),
		errors.Is(err, errs.ErrInvalidAction),
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
