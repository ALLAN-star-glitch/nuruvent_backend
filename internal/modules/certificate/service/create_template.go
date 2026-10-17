// internal/modules/certificate/service/create_template.go

package service

import (
	"context"
	"errors"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/certificate/domain"
)

// CreateTemplate implements Service.CreateTemplate.
//
// Two modes:
//
//   - System templates (in.IsSystem == true): only platform admins can
//     create them. They have no AccountID and are globally visible to
//     every account. Used by Nuruvent staff to seed pricing tiers.
//
//   - Account templates (in.IsSystem == false): created by an account
//     admin. AccountID is required; the caller must hold
//     certificate:create/update/delete in that account.
//
// In both modes the chosen tier determines which design fields are
// permitted. Fields outside the tier's allowlist are stripped before
// persistence — the UI hides them, but the enforcement happens here so
// no API call can bypass it.
func (s *serviceImpl) CreateTemplate(
	ctx context.Context,
	userID string,
	in CreateTemplateInput,
) (*domain.Template, error) {
	// ---- 1. Authorization ----
	var accountID *string

	if in.IsSystem {
		// System templates: platform-level check, no account required.
		ok, err := s.deps.Authz.CanManageSystemTemplates(ctx, userID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domain.ErrForbidden
		}
	} else {
		// Account templates: standard account-level check.
		if in.AccountID == "" {
			return nil, errors.New("account_id is required for account templates")
		}
		accountID = &in.AccountID
		
		ok, err := s.deps.Authz.CanManageTemplates(ctx, userID, in.AccountID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, domain.ErrForbidden
		}
	}

	// ---- 2. Resolve the tier ----
	//
	// For account templates, the tier is a system template the organizer
	// picked. For system templates, there is no parent tier — the admin
	// is defining the tier itself.
	var tier domain.SystemTemplateSlug
	if !in.IsSystem {
		var err error
		tier, err = s.resolveTierSlug(ctx, in.TierTemplateID)
		if err != nil {
			return nil, err
		}
	}

	// ---- 3. Enforce the tier's design restrictions ----
	design := domain.SanitizeDesign(tier, in.Design)

	// ---- 4. Build the template ----
	now := s.deps.Clock.Now().UTC()
	tmpl := &domain.Template{
		ID:          s.deps.IDs.NewID(),
		AccountID:   accountID, // <-- Uses the pointer we set above
		Name:        in.Name,
		Description: in.Description,
		Design:      design,
		PriceMinor:  in.PriceMinor,
		IsDefault:   false,
		IsSystem:    in.IsSystem,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// ---- 5. Validate ----
	if err := tmpl.Validate(); err != nil {
		return nil, err
	}

	// ---- 6. Persist ----
	if err := s.deps.Templates.Create(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

// resolveTierSlug looks up the system template by ID and returns its
// SystemSlug. Empty ID → empty slug → no restrictions.
func (s *serviceImpl) resolveTierSlug(
	ctx context.Context,
	tierTemplateID string,
) (domain.SystemTemplateSlug, error) {
	if tierTemplateID == "" {
		return "", nil
	}
	t, err := s.deps.Templates.FindByID(ctx, tierTemplateID)
	if err != nil {
		return "", err
	}
	if !t.IsSystem {
		return "", nil
	}
	return t.SystemSlug, nil
}