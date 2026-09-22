// internal/modules/payment/delivery/http/errors.go

package http

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// mapDomainError translates a domain error to the correct HTTP response.
// Unknown errors map to 500 with a generic message.
func mapDomainError(c fiber.Ctx, err error) error {
	switch {
	// 404
	case errors.Is(err, paymentdomain.ErrOrderNotFound):
		return response.NotFound(c, "Order not found", nil)
	case errors.Is(err, paymentdomain.ErrPaymentNotFound):
		return response.NotFound(c, "Payment not found", nil)
	case errors.Is(err, paymentdomain.ErrRefundNotFound):
		return response.NotFound(c, "Refund not found", nil)

	// 400
	case errors.Is(err, paymentdomain.ErrInvalidAmount):
		return response.BadRequest(c, "Invalid amount", nil)
	case errors.Is(err, paymentdomain.ErrInvalidCurrency):
		return response.BadRequest(c, "Invalid currency", nil)
	case errors.Is(err, paymentdomain.ErrIdentityRequired):
		return response.BadRequest(c, "Identity is required", nil)
	case errors.Is(err, paymentdomain.ErrMalformedWebhook):
		return response.BadRequest(c, "Malformed webhook payload", nil)

	// 401
	case errors.Is(err, paymentdomain.ErrInvalidWebhookSignature):
		return response.Unauthorized(c, "Invalid webhook signature", nil)

	// 403
	case errors.Is(err, paymentdomain.ErrNotOwner):
		return response.Forbidden(c, "You do not own this resource", nil)
	case errors.Is(err, paymentdomain.ErrNotAuthorized):
		return response.Forbidden(c, "You are not authorized for this action", nil)

	// 409
	case errors.Is(err, paymentdomain.ErrOrderNotPending):
		return response.Conflict(c, "Order is not in a pending state", nil)
	case errors.Is(err, paymentdomain.ErrOrderExpired):
		return response.Conflict(c, "Order has expired", nil)
	case errors.Is(err, paymentdomain.ErrOrderAlreadyPaid):
		return response.Conflict(c, "Order is already paid", nil)
	case errors.Is(err, paymentdomain.ErrPaymentAlreadySucceeded):
		return response.Conflict(c, "Payment has already succeeded", nil)
	case errors.Is(err, paymentdomain.ErrPaymentNotPending):
		return response.Conflict(c, "Payment is not pending", nil)
	case errors.Is(err, paymentdomain.ErrPaymentNotSucceeded):
		return response.Conflict(c, "Payment has not succeeded", nil)
	case errors.Is(err, paymentdomain.ErrInvalidStatusTransition):
		return response.Conflict(c, "Invalid status transition", nil)

	// 422
	case errors.Is(err, paymentdomain.ErrRefundExceedsPayment):
		return response.UnprocessableEntity(c, "Refund exceeds the paid amount", nil)
	case errors.Is(err, paymentdomain.ErrPricingMismatch):
		return response.UnprocessableEntity(c, "Pricing mismatch", nil)

	// 502
	case errors.Is(err, paymentdomain.ErrProviderRejected):
		return response.BadGateway(c, "Payment provider rejected the request", nil)

	// 503
	case errors.Is(err, paymentdomain.ErrProviderUnavailable):
		return response.ServiceUnavailable(c, "Payment provider is unavailable", nil)
	case errors.Is(err, paymentdomain.ErrProviderNotFound):
		return response.ServiceUnavailable(c, "No payment provider available for this method", nil)
	}

	return response.InternalError(c, "Something went wrong", nil)
}