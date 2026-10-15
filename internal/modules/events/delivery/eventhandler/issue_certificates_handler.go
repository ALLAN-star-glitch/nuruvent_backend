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
// Starts a certificate batch for the event. Free batches issue
// immediately; paid batches create a pending payment order that the
// organizer must settle before certificates are issued.
//
// Idempotent per event.
func (h *EventHandler) IssueCertificates(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
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

	// Validate mode.
	mode := body.Mode
	if mode == "" {
		mode = service.IssueModeAdd
	}
	if mode != service.IssueModeAdd && mode != service.IssueModeReplace {
		return response.BadRequest(c, "Invalid mode — must be 'add' or 'replace'", fiber.Map{
			"field": "mode",
		})
	}
	if mode == service.IssueModeReplace && len(body.Recipients) == 0 {
		return response.BadRequest(c, "recipients is required when mode is 'replace'", fiber.Map{
			"field": "recipients",
		})
	}

	// Parse recipients.
	recipients := make([]service.IssueCertificateRecipientInput, 0, len(body.Recipients))
	for i, r := range body.Recipients {
		if r.Name == "" || r.Email == "" {
			return response.BadRequest(c, "Each recipient must have name and email", fiber.Map{
				"index": i,
			})
		}
		var completedAt time.Time
		if r.CompletedAt != "" {
			t, err := time.Parse(time.RFC3339, r.CompletedAt)
			if err != nil {
				return response.BadRequest(c, "Invalid completed_at — expected RFC3339", fiber.Map{
					"index": i,
					"email": r.Email,
				})
			}
			completedAt = t
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

	result, err := h.svc.IssueCertificatesForEvent(c.Context(), userID, service.IssueCertificatesForEventCommand{
		EventID:       eventID,
		TemplateID:    body.TemplateID,
		EventDuration: body.EventDuration,
		IssuerName:    issuerName,
		Mode:          mode,
		Recipients:    recipients,
	})
	if err != nil {
		return respondClassifiedError(c, "issue certificates for this event", err)
	}

	resp := IssueCertificatesResponse{
		Type:          result.Type,
		Issued:        result.Issued,
		Reused:        result.Reused,
		Failed:        result.Failed,
		OrderID:       result.OrderID,
		AmountMinor:   result.AmountMinor,
		Currency:      result.Currency,
		AttendeeCount: result.AttendeeCount,
	}

	status := fiber.StatusOK
	if result.Type == "free" {
		status = fiber.StatusCreated
	}

	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": "Certificate batch initiated",
		"data":    resp,
	})
}