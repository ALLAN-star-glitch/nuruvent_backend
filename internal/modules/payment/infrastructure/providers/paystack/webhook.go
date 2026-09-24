// internal/modules/payment/infrastructure/providers/paystack/webhook.go

package paystack

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/payment/paymentdomain"
)

// ParseWebhook verifies the signature and parses a Paystack webhook.
//
// Paystack signs webhooks with HMAC-SHA512:
//   - Header: x-paystack-signature
//   - Algorithm: HMAC-SHA512 of the raw body
//   - Key: the Paystack secret key (same one used for API calls)
//   - Encoding: hex
//
// We compute the same signature and compare in constant time.
func (p *Provider) ParseWebhook(
	ctx context.Context,
	payload []byte,
	headers map[string]string,
) (*paymentdomain.WebhookEventData, error) {
	// 1. Verify signature (must happen before trusting the body).
	if err := p.verifySignature(payload, headers); err != nil {
		return nil, err
	}

	// 2. Parse body.
	var body webhookPayload
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, fmt.Errorf("%w: %v", paymentdomain.ErrMalformedWebhook, err)
	}

	// 3. Only handle charge events. Paystack also sends
	// transfer.success, refund.processed, etc. — we ignore those.
	if body.Event != "charge.success" && body.Event != "charge.failed" {
		return nil, fmt.Errorf("%w: unsupported event %q",
			paymentdomain.ErrMalformedWebhook, body.Event)
	}

	// 4. Map status.
	status := mapPaystackStatus(body.Data.Status)

	return &paymentdomain.WebhookEventData{
		ProviderEventID:   body.Data.Reference,
		ProviderReference: body.Data.Reference,
		OurPaymentID:      body.Data.Reference,
		Status:            status,
		Amount:            body.Data.Amount,
		Currency:          body.Data.Currency,
		Message:           body.Data.GatewayResponse,
		OccurredAt:        parseTime(body.Data.PaidAt),
	}, nil
}

// verifySignature compares the x-paystack-signature header against an
// HMAC-SHA512 of the raw body, keyed with the secret key.
//
// Uses hmac.Equal for constant-time comparison to prevent timing attacks.
func (p *Provider) verifySignature(payload []byte, headers map[string]string) error {
	expected := p.cfg.SecretKey
	if expected == "" {
		return fmt.Errorf("%w: no secret key configured",
			paymentdomain.ErrInvalidWebhookSignature)
	}

	// Paystack sends the header lowercased, but some HTTP clients
	// normalize to Title-Case. Check both.
	got := headers["x-paystack-signature"]
	if got == "" {
		got = headers["X-Paystack-Signature"]
	}
	if got == "" {
		return fmt.Errorf("%w: missing x-paystack-signature header",
			paymentdomain.ErrInvalidWebhookSignature)
	}

	hash := hmac.New(sha512.New, []byte(expected))
	hash.Write(payload)
	computed := hex.EncodeToString(hash.Sum(nil))

	if !hmac.Equal([]byte(computed), []byte(got)) {
		return paymentdomain.ErrInvalidWebhookSignature
	}
	return nil
}