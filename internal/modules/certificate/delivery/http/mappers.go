// internal/modules/certificate/delivery/http/mappers.go

package http

import (
	certificateService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// ============================================================
// CERTIFICATE MAPPERS
// ============================================================

func toCertificateResponse(c *domain.Certificate) CertificateResponse {
	if c == nil {
		return CertificateResponse{}
	}
	return CertificateResponse{
		ID:               c.ID,
		AccountID:        c.AccountID,
		TemplateID:       c.TemplateID,
		VerificationCode: c.VerificationCode.String(),
		RecipientName:    c.RecipientName,
		RecipientEmail:   c.RecipientEmail,
		RecipientUserID:  c.RecipientUserID,
		ContextType:      c.ContextType,
		ContextID:        c.ContextID,
		Metadata:         map[string]string(c.Metadata),
		CompletedAt:      c.CompletedAt,
		IssuedAt:         c.IssuedAt,
		Status:           string(c.Status),
		RevokedAt:        c.RevokedAt,
		RevocationReason: c.RevocationReason,
		PDFURL:           c.PDFURL,
	}
}

func toVerifyResponse(c *domain.Certificate) VerifyResponse {
	if c == nil {
		return VerifyResponse{}
	}
	return VerifyResponse{
		VerificationCode: c.VerificationCode.String(),
		RecipientName:    c.RecipientName,
		ContextType:      c.ContextType,
		ContextID:        c.ContextID,
		Metadata:         map[string]string(c.Metadata),
		CompletedAt:      c.CompletedAt,
		IssuedAt:         c.IssuedAt,
		Status:           string(c.Status),
	}
}

func toCertificateListResponse(certs []*domain.Certificate) map[string]any {
	out := make([]CertificateResponse, 0, len(certs))
	for _, c := range certs {
		out = append(out, toCertificateResponse(c))
	}
	return map[string]any{
		"count":        len(out),
		"certificates": out,
	}
}

func toBulkResponse(r *certificateService.BulkResult) IssueBulkResponse {
	if r == nil {
		return IssueBulkResponse{}
	}
	items := make([]IssueBulkItemResponse, 0, len(r.Items))
	for _, it := range r.Items {
		item := IssueBulkItemResponse{
			RecipientEmail: it.Completion.Recipient.Email,
			AlreadyHad:     it.AlreadyHad,
		}
		if it.Certificate != nil {
			item.CertificateID = it.Certificate.ID
			item.VerificationCode = it.Certificate.VerificationCode.String()
		}
		if it.Err != nil {
			item.Error = it.Err.Error()
		}
		items = append(items, item)
	}
	return IssueBulkResponse{
		Succeeded: r.Succeeded,
		Failed:    r.Failed,
		Reused:    r.Reused,
		Items:     items,
	}
}

// ============================================================
// TEMPLATE MAPPERS
// ============================================================

func toTemplateResponse(t *domain.Template) TemplateResponse {
	if t == nil {
		return TemplateResponse{}
	}
	return TemplateResponse{
		ID:          t.ID,
		AccountID:   t.AccountID,
		Name:        t.Name,
		Description: t.Description,
		Design:      designToDTO(t.Design),
		PriceMinor:  t.PriceMinor, // ← NEW
		SystemSlug:  string(t.SystemSlug),
		IsDefault:   t.IsDefault,
		IsSystem:    t.IsSystem,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}