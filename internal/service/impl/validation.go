package impl

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"
	"strings"
	"time"
)

const minBookingTTL = 1 * time.Minute
const maxBookingTTL = 24 * time.Hour

func validateUser(user models.User) error {
	if user.Login == "" {
		return errs.ErrEmptyLogin
	}
	if user.Password == "" {
		return errs.ErrEmptyPassword
	}
	return nil
}

func validateEvent(event *models.Event) error {

	if err := validateTitle(event.Title); err != nil {
		return err
	}

	if err := validateDescription(event.Description); err != nil {
		return err
	}

	if err := validateDate(event.Date); err != nil {
		return err
	}

	if err := validateSeats(event.Seats); err != nil {
		return err
	}

	if err := validateBookingTTL(event.BookingTTL); err != nil {
		return err
	}

	return nil
}

func validateTitle(title string) error {

	title = strings.TrimSpace(title)

	if title == "" {
		return errs.ErrMissingTitle
	}

	if len(title) < 3 {
		return errs.ErrTitleTooShort
	}

	if len(title) > 200 {
		return errs.ErrTitleTooLong
	}

	return nil

}

func validateDescription(desc string) error {
	if len(desc) > 2000 {
		return errs.ErrDescriptionTooLong
	}
	return nil
}

func validateDate(t time.Time) error {

	if t.IsZero() {
		return errs.ErrMissingDate
	}

	now := time.Now().UTC()

	if t.Before(now) {
		return errs.ErrDateInPast
	}
	if t.Before(now.Add(24 * time.Hour)) {
		return errs.ErrDateTooSoon
	}

	if t.After(now.AddDate(1, 0, 0)) {
		return errs.ErrDateTooFar
	}

	return nil

}

func validateSeats(seats int) error {
	if seats <= 0 {
		return errs.ErrInvalidSeatCount
	}
	if seats > 10000 {
		return errs.ErrTooManySeats
	}
	return nil
}

func validateBookingTTL(ttl time.Duration) error {

	if ttl < minBookingTTL {
		return errs.ErrBookingTTLTooShort
	}

	if ttl > maxBookingTTL {
		return errs.ErrBookingTTLTooLong
	}

	return nil

}
