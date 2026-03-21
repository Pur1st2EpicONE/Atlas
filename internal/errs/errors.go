package errs

import "errors"

var (
	ErrInvalidJSON               = errors.New("invalid JSON format")                                // invalid JSON format
	ErrEmptyLogin                = errors.New("login field cannot be empty")                        // login field cannot be empty
	ErrEmptyPassword             = errors.New("password field cannot be empty")                     // password field cannot be empty
	ErrEmptyRole                 = errors.New("role field cannot be empty")                         // role field cannot be empty
	ErrInvalidRole               = errors.New("only admin, maganer and viewer roles are availible") // only admin, maganer and viewer roles are availible
	ErrPasswordTooLong           = errors.New("password is too long")                               // password is too long
	ErrPasswordTooShort          = errors.New("password is too short")                              // password is too short
	ErrLoginTooShort             = errors.New("login is too short")                                 // login is too short
	ErrLoginTooLong              = errors.New("login is too long")                                  // login is too long
	ErrLoginInvalidFormat        = errors.New("login contains invalid characters")                  // login contains invalid characters
	ErrMissingItemName           = errors.New("item name is required")                              // item name is required
	ErrItemNameTooShort          = errors.New("item name is too short")                             // item name is too short
	ErrItemNameTooLong           = errors.New("item name is too long")                              // item name is too long
	ErrItemDescriptionTooLong    = errors.New("item description is too long")                       // item description is too long
	ErrItemQuantityTooLow        = errors.New("quantity must be at least 1")                        // quantity must be at least 1
	ErrItemQuantityTooHigh       = errors.New("quantity is too large")                              // quantity is too large
	ErrNegativeItemPrice         = errors.New("price cannot be negative")                           // price cannot be negative
	ErrItemZeroPrice             = errors.New("price cannot be zero")                               // price cannot be zero
	ErrItemPriceTooLarge         = errors.New("price is too high")                                  // price is too high
	ErrItemPriceInvalidPrecision = errors.New("price must have at most 2 decimal places")           // price must have at most 2 decimal places

	ErrMissingRequiredField = errors.New("missing required field")      // missing required field
	ErrInvalidFieldFormat   = errors.New("invalid field format")        // invalid field format
	ErrInvalidUserID        = errors.New("invalid userID")              // invalid userID
	ErrInvalidItemID        = errors.New("item_id is empty or invalid") // item_id is empty or invalid

	ErrEmptyAuthHeader    = errors.New("authorization header is empty")       // authorization header is empty
	ErrInvalidAuthHeader  = errors.New("invalid authorization header format") // invalid authorization header format
	ErrInvalidToken       = errors.New("invalid or expired token")            // invalid or expired token
	ErrInvalidCredentials = errors.New("invalid login or password")           // invalid login or password

	ErrInsufficientPermissions = errors.New("insufficient permissions")            // insufficient permissions
	ErrActionNotAllowedForRole = errors.New("action not allowed for current role") // action not allowed for current role

	ErrUserNotFound     = errors.New("user not found")     // user not found
	ErrItemNotFound     = errors.New("item not found")     // item not found
	ErrResourceNotFound = errors.New("resource not found") // resource not found

	ErrUserAlreadyExists      = errors.New("user already exists")                // user already exists
	ErrItemAlreadyExists      = errors.New("item with this name already exists") // item with this name already exists
	ErrCannotDeleteActiveItem = errors.New("cannot delete active/used item")     // cannot delete active/used item

	ErrInternal         = errors.New("internal server error")                               // internal server error
	ErrMissingDate      = errors.New("missing date")                                        // missing date
	ErrInvalidDate      = errors.New("invalid date format, expected RFC3339 or YYYY-MM-DD") // invalid date format, expected RFC3339 or YYYY-MM-DD
	ErrInvalidDateRange = errors.New("from date must be before or equal to to date")        // from date must be before or equal to to date
	ErrInvalidLimit     = errors.New("invalid limit")                                       // invalid limit
	ErrInvalidAction    = errors.New("invalid action")                                      // invalid action
)
