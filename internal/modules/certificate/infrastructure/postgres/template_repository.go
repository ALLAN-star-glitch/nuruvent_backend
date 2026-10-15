// internal/modules/certificate/infrastructure/postgres/template_repository.go

package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
	"gorm.io/gorm"
)

// TemplateRepository implements domain.TemplateRepository against Postgres.
type TemplateRepository struct {
	db *gorm.DB
}

// NewTemplateRepository constructs a TemplateRepository.
func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// ============================================================
// WRITE
// ============================================================

func (r *TemplateRepository) Create(ctx context.Context, t *domain.Template) error {
	model, err := toTemplateModel(t)
	if err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return translateError(err, "create template")
	}
	return nil
}

func (r *TemplateRepository) Update(ctx context.Context, t *domain.Template) error {
	design, err := json.Marshal(t.Design)
	if err != nil {
		return err
	}
	res := r.db.WithContext(ctx).
		Model(&TemplateModel{}).
		Where("id = ?", t.ID).
		Updates(map[string]any{
			"name":        t.Name,
			"description": t.Description,
			"design":      JSONB(design),
			"price_minor": t.PriceMinor,
			"is_default":  t.IsDefault,
			"updated_at":  t.UpdatedAt,
		})
	if res.Error != nil {
		return translateError(res.Error, "update template")
	}
	if res.RowsAffected == 0 {
		return domain.ErrTemplateNotFound
	}
	return nil
}

// SoftDelete marks the template as deleted. System templates are never
// soft-deleted (the DB filter is defensive).
func (r *TemplateRepository) SoftDelete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).
		Model(&TemplateModel{}).
		Where("id = ? AND is_system = FALSE", id).
		Update("deleted_at", gorm.Expr("NOW()"))
	if res.Error != nil {
		return translateError(res.Error, "delete template")
	}
	if res.RowsAffected == 0 {
		return domain.ErrTemplateNotFound
	}
	return nil
}

// ClearDefault unsets any existing default for the account.
func (r *TemplateRepository) ClearDefault(ctx context.Context, accountID string) error {
	return r.db.WithContext(ctx).
		Model(&TemplateModel{}).
		Where("account_id = ? AND is_default = TRUE", accountID).
		Update("is_default", false).Error
}

// ============================================================
// READ
// ============================================================

func (r *TemplateRepository) FindByID(ctx context.Context, id string) (*domain.Template, error) {
	var model TemplateModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTemplateNotFound
	}
	if err != nil {
		return nil, translateError(err, "find template")
	}
	return toTemplateDomain(&model)
}

func (r *TemplateRepository) FindByAccount(ctx context.Context, accountID string) ([]*domain.Template, error) {
	var models []TemplateModel
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND deleted_at IS NULL", accountID).
		Order("is_default DESC, created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "list templates")
	}
	out := make([]*domain.Template, 0, len(models))
	for i := range models {
		t, err := toTemplateDomain(&models[i])
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

func (r *TemplateRepository) FindDefaultForAccount(ctx context.Context, accountID string) (*domain.Template, error) {
	var model TemplateModel
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND is_default = TRUE AND deleted_at IS NULL", accountID).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTemplateNotFound
	}
	if err != nil {
		return nil, translateError(err, "find default template")
	}
	return toTemplateDomain(&model)
}

func (r *TemplateRepository) FindSystemBySlug(ctx context.Context, slug domain.SystemTemplateSlug) (*domain.Template, error) {
	var model TemplateModel
	err := r.db.WithContext(ctx).
		Where("system_slug = ? AND is_system = TRUE AND deleted_at IS NULL", string(slug)).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTemplateNotFound
	}
	if err != nil {
		return nil, translateError(err, "find system template")
	}
	return toTemplateDomain(&model)
}

// ============================================================
// MAPPERS
// ============================================================

func toTemplateModel(t *domain.Template) (*TemplateModel, error) {
    design, err := json.Marshal(t.Design)
    if err != nil {
        return nil, err
    }
    return &TemplateModel{
        ID:          t.ID,
        Slug:        slugPtr(t.SystemSlug),
        Name:        t.Name,
        DisplayName: nil, // domain doesn't expose display_name; set to NULL
        Description: nullableString(t.Description),
        PreviewURL:  nil,
        IsActive:    true,
        AccountID:   nullableString(t.AccountID),
        Design:      JSONB(design),
		PriceMinor: t.PriceMinor,
        IsDefault:   t.IsDefault,
        IsSystem:    t.IsSystem,
        CreatedAt:   t.CreatedAt,
        UpdatedAt:   t.UpdatedAt,
        DeletedAt:   t.DeletedAt,
    }, nil
}

func toTemplateDomain(m *TemplateModel) (*domain.Template, error) {
    design := domain.TemplateDesign{}
    if len(m.Design) > 0 {
        if err := json.Unmarshal(m.Design, &design); err != nil {
            return nil, err
        }
    }
    return &domain.Template{
        ID:          m.ID,
        AccountID:   derefString(m.AccountID),
        Name:        m.Name,
        Description: derefString(m.Description),
        Design:      design,
		PriceMinor: m.PriceMinor,
        SystemSlug:  systemSlugValue(m.Slug),
        IsDefault:   m.IsDefault,
        IsSystem:    m.IsSystem,
        CreatedAt:   m.CreatedAt,
        UpdatedAt:   m.UpdatedAt,
        DeletedAt:   m.DeletedAt,
    }, nil
}

func slugPtr(s domain.SystemTemplateSlug) *string {
    if s == "" {
        return nil
    }
    v := string(s)
    return &v
}

func systemSlugValue(p *string) domain.SystemTemplateSlug {
    if p == nil {
        return ""
    }
    return domain.SystemTemplateSlug(*p)
}

func systemSlugPtr(s domain.SystemTemplateSlug) *string {
	if s == "" {
		return nil
	}
	v := string(s)
	return &v
}

