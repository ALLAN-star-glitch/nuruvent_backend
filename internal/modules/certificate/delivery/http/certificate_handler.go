// internal/modules/certificate/delivery/http/certificate_handler.go

package http

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// Issue handles POST /certificates.
func (h *Handler) Issue(c fiber.Ctx) error {
	if _, err := handlerhelper.GetUserID(c); err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	var body IssueCertificateRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	completedAt, err := time.Parse(time.RFC3339, body.CompletedAt)
	if err != nil {
		return response.BadRequest(c, "Invalid completed_at — expected RFC3339", fiber.Map{
			"field": "completed_at",
		})
	}

	cert, err := h.svc.Issue(c.Context(), domain.ExternalCompletion{
		ContextType: body.ContextType,
		ContextID:   body.ContextID,
		Recipient: domain.Recipient{
			UserID: body.RecipientUserID,
			Name:   body.RecipientName,
			Email:  body.RecipientEmail,
		},
		CompletedAt:      completedAt,
		Metadata:         domain.Metadata(body.Metadata),
		TemplateID:       body.TemplateID,
		AccountID:        body.AccountID,
		PaymentReference: body.PaymentReference,
	})
	if err != nil {
		return respondClassifiedError(c, "issue certificate", err)
	}

	return response.Created(c, "Certificate issued", toCertificateResponse(cert))
}

// IssueBulk handles POST /certificates/bulk.
func (h *Handler) IssueBulk(c fiber.Ctx) error {
	if _, err := handlerhelper.GetUserID(c); err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	var body IssueBulkRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	completions := make([]domain.ExternalCompletion, 0, len(body.Completions))
	for _, item := range body.Completions {
		completedAt, err := time.Parse(time.RFC3339, item.CompletedAt)
		if err != nil {
			return response.BadRequest(c, "Invalid completed_at — expected RFC3339", fiber.Map{
				"field":  "completed_at",
				"reason": item.RecipientEmail,
			})
		}
		completions = append(completions, domain.ExternalCompletion{
			ContextType: item.ContextType,
			ContextID:   item.ContextID,
			Recipient: domain.Recipient{
				UserID: item.RecipientUserID,
				Name:   item.RecipientName,
				Email:  item.RecipientEmail,
			},
			CompletedAt:      completedAt,
			Metadata:         domain.Metadata(item.Metadata),
			TemplateID:       item.TemplateID,
			AccountID:        item.AccountID,
			PaymentReference: item.PaymentReference,
		})
	}

	result, err := h.svc.IssueBulk(c.Context(), completions)
	if err != nil {
		return respondClassifiedError(c, "issue certificates", err)
	}

	return response.Created(c, "Bulk issuance complete", toBulkResponse(result))
}

// Verify handles GET /certificates/verify/:code — public, no auth.
//
// TODO(BRD FR-CV-006): rate-limit this route once a limiter exists
// in the shared middleware stack.
func (h *Handler) Verify(c fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return response.BadRequest(c, "Verification code is required", nil)
	}

	cert, err := h.svc.GetByCode(c.Context(), code)
	if err != nil {
		return respondClassifiedError(c, "verify certificate", err)
	}

	return response.Success(c, "Certificate verified", toVerifyResponse(cert))
}

// RevokeByUser handles POST /certificates/:id/revoke.
func (h *Handler) RevokeByUser(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Certificate ID is required", nil)
	}

	var body RevokeCertificateRequest
	_ = c.Bind().Body(&body)

	if err := h.svc.RevokeByUser(c.Context(), userID, id, body.Reason); err != nil {
		return respondClassifiedError(c, "revoke certificate", err)
	}

	return response.Success(c, "Certificate revoked", nil)
}

// ListByAccount handles GET /accounts/:id/certificates.
func (h *Handler) ListByAccount(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	accountID := c.Params("id")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	filter := parseCertificateFilter(c)

	certs, err := h.svc.ListByAccount(c.Context(), userID, accountID, filter)
	if err != nil {
		return respondClassifiedError(c, "list certificates", err)
	}

	return response.Success(c, "Certificates retrieved", toCertificateListResponse(certs))
}

// ListMine handles GET /me/certificates.
func (h *Handler) ListMine(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	filter := parseCertificateFilter(c)

	certs, err := h.svc.ListMine(c.Context(), userID, filter)
	if err != nil {
		return respondClassifiedError(c, "list certificates", err)
	}

	return response.Success(c, "Certificates retrieved", toCertificateListResponse(certs))
}

// Download handles GET /certificates/:id/download.
//
// Returns a 302 redirect to the certificate's PDF URL. Triggers
// rendering on first call.
func (h *Handler) Download(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Certificate ID is required", nil)
	}

	url, err := h.svc.DownloadURL(c.Context(), userID, id)
	if err != nil {
		return respondClassifiedError(c, "download certificate", err)
	}

	return c.Redirect().To(url)
}

// ============================================================
// BATCH ISSUANCE (organizer-pays)
// ============================================================

// InitiateBatchIssuance handles POST /events/:id/certificates/issue.
//
// Resolves the template's price:
//
//   - price == 0 → issues all certificates immediately,
//                  returns {type:"free", issued, reused, failed}
//   - price > 0  → creates a pending payment order for the batch,
//                  returns {type:"paid", order_id, amount_minor, currency}
//
// Idempotent: a repeat call returns the existing batch (free) or the
// existing pending order (paid).
func (h *Handler) InitiateBatchIssuance(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	eventID := c.Params("id")
	if eventID == "" {
		return response.BadRequest(c, "Event ID is required", nil)
	}

	var body InitiateBatchIssuanceRequest
	if err := c.Bind().Body(&body); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	if len(body.Recipients) == 0 {
		return response.BadRequest(c, "At least one recipient is required", fiber.Map{
			"field": "recipients",
		})
	}

	recipients := make([]domain.BatchRecipient, 0, len(body.Recipients))
	for i, r := range body.Recipients {
		completedAt, err := time.Parse(time.RFC3339, r.CompletedAt)
		if err != nil {
			return response.BadRequest(c, "Invalid completed_at — expected RFC3339", fiber.Map{
				"index": i,
				"email": r.Email,
			})
		}
		recipients = append(recipients, domain.BatchRecipient{
			UserID:      r.UserID,
			Name:        r.Name,
			Email:       r.Email,
			CompletedAt: completedAt,
		})
	}

	eventDate, err := time.Parse(time.RFC3339, body.EventDate)
	if err != nil {
		return response.BadRequest(c, "Invalid event_date — expected RFC3339", fiber.Map{
			"field": "event_date",
		})
	}

	result, err := h.svc.InitiateBatchIssuance(c.Context(), userID, service.InitiateBatchIssuanceInput{
		EventID:       eventID,
		AccountID:     body.AccountID,
		TemplateID:    body.TemplateID,
		EventTitle:    body.EventTitle,
		EventDate:     eventDate,
		EventDuration: body.EventDuration,
		IssuerName:    body.IssuerName,
		Recipients:    recipients,
	})
	if err != nil {
		return respondClassifiedError(c, "initiate batch issuance", err)
	}

	resp := InitiateBatchIssuanceResponse{
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

// ============================================================
// HELPERS
// ============================================================

func parseCertificateFilter(c fiber.Ctx) domain.CertificateFilter {
	f := domain.CertificateFilter{
		ContextType: handlerhelper.GetQueryString(c, "context_type", ""),
		ContextID:   handlerhelper.GetQueryString(c, "context_id", ""),
		Limit:       handlerhelper.GetQueryInt(c, "limit", 50),
		Offset:      handlerhelper.GetQueryInt(c, "offset", 0),
	}
	if status := handlerhelper.GetQueryString(c, "status", ""); status != "" {
		s := domain.Status(status)
		f.Status = &s
	}
	if from := handlerhelper.GetQueryString(c, "from", ""); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			f.From = &t
		}
	}
	if to := handlerhelper.GetQueryString(c, "to", ""); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			f.To = &t
		}
	}
	return f
}

// ============================================================
// DOWNLOAD BY TOKEN (public)
// ============================================================

// DownloadByToken handles GET /certificates/download/:token.
//
// Public — the signed token in the URL is the credential. Used by
// certificate emails so recipients (registered or guest) can download
// their PDF without authenticating.
func (h *Handler) DownloadByToken(c fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return response.BadRequest(c, "Token is required", nil)
	}

	url, err := h.svc.DownloadByToken(c.Context(), token)
	if err != nil {
		return respondClassifiedError(c, "download certificate", err)
	}

	return c.Redirect().To(url)
}

// ============================================================
// SEND (manual re-send)
// ============================================================

// SendCertificate handles POST /certificates/:id/send.
//
// Regenerates the download link and re-enqueues the certificate
// email. Authorized for the recipient and for account admins.
func (h *Handler) SendCertificate(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "Authentication required", nil)
	}

	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Certificate ID is required", nil)
	}

	if err := h.svc.SendCertificateEmail(c.Context(), userID, id); err != nil {
		return respondClassifiedError(c, "send certificate email", err)
	}

	return response.Success(c, "Certificate email sent", nil)
}