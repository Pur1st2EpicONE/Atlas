package errs

import "errors"

var (
	ErrInvalidJSON = errors.New("invalid JSON format") // invalid JSON format

	ErrMissingDate = errors.New("event date is required")                              // event date is required
	ErrInvalidDate = errors.New("invalid event date format, expected RFC3339")         // invalid event date format, expected RFC3339
	ErrDateInPast  = errors.New("event date cannot be in the past")                    // event date cannot be in the past
	ErrDateTooFar  = errors.New("event date cannot be more than 1 year in the future") // event date cannot be more than 1 year in the future
	ErrDateTooSoon = errors.New("event date must be at least 24 hours in the future")  // event date must be at least 24 hours in the future

	ErrMissingTitle       = errors.New("event title is required")                                     // event title is required
	ErrTitleTooShort      = errors.New("event title must be at least 3 characters long")              // event title must be at least 3 characters long
	ErrTitleTooLong       = errors.New("event title exceeds maximum length of 100 characters")        // event title exceeds maximum length of 100 characters
	ErrDescriptionTooLong = errors.New("event description exceeds maximum length of 1000 characters") // event description exceeds maximum length of 1000 characters
	ErrBookingTTLTooLong  = errors.New("booking TTL is too long: maximum 1 day")                      // booking TTL is too long: maximum 1 day
	ErrBookingTTLTooShort = errors.New("booking TTL is too short: minimum 1 minute")                  // booking TTL is too short: minimum 1 minute

	ErrInvalidSeatCount = errors.New("total seats must be greater than zero")       // total seats must be greater than zero
	ErrTooManySeats     = errors.New("total seats exceeds maximum allowed of 1000") // total seats exceeds maximum allowed of 1000

	ErrInternal = errors.New("internal server error") // internal server error

	ErrUserAlreadyExists  = errors.New("user already exists")                 // user already exists
	ErrPasswordTooLong    = errors.New("password is too long")                // password is too long
	ErrInvalidCredentials = errors.New("invalid login or password")           // invalid login or password
	ErrEmptyLogin         = errors.New("login field can not be empty")        // login field can not be empty
	ErrEmptyPassword      = errors.New("password field can not be empty")     // password field can not be empty
	ErrEmptyAuthHeader    = errors.New("authorization header is empty")       // authorization header is empty
	ErrInvalidAuthHeader  = errors.New("invalid authorization header format") // invalid authorization header format
	ErrInvalidToken       = errors.New("invalid or expired token")            // invalid or expired token
	ErrInvalidUserID      = errors.New("invalid userID")                      // invalid userID

	ErrInvalidEventID = errors.New("event_id is empty or invalid") // event_id is empty or invalid
	ErrEventFull      = errors.New("event is full: no seats left") // event is full: no seats left
	ErrEventNotFound  = errors.New("event not found")              // event not found

	ErrBookingAlreadyExists = errors.New("booking already exists for this user and event") // booking already exists for this user and event
	ErrAlreadyConfirmed     = errors.New("booking is already confirmed")                   // booking is already confirmed
	ErrInvalidBookingTTL    = errors.New("invalid booking expiration time")                // invalid booking expiration time
	ErrBookingExpired       = errors.New("booking has expired")                            // booking has expired
	ErrBookingNotFound      = errors.New("booking not found")                              // booking not found
)
