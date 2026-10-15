// internal/modules/payment/infrastructure/postgres/certificate_context_json.go

package postgres

import (
	"encoding/json"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// certificateContextJSON is the persistence shape for
// paymentdomain.CertificateContext. It carries the json tags the
// JSONB column stores; the domain type stays tag-free.
type certificateContextJSON struct {
	ContextType     string    `json:"context_type"`
	ContextID       string    `json:"context_id"`
	TemplateID      string    `json:"template_id,omitempty"`
	RecipientName   string    `json:"recipient_name"`
	RecipientEmail  string    `json:"recipient_email"`
	RecipientUserID string    `json:"recipient_user_id,omitempty"`
	CompletedAt     time.Time `json:"completed_at"`
	AccountID       string    `json:"account_id"`
}

// certContextToJSON marshals a domain CertificateContext into the
// persistence shape. Returns nil for a nil input so the column stays
// NULL for registration orders.
func certContextToJSON(c *paymentdomain.CertificateContext) (JSONB, error) {
	if c == nil {
		return nil, nil
	}
	raw, err := json.Marshal(certificateContextJSON{
		ContextType:     c.ContextType,
		ContextID:       c.ContextID,
		TemplateID:      c.TemplateID,
		RecipientName:   c.RecipientName,
		RecipientEmail:  c.RecipientEmail,
		RecipientUserID: c.RecipientUserID,
		CompletedAt:     c.CompletedAt,
		AccountID:       c.AccountID,
	})
	if err != nil {
		return nil, err
	}
	return JSONB(raw), nil
}

// certContextFromJSON unmarshals the persistence shape into a domain
// CertificateContext. Returns nil for empty input.
func certContextFromJSON(raw JSONB) (*paymentdomain.CertificateContext, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var p certificateContextJSON
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	return &paymentdomain.CertificateContext{
		ContextType:     p.ContextType,
		ContextID:       p.ContextID,
		TemplateID:      p.TemplateID,
		RecipientName:   p.RecipientName,
		RecipientEmail:  p.RecipientEmail,
		RecipientUserID: p.RecipientUserID,
		CompletedAt:     p.CompletedAt,
		AccountID:       p.AccountID,
	}, nil
}