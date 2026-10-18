// internal/modules/certificate/service/issue_batch.go

package service

import (
	"context"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// IssueBatch is the public entry point invoked by the payment webhook
// adapter after a batch order is paid. It builds the ExternalCompletion
// for each recipient, delegates to IssueBulk, then enqueues a
// certificate-ready email for each newly-issued certificate.
//
// Idempotent per recipient — existing certificates are returned
// unchanged, so a webhook re-delivery cannot duplicate issuance. Email
// sending is also idempotent-friendly: certificates that were already
// issued are skipped for email so retries don't re-send.
func (s *serviceImpl) IssueBatch(
	ctx context.Context,
	req domain.BatchIssuanceRequest,
) (*domain.BatchIssuanceResult, error) {
	if req.EventID == "" {
		return nil, domain.ErrContextIDRequired
	}
	if req.AccountID == "" {
		return nil, domain.ErrAccountIDRequired
	}
	if len(req.Recipients) == 0 {
		return &domain.BatchIssuanceResult{}, nil
	}

	// ------------------------------------------------------------
	// Resolve the issuer name.
	//
	// Precedence:
	//   1. Explicit req.IssuerName (caller override)
	//   2. Account's display name (from the account module)
	//   3. Safe fallback to the platform name
	// ------------------------------------------------------------
	issuerName := req.IssuerName
	if issuerName == "" && s.deps.Accounts != nil {
		if name, err := s.deps.Accounts.GetAccountName(ctx, req.AccountID); err == nil && name != "" {
			issuerName = name
		}
	}
	if issuerName == "" {
		issuerName = "Nuruvent"
	}

	completions := make([]domain.ExternalCompletion, 0, len(req.Recipients))
	for _, r := range req.Recipients {
		completions = append(completions, domain.ExternalCompletion{
			ContextType: "event",
			ContextID:   req.EventID,
			Recipient: domain.Recipient{
				UserID: r.UserID,
				Name:   r.Name,
				Email:  r.Email,
			},
			CompletedAt: r.CompletedAt,
			Metadata: domain.Metadata{
				domain.MetadataKeyContextTitle:    req.EventTitle,
				domain.MetadataKeyContextDate:     req.EventDate.UTC().Format("2006-01-02"),
				domain.MetadataKeyContextDuration: req.EventDuration,
				domain.MetadataKeyIssuerName:      issuerName, // resolved name
				domain.MetadataKeyRecipientName:   r.Name,
			},
			TemplateID:       req.TemplateID,
			AccountID:        req.AccountID,
			PaymentReference: req.PaymentReference,
		})
	}

	bulk, err := s.IssueBulk(ctx, completions)
	if err != nil {
		return nil, err
	}

	result := &domain.BatchIssuanceResult{
		Issued: bulk.Succeeded - bulk.Reused,
		Reused: bulk.Reused,
		Failed: bulk.Failed,
	}
	for _, item := range bulk.Items {
		if item.Err != nil {
			result.Errors = append(result.Errors, item.Err.Error())
		}
	}

	// Enqueue an email per newly-issued certificate. Already-issued
	// certificates are skipped so retries don't spam recipients.
	s.enqueueCertificateEmails(ctx, bulk)

	return result, nil
}

// enqueueCertificateEmails enqueues a "certificate ready" email for
// every certificate in the bulk result that was newly issued (not
// reused) and has a valid recipient email.
//
// Failures are logged but never returned — email delivery must not
// fail the batch.
func (s *serviceImpl) enqueueCertificateEmails(
	ctx context.Context,
	bulk *BulkResult,
) {
	if s.deps.Notifier == nil {
		return
	}
	for _, item := range bulk.Items {
		if item.Certificate == nil || item.AlreadyHad || item.Err != nil {
			continue
		}
		s.sendCertificateEmail(ctx, item.Certificate)
	}
}

// sendCertificateEmail generates a signed download token and enqueues
// the email. Errors are logged, never returned.
func (s *serviceImpl) sendCertificateEmail(
	ctx context.Context,
	cert *domain.Certificate,
) {
	token, err := s.GenerateDownloadToken(ctx, cert)
	if err != nil {
		log.Printf("[CertificateService] generate download token for %s failed: %v", cert.ID, err)
		return
	}

	downloadURL := s.buildDownloadURL(token)

	if err := s.deps.Notifier.SendCertificate(ctx, domain.SendCertificateInput{
		Certificate: cert,
		DownloadURL: downloadURL,
	}); err != nil {
		log.Printf("[CertificateService] send certificate email for %s failed: %v", cert.ID, err)
	}
}

// buildDownloadURL composes the public signed-download URL embedded in
// the email.
func (s *serviceImpl) buildDownloadURL(token string) string {
	base := trimTrailingSlash(s.deps.DownloadBaseURL)
	return base + "/api/v1/certificates/download/" + token
}