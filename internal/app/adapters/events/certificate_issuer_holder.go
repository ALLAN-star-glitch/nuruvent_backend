// internal/app/adapters/events/certificate_issuer_holder.go

package events

import (
	"context"
	"fmt"
	"sync"

	certdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	certificateService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
	eventsDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// LAZY HOLDER — breaks the events ↔ certificate cycle
// ============================================================

type CertificateServiceHolder struct {
	mu  sync.RWMutex
	svc certificateService.Service
}

func NewCertificateServiceHolder() *CertificateServiceHolder {
	return &CertificateServiceHolder{}
}

func (h *CertificateServiceHolder) Set(svc certificateService.Service) {
	h.mu.Lock()
	h.svc = svc
	h.mu.Unlock()
}

func (h *CertificateServiceHolder) Resolver() certificateService.Service {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.svc
}

// ============================================================
// ADAPTER
// ============================================================

// CertificateIssuerAdapter bridges the certificate service to the
// events module's CertificateIssuer port, resolving the service
// lazily through the holder.
type CertificateIssuerAdapter struct {
	holder *CertificateServiceHolder
}

func NewCertificateIssuerAdapter(holder *CertificateServiceHolder) *CertificateIssuerAdapter {
	return &CertificateIssuerAdapter{holder: holder}
}

// InitiateBatchIssuance implements eventsDomain.CertificateIssuer.
func (a *CertificateIssuerAdapter) InitiateBatchIssuance(
	ctx context.Context,
	req eventsDomain.InitiateBatchIssuanceRequest,
) (*eventsDomain.InitiateBatchIssuanceResult, error) {
	svc := a.holder.Resolver()
	if svc == nil {
		return nil, fmt.Errorf("certificate service not yet available")
	}

	// Translate events recipients into certificate recipients.
	recipients := make([]certdomain.BatchRecipient, 0, len(req.Recipients))
	for _, r := range req.Recipients {
		recipients = append(recipients, certdomain.BatchRecipient{
			UserID:      r.UserID,
			Name:        r.Name,
			Email:       r.Email,
			CompletedAt: r.CompletedAt,
		})
	}

	result, err := svc.InitiateBatchIssuance(ctx, req.UserID, certificateService.InitiateBatchIssuanceInput{
		EventID:       req.EventID,
		AccountID:     req.AccountID,
		TemplateID:    req.TemplateID,
		EventTitle:    req.EventTitle,
		EventDate:     req.EventDate,
		EventDuration: req.EventDuration,
		IssuerName:    req.IssuerName,
		Recipients:    recipients,
	})
	if err != nil {
		return nil, err
	}

	return &eventsDomain.InitiateBatchIssuanceResult{
		Type:          result.Type,
		Issued:        result.Issued,
		Reused:        result.Reused,
		Failed:        result.Failed,
		OrderID:       result.OrderID,
		AmountMinor:   result.AmountMinor,
		Currency:      result.Currency,
		AttendeeCount: result.AttendeeCount,
	}, nil
}

// compile-time assertion
var _ eventsDomain.CertificateIssuer = (*CertificateIssuerAdapter)(nil)