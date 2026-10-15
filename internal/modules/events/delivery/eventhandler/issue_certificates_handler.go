// internal/modules/events/delivery/eventhandler/issue_certificates_handler.go

package eventhandler

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// IssueCertificates handles POST /events/:id/certificates/issue.
//
// Issues certificates for the supplied recipients on behalf of the
// event. The certificate module enforces idempotency per
// (context_type, context_id, recipient) triple — re-issuing returns
// the existing certificate rather than creating a duplicate.
//
// Requires certificate:issue on the event's account. The route is
// gated by authzMiddleware, which resolves the permission from the
// path and method.
func (h *EventHandler) IssueCertificates(c fiber.Ctx) error {
	if _, err := handlerhelper.GetUserID(c); err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	var body IssueCertificatesRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	recipients := make([]service.IssueCertificateRecipientInput, 0, len(body.Recipients))
	for i, r := range body.Recipients {
		completedAt, err := time.Parse(time.RFC3339, r.CompletedAt)
		if err != nil {
			return response.BadRequest(c, "Invalid completed_at — expected RFC3339", fiber.Map{
				"index": i,
				"email": r.Email,
			})
		}
		recipients = append(recipients, service.IssueCertificateRecipientInput{
			UserID:      r.UserID,
			Name:        r.Name,
			Email:       r.Email,
			CompletedAt: completedAt,
		})
	}

	issuerName := body.IssuerName
	if issuerName == "" {
		issuerName = "Nuruvent"
	}

	result, err := h.svc.IssueCertificatesForEvent(c.Context(), service.IssueCertificatesForEventCommand{
		EventID:       eventID,
		TemplateID:    body.TemplateID,
		EventDuration: body.EventDuration,
		IssuerName:    issuerName,
		Recipients:    recipients,
	})
	if err != nil {
		return respondClassifiedError(c, "issue certificates for this event", err)
	}

	return response.Created(c, "Certificates issued", IssueCertificatesResponse{
		Issued: result.Issued,
		Reused: result.Reused,
		Failed: result.Failed,
		Errors: result.Errors,
	})
}