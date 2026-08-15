// internal/modules/auth/service/token.go

package service

import (
	"context"
	"fmt"
	"time"

	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// GenerateTokens generates both access and refresh tokens for a user
// This matches the Service interface signature
func (s *service) GenerateTokens(
	ctx context.Context,
	account *authdomain.Account,
) (string, string, error) {
	// Get account type
	accountType, err := s.repo.GetAccountTypeByID(account.AccountTypeID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get account type: %w", err)
	}
	if accountType == nil {
		return "", "", fmt.Errorf("account type not found")
	}

	// Determine role
	role := s.determineRole(ctx, account)

	// Build TokenContext
	tokenCtx := &authdomain.TokenContext{
		UserID:          account.ID,
		Email:           account.Email,
		Role:            role,
		AccountTypeID:   accountType.ID,
		AccountTypeSlug: accountType.Slug,
		AccountID:       account.ID,
	}

	// Add institution ID if present
	if account.InstitutionID != nil && *account.InstitutionID != "" {
		tokenCtx.InstitutionID = *account.InstitutionID
	}

	// Generate access token
	accessToken, err := s.tokenSvc.GenerateAccessToken(tokenCtx)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.tokenSvc.GenerateRefreshToken(account.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in database
	newToken, err := authdomain.NewRefreshToken(
		account.ID,
		refreshToken,
		"", // userAgent
		"", // ipAddress
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

	// 4. Get account
	account, err := s.repo.GetAccountByID(token.AccountID)
	if err != nil {
		return "", "", err
	}
	if account == nil {
		return "", "", authdomain.ErrAccountNotFound
	}

	// 5. Generate new tokens
	accessToken, newRefreshToken, err := s.GenerateTokens(ctx, account)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// RevokeToken revokes a refresh token
func (s *service) RevokeToken(ctx context.Context, refreshToken string) error {
	return s.repo.RevokeRefreshToken(refreshToken)
}

// determineRole determines the user's role using authdomain constants
// Users have ONE role (with inheritance), so we check in priority order
// Priority: super_admin > admin > account_admin > event_manager > team_member > guest
func (s *service) determineRole(ctx context.Context, account *authdomain.Account) string {
	// ============================================================
	// 1. CHECK PLATFORM ROLES (HIGHEST PRIORITY)
	// ============================================================
	
	// Check if user is super admin (platform-wide)
	isSuperAdmin, err := s.repo.IsSuperAdmin(ctx, account.ID)
	if err == nil && isSuperAdmin {
		return authdomain.RoleSuperAdmin.String()
	}

	// Check if user is platform admin
	isPlatformAdmin, err := s.repo.IsPlatformAdmin(ctx, account.ID)
	if err == nil && isPlatformAdmin {
		return authdomain.RoleAdmin.String()
	}

	// ============================================================
	// 2. CHECK ACCOUNT ROLES (via Casbin)
	// ============================================================
	
	// Check account_admin role (highest account role)
	if s.permService.IsAccountAdmin(ctx, account.ID, account.ID) {
		return authdomain.RoleAccountAdmin.String()
	}

	// Check event_manager role
	if s.permService.IsEventManager(ctx, account.ID, account.ID) {
		return authdomain.RoleEventManager.String()
	}

	// Check team_member role
	if s.permService.IsTeamMember(ctx, account.ID, account.ID) {
		return authdomain.RoleTeamMember.String()
	}

	// ============================================================
	// 3. CHECK INSTITUTION MEMBERSHIP
	// ============================================================
	
	// If user has an institution, check their team member role
	if account.InstitutionID != nil && *account.InstitutionID != "" {
		// Check if user is an admin of the institution
		isAdmin, err := s.repo.IsInstitutionAdmin(ctx, account.ID, *account.InstitutionID)
		if err == nil && isAdmin {
			return authdomain.RoleAccountAdmin.String()
		}

		// Get team member role from repository
		member, err := s.repo.GetTeamMemberByAccountAndInstitution(account.ID, *account.InstitutionID)
		if err == nil && member != nil {
			// Map team member role to authdomain role
			switch member.Role {
			case authdomain.RoleAccountAdmin.String():
				return authdomain.RoleAccountAdmin.String()
			case authdomain.RoleEventManager.String():
				return authdomain.RoleEventManager.String()
			case authdomain.RoleTeamMember.String():
				return authdomain.RoleTeamMember.String()
			default:
				return authdomain.RoleTeamMember.String()
			}
		}
		return authdomain.RoleTeamMember.String()
	}

	// ============================================================
	// 4. DEFAULT ROLE
	// ============================================================
	
	// Default role for individual accounts
	return authdomain.RoleGuest.String()
}