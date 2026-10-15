// internal/modules/certificate/delivery/http/errors.go

package http

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// classifiedError is the internal shape produced by
// classifyCertificateError.
type classifiedError struct {
	Status      int
	Message     string
	UserMessage string
	Details     fiber.Map
}

// classifyCertificateError maps a service-layer error to an HTTP
// status and message. All certificate-domain sentinels are handled
// here so handlers stay free of error chains.
func classifyCertificateError(action string, err error) classifiedError {
	if err == nil {
		return classifiedError{
			Status:  fiber.StatusInternalServerError,
			Message: "Unexpected error",
		}
	}

	// ---------- AUTH / PERMISSION ----------
	if errors.Is(err, domain.ErrForbidden) {
		return classifiedError{
			Status:  fiber.StatusForbidden,
			Message: "You don't have permission to " + action,
		}
	}

	// ---------- NOT FOUND ----------
	if errors.Is(err, domain.ErrCertificateNotFound) {
		return classifiedError{
			Status:  fiber.StatusNotFound,
			Message: "Certificate not found",
		}
	}
	if errors.Is(err, domain.ErrTemplateNotFound) {
		return classifiedError{
			Status:  fiber.StatusNotFound,
			Message: "Certificate template not found",
		}
	}

	// ---------- CONFLICT ----------
	if errors.Is(err, domain.ErrAlreadyIssued) {
		return classifiedError{
			Status:  fiber.StatusConflict,
			Message: "A certificate already exists for this completion",
		}
	}
	if errors.Is(err, domain.ErrCertificateAlreadyRevoked) {
		return classifiedError{
			Status:  fiber.StatusConflict,
			Message: "Certificate is already revoked",
		}
	}
	if errors.Is(err, domain.ErrDefaultTemplateExists) {
		return classifiedError{
			Status:  fiber.StatusConflict,
			Message: "Account already has a default template",
		}
	}


		if errors.Is(err, domain.ErrInvalidDownloadToken) {
		return classifiedError{
			Status:  fiber.StatusUnauthorized,
			Message: "This download link is invalid or has expired",
		}
	}
	if errors.Is(err, domain.ErrExpiredDownloadToken) {
		return classifiedError{
			Status:  fiber.StatusUnauthorized,
			Message: "This download link has expired. Please request a new one.",
		}
	}
	if errors.Is(err, domain.ErrDownloadTokenEmailMismatch) {
		return classifiedError{
			Status:  fiber.StatusUnauthorized,
			Message: "This download link is not valid for this recipient.",
		}
	}

	// ---------- VALIDATION ----------
	if ce, ok := classifyValidationError(err); ok {
		return ce
	}

	// ---------- IMMUTABLE ----------
	if errors.Is(err, domain.ErrCannotDeleteSystemTemplate) {
		return classifiedError{
			Status:  fiber.StatusUnprocessableEntity,
			Message: "System templates cannot be deleted",
		}
	}
	if errors.Is(err, domain.ErrCannotEditSystemTemplate) {
		return classifiedError{
			Status:  fiber.StatusUnprocessableEntity,
			Message: "System templates cannot be edited",
		}
	}
	if errors.Is(err, domain.ErrCertificateNotRevocable) {
		return classifiedError{
			Status:  fiber.StatusUnprocessableEntity,
			Message: "Certificate cannot be revoked",
		}
	}

	// ---------- FALLBACK ----------
	return classifiedError{
		Status:  fiber.StatusInternalServerError,
		Message: "Failed to " + action,
		Details: fiber.Map{"error": err.Error()},
	}
}

func classifyValidationError(err error) (classifiedError, bool) {
	type rule struct {
		target  error
		field   string
		message string
	}

	rules := []rule{
		{domain.ErrAccountIDRequired, "account_id", "Account ID is required"},
		{domain.ErrContextTypeRequired, "context_type", "Context type is required"},
		{domain.ErrContextIDRequired, "context_id", "Context ID is required"},
		{domain.ErrCompletedAtRequired, "completed_at", "Completed at is required"},
		{domain.ErrCompletedAtInFuture, "completed_at", "Completed at cannot be in the future"},
		{domain.ErrRecipientNameRequired, "recipient_name", "Recipient name is required"},
		{domain.ErrRecipientEmailRequired, "recipient_email", "Recipient email is required"},
		{domain.ErrTemplateNameRequired, "name", "Template name is required"},
		{domain.ErrVerificationCodeRequired, "code", "Verification code is required"},
		{domain.ErrVerificationCodeTooShort, "code", "Verification code is too short"},
		{domain.ErrVerificationCodeTooLong, "code", "Verification code is too long"},
		{domain.ErrVerificationCodeInvalidChars, "code", "Verification code contains invalid characters"},
	}

	for _, r := range rules {
		if errors.Is(err, r.target) {
			return classifiedError{
				Status:  fiber.StatusBadRequest,
				Message: r.message,
				Details: fiber.Map{
					"field":  r.field,
					"reason": r.message,
				},
			}, true
		}
	}
	return classifiedError{}, false
}

// respondClassifiedError converts a service error into an HTTP response.
func respondClassifiedError(c fiber.Ctx, action string, err error) error {
	ce := classifyCertificateError(action, err)

	details := ce.Details
	if details == nil {
		details = fiber.Map{}
	}

	msg := ce.Message
	if ce.UserMessage != "" {
		msg = ce.UserMessage
	}

	switch ce.Status {
	case fiber.StatusBadRequest:
		return response.BadRequest(c, msg, details)
	case fiber.StatusUnauthorized:
		return response.Unauthorized(c, msg, details)
	case fiber.StatusForbidden:
		return response.Forbidden(c, msg, details)
	case fiber.StatusNotFound:
		return response.NotFound(c, msg, details)
	case fiber.StatusConflict:
		return response.Conflict(c, msg, details)
	case fiber.StatusUnprocessableEntity:
		return response.UnprocessableEntity(c, msg, details)
	default:
		return response.InternalError(c, msg, details)
	}
}
