// internal/app/adapters/profile/permission_adapter.go

package profile

import (
	"context"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

// PermissionAdapter adapts the auth module's PermissionChecker to the profile domain's PermissionChecker
type PermissionAdapter struct {
	authPermChecker authdomain.PermissionChecker
}

// NewPermissionAdapter creates a new permission adapter
func NewPermissionAdapter(authPermChecker authdomain.PermissionChecker) domain.PermissionChecker {
	return &PermissionAdapter{
		authPermChecker: authPermChecker,
	}
}

// ============================================================
// CORE PERMISSION METHODS
// ============================================================

func (a *PermissionAdapter) HasPermission(ctx context.Context, userID string, domain string, resource, action string) (bool, error) {
	return a.authPermChecker.HasPermission(ctx, userID, domain, resource, action)
}

func (a *PermissionAdapter) HasAnyPermission(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error) {
	return a.authPermChecker.HasAnyPermission(ctx, userID, domain, resource, actions...)
}

func (a *PermissionAdapter) HasAllPermissions(ctx context.Context, userID string, domain string, resource string, actions ...string) (bool, error) {
	return a.authPermChecker.HasAllPermissions(ctx, userID, domain, resource, actions...)
}

// ============================================================
// PROFILE PERMISSIONS - READ
// ============================================================

func (a *PermissionAdapter) CanReadAllProfiles(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanReadAllProfiles(ctx, userID, domain)
}

func (a *PermissionAdapter) CanReadOwnProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanReadOwnProfile(ctx, userID, domain)
}

func (a *PermissionAdapter) CanReadProfile(ctx context.Context, userID string, domain string) (bool, error) {
	// CanReadProfile = CanReadAllProfiles OR CanReadOwnProfile
	canReadAll, err := a.authPermChecker.CanReadAllProfiles(ctx, userID, domain)
	if err != nil {
		return false, err
	}
	if canReadAll {
		return true, nil
	}
	return a.authPermChecker.CanReadOwnProfile(ctx, userID, domain)
}

// ============================================================
// PROFILE PERMISSIONS - UPDATE
// ============================================================

func (a *PermissionAdapter) CanUpdateAllProfiles(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanUpdateAllProfiles(ctx, userID, domain)
}

func (a *PermissionAdapter) CanUpdateOwnProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanUpdateOwnProfile(ctx, userID, domain)
}

func (a *PermissionAdapter) CanUpdateProfile(ctx context.Context, userID string, domain string) (bool, error) {
	// CanUpdateProfile = CanUpdateAllProfiles OR CanUpdateOwnProfile
	canUpdateAll, err := a.authPermChecker.CanUpdateAllProfiles(ctx, userID, domain)
	if err != nil {
		return false, err
	}
	if canUpdateAll {
		return true, nil
	}
	return a.authPermChecker.CanUpdateOwnProfile(ctx, userID, domain)
}

// ============================================================
// PROFILE PERMISSIONS - MANAGEMENT (Convenience)
// ============================================================

func (a *PermissionAdapter) CanManageProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanManageProfile(ctx, userID, domain)
}

func (a *PermissionAdapter) CanViewProfile(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.CanViewProfile(ctx, userID, domain)
}

// ============================================================
// ACCOUNT ROLE CHECKS
// ============================================================

func (a *PermissionAdapter) IsAccountAdmin(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.IsAccountAdmin(ctx, userID, domain)
}

func (a *PermissionAdapter) IsTeamAdmin(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.IsAccountAdmin(ctx, userID, domain)
}

func (a *PermissionAdapter) IsTrainer(ctx context.Context, userID string, domain string) (bool, error) {
	return a.authPermChecker.IsTrainer(ctx, userID, domain)
}

// ============================================================
// USER INFORMATION METHODS
// ============================================================

func (a *PermissionAdapter) GetUserTeamDomains(ctx context.Context, userID string) ([]string, error) {
	return a.authPermChecker.GetUserTeamDomains(ctx, userID)
}

func (a *PermissionAdapter) GetUserPersonalTeamIDs(ctx context.Context, userID string) ([]string, error) {
	domains, err := a.authPermChecker.GetUserTeamDomains(ctx, userID)
	if err != nil {
		return nil, err
	}
	var personalIDs []string
	for _, d := range domains {
		if domain.IsPersonalTeamDomain(d) {
			personalIDs = append(personalIDs, domain.ExtractTeamID(d))
		}
	}
	return personalIDs, nil
}

func (a *PermissionAdapter) GetUserInstitutionTeamIDs(ctx context.Context, userID string) ([]string, error) {
	domains, err := a.authPermChecker.GetUserTeamDomains(ctx, userID)
	if err != nil {
		return nil, err
	}
	var institutionIDs []string
	for _, d := range domains {
		if domain.IsInstitutionTeamDomain(d) {
			institutionIDs = append(institutionIDs, domain.ExtractTeamID(d))
		}
	}
	return institutionIDs, nil
}

func (a *PermissionAdapter) GetUserAccountIDs(ctx context.Context, userID string) ([]string, error) {
	return a.authPermChecker.GetUserAccountIDs(ctx, userID)
}

func (a *PermissionAdapter) GetUserRoles(ctx context.Context, userID string, domain string) ([]string, error) {
	return a.authPermChecker.GetUserRoles(ctx, userID, domain)
}

func (a *PermissionAdapter) HasTeamAccess(ctx context.Context, userID string) (bool, error) {
	return a.authPermChecker.HasTeamAccess(ctx, userID)
}