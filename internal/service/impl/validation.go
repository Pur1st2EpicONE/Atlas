package impl

import (
	"Atlas/internal/errs"
	"Atlas/internal/models"

	"github.com/shopspring/decimal"
)

const bcryptMaxLen = 72 // bcryptMaxLen is the maximum allowed password length for bcrypt hashing

// validateNewUser performs validation on a new user before creation.
// It checks login, password, and role.
func (a *AuthService) validateNewUser(user models.User) error {

	if err := a.validateLogin(user.Login); err != nil {
		return err
	}
	if err := a.validatePassword(user.Password); err != nil {
		return err
	}
	if err := a.validateRole(user.Role); err != nil {
		return err
	}

	return nil
}

// validateLogin checks that the login meets length constraints.
func (a *AuthService) validateLogin(login string) error {

	length := len(login)

	if length == 0 {
		return errs.ErrEmptyLogin
	}
	if length < a.config.MinLoginLength {
		return errs.ErrLoginTooShort
	}
	if length > a.config.MaxLoginLength {
		return errs.ErrLoginTooLong
	}

	return nil

}

// validatePassword checks that the password is non‑empty,
// meets minimum length, and does not exceed bcrypt limit.
func (a *AuthService) validatePassword(password string) error {

	length := len(password)

	if length == 0 {
		return errs.ErrEmptyPassword
	}
	if length < a.config.MinPasswordLength {
		return errs.ErrPasswordTooShort
	}
	if length > bcryptMaxLen {
		return errs.ErrPasswordTooLong
	}

	return nil

}

// validateRole ensures the role is one of the predefined values (admin, manager, viewer).
func (a *AuthService) validateRole(role string) error {

	if role == "" {
		return errs.ErrEmptyRole
	}

	if role != models.Admin &&
		role != models.Manager &&
		role != models.Viewer {
		return errs.ErrInvalidRole
	}

	return nil

}

// validateUser checks that a user object has non‑empty login and password.
// Used when authenticating an existing user.
func (a *AuthService) validateUser(user models.User) error {
	if user.Login == "" {
		return errs.ErrEmptyLogin
	}
	if user.Password == "" {
		return errs.ErrEmptyPassword
	}
	return nil
}

// validateItem validates all fields of an item (name, description, quantity, price).
func (s *CoreService) validateItem(item models.Item) error {

	if err := s.validateName(item.Name); err != nil {
		return err
	}
	if err := s.validateDescription(item.Description); err != nil {
		return err
	}
	if err := s.validateQuantity(item.Quantity); err != nil {
		return err
	}
	if err := s.validatePrice(item.Price); err != nil {
		return err
	}

	return nil

}

// validateName checks that the item name is non‑empty and within length limits.
func (s *CoreService) validateName(name string) error {

	length := len(name)

	if length == 0 {
		return errs.ErrMissingItemName
	}
	if length < s.config.MinItemNameLength {
		return errs.ErrItemNameTooShort
	}
	if length > s.config.MaxItemNameLength {
		return errs.ErrItemNameTooLong
	}

	return nil

}

// validateDescription checks that the description does not exceed the maximum allowed length.
func (s *CoreService) validateDescription(description string) error {
	if len(description) > s.config.MaxItemDescriptionLength {
		return errs.ErrItemDescriptionTooLong
	}
	return nil
}

// validateQuantity checks that the quantity is within defined minimum and maximum.
func (s *CoreService) validateQuantity(quantity int) error {
	if quantity < s.config.MinItemQuantity {
		return errs.ErrItemQuantityTooLow
	}
	if quantity > s.config.MaxItemQuantity {
		return errs.ErrItemQuantityTooHigh
	}
	return nil
}

// validatePrice ensures the price is non‑negative, not zero, does not exceed the maximum,
// and has at most two decimal places.
func (s *CoreService) validatePrice(price decimal.Decimal) error {

	if price.IsNegative() {
		return errs.ErrNegativeItemPrice
	}
	if price.IsZero() {
		return errs.ErrItemZeroPrice
	}
	if price.GreaterThan(decimal.NewFromInt(s.config.MaxItemPrice)) {
		return errs.ErrItemPriceTooLarge
	}
	if !price.Sub(price.Truncate(2)).Equal(decimal.Zero) {
		return errs.ErrItemPriceInvalidPrecision
	}

	return nil

}

// validateFilter checks that the From date is not after the To date.
func (s *CoreService) validateFilter(filter models.HistoryFilter) error {

	if !filter.From.IsZero() && !filter.To.IsZero() && filter.From.After(filter.To) {
		return errs.ErrInvalidDateRange
	}

	return nil

}
