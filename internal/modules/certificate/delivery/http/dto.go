// internal/modules/certificate/delivery/http/dto.go

package http

import "time"

// ============================================================
// CERTIFICATE REQUESTS
// ============================================================

// IssueCertificateRequest is the body for POST /certificates.
//
// In practice this endpoint is called by admin tools and other
// modules via the service interface; the HTTP surface is a thin
// wrapper around the same operation.
type IssueCertificateRequest struct {
	AccountID        string            `json:"account_id" binding:"required,uuid"`
	TemplateID       string            `json:"template_id,omitempty"`
	ContextType      string            `json:"context_type" binding:"required"`
	ContextID        string            `json:"context_id" binding:"required,uuid"`
	RecipientName    string            `json:"recipient_name" binding:"required"`
	RecipientEmail   string            `json:"recipient_email" binding:"required,email"`
	RecipientUserID  string            `json:"recipient_user_id,omitempty"`
	CompletedAt      string            `json:"completed_at" binding:"required"` // RFC3339
	Metadata         map[string]string `json:"metadata,omitempty"`
	PaymentReference string            `json:"payment_reference,omitempty"`
}

// IssueBulkRequest is the body for POST /certificates/bulk.
type IssueBulkRequest struct {
	Completions []IssueCertificateRequest `json:"completions" binding:"required,min=1"`
}

// RevokeCertificateRequest is the body for POST /certificates/:id/revoke.
type RevokeCertificateRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ============================================================
// CERTIFICATE RESPONSES
// ============================================================

// CertificateResponse is the wire format for a certificate.
type CertificateResponse struct {
	ID               string            `json:"id"`
	AccountID        string            `json:"account_id"`
	TemplateID       string            `json:"template_id"`
	VerificationCode string            `json:"verification_code"`
	RecipientName    string            `json:"recipient_name"`
	RecipientEmail   string            `json:"recipient_email"`
	RecipientUserID  string            `json:"recipient_user_id,omitempty"`
	ContextType      string            `json:"context_type"`
	ContextID        string            `json:"context_id"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	CompletedAt      time.Time         `json:"completed_at"`
	IssuedAt         time.Time         `json:"issued_at"`
	Status           string            `json:"status"`
	RevokedAt        *time.Time        `json:"revoked_at,omitempty"`
	RevocationReason string            `json:"revocation_reason,omitempty"`
	PDFURL           string            `json:"pdf_url,omitempty"`
	VerifyURL        string            `json:"verify_url,omitempty"`
}

// IssueBulkResponse summarizes a bulk issuance.
type IssueBulkResponse struct {
	Succeeded int                        `json:"succeeded"`
	Failed    int                        `json:"failed"`
	Reused    int                        `json:"reused"`
	Items     []IssueBulkItemResponse    `json:"items"`
}

// IssueBulkItemResponse is one result inside a bulk issuance.
type IssueBulkItemResponse struct {
	CertificateID string `json:"certificate_id,omitempty"`
	VerificationCode string `json:"verification_code,omitempty"`
	RecipientEmail   string `json:"recipient_email"`
	AlreadyHad       bool   `json:"already_had,omitempty"`
	Error            string `json:"error,omitempty"`
}

// VerifyResponse is the public verification response. Deliberately
// smaller than CertificateResponse — no internal IDs.
type VerifyResponse struct {
	VerificationCode string            `json:"verification_code"`
	RecipientName    string            `json:"recipient_name"`
	ContextType      string            `json:"context_type"`
	ContextID        string            `json:"context_id"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	CompletedAt      time.Time         `json:"completed_at"`
	IssuedAt         time.Time         `json:"issued_at"`
	Status           string            `json:"status"`
}

// ============================================================
// TEMPLATE REQUESTS
// ============================================================

// CreateTemplateRequest is the body for POST /certificate-templates.
type CreateTemplateRequest struct {
	AccountID   string             `json:"account_id" binding:"required,uuid"`
	Name        string             `json:"name" binding:"required,max=100"`
	Description string             `json:"description,omitempty"`
	Design      TemplateDesignDTO  `json:"design"`
	PriceMinor  *int64             `json:"price_minor,omitempty"` // ← NEW
}

type UpdateTemplateRequest struct {
	Name        *string            `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	Design      *TemplateDesignDTO `json:"design,omitempty"`
	PriceMinor  *int64             `json:"price_minor,omitempty"` // ← NEW
}

// TemplateDesignDTO mirrors domain.TemplateDesign on the wire.
type TemplateDesignDTO struct {
	BackgroundColor string `json:"background_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
	AccentColor     string `json:"accent_color,omitempty"`
	FontFamily      string `json:"font_family,omitempty"`

	LogoURL       string `json:"logo_url,omitempty"`
	SealURL       string `json:"seal_url,omitempty"`
	SignatureURL  string `json:"signature_url,omitempty"`
	BackgroundURL string `json:"background_url,omitempty"`

	SignatoryName  string `json:"signatory_name,omitempty"`
	SignatoryTitle string `json:"signatory_title,omitempty"`

	TitleText    string `json:"title_text,omitempty"`
	SubtitleText string `json:"subtitle_text,omitempty"`
	BodyText     string `json:"body_text,omitempty"`
	FooterText   string `json:"footer_text,omitempty"`

	QRPosition string `json:"qr_position,omitempty"`

	BorderStyle   string `json:"border_style,omitempty"`
	BorderColor   string `json:"border_color,omitempty"`
	SignatureLine bool   `json:"signature_line,omitempty"`
	WatermarkText string `json:"watermark_text,omitempty"`
	VerifyBadge   bool   `json:"verify_badge,omitempty"`
}

// ============================================================
// TEMPLATE RESPONSES
// ============================================================

// TemplateResponse is the wire format for a template.
type TemplateResponse struct {
	ID          string            `json:"id"`
	AccountID   string            `json:"account_id,omitempty"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Design      TemplateDesignDTO `json:"design"`
	PriceMinor  *int64            `json:"price_minor,omitempty"` // ← NEW
	SystemSlug  string            `json:"system_slug,omitempty"`
	IsDefault   bool              `json:"is_default"`
	IsSystem    bool              `json:"is_system"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ============================================================
// BATCH ISSUANCE (organizer-pays)
// ============================================================

// InitiateBatchIssuanceRequest is the body for
// POST /events/:id/certificates/issue.
type InitiateBatchIssuanceRequest struct {
	AccountID     string             `json:"account_id" binding:"required,uuid"`
	TemplateID    string             `json:"template_id,omitempty"`
	EventTitle    string             `json:"event_title,omitempty"`
	EventDate     string             `json:"event_date" binding:"required"` // RFC3339
	EventDuration string             `json:"event_duration,omitempty"`
	IssuerName    string             `json:"issuer_name,omitempty"`
	Recipients    []BatchRecipientDTO `json:"recipients" binding:"required,min=1,dive"`
}

// BatchRecipientDTO is one recipient in the batch request.
type BatchRecipientDTO struct {
	UserID      string `json:"user_id,omitempty"`
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	CompletedAt string `json:"completed_at" binding:"required"` // RFC3339
}

// InitiateBatchIssuanceResponse is the response for
// POST /events/:id/certificates/issue.
//
// Exactly one set of fields is populated:
//   - type = "free" → issued, reused, failed, attendee_count
//   - type = "paid" → order_id, amount_minor, currency, attendee_count
type InitiateBatchIssuanceResponse struct {
	Type string `json:"type"` // "free" | "paid"

	// Free path
	Issued int `json:"issued,omitempty"`
	Reused int `json:"reused,omitempty"`
	Failed int `json:"failed,omitempty"`

	// Paid path
	OrderID     string `json:"order_id,omitempty"`
	AmountMinor int64  `json:"amount_minor,omitempty"`
	Currency    string `json:"currency,omitempty"`

	// Shared
	AttendeeCount int `json:"attendee_count"`
}