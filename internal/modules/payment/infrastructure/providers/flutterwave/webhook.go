package flutterwave

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// ParseWebhook verifies the signature and parses a Flutterwave webhook.
//
// Signature verification uses the plain `verif-hash` header compared to
// our configured secret hash. Constant-time comparison prevents timing
// attacks.
//
// IMPORTANT: callers must verify the transaction server-side after
// parsing. See Provider.Verify.
func (p *Provider) ParseWebhook(
	ctx context.Context,
	payload []byte,
	headers map[string]string,
) (*paymentdomain.WebhookEventData, error) {
	// 1. Signature verification.
	if err := p.verifySignature(headers); err != nil {
		return nil, err
	}

	// 2. Parse body.
	var body webhookPayload
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, fmt.Errorf("%w: %v", paymentdomain.ErrMalformedWebhook, err)
	}

	// 3. Only handle charge.completed for MVP1.
	if body.Event != "charge.completed" {
		return nil, fmt.Errorf("%w: unsupported event %q", paymentdomain.ErrMalformedWebhook, body.Event)
	}

	// 4. Map status.
	status := mapChargeStatus(body.Data.Status)

	return &paymentdomain.WebhookEventData{
		ProviderEventID:   body.Data.FlwRef,
		ProviderReference: body.Data.TxRef,
		OurPaymentID:      body.Data.TxRef,
		Status:            status,
		Amount:            int64(body.Data.Amount * 100),
		Currency:          body.Data.Currency,
		Message:           body.Data.ProcessorResponse,
		OccurredAt:        time.Now().UTC(),
	}, nil
}

// verifySignature compares the verif-hash header against our configured
// secret hash using constant-time comparison.
func (p *Provider) verifySignature(headers map[string]string) error {
	expected := p.cfg.SecretHash
	if expected == "" {
		return fmt.Errorf("%w: no secret hash configured", paymentdomain.ErrInvalidWebhookSignature)
	}

	got := headers["verif-hash"]
	if got == "" {
		// Try case-insensitive lookup — some clients lowercase headers.
		got = headers["Verif-Hash"]
	}
	if got == "" {
		return fmt.Errorf("%w: missing verif-hash header", paymentdomain.ErrInvalidWebhookSignature)
	}

	if subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
		return paymentdomain.ErrInvalidWebhookSignature
	}
	return nil
}