// internal/modules/certificate/domain/template_repository.go

package domain

import "context"

// TemplateRepository persists certificate templates.
type TemplateRepository interface {
	Create(ctx context.Context, t *Template) error
	FindByID(ctx context.Context, id string) (*Template, error)
	FindByAccount(ctx context.Context, accountID string) ([]*Template, error)

	// FindDefaultForAccount returns the account's default template, or
	// ErrTemplateNotFound if none is set.
	FindDefaultForAccount(ctx context.Context, accountID string) (*Template, error)

	// FindSystemBySlug returns a system template by slug.
	FindSystemBySlug(ctx context.Context, slug SystemTemplateSlug) (*Template, error)

	Update(ctx context.Context, t *Template) error
	SoftDelete(ctx context.Context, id string) error

	// ClearDefault unsets any existing default for the account.
	ClearDefault(ctx context.Context, accountID string) error
}