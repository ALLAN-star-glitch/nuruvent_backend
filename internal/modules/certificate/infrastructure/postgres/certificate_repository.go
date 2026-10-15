// internal/modules/certificate/infrastructure/postgres/certificate_repository.go

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	"gorm.io/gorm"
)

// CertificateRepository implements domain.CertificateRepository against
// Postgres.
type CertificateRepository struct {
	db *gorm.DB
}

// NewCertificateRepository constructs a CertificateRepository.
func NewCertificateRepository(db *gorm.DB) *CertificateRepository {
	return &CertificateRepository{db: db}
}

// ============================================================
// WRITE
// ============================================================

// Create inserts a new certificate. Uniqueness on the completion
// identity is enforced by the DB partial unique indexes; a violation
// maps to domain.ErrAlreadyIssued.
func (r *CertificateRepository) Create(ctx context.Context, c *domain.Certificate) error {
	model, err := toCertificateModel(c)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return translateError(err, "create certificate")
	}
	return nil
}

// Update persists scalar changes on an existing certificate.
func (r *CertificateRepository) Update(ctx context.Context, c *domain.Certificate) error {
	res := r.db.WithContext(ctx).
		Model(&CertificateModel{}).
		Where("id = ?", c.ID).
		Updates(map[string]any{
			"status":            string(c.Status),
			"revoked_at":        c.RevokedAt,
			"revocation_reason": nullableString(c.RevocationReason),
			"pdf_url":           nullableString(c.PDFURL),
			"pdf_generated_at":  c.PDFGeneratedAt,
			"updated_at":        time.Now().UTC(),
		})
	if res.Error != nil {
		return translateError(res.Error, "update certificate")
	}
	if res.RowsAffected == 0 {
		return domain.ErrCertificateNotFound
	}
	return nil
}

// RecordDelivery inserts a delivery attempt row.
func (r *CertificateRepository) RecordDelivery(ctx context.Context, d *domain.Delivery) error {
	model := &DeliveryModel{
		ID:            d.ID,
		CertificateID: d.CertificateID,
		Channel:       string(d.Channel),
		Recipient:     d.Recipient,
		Status:        string(d.Status),
		Error:         nullableString(d.Error),
		Attempts:      d.Attempts,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return translateError(err, "record delivery")
	}
	return nil
}

// ============================================================
// READ
// ============================================================

func (r *CertificateRepository) FindByID(ctx context.Context, id string) (*domain.Certificate, error) {
	var model CertificateModel
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCertificateNotFound
	}
	if err != nil {
		return nil, translateError(err, "find certificate")
	}
	return toCertificateDomain(&model)
}

func (r *CertificateRepository) FindByVerificationCode(ctx context.Context, code domain.VerificationCode) (*domain.Certificate, error) {
	var model CertificateModel
	err := r.db.WithContext(ctx).
		Where("verification_code = ?", code.String()).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCertificateNotFound
	}
	if err != nil {
		return nil, translateError(err, "find certificate by code")
	}
	return toCertificateDomain(&model)
}

// FindByIdentity uses the two partial unique indexes. Registered users
// look up by recipient_user_id; guests look up by recipient_email.
func (r *CertificateRepository) FindByIdentity(ctx context.Context, id domain.CompletionIdentity) (*domain.Certificate, error) {
	q := r.db.WithContext(ctx).
		Where("context_type = ? AND context_id = ?", id.ContextType, id.ContextID)

	switch {
	case strings.HasPrefix(id.RecipientID, "user:"):
		q = q.Where("recipient_user_id = ?", strings.TrimPrefix(id.RecipientID, "user:"))
	case strings.HasPrefix(id.RecipientID, "email:"):
		q = q.Where("recipient_user_id IS NULL AND recipient_email = ?", strings.TrimPrefix(id.RecipientID, "email:"))
	default:
		return nil, domain.ErrCertificateNotFound
	}

	var model CertificateModel
	err := q.First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrCertificateNotFound
	}
	if err != nil {
		return nil, translateError(err, "find certificate by identity")
	}
	return toCertificateDomain(&model)
}

func (r *CertificateRepository) ListByAccount(ctx context.Context, accountID string, f domain.CertificateFilter) ([]*domain.Certificate, error) {
	return r.list(ctx, r.applyFilters(r.db.WithContext(ctx).
		Where("account_id = ?", accountID), f))
}

func (r *CertificateRepository) ListByRecipient(ctx context.Context, userID string, f domain.CertificateFilter) ([]*domain.Certificate, error) {
	return r.list(ctx, r.applyFilters(r.db.WithContext(ctx).
		Where("recipient_user_id = ?", userID), f))
}

// ============================================================
// INTERNAL HELPERS
// ============================================================

func (r *CertificateRepository) applyFilters(q *gorm.DB, f domain.CertificateFilter) *gorm.DB {
	if f.ContextType != "" {
		q = q.Where("context_type = ?", f.ContextType)
	}
	if f.ContextID != "" {
		q = q.Where("context_id = ?", f.ContextID)
	}
	if f.From != nil {
		q = q.Where("issued_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("issued_at <= ?", *f.To)
	}
	if f.Status != nil {
		q = q.Where("status = ?", string(*f.Status))
	}
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if f.Offset > 0 {
		q = q.Offset(f.Offset)
	}
	return q.Order("issued_at DESC")
}

func (r *CertificateRepository) list(ctx context.Context, q *gorm.DB) ([]*domain.Certificate, error) {
	var models []CertificateModel
	if err := q.Find(&models).Error; err != nil {
		return nil, translateError(err, "list certificates")
	}
	out := make([]*domain.Certificate, 0, len(models))
	for i := range models {
		c, err := toCertificateDomain(&models[i])
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// ============================================================
// MAPPERS
// ============================================================

func toCertificateModel(c *domain.Certificate) (*CertificateModel, error) {
	meta, err := json.Marshal(c.Metadata)
	if err != nil {
		return nil, err
	}
	return &CertificateModel{
		ID:               c.ID,
		AccountID:        c.AccountID,
		TemplateID:       c.TemplateID,
		VerificationCode: c.VerificationCode.String(),
		RecipientUserID:  nullableString(c.RecipientUserID),
		RecipientName:    c.RecipientName,
		RecipientEmail:   c.RecipientEmail,
		ContextType:      c.ContextType,
		ContextID:        c.ContextID,
		Metadata:         JSONB(meta),
		CompletedAt:      c.CompletedAt,
		IssuedAt:         c.IssuedAt,
		PaymentReference: nullableString(c.PaymentReference),
		Status:           string(c.Status),
		RevokedAt:        c.RevokedAt,
		RevocationReason: nullableString(c.RevocationReason),
		PDFURL:           nullableString(c.PDFURL),
		PDFGeneratedAt:   c.PDFGeneratedAt,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}, nil
}

func toCertificateDomain(m *CertificateModel) (*domain.Certificate, error) {
	meta := domain.Metadata{}
	if len(m.Metadata) > 0 {
		if err := json.Unmarshal(m.Metadata, &meta); err != nil {
			return nil, err
		}
	}
	return &domain.Certificate{
		ID:               m.ID,
		AccountID:        m.AccountID,
		TemplateID:       m.TemplateID,
		VerificationCode: domain.VerificationCode(m.VerificationCode),
		RecipientUserID:  derefString(m.RecipientUserID),
		RecipientName:    m.RecipientName,
		RecipientEmail:   m.RecipientEmail,
		ContextType:      m.ContextType,
		ContextID:        m.ContextID,
		Metadata:         meta,
		CompletedAt:      m.CompletedAt,
		IssuedAt:         m.IssuedAt,
		PaymentReference: derefString(m.PaymentReference),
		Status:           domain.Status(m.Status),
		RevokedAt:        m.RevokedAt,
		RevocationReason: derefString(m.RevocationReason),
		PDFURL:           derefString(m.PDFURL),
		PDFGeneratedAt:   m.PDFGeneratedAt,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}, nil
}

// ============================================================
// SMALL HELPERS
// ============================================================

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}