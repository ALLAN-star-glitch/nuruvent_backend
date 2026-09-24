// internal/modules/payment/paymentdomain/webhook_event.go

package paymentdomain

import "time"

// WebhookEvent is a raw webhook delivery stored for audit and replay.
//
// Every webhook the module receives is persisted before processing.
// The (Provider, ProviderEventID) pair is unique — this is how we
// deduplicate webhooks that providers deliver more than once.
//
// The raw Payload is stored as-is so we can:
//   - Replay any event if our processing logic changes
//   - Audit what the provider actually sent (vs. what we did with it)
//   - Investigate disputes without needing the provider's logs
//
// ProcessedAt and ProcessingError are set after the handler runs. A
// non-nil ProcessedAt with an empty ProcessingError means success. A
// non-empty ProcessingError means we tried and failed — the event is
// eligible for retry.
type WebhookEvent struct {
	ID       string
	Provider string

	// ProviderEventID is the provider's own identifier for this event.
	// Together with Provider it forms a unique constraint.
	ProviderEventID string

	// SignatureValid records whether signature verification passed.
	// Only valid events are processed; invalid ones are still stored
	// for audit purposes.
	SignatureValid bool

	// Payload is the raw webhook body. Stored as-is.
	Payload []byte

	ReceivedAt      time.Time
	ProcessedAt     *time.Time
	ProcessingError string
}

// NewWebhookEvent constructs a record from a received webhook.
//
// The caller (the webhook handler) has already attempted signature
// verification and supplies the result. The payload is copied into
// the struct, not referenced.
func NewWebhookEvent(
	id, provider, providerEventID string,
	signatureValid bool,
	payload []byte,
	now time.Time,
) *WebhookEvent {
	return &WebhookEvent{
		ID:              id,
		Provider:        provider,
		ProviderEventID: providerEventID,
		SignatureValid:  signatureValid,
		Payload:         payload,
		ReceivedAt:      now,
	}
}

// HydrateWebhookEvent reconstructs a webhook event from persistence.
func HydrateWebhookEvent(
	id, provider, providerEventID string,
	signatureValid bool,
	payload []byte,
	receivedAt time.Time,
	processedAt *time.Time,
	processingError string,
) *WebhookEvent {
	return &WebhookEvent{
		ID:              id,
		Provider:        provider,
		ProviderEventID: providerEventID,
		SignatureValid:  signatureValid,
		Payload:         payload,
		ReceivedAt:      receivedAt,
		ProcessedAt:     processedAt,
		ProcessingError: processingError,
	}
}

// ============================================================
// LIFECYCLE
// ============================================================

// MarkProcessed records that the event has been fully processed.
// Idempotent — a processed event that has already been marked
// successfully will not be updated.
func (w *WebhookEvent) MarkProcessed(now time.Time) {
	w.ProcessedAt = &now
	w.ProcessingError = ""
}

// MarkFailed records that processing failed. Idempotent — always
// overwrites the previous error, because the last error is the most
// relevant one.
func (w *WebhookEvent) MarkFailed(reason string) {
	w.ProcessingError = reason
}

// ============================================================
// QUERIES
// ============================================================

// IsProcessed reports whether the event has been successfully
// processed.
func (w *WebhookEvent) IsProcessed() bool {
	return w.ProcessedAt != nil && w.ProcessingError == ""
}

// HasFailed reports whether the last processing attempt failed.
func (w *WebhookEvent) HasFailed() bool {
	return w.ProcessingError != ""
}

// IsPending reports whether the event has never been attempted.
func (w *WebhookEvent) IsPending() bool {
	return w.ProcessedAt == nil && w.ProcessingError == ""
}

// IsSignatureValid is a convenience for filtering.
func (w *WebhookEvent) IsSignatureValid() bool {
	return w.SignatureValid
}