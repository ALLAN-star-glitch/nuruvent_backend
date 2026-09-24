// internal/modules/payment/paymentdomain/provider.go

package paymentdomain

import (
	"context"
	"time"
)

// ============================================================
// PAYMENT PROVIDER INTERFACE
// ============================================================

// PaymentProvider is the outbound port implemented by every payment
// gateway (M-Pesa, card processor, stub).
//
// Implementations must:
//   - Be safe for concurrent use
//   - Return ErrProviderTimeout on network timeouts
//   - Return ErrProviderRejected on 4xx from the provider
//   - Return ErrProviderUnavailable on 5xx from the provider
//   - Never panic on malformed input
//   - Never log sensitive data (card numbers, PINs, credentials)
//
// Contract tests in the infrastructure layer verify every
// implementation against a shared suite. Adding a new provider means
// implementing this interface and passing those tests — no changes to
// the service or handler layers.
type PaymentProvider interface {
	// Name is a stable identifier for the provider ("mpesa",
	// "stripe", "stub"). Used for routing and audit records.
	Name() string

	// Method is the payment method this provider serves. The registry
	// uses this to dispatch by method when a specific provider isn't
	// named.
	Method() PaymentMethod

	// Initiate begins a payment. Returns a reference and instructions
	// for the caller to complete the flow (a redirect URL for card,
	// an STK push prompt for M-Pesa).
	//
	// On success, the returned ProviderReference may already be set
	// (some providers return an ID immediately) or empty (providers
	// that only assign a reference when the callback arrives).
	Initiate(ctx context.Context, req InitiateRequest) (*InitiateResult, error)

	// Verify fetches the current state of a transaction at the
	// provider. Used for reconciliation and for polling when webhooks
	// are delayed or lost.
	Verify(ctx context.Context, reference string) (*ProviderStatus, error)

	// Refund reverses a payment (full or partial). The provider
	// determines whether partial refunds are supported; if not, it
	// returns ErrProviderRejected with an explanatory message.
	Refund(ctx context.Context, req RefundRequest) (*RefundResult, error)

	// ParseWebhook normalizes a provider callback into a
	// WebhookEventData. Implementations must verify signatures and
	// return ErrInvalidWebhookSignature on failure. Malformed
	// payloads return ErrMalformedWebhook.
	//
	// ParseWebhook MUST NOT perform any state changes or call back
	// into the service. It is a pure translation function.
	ParseWebhook(
		ctx context.Context,
		payload []byte,
		headers map[string]string,
	) (*WebhookEventData, error)
}

// ============================================================
// INITIATE
// ============================================================

// InitiateRequest is the input to Initiate.
// InitiateRequest is the input to Initiate.
type InitiateRequest struct {
	// PaymentID is our internal ID. Some providers echo it back in
	// webhooks, which helps reconciliation.
	PaymentID string

	// OrderID is the parent order's ID. Included for provider
	// reference fields that support merchant-supplied metadata.
	OrderID string

	// Amount is in minor units. The provider converts as needed.
	Amount int64

	// Currency is an ISO 4217 code. Providers that only support one
	// currency ignore it; multi-currency providers require it.
	Currency string

	// Method is the payment method the user chose. Required when a
	// provider serves more than one method (e.g. Flutterwave serves
	// both M-Pesa and card).
	Method PaymentMethod

	// PayerPhone is required for M-Pesa. Format should be normalized
	// before calling (e.g. 254712345678, no +).
	PayerPhone string

	// PayerEmail is required for card processors.
	PayerEmail string

	// Card carries card details when Method is PaymentMethodCard.
	// Nil for M-Pesa and other non-card methods.
	//
	// SECURITY: the values in this struct must never be logged,
	// persisted, or returned to clients. They exist only to be
	// encrypted and forwarded to the provider.
	Card *CardDetails

	// IdempotencyKey is client-supplied. Providers that support
	// idempotency (Stripe, Flutterwave v4) use it directly. Providers
	// that don't (M-Pesa) ignore it — our own deduplication handles
	// retries.
	IdempotencyKey string

	// Description is a short string shown to the user in the
	// provider's checkout page.
	Description string

	// ReturnURL is where the provider redirects after payment.
	// Set for card providers; ignored by M-Pesa.
	ReturnURL string
}


// CardDetails carries raw card information for card charges.
//
// SECURITY: this struct holds PCI-sensitive data. Handle it carefully:
//   - Never log these values, even at debug level
//   - Never persist them (no DB column, no file, no cache)
//   - Never return them in an API response
//   - Zero them out as soon as the request completes
//
// The provider is responsible for encrypting these values before
// sending them to the payment gateway.
type CardDetails struct {
	// Number is the card PAN (Primary Account Number), digits only,
	// no spaces or dashes.
	Number string

	// CVV is the 3- or 4-digit security code.
	CVV string

	// Expiry is the card expiry in "MM/YY" format.
	Expiry string
}

// InitiateResult describes how the caller should complete the payment.
type InitiateResult struct {
	// ProviderReference is the provider's transaction ID. May be
	// empty for providers that only assign it on callback.
	ProviderReference string

	// Status reflects what the provider says happened at initiation.
	// Usually PaymentStatusPending — the user must still complete
	// the flow. Some providers return Succeeded immediately for
	// pre-authorized charges.
	Status PaymentStatus

	// RedirectURL is set for providers that redirect the user to a
	// hosted checkout page (card). Empty for M-Pesa.
	RedirectURL string

	// CustomerMessage is a short human-readable instruction, e.g.
	// "Enter your M-Pesa PIN on your phone to complete payment".
	CustomerMessage string

	// ExpiresAt is when the provider's checkout session expires.
	// Zero if the provider doesn't expose this.
	ExpiresAt time.Time

	// Raw is the provider's raw response for logging and debugging.
	// Never returned to end users.
	Raw map[string]any


	AccessCode        string
}

// ============================================================
// VERIFY
// ============================================================

// ProviderStatus is the provider's current view of a transaction.
type ProviderStatus struct {
	// Reference is the provider transaction ID that was queried.
	Reference string

	// Status is our normalized interpretation of the provider's
	// state.
	Status PaymentStatus

	// Amount and Currency come from the provider for cross-checking
	// against our local record.
	Amount   int64
	Currency string

	// Message is a human-readable description of the current state,
	// useful for support ("payment confirmed", "insufficient funds").
	Message string

	// UpdatedAt is when the provider last changed this transaction.
	UpdatedAt time.Time
}

// ============================================================
// REFUND
// ============================================================

// RefundRequest is the input to Refund.
type RefundRequest struct {
	// PaymentID is our internal ID for traceability.
	PaymentID string

	// ProviderReference is the provider transaction ID being
	// reversed.
	ProviderReference string

	// Amount is in minor units. For full refunds, this equals the
	// original payment amount.
	Amount int64

	// Currency must match the original payment's currency.
	Currency string

	// Reason is a short explanation for audit. Providers that require
	// a reason code map this internally.
	Reason string
}

// RefundResult describes the outcome of a refund attempt.
type RefundResult struct {
	// ProviderReference is the provider's ID for the refund
	// transaction. Distinct from the original payment's reference.
	ProviderReference string

	// Status is the refund's current state. Usually Pending (the
	// provider processes asynchronously) or Succeeded (immediate).
	Status PaymentStatus

	// Message is a human-readable description.
	Message string
}

// ============================================================
// WEBHOOK
// ============================================================

// WebhookEventData is the normalized form of a provider webhook.
//
// Providers send wildly different payloads; ParseWebhook translates
// each one into this shape, so the service layer never sees provider-
// specific JSON.
type WebhookEventData struct {
	// ProviderEventID is a stable ID from the provider. Used for
	// deduplication — the same event delivered twice must produce
	// the same ProviderEventID.
	//
	// Some providers don't offer one; in that case, implementations
	// must synthesize a stable value (e.g. hash of a subset of the
	// payload).
	ProviderEventID string

	// ProviderReference ties the event to a transaction at the
	// provider.
	ProviderReference string

	// OurPaymentID is populated when the provider echoes our internal
	// ID back (via a metadata field, merchant reference, or similar).
	// Empty otherwise.
	OurPaymentID string

	// Status is the new state this event implies.
	Status PaymentStatus

	// Amount and Currency come from the provider for cross-checking
	// against our local record. If they disagree, the service logs
	// and rejects the event.
	Amount   int64
	Currency string

	// Message is an optional human-readable string, e.g. a failure
	// reason like "insufficient funds".
	Message string

	// OccurredAt is when the event happened at the provider. May
	// differ from when we received it — networks delay delivery.
	OccurredAt time.Time
}

// ============================================================
// REGISTRY
// ============================================================

// ProviderRegistry resolves a provider by name and by method.
//
// The service layer depends on this interface, not on concrete
// providers. Adding a new provider means registering it with the
// registry — the service code never changes.
type ProviderRegistry interface {
	// ByName returns the provider registered under the given name.
	// Returns ErrProviderNotFound if no such provider exists.
	ByName(name string) (PaymentProvider, error)

	// ByMethod returns the provider that serves the given payment
	// method. If multiple providers serve the same method, the
	// registry returns the default one.
	// Returns ErrProviderNotFound if no provider serves the method.
	ByMethod(method PaymentMethod) (PaymentProvider, error)
}