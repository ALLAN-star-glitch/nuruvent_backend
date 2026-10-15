// internal/modules/certificate/domain/external_completion.go

package domain

import (
	"strings"
	"time"
)

// ExternalCompletion is the input to Issuer.Issue.
//
// It carries everything the certificate module needs to produce a
// certificate — without any knowledge of what "event" or "course"
// means. ContextType and ContextID are opaque to this module.
type ExternalCompletion struct {
	// ContextType is an opaque string identifying the kind of context
	// ("event", "course", "workshop", ...). Not interpreted here.
	ContextType string

	// ContextID is the caller's own identifier for the context.
	// Opaque to the certificate module.
	ContextID string

	// Recipient is who receives the certificate.
	Recipient Recipient

	// CompletedAt is when the completion occurred. Distinct from the
	// issue time. Must be in the past or present.
	CompletedAt time.Time

	// Metadata is free-form key-value data available to the template.
	Metadata Metadata

	// TemplateID is optional. Empty means: use the account's default
	// template, or a system template if the account has none.
	TemplateID string

	// AccountID is the issuing account.
	AccountID string

	// PaymentReference is optional. Set when a paid certificate was
	// purchased. Empty for free certificates.
	PaymentReference string
}

// Validate checks the minimum invariants for issuing.
func (e ExternalCompletion) Validate() error {
	if strings.TrimSpace(e.AccountID) == "" {
		return ErrAccountIDRequired
	}
	if strings.TrimSpace(e.ContextType) == "" {
		return ErrContextTypeRequired
	}
	if strings.TrimSpace(e.ContextID) == "" {
		return ErrContextIDRequired
	}
	if err := e.Recipient.Validate(); err != nil {
		return err
	}
	if e.CompletedAt.IsZero() {
		return ErrCompletedAtRequired
	}
	if e.CompletedAt.After(time.Now().UTC().Add(time.Minute)) {
		return ErrCompletedAtInFuture
	}
	return nil
}

// Identity returns the tuple that makes issuance idempotent:
// (context_type, context_id, recipient).
func (e ExternalCompletion) Identity() CompletionIdentity {
	return CompletionIdentity{
		ContextType: e.ContextType,
		ContextID:   e.ContextID,
		RecipientID: e.recipientKey(),
	}
}

func (e ExternalCompletion) recipientKey() string {
	if e.Recipient.UserID != "" {
		return "user:" + e.Recipient.UserID
	}
	return "email:" + e.Recipient.normalizedEmail()
}

// CompletionIdentity is the uniqueness key for a certificate.
type CompletionIdentity struct {
	ContextType string
	ContextID   string
	RecipientID string
}

// Equal reports whether two identities refer to the same completion.
func (a CompletionIdentity) Equal(b CompletionIdentity) bool {
	return a.ContextType == b.ContextType &&
		a.ContextID == b.ContextID &&
		a.RecipientID == b.RecipientID
}