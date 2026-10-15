// internal/modules/certificate/service/download_certificate.go

package service

import (
	"context"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// DownloadURL returns a URL the caller can use to download the
// certificate PDF.
//
// If the PDF has already been rendered, the cached URL is returned.
// Otherwise the PDF is rendered, uploaded to object storage, and the
// URL is persisted on the certificate.
//
// Authorization:
//   - The issuing account's admin (certificate:download permission)
//   - OR the recipient themselves (recipient_user_id matches)
//
// Guests (no account) are out of scope for Phase 2 — signed download
// links are a Phase 3 concern.
func (s *serviceImpl) DownloadURL(
	ctx context.Context,
	userID, certificateID string,
) (string, error) {
	cert, err := s.deps.Certificates.FindByID(ctx, certificateID)
	if err != nil {
		return "", err
	}

	if err := s.authorizeDownload(ctx, userID, cert); err != nil {
		return "", err
	}

	// Cache hit: return the existing URL.
	if cert.PDFURL != "" {
		return cert.PDFURL, nil
	}

	return s.renderAndStore(ctx, cert)
}

// ============================================================
// INTERNAL
// ============================================================

func (s *serviceImpl) authorizeDownload(
	ctx context.Context,
	userID string,
	cert *domain.Certificate,
) error {
	// Recipient self-access.
	if cert.RecipientUserID != "" && cert.RecipientUserID == userID {
		return nil
	}

	// Account-admin access.
	ok, err := s.deps.Authz.CanDownload(ctx, userID, cert.AccountID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

// renderAndStore renders the PDF, uploads it, and updates the
// certificate row with the resulting URL.
func (s *serviceImpl) renderAndStore(
	ctx context.Context,
	cert *domain.Certificate,
) (string, error) {
	tmpl, err := s.deps.Templates.FindByID(ctx, cert.TemplateID)
	if err != nil {
		return "", err
	}

	verifyURL := fmt.Sprintf("%s/verify/%s",
		trimTrailingSlash(s.deps.VerifyBaseURL),
		cert.VerificationCode.String(),
	)

	pdfBytes, err := s.deps.PDF.Render(ctx, domain.RenderInput{
		Certificate: cert,
		Template:    tmpl,
		VerifyURL:   verifyURL,
	})
	if err != nil {
		return "", fmt.Errorf("render pdf: %w", err)
	}

	objectKey := fmt.Sprintf("certificates/%s/%s.pdf", cert.AccountID, cert.ID)
	url, err := s.deps.Storage.Put(ctx, objectKey, pdfBytes, "application/pdf")
	if err != nil {
		return "", fmt.Errorf("upload pdf: %w", err)
	}

	now := s.deps.Clock.Now().UTC()
	cert.PDFURL = url
	cert.PDFGeneratedAt = &now
	cert.UpdatedAt = now
	if err := s.deps.Certificates.Update(ctx, cert); err != nil {
		return "", fmt.Errorf("persist pdf url: %w", err)
	}

	return url, nil
}

func trimTrailingSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}