// internal/modules/certificate/domain/errors.go

package domain

import "errors"

var (
	// Input validation
	ErrAccountIDRequired      = errors.New("account id is required")
	ErrContextTypeRequired    = errors.New("context type is required")
	ErrContextIDRequired      = errors.New("context id is required")
	ErrCompletedAtRequired    = errors.New("completed_at is required")
	ErrCompletedAtInFuture    = errors.New("completed_at cannot be in the future")
	ErrRecipientNameRequired  = errors.New("recipient name is required")
	ErrRecipientEmailRequired = errors.New("recipient email is required")
	ErrTemplateNameRequired   = errors.New("template name is required")

	// Templates
	ErrTemplateNotFound           = errors.New("certificate template not found")
	ErrDefaultTemplateExists      = errors.New("account already has a default template")
	ErrCannotDeleteSystemTemplate = errors.New("system templates cannot be deleted")
	ErrInvalidSystemTemplateSlug  = errors.New("invalid system template slug")

	// Verification codes
	ErrVerificationCodeRequired     = errors.New("verification code is required")
	ErrVerificationCodeTooShort     = errors.New("verification code is too short")
	ErrVerificationCodeTooLong      = errors.New("verification code is too long")
	ErrVerificationCodeInvalidChars = errors.New("verification code contains invalid characters")
	ErrVerificationCodeCollision    = errors.New("verification code collision")

	// Certificates
	ErrCertificateNotFound       = errors.New("certificate not found")
	ErrCertificateAlreadyRevoked = errors.New("certificate is already revoked")
	ErrCertificateNotRevocable   = errors.New("certificate cannot be revoked")
	ErrAlreadyIssued             = errors.New("certificate already issued for this completion")


	// Authorization
	ErrForbidden = errors.New("forbidden")


		// Templates (additional)
	ErrCannotEditSystemTemplate = errors.New("system templates cannot be edited")



		// Payment / request flow
	ErrPaymentRequired       = errors.New("payment is required before the certificate can be issued")
	ErrPaymentChargeFailed   = errors.New("failed to create certificate order")
	ErrCertificateNotPayable = errors.New("certificate is not payable for this context")
	ErrAccountPriceLookup    = errors.New("failed to resolve account certificate price")


		// Download tokens
	ErrInvalidDownloadToken = errors.New("invalid download token")
	ErrExpiredDownloadToken = errors.New("download token has expired")
	ErrDownloadTokenEmailMismatch = errors.New("download token recipient does not match certificate")
	
)

var ErrPDFPending = errors.New("pdf: rendering in progress")


var (
	ErrSystemTemplateSlugRequired = errors.New("system template requires a valid slug")
)

var ErrSystemTemplateAlreadyExists = errors.New("a system template already exists for this tier")