package v1

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/wb-go/wbf/ginext"
)

const defaultLimit = 100
const dateLayoutCSV = time.RFC3339
const defaultFilename = "history"

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

func fmtRespond(c *ginext.Context, data any) {

	if c.Query("export") != "csv" {
		respondOK(c, data)
		return
	}

	var filename string
	switch v := data.(type) {
	case []models.ItemHistory:
		if len(v) > 0 {
			filename = fmt.Sprintf("item_%d_history", v[0].ItemID)
		} else {
			filename = "item_history"
		}
	default:
		filename = defaultFilename
	}

	filename = fmt.Sprintf("%s_%s.csv", filename, time.Now().UTC().Format("20060102_150405"))

	c.Writer.Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	writer := csv.NewWriter(c.Writer)

	switch values := data.(type) {
	case []models.ItemHistory:
		if err := writeItemHistoryCSV(writer, values); err != nil {
			RespondError(c, err)
			return
		}
	default:
		RespondError(c, errs.ErrInternal)
		return
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		RespondError(c, err)
	}

}

func writeItemHistoryCSV(writer *csv.Writer, history []models.ItemHistory) error {

	header := []string{"ID", "ItemID", "UserID",
		"Action", "ChangedAt", "OldData", "NewData"}

	if err := writer.Write(header); err != nil {
		return err
	}

	for _, h := range history {
		oldDataStr := "-"
		if len(h.OldData) > 0 {
			oldDataStr = string(h.OldData)
			if len(oldDataStr) > 500 {
				oldDataStr = oldDataStr[:497] + "..."
			}
		}

		newDataStr := "-"
		if len(h.NewData) > 0 {
			newDataStr = string(h.NewData)
			if len(newDataStr) > 500 {
				newDataStr = newDataStr[:497] + "..."
			}
		}

		row := []string{
			fmt.Sprintf("%d", h.ID),
			fmt.Sprintf("%d", h.ItemID),
			fmt.Sprintf("%d", h.UserID),
			h.Action, h.ChangedAt.Format(dateLayoutCSV),
			oldDataStr, newDataStr,
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil

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
		errors.Is(err, errs.ErrInvalidUserID),
		errors.Is(err, errs.ErrMissingDate),
		errors.Is(err, errs.ErrInvalidDate),
		errors.Is(err, errs.ErrInvalidDateRange),
		errors.Is(err, errs.ErrInvalidLimit),
		errors.Is(err, errs.ErrInvalidAction),
		errors.Is(err, errs.ErrInvalidItemID):
		return http.StatusBadRequest, err.Error()

	case errors.Is(err, errs.ErrEmptyAuthHeader),
		errors.Is(err, errs.ErrInvalidToken),
		errors.Is(err, errs.ErrInvalidCredentials):
		return http.StatusUnauthorized, err.Error()

	case errors.Is(err, errs.ErrInsufficientPermissions):
		return http.StatusForbidden, err.Error()

	case errors.Is(err, errs.ErrItemNotFound):
		return http.StatusNotFound, err.Error()

	case errors.Is(err, errs.ErrUserAlreadyExists):
		return http.StatusConflict, err.Error()

	default:
		return http.StatusInternalServerError, errs.ErrInternal.Error()
	}

}
