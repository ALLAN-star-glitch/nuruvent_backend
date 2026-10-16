// internal/modules/certificate/service/service.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// Service is the certificate module's public API.
//
// Consumers depend on this interface, not on the concrete
// implementation.
type Service interface {
	// ============================================================
	// ISSUANCE (machine-to-machine)
	// ============================================================

	// Issue issues a single certificate. Idempotent.
	Issue(ctx context.Context, completion domain.ExternalCompletion) (*domain.Certificate, error)

	// IssueBulk issues many certificates.
	IssueBulk(ctx context.Context, completions []domain.ExternalCompletion) (*BulkResult, error)

	// ============================================================
	// VERIFICATION (public)
	// ============================================================

	// GetByCode looks up a certificate by its verification code.
	// Public — no auth required at the HTTP layer.
	GetByCode(ctx context.Context, code string) (*domain.Certificate, error)

	// ============================================================
	// ACCOUNT-SCOPED CERTIFICATE OPS (user-facing)
	// ============================================================

	// ListByAccount returns all certificates issued by an account.
	ListByAccount(
		ctx context.Context,
		userID, accountID string,
		f domain.CertificateFilter,
	) ([]*domain.Certificate, error)

	// Revoke revokes a certificate. Authorization is the caller's
	// responsibility.
	Revoke(ctx context.Context, certificateID, reason string) error

	// RevokeByUser revokes a certificate on behalf of a user.
	RevokeByUser(ctx context.Context, userID, certificateID, reason string) error

	// ============================================================
	// RECIPIENT-SCOPED CERTIFICATE OPS (user-facing)
	// ============================================================

	// ListMine returns the caller's own certificates.
	ListMine(
		ctx context.Context,
		userID string,
		f domain.CertificateFilter,
	) ([]*domain.Certificate, error)

	// ============================================================
	// TEMPLATES (user-facing)
	// ============================================================

	// CreateTemplate creates a new certificate template for an account.
	// Requires certificate:create in the account.
	CreateTemplate(
		ctx context.Context,
		userID string,
		in CreateTemplateInput,
	) (*domain.Template, error)

	// ListTemplates returns all templates visible to an account: its
	// own plus system templates. Requires certificate:read.
	ListTemplates(
		ctx context.Context,
		userID, accountID string,
	) ([]*domain.Template, error)

	// UpdateTemplate updates a template's editable fields.
	// Requires certificate:update in the owning account.
	UpdateTemplate(
		ctx context.Context,
		userID, templateID string,
		in UpdateTemplateInput,
	) (*domain.Template, error)

	// DeleteTemplate soft-deletes a template. System templates cannot
	// be deleted. Requires certificate:delete in the owning account.
	DeleteTemplate(
		ctx context.Context,
		userID, templateID string,
	) error

	// SetDefaultTemplate marks a template as the account's default.
	// Requires certificate:update in the owning account.
	SetDefaultTemplate(
		ctx context.Context,
		userID, templateID string,
	) error


		// DownloadURL returns a URL the caller can use to download the
	// certificate PDF. Renders the PDF on first call; returns the
	// cached URL afterwards.
	//
	// Authorized for the issuing account's admins and for the
	// certificate's own recipient.
	DownloadURL(
		ctx context.Context,
		userID, certificateID string,
	) (string, error)


	

		// ============================================================
	// BATCH ISSUANCE (organizer-pays)
	// ============================================================

	// InitiateBatchIssuance starts a certificate batch for an event.
	// Free batches issue immediately; paid batches create a pending
	// payment order. Idempotent per event.
	InitiateBatchIssuance(
		ctx context.Context,
		userID string,
		in InitiateBatchIssuanceInput,
	) (*InitiateBatchIssuanceResult, error)

	// IssueBatch issues all certificates for a paid batch. Invoked by
	// the payment webhook adapter on successful payment.
	IssueBatch(
		ctx context.Context,
		req domain.BatchIssuanceRequest,
	) (*domain.BatchIssuanceResult, error)

		// DownloadByToken returns a URL to download the certificate
	// authorized by a signed token. No authentication is required —
	// the token itself is the credential.
	DownloadByToken(ctx context.Context, token string) (string, error)

		// SendCertificateEmail regenerates and re-sends the certificate
	// email to the recipient.
	//
	// Authorized for the issuing account's admins and for the
	// certificate's own recipient.
	SendCertificateEmail(
		ctx context.Context,
		userID, certificateID string,
	) error
}

// ============================================================
// RESULT TYPES
// ============================================================

// BulkItemResult is the per-recipient outcome of a bulk issuance.
type BulkItemResult struct {
	Completion  domain.ExternalCompletion
	Certificate *domain.Certificate
	AlreadyHad  bool
	Err         error
}

// BulkResult aggregates a bulk issuance call.
type BulkResult struct {
	Items     []BulkItemResult
	Succeeded int
	Failed    int
	Reused    int
}

// ============================================================
// INPUT TYPES
// ============================================================

// CreateTemplateInput carries the caller-supplied fields for a new
// account template.
type CreateTemplateInput struct {
	AccountID      string
	Name           string
	Description    string
	Design         domain.TemplateDesign
	PriceMinor     *int64
	TierTemplateID string // ← NEW: system template ID chosen as the tier
}

// UpdateTemplateInput carries the patch for an existing account
// template. Nil fields are left unchanged.
type UpdateTemplateInput struct {
	Name           *string
	Description    *string
	Design         *domain.TemplateDesign
	PriceMinor     *int64
	TierTemplateID string // ← NEW
}
