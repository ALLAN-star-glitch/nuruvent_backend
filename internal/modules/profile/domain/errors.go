// internal/modules/profile/domain/errors.go

package domain

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInstitutionNotFound = errors.New("institution not found")
	ErrInvalidUserID       = errors.New("invalid user ID")
	ErrInvalidInstitutionID = errors.New("invalid institution ID")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrInvalidScope		 = errors.New("invalid scope")
)