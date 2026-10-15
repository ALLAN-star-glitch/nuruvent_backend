// internal/modules/payment/infrastructure/postgres/batch_context_json.go

package postgres

import (
	"encoding/json"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// batchContextJSON is the persistence shape for
// paymentdomain.BatchContext. Carries the json tags the JSONB column
// stores; the domain type stays tag-free.
type batchContextJSON struct {
	EventID           string                `json:"event_id"`
	AccountID         string                `json:"account_id"`
	TemplateID        string                `json:"template_id"`
	PricePerCertMinor int64                 `json:"price_per_cert_minor"`
	EventTitle        string                `json:"event_title"`
	EventDate         time.Time             `json:"event_date"`
	EventDuration     string                `json:"event_duration,omitempty"`
	IssuerName        string                `json:"issuer_name"`
	AttendeeCount     int                   `json:"attendee_count"`
	Recipients        []batchRecipientJSON  `json:"recipients"`
}

type batchRecipientJSON struct {
	UserID      string    `json:"user_id,omitempty"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	CompletedAt time.Time `json:"completed_at"`
}

// batchContextToJSON marshals a domain BatchContext into the
// persistence shape. Returns nil for nil input.
func batchContextToJSON(c *paymentdomain.BatchContext) (JSONB, error) {
	if c == nil {
		return nil, nil
	}

	recipients := make([]batchRecipientJSON, 0, len(c.Recipients))
	for _, r := range c.Recipients {
		recipients = append(recipients, batchRecipientJSON{
			UserID:      r.UserID,
			Name:        r.Name,
			Email:       r.Email,
			CompletedAt: r.CompletedAt,
		})
	}

	raw, err := json.Marshal(batchContextJSON{
		EventID:           c.EventID,
		AccountID:         c.AccountID,
		TemplateID:        c.TemplateID,
		PricePerCertMinor: c.PricePerCertMinor,
		EventTitle:        c.EventTitle,
		EventDate:         c.EventDate,
		EventDuration:     c.EventDuration,
		IssuerName:        c.IssuerName,
		AttendeeCount:     c.AttendeeCount,
		Recipients:        recipients,
	})
	if err != nil {
		return nil, err
	}
	return JSONB(raw), nil
}

// batchContextFromJSON unmarshals the persistence shape into a domain
// BatchContext. Returns nil for empty input.
func batchContextFromJSON(raw JSONB) (*paymentdomain.BatchContext, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var p batchContextJSON
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}

	recipients := make([]paymentdomain.BatchRecipient, 0, len(p.Recipients))
	for _, r := range p.Recipients {
		recipients = append(recipients, paymentdomain.BatchRecipient{
			UserID:      r.UserID,
			Name:        r.Name,
			Email:       r.Email,
			CompletedAt: r.CompletedAt,
		})
	}

	return &paymentdomain.BatchContext{
		EventID:           p.EventID,
		AccountID:         p.AccountID,
		TemplateID:        p.TemplateID,
		PricePerCertMinor: p.PricePerCertMinor,
		EventTitle:        p.EventTitle,
		EventDate:         p.EventDate,
		EventDuration:     p.EventDuration,
		IssuerName:        p.IssuerName,
		AttendeeCount:     p.AttendeeCount,
		Recipients:        recipients,
	}, nil
}