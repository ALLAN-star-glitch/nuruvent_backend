// internal/app/adapters/payment/certificate_issuer_holder.go

package payment

import (
	"context"
	"fmt"
	"sync"

	certdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	certificateService "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// ============================================================
// LAZY HOLDER — breaks the payment ↔ certificate cycle
// ============================================================
//
// The full cycle is:
//
//   certificate → payment → registration → events → certificate
//
// Breaking at any single edge is valid. This holder breaks the
// payment → certificate edge: payment receives the holder eagerly
// (empty), and the certificate service is populated AFTER the Wire
// graph is constructed (see internal/app/app.go).
//
// Mirrors the pattern used by team.AccountServiceHolder and
// attendanceadapters.EventsServiceHolder.
type CertificateServiceHolder struct {
	mu  sync.RWMutex
	svc certificateService.Service
}

// NewCertificateServiceHolder constructs the empty holder. Wire
// populates it after the graph is built, via app.go.
func NewCertificateServiceHolder() *CertificateServiceHolder {
	return &CertificateServiceHolder{}
}

// Set stores the certificate service. Called once by app.go after
// InitializeApp returns.
func (h *CertificateServiceHolder) Set(svc certificateService.Service) {
	h.mu.Lock()
	h.svc = svc
	h.mu.Unlock()
}

// Resolver returns the certificate service, or nil if not yet set.
func (h *CertificateServiceHolder) Resolver() certificateService.Service {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.svc
}

// ============================================================
// ADAPTER
// ============================================================

// CertificateIssuerAdapter bridges the certificate service to the
// payment module's CertificateIssuer port, resolving the service
// lazily through the holder.
//
// Invoked by ConfirmPayment when a certificate order is paid.
type CertificateIssuerAdapter struct {
	holder *CertificateServiceHolder
}

// NewCertificateIssuerAdapter constructs a CertificateIssuerAdapter.
func NewCertificateIssuerAdapter(holder *CertificateServiceHolder) *CertificateIssuerAdapter {
	return &CertificateIssuerAdapter{holder: holder}
}

// IssueFromPaidOrder implements paymentdomain.CertificateIssuer.
func (a *CertificateIssuerAdapter) IssueFromPaidOrder(
	ctx context.Context,
	order *paymentdomain.Order,
	payment *paymentdomain.Payment,
) error {
	svc := a.holder.Resolver()
	if svc == nil {
		return fmt.Errorf("certificate service not yet available")
	}

	if order == nil || order.CertificateCtx == nil {
		return fmt.Errorf("certificate order has no context")
	}

	cc := order.CertificateCtx

	completion := certdomain.ExternalCompletion{
		ContextType: cc.ContextType,
		ContextID:   cc.ContextID,
		Recipient: certdomain.Recipient{
			UserID: cc.RecipientUserID,
			Name:   cc.RecipientName,
			Email:  cc.RecipientEmail,
		},
		CompletedAt:      cc.CompletedAt,
		Metadata:         certdomain.Metadata{},
		TemplateID:       cc.TemplateID,
		AccountID:        cc.AccountID,
		PaymentReference: payment.ID,
	}

	if _, err := svc.Issue(ctx, completion); err != nil {
		return fmt.Errorf("issue certificate for order %s: %w", order.ID, err)
	}
	return nil
}

// Add to the CertificateIssuerAdapter struct — no new field, same holder.

// IssueBatch implements paymentdomain.CertificateBatchIssuer.
//
// Invoked by ConfirmPayment when a certificate_batch order is paid.
// Forwards the frozen batch context to the certificate service.
func (a *CertificateIssuerAdapter) IssueBatch(
	ctx context.Context,
	order *paymentdomain.Order,
	payment *paymentdomain.Payment,
) error {
	svc := a.holder.Resolver()
	if svc == nil {
		return fmt.Errorf("certificate service not yet available")
	}

	if order == nil || order.BatchCtx == nil {
		return fmt.Errorf("certificate batch order has no context")
	}

	bc := order.BatchCtx

	// Convert recipients.
	recipients := make([]certdomain.BatchRecipient, 0, len(bc.Recipients))
	for _, r := range bc.Recipients {
		recipients = append(recipients, certdomain.BatchRecipient{
			UserID:      r.UserID,
			Name:        r.Name,
			Email:       r.Email,
			CompletedAt: r.CompletedAt,
		})
	}

	_, err := svc.IssueBatch(ctx, certdomain.BatchIssuanceRequest{
		EventID:          bc.EventID,
		AccountID:        bc.AccountID,
		TemplateID:       bc.TemplateID,
		EventTitle:       bc.EventTitle,
		EventDate:        bc.EventDate,
		EventDuration:    bc.EventDuration,
		IssuerName:       bc.IssuerName,
		Recipients:       recipients,
		PaymentReference: payment.ID,
	})
	if err != nil {
		return fmt.Errorf("issue batch for order %s: %w", order.ID, err)
	}
	return nil
}


// compile-time assertion
var _ paymentdomain.CertificateIssuer = (*CertificateIssuerAdapter)(nil)