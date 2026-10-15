// internal/modules/certificate/service/issue.go

package service

import (
	"context"
	"errors"
	"strings"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// Issue implements Service.Issue.
func (s *serviceImpl) Issue(
	ctx context.Context,
	completion domain.ExternalCompletion,
) (*domain.Certificate, error) {
	if err := completion.Validate(); err != nil {
		return nil, err
	}

	// Idempotency: if a certificate already exists for this completion,
	// return it untouched.
	if existing, err := s.deps.Certificates.FindByIdentity(ctx, completion.Identity()); err == nil {
		return existing, nil
	} else if !errors.Is(err, domain.ErrCertificateNotFound) {
		return nil, err
	}

	tmpl, err := s.resolveTemplate(ctx, completion)
	if err != nil {
		return nil, err
	}

	code, err := s.uniqueVerificationCode(ctx)
	if err != nil {
		return nil, err
	}

	now := s.deps.Clock.Now().UTC()
	cert := &domain.Certificate{
		ID:               s.deps.IDs.NewID(),
		AccountID:        completion.AccountID,
		TemplateID:       tmpl.ID,
		VerificationCode: code,
		RecipientUserID:  completion.Recipient.UserID,
		RecipientName:    completion.Recipient.Name,
		RecipientEmail:   strings.ToLower(strings.TrimSpace(completion.Recipient.Email)),
		ContextType:      completion.ContextType,
		ContextID:        completion.ContextID,
		Metadata:         completion.Metadata.Clone(),
		CompletedAt:      completion.CompletedAt.UTC(),
		IssuedAt:         now,
		PaymentReference: completion.PaymentReference,
		Status:           domain.StatusActive,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.deps.Certificates.Create(ctx, cert); err != nil {
		return nil, err
	}
	return cert, nil
}

// resolveTemplate picks the template for the completion, in order:
// explicit TemplateID → account default → system default.
func (s *serviceImpl) resolveTemplate(
	ctx context.Context,
	c domain.ExternalCompletion,
) (*domain.Template, error) {
	if c.TemplateID != "" {
		t, err := s.deps.Templates.FindByID(ctx, c.TemplateID)
		if err != nil {
			return nil, err
		}
		if t.AccountID != c.AccountID && !t.IsSystem {
			return nil, domain.ErrTemplateNotFound
		}
		return t, nil
	}

	if t, err := s.deps.Templates.FindDefaultForAccount(ctx, c.AccountID); err == nil {
		return t, nil
	} else if !errors.Is(err, domain.ErrTemplateNotFound) {
		return nil, err
	}

	return s.deps.Templates.FindSystemBySlug(ctx, domain.SystemTemplateDefault)
}