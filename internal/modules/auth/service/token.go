// internal/modules/auth/service/token.go

package service

import (
	"context"
	"fmt"
	"time"

	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// GenerateTokens generates both access and refresh tokens for a user
func (s *service) GenerateTokens(
	ctx context.Context,
	user *authdomain.User,
) (string, string, error) {
	// Determine role and team type using Casbin
	role, teamTypeSlug, teamID := s.determineRoleAndTeamType(ctx, user)

	// Get account type for account type slug
	accountType, err := s.repo.GetAccountTypeByID(user.AccountTypeID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get account type: %w", err)
	}
	if accountType == nil {
		return "", "", fmt.Errorf("account type not found")
	}

	// Build TokenContext with all needed fields
	tokenCtx := &authdomain.TokenContext{
		UserID:          user.ID,
		Email:           user.Email,
		DisplayName:     user.DisplayName,
		Role:            role,
		AccountTypeSlug: accountType.Slug, // snake_case: "personal" or "institution"
		TeamTypeSlug:    teamTypeSlug,     // kebab-case: "personal-team" or "institution-team"
		TeamID:          teamID,           // user_id or institution_id
		InstitutionID:   teamID,           // Keep for backward compatibility (only for institution)
		IsVerified:      user.EmailVerified,
		IsActive:        user.IsActive,
	}

	// Generate access token
	accessToken, err := s.tokenSvc.GenerateAccessToken(tokenCtx)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.tokenSvc.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in database
	newToken, err := authdomain.NewRefreshToken(
		user.ID,
		refreshToken,
		"", // userAgent - can be passed from context if needed
		"", // ipAddress - can be passed from context if needed
		time.Now().Add(s.config.JWT.RefreshExpiration),
	)
	if err != nil {
		return "", "", err
	}

	if err := s.repo.CreateRefreshToken(newToken); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// GenerateTokensWithContext generates tokens with user agent and IP for tracking
func (s *service) GenerateTokensWithContext(
	ctx context.Context,
	user *authdomain.User,
	userAgent, ipAddress string,
) (string, string, error) {
	accessToken, refreshToken, err := s.GenerateTokens(ctx, user)
	if err != nil {
		return "", "", err
	}

	// Update refresh token with user agent and IP
	if err := s.repo.UpdateRefreshTokenContext(refreshToken, userAgent, ipAddress); err != nil {
		// Log but don't fail - token is still valid
		fmt.Printf("Failed to update refresh token context: %v\n", err)
	}

	return accessToken, refreshToken, nil
}

// RefreshTokens refreshes an expired access token using a refresh token
func (s *service) RefreshTokens(ctx context.Context, refreshToken, userAgent, ip string) (string, string, error) {
	// 1. Get token from database
	token, err := s.repo.GetRefreshTokenByToken(refreshToken)
	if err != nil {
		return "", "", err
	}
	if token == nil {
		return "", "", authdomain.ErrInvalidToken
	}

	// 2. Check if valid
	if token.IsExpired() || token.IsRevoked() {
		return "", "", authdomain.ErrInvalidToken
	}

	// 3. Revoke old token
	if err := s.repo.RevokeRefreshToken(refreshToken); err != nil {
		return "", "", err
	}

	// 4. Get user
	user, err := s.repo.GetUserByID(token.UserID)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", authdomain.ErrUserNotFound
	}

	// 5. Generate new tokens with context
	accessToken, newRefreshToken, err := s.GenerateTokensWithContext(ctx, user, userAgent, ip)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// RevokeToken revokes a refresh token
func (s *service) RevokeToken(ctx context.Context, refreshToken string) error {
	return s.repo.RevokeRefreshToken(refreshToken)
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (s *service) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return s.repo.RevokeAllRefreshTokensForUser(userID)
}

// determineRoleAndTeamType determines the user's role and team type using Casbin
// Returns: role, teamTypeSlug, teamID (user_id for personal, institution_id for institution)
func (s *service) determineRoleAndTeamType(ctx context.Context, user *authdomain.User) (string, string, string) {
	// ============================================================
	// 1. CHECK PLATFORM ROLES (HIGHEST PRIORITY)
	// ============================================================

	// Check if user is super admin (platform-wide)
	isSuperAdmin, err := s.repo.IsSuperAdmin(ctx, user.ID)
	if err == nil && isSuperAdmin {
		return authdomain.RoleSuperAdmin.String(), "", ""
	}

	// Check if user is platform admin
	isPlatformAdmin, err := s.repo.IsPlatformAdmin(ctx, user.ID)
	if err == nil && isPlatformAdmin {
		return authdomain.RoleAdmin.String(), "", ""
	}

	// ============================================================
	// 2. CHECK PERSONAL TEAM ROLES
	// Domain: personal:team:{user_id}
	// ============================================================

	// Check account_admin role in personal team
	if s.permService.IsPersonalTeamAdmin(ctx, user.ID, user.ID) {
		return authdomain.RoleAccountAdmin.String(), "personal-team", user.ID
	}

	// Check event_manager role in personal team
	if s.permService.IsPersonalTeamEventManager(ctx, user.ID, user.ID) {
		return authdomain.RoleEventManager.String(), "personal-team", user.ID
	}

	// Check team_member role in personal team
	if s.permService.IsPersonalTeamMember(ctx, user.ID, user.ID) {
		return authdomain.RoleTeamMember.String(), "personal-team", user.ID
	}

	// ============================================================
	// 3. CHECK INSTITUTION ROLES
	// Domain: institution:team:{institution_id}
	// ============================================================

	if user.InstitutionID != nil && *user.InstitutionID != "" {
		institutionID := *user.InstitutionID

		// Check account_admin role in institution
		if s.permService.IsInstitutionTeamAdmin(ctx, user.ID, institutionID) {
			return authdomain.RoleAccountAdmin.String(), "institution-team", institutionID
		}

		// Check event_manager role in institution
		if s.permService.IsInstitutionTeamEventManager(ctx, user.ID, institutionID) {
			return authdomain.RoleEventManager.String(), "institution-team", institutionID
		}

		// Check team_member role in institution
		if s.permService.IsInstitutionTeamMember(ctx, user.ID, institutionID) {
			return authdomain.RoleTeamMember.String(), "institution-team", institutionID
		}
	}

	// ============================================================
	// 4. CHECK IF USER HAS ANY TEAM ACCESS (via Casbin)
	// ============================================================

	if s.permService.HasTeamAccess(ctx, user.ID) {
		// Get user's teams and determine type
		personalTeams := s.permService.GetUserPersonalTeamIDs(ctx, user.ID)
		if len(personalTeams) > 0 {
			return authdomain.RoleTeamMember.String(), "personal-team", personalTeams[0]
		}

		institutionTeams := s.permService.GetUserInstitutionTeamIDs(ctx, user.ID)
		if len(institutionTeams) > 0 {
			return authdomain.RoleTeamMember.String(), "institution-team", institutionTeams[0]
		}

		return authdomain.RoleTeamMember.String(), "personal-team", user.ID
	}

	// ============================================================
	// 5. DEFAULT ROLE
	// ============================================================

	// Default role for users with no special permissions
	return authdomain.RoleGuest.String(), "", ""
}

// GetUserRolesForDomain gets all roles for a user in a specific domain
func (s *service) GetUserRolesForDomain(ctx context.Context, userID, domain string) ([]string, error) {
	return s.permService.GetUserRoles(ctx, userID, domain)
}

// GetUserTeamRoles returns all roles for a user in a specific team
func (s *service) GetUserTeamRoles(ctx context.Context, userID, teamID string) ([]string, error) {
	// Try personal team domain first
	personalDomain := authdomain.PersonalTeamDomain(teamID)
	roles, err := s.permService.GetUserRoles(ctx, userID, personalDomain)
	if err == nil && len(roles) > 0 {
		return roles, nil
	}

	// Try institution team domain
	institutionDomain := authdomain.InstitutionTeamDomain(teamID)
	return s.permService.GetUserRoles(ctx, userID, institutionDomain)
}