package v1

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"time"

	"github.com/wb-go/wbf/ginext"
)

func (h *Handler) CreateEvent(c *ginext.Context) {

	userID, ok := c.Request.Context().Value("userID").(int64)
	if !ok {
		RespondError(c, errs.ErrInvalidToken)
		return
	}

	var request CreateEventDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		RespondError(c, errs.ErrInvalidJSON)
		return
	}

	date, err := parseTime(request.Date)
	if err != nil {
		RespondError(c, err)
		return
	}

	bookingTTL, err := time.ParseDuration(request.BookingTTL)
	if err != nil {
		RespondError(c, errs.ErrInvalidBookingTTL)
		return
	}

	eventID, err := h.service.CreateEvent(c.Request.Context(), &models.Event{
		UserID:      userID,
		Title:       request.Title,
		Description: request.Description,
		Date:        date,
		Seats:       request.Seats,
		BookingTTL:  bookingTTL})

	if err != nil {
		RespondError(c, err)
		return
	}

	respondOK(c, eventID)

}
