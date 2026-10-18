// internal/modules/certificate/service/service.go

package service

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// Service is the certificate module's public API.
type Service interface {
	// ============================================================
	// ISSUANCE (machine-to-machine)
	// ============================================================
	Issue(ctx context.Context, completion domain.ExternalCompletion) (*domain.Certificate, error)
	IssueBulk(ctx context.Context, completions []domain.ExternalCompletion) (*BulkResult, error)

	// ============================================================
	// VERIFICATION (public)
	// ============================================================
	GetByCode(ctx context.Context, code string) (*domain.Certificate, error)

	// ============================================================
	// ACCOUNT-SCOPED CERTIFICATE OPS (user-facing)
	// ============================================================
	ListByAccount(ctx context.Context, userID, accountID string, f domain.CertificateFilter) ([]*domain.Certificate, error)
	Revoke(ctx context.Context, certificateID, reason string) error
	RevokeByUser(ctx context.Context, userID, certificateID, reason string) error

	// ============================================================
	// RECIPIENT-SCOPED CERTIFICATE OPS (user-facing)
	// ============================================================
	ListMine(ctx context.Context, userID string, f domain.CertificateFilter) ([]*domain.Certificate, error)

	// ============================================================
	// TEMPLATES (user-facing)
	// ============================================================

	// CreateTemplate creates a new certificate template.
	// For system templates (in.IsSystem = true), requires platform admin.
	// For account templates (in.IsSystem = false), requires certificate:create in the account.
	CreateTemplate(ctx context.Context, userID string, in CreateTemplateInput) (*domain.Template, error)

	ListTemplates(ctx context.Context, userID, accountID string) ([]*domain.Template, error)
	UpdateTemplate(ctx context.Context, userID, templateID string, in UpdateTemplateInput) (*domain.Template, error)
	DeleteTemplate(ctx context.Context, userID, templateID string) error
	SetDefaultTemplate(ctx context.Context, userID, templateID string) error

	// ============================================================
	// DOWNLOADS
	// ============================================================
	DownloadURL(ctx context.Context, userID, certificateID string) (string, error)
	DownloadByToken(ctx context.Context, token string) (string, error)
	SendCertificateEmail(ctx context.Context, userID, certificateID string) error

	// ============================================================
	// BATCH ISSUANCE (organizer-pays)
	// ============================================================
	InitiateBatchIssuance(ctx context.Context, userID string, in InitiateBatchIssuanceInput) (*InitiateBatchIssuanceResult, error)
	IssueBatch(ctx context.Context, req domain.BatchIssuanceRequest) (*domain.BatchIssuanceResult, error)

	// ============================================================
	// MEDIA
	// ============================================================
	UploadTemplateImage(ctx context.Context, templateID, filename, contentType string, data []byte) (string, error)
	DeleteTemplateImage(ctx context.Context, url string) error

	ListSystemTemplates(ctx context.Context) ([]*domain.Template, error)

	RenderCertificatePDF(ctx context.Context, certificateID string) error
}

// ============================================================
// RESULT TYPES
// ============================================================

type BulkItemResult struct {
	Completion  domain.ExternalCompletion
	Certificate *domain.Certificate
	AlreadyHad  bool
	Err         error
}

type BulkResult struct {
	Items     []BulkItemResult
	Succeeded int
	Failed    int
	Reused    int
}

// ============================================================
// INPUT TYPES
// ============================================================

// CreateTemplateInput carries the caller-supplied fields for a new template.
type CreateTemplateInput struct {
	AccountID      string
	Name           string
	Description    string
	Design         domain.TemplateDesign
	PriceMinor     *int64
	TierTemplateID string 
	
	// IsSystem determines if this is a global system template (created by an admin).
	// If true, AccountID is ignored, and platform-admin authorization is required.
	IsSystem       bool   
}

// UpdateTemplateInput carries the patch for an existing account template.
type UpdateTemplateInput struct {
	Name           *string
	Description    *string
	Design         *domain.TemplateDesign
	PriceMinor     *int64
	TierTemplateID string
}