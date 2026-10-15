// internal/modules/certificate/service/download_by_token.go

package service

import (
	"context"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// DownloadByToken validates a signed download token and returns the
// PDF URL for the certificate it authorizes.
//
// The token is the only credential — no authentication is required.
// This is the guest-recipient path: certificate emails contain a
// signed link that works without a Nuruvent account.
//
// Validation:
//  1. Verify the token's signature and expiry.
//  2. Load the certificate.
//  3. Confirm the token's email matches the certificate's recipient.
//  4. Confirm the certificate is active (revoked → not found).
//
// Renders the PDF on first call; returns the cached URL afterwards.
func (s *serviceImpl) DownloadByToken(
	ctx context.Context,
	token string,
) (string, error) {
	claims, err := s.deps.DownloadTokens.Verify(ctx, token)
	if err != nil {
		return "", err
	}

	cert, err := s.deps.Certificates.FindByID(ctx, claims.CertificateID)
	if err != nil {
		return "", err
	}

	// Revoked → treat as not found.
	if cert.Status != domain.StatusActive {
		return "", domain.ErrCertificateNotFound
	}

	// The token is bound to the recipient's email. Enforce it.
	if !strings.EqualFold(cert.RecipientEmail, claims.Email) {
		return "", domain.ErrDownloadTokenEmailMismatch
	}

	// Cache hit.
	if cert.PDFURL != "" {
		return cert.PDFURL, nil
	}

	return s.renderAndStore(ctx, cert)
}

// GenerateDownloadToken signs a token for the given certificate.
//
// Called when issuing a batch (to embed in email links) and by the
// manual resend endpoint.
func (s *serviceImpl) GenerateDownloadToken(
	ctx context.Context,
	cert *domain.Certificate,
) (string, error) {
	if cert == nil {
		return "", domain.ErrCertificateNotFound
	}
	return s.deps.DownloadTokens.Sign(
		ctx,
		cert.ID,
		cert.RecipientEmail,
		s.deps.DownloadTokenTTL,
	)
}