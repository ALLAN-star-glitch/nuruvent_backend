// internal/modules/payment/infrastructure/providers/intasend/webhook.go

package intasend

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// ParseWebhook verifies the challenge and parses an IntaSend webhook.
//
// IntaSend does not sign the payload with an HMAC. Instead, every
// webhook includes a `challenge` field, whose value you set in the
// IntaSend dashboard. We compare that value to our configured
// challenge using constant-time comparison to prevent timing attacks.
//
// IMPORTANT: callers should still verify the transaction server-side
// after parsing. See Provider.Verify.
func (p *Provider) ParseWebhook(
	ctx context.Context,
	payload []byte,
	headers map[string]string,
) (*paymentdomain.WebhookEventData, error) {
	// 1. Parse body first — we need it to read the challenge.
	var body webhookPayload
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, fmt.Errorf("%w: %v", paymentdomain.ErrMalformedWebhook, err)
	}

	// 2. Verify challenge.
	if err := p.verifyChallenge(body.Challenge); err != nil {
		return nil, err
	}

	// 3. Determine which event this is.
	//
	// IntaSend webhooks don't have a top-level "event" string like
	// Flutterwave. The event type is implied by which fields are
	// present. For our MVP we only care about collection events
	// (payments). We detect them by the presence of invoice_id.
	if body.InvoiceID == "" {
		return nil, fmt.Errorf("%w: not a collection event", paymentdomain.ErrMalformedWebhook)
	}

	// 4. Map state → domain status.
	status := mapInvoiceState(body.State)

	// 5. Build the event.
	//
	// ProviderReference is IntaSend's canonical invoice_id (the short
	// code like "RLZ7Z8Y"). This is what Verify() needs — NOT our
	// internal api_ref. Our api_ref is echoed back in the payload's
	// `api_ref` field, which we expose separately as OurPaymentID.
	return &paymentdomain.WebhookEventData{
		ProviderEventID:   body.InvoiceID,
		ProviderReference: body.InvoiceID,
		OurPaymentID:      body.APIRef,
		Status:            status,
		Amount:            parseAmount(body.Amount),
		Currency:          body.Currency,
		Message:           body.FailedReason,
		OccurredAt:        parseTime(body.UpdatedAt),
	}, nil
}

// verifyChallenge compares the challenge field from the webhook body
// against our configured challenge string using constant-time
// comparison.
func (p *Provider) verifyChallenge(got string) error {
	expected := p.cfg.Challenge
	if expected == "" {
		return fmt.Errorf("%w: no challenge configured", paymentdomain.ErrInvalidWebhookSignature)
	}
	if got == "" {
		return fmt.Errorf("%w: missing challenge in payload", paymentdomain.ErrInvalidWebhookSignature)
	}
	if subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
		return paymentdomain.ErrInvalidWebhookSignature
	}
	return nil
}

// Compile-time safety: ensure `time` import stays used even if the
// file evolves. Currently used for time.Time defaults.
var _ = time.Now