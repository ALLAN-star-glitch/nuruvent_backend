package validation

import "errors"

var (
	ErrInvalidPhone            = errors.New("invalid phone number")
	ErrPhoneRegionNotSupported = errors.New("phone region not supported")
)