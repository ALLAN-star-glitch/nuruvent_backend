// internal/modules/certificate/delivery/http/mappers.go

package http

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	certificateService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
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
		IssuerName:       c.Metadata.Get(domain.MetadataKeyIssuerName), // <-- MAP THIS
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

	// AccountID is a pointer. System templates have nil AccountID,
	// which we surface as an empty string in the response.
	accountID := ""
	if t.AccountID != nil {
		accountID = *t.AccountID
	}

	return TemplateResponse{
		ID:          t.ID,
		AccountID:   accountID,
		Name:        t.Name,
		Description: t.Description,
		Design:      designToDTO(t.Design),
		PriceMinor:  t.PriceMinor,
		SystemSlug:  string(t.SystemSlug),
		IsDefault:   t.IsDefault,
		IsSystem:    t.IsSystem,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// designToDTO converts a domain design to its wire form.
func designToDTO(d domain.TemplateDesign) TemplateDesignDTO {
	return TemplateDesignDTO{
		BackgroundColor:    d.BackgroundColor,
		TextColor:          d.TextColor,
		AccentColor:        d.AccentColor,
		FontFamily:         d.FontFamily,
		LogoURL:            d.LogoURL,
		SealURL:            d.SealURL,
		SignatureURL:       d.SignatureURL,
		BackgroundURL:      d.BackgroundURL,
		BackgroundOpacity:  d.BackgroundOpacity,
		BackgroundSize:     d.BackgroundSize,
		BackgroundPosition: d.BackgroundPosition,
		SignatoryName:      d.SignatoryName,
		SignatoryTitle:     d.SignatoryTitle,
		TitleText:          d.TitleText,
		SubtitleText:       d.SubtitleText,
		BodyText:           d.BodyText,
		FooterText:         d.FooterText,
		QRPosition:         d.QRPosition,
		BorderStyle:        d.BorderStyle,
		BorderColor:        d.BorderColor,
		SignatureLine:      d.SignatureLine,
		WatermarkText:      d.WatermarkText,
		VerifyBadge:        d.VerifyBadge,
	}
}

// designFromDTO converts the wire form back to a domain design.
func designFromDTO(d TemplateDesignDTO) domain.TemplateDesign {
	return domain.TemplateDesign{
		BackgroundColor:    d.BackgroundColor,
		TextColor:          d.TextColor,
		AccentColor:        d.AccentColor,
		FontFamily:         d.FontFamily,
		LogoURL:            d.LogoURL,
		SealURL:            d.SealURL,
		SignatureURL:       d.SignatureURL,
		BackgroundURL:      d.BackgroundURL,
		BackgroundOpacity:  d.BackgroundOpacity,
		BackgroundSize:     d.BackgroundSize,
		BackgroundPosition: d.BackgroundPosition,
		SignatoryName:      d.SignatoryName,
		SignatoryTitle:     d.SignatoryTitle,
		TitleText:          d.TitleText,
		SubtitleText:       d.SubtitleText,
		BodyText:           d.BodyText,
		FooterText:         d.FooterText,
		QRPosition:         d.QRPosition,
		BorderStyle:        d.BorderStyle,
		BorderColor:        d.BorderColor,
		SignatureLine:      d.SignatureLine,
		WatermarkText:      d.WatermarkText,
		VerifyBadge:        d.VerifyBadge,
	}
}