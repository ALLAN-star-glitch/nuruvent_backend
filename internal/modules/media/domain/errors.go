package domain

import "errors"

var (
	// Media errors
	ErrMediaNotFound   = errors.New("media not found")
	ErrMediaExists     = errors.New("media already exists")
	ErrInvalidMedia    = errors.New("invalid media data")
	ErrUploadFailed    = errors.New("file upload failed")

	// Media type errors
	ErrMediaTypeNotFound = errors.New("media type not found")
	ErrInvalidMediaType  = errors.New("invalid media type")

	// Permission errors
	ErrUnauthorizedMediaAccess = errors.New("unauthorized access to media")
)