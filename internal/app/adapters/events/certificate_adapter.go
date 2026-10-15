// internal/app/adapters/events/certificate_adapter.go

package events

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	certificateService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// CertificateIssuerAdapter bridges the certificate module's Service to
// the events module's CertificateIssuer port.
//
// Translation layer between two module vocabularies:
//   - events speaks IssueCertificatesRequest / CertificateRecipient
//   - certificate speaks ExternalCompletion / Recipient
type CertificateIssuerAdapter struct {
	svc certificateService.Service
}

// NewCertificateIssuerAdapter constructs a CertificateIssuerAdapter.
func NewCertificateIssuerAdapter(svc certificateService.Service) *CertificateIssuerAdapter {
	return &CertificateIssuerAdapter{svc: svc}
}

// IssueForEvent implements eventsDomain.CertificateIssuer.
func (a *CertificateIssuerAdapter) IssueForEvent(
	ctx context.Context,
	req eventsDomain.IssueCertificatesRequest,
) (*eventsDomain.IssueCertificatesResult, error) {
	// Build one ExternalCompletion per recipient.
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
				domain.MetadataKeyIssuerName:      req.IssuerName,
				domain.MetadataKeyRecipientName:   r.Name,
			},
			TemplateID: req.TemplateID,
			AccountID:  req.AccountID,
		})
	}

	bulk, err := a.svc.IssueBulk(ctx, completions)
	if err != nil {
		return nil, err
	}

	// Translate the certificate module's BulkResult into events terms.
	result := &eventsDomain.IssueCertificatesResult{
		Issued: bulk.Succeeded - bulk.Reused,
		Reused: bulk.Reused,
		Failed: bulk.Failed,
	}
	for _, item := range bulk.Items {
		if item.Err != nil {
			result.Errors = append(result.Errors, item.Err.Error())
		}
	}
	return result, nil
}