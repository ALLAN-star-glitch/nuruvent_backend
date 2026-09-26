// internal/modules/attendance/delivery/http/webhook_handler.go

package http

import (
	"github.com/gofiber/fiber/v3"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// WebhookHandler ingests provider webhooks. No authentication —
// signature verification inside the provider adapter is the
// credential.
type WebhookHandler struct {
	svc       service.Service
	providers map[attendance.SessionProvider]service.ProviderAdapter
}

func NewWebhookHandler(
	svc service.Service,
	providers map[attendance.SessionProvider]service.ProviderAdapter,
) *WebhookHandler {
	return &WebhookHandler{svc: svc, providers: providers}
}

// ZoomWebhook handles POST /webhooks/zoom.
func (h *WebhookHandler) ZoomWebhook(c fiber.Ctx) error {
	return h.ingest(c, attendance.ProviderZoom)
}

// GoogleMeetWebhook handles POST /webhooks/google-meet.
func (h *WebhookHandler) GoogleMeetWebhook(c fiber.Ctx) error {
	return h.ingest(c, attendance.ProviderGoogleMeet)
}

// ingest is the shared handler body for all providers.
//
// Flow:
//  1. Read the raw body and headers.
//  2. Resolve the adapter for the provider.
//  3. Ask the adapter to parse + verify the webhook.
//  4. If the parse returns a provider-specific verification request
//     (e.g. Zoom's endpoint.url_validation), respond with the
//     verification body instead of processing the event.
//  5. Otherwise, hand the normalized event to the service.
func (h *WebhookHandler) ingest(c fiber.Ctx, provider attendance.SessionProvider) error {
	payload := c.Body()
	if len(payload) == 0 {
		return response.BadRequest(c, "Empty webhook payload", nil)
	}

	// Collect headers as a map.
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(k, v []byte) {
		headers[string(k)] = string(v)
	})

	// Resolve the adapter for this provider.
	adapter, ok := h.providers[provider]
	if !ok {
		return response.BadRequest(c, "Unknown provider", fiber.Map{
			"error": "no adapter registered for provider " + string(provider),
		})
	}

	// Parse + verify.
	event, err := adapter.ParseWebhook(c.Context(), payload, headers)
	if err != nil {
		// Provider-specific verification handshake?
		//
		// Zoom's app registration flow sends an endpoint.url_validation
		// event. The adapter returns a *zoom.URLValidationError as a
		// signal to the handler that a specific JSON response is
		// expected. Other providers may have similar flows; they opt
		// in by implementing service.URLValidator.
		if uv, ok := adapter.(service.URLValidator); ok && uv.IsURLValidationError(err) {
			body := uv.HandleURLValidation(err)
			if body == nil {
				return response.BadRequest(c, "Webhook verification failed", fiber.Map{
					"error": "provider could not produce a verification response",
				})
			}
			return c.JSON(body)
		}

		// Any other parse error → reject.
		return response.BadRequest(c, "Webhook rejected", fiber.Map{
			"error": err.Error(),
		})
	}

	// Valid event → let the service record it.
	if err := h.svc.ProcessWebhookEvent(c.Context(), provider, event); err != nil {
		return mapDomainError(c, err)
	}

	return response.Success(c, "Webhook received", nil)
}