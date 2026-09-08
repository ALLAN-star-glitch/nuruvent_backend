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
	// Determine role and team type using account-level roles only
	role, teamTypeSlug, teamID, accountID := s.determineRoleAndTeamType(ctx, user)

	// Get account type for account type slug
	accountType, err := s.repo.GetAccountTypeByID(ctx, user.AccountTypeID)
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
		AccountID:       accountID,
		AccountTypeSlug: accountType.Slug,
		TeamID:          teamID,
		TeamTypeSlug:    teamTypeSlug,
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
		"",
		"",
		time.Now().Add(s.config.JWT.RefreshExpiration),
	)
	if err != nil {
		return "", "", err
	}

	if err := s.repo.CreateRefreshToken(ctx, newToken); err != nil {
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

	if err := s.repo.UpdateRefreshTokenContext(ctx, refreshToken, userAgent, ipAddress); err != nil {
		fmt.Printf("Failed to update refresh token context: %v\n", err)
	}

	return accessToken, refreshToken, nil
}

// RefreshTokens refreshes an expired access token using a refresh token
func (s *service) RefreshTokens(ctx context.Context, refreshToken, userAgent, ip string) (string, string, error) {
	token, err := s.repo.GetRefreshTokenByToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}
	if token == nil {
		return "", "", authdomain.ErrInvalidToken
	}

	if token.IsExpired() || token.IsRevoked() {
		return "", "", authdomain.ErrInvalidToken
	}

	if err := s.repo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		return "", "", err
	}

	user, err := s.repo.GetUserByID(ctx, token.UserID)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", authdomain.ErrUserNotFound
	}

	accessToken, newRefreshToken, err := s.GenerateTokensWithContext(ctx, user, userAgent, ip)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// RevokeToken revokes a refresh token
func (s *service) RevokeToken(ctx context.Context, refreshToken string) error {
	return s.repo.RevokeRefreshToken(ctx, refreshToken)
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (s *service) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return s.repo.RevokeAllRefreshTokensForUser(ctx, userID)
}

// determineRoleAndTeamType determines the user's role, team type, and IDs using account-level roles only
// Returns: role, teamTypeSlug, teamID (UUID from teams table), accountID
func (s *service) determineRoleAndTeamType(ctx context.Context, user *authdomain.User) (string, string, string, string) {
    // ============================================================
    // 1. CHECK PLATFORM ROLES (HIGHEST PRIORITY)
    // ============================================================

    isSuperAdmin, err := s.repo.IsSuperAdmin(ctx, user.ID)
    if err == nil && isSuperAdmin {
        return authdomain.RoleSuperAdmin.String(), "", "", ""
    }

    isPlatformAdmin, err := s.repo.IsPlatformAdmin(ctx, user.ID)
    if err == nil && isPlatformAdmin {
        return authdomain.RoleAdmin.String(), "", "", ""
    }

    // ============================================================
    // 2. CHECK ACCOUNT-LEVEL ROLE (PRIMARY SOURCE OF TRUTH)
    // ============================================================

    accountRole, err := s.getAccountRole(ctx, user.ID)
    if err != nil || accountRole == "" {
        return authdomain.RoleGuest.String(), "", "", ""
    }

    // ============================================================
    // 3. FIND USER'S TEAM
    // ============================================================

    // Try personal team first (priority)
    personalTeam, err := s.teamSvc.GetPersonalTeamByUserID(ctx, user.ID)
    if err == nil && personalTeam != nil {
        return accountRole, "personal", personalTeam.ID, personalTeam.AccountID
    }

    // If no personal team, try institution teams
    institutionTeams, err := s.teamSvc.GetUserInstitutionTeamIDs(ctx, user.ID)
    if err == nil && len(institutionTeams) > 0 {
        team, err := s.teamSvc.GetTeamByID(ctx, institutionTeams[0])
        if err == nil && team != nil {
            return accountRole, "institution", team.ID, team.AccountID
        }
    }

    // ============================================================
    // 4. NO TEAM FOUND - RETURN ROLE WITHOUT TEAM CONTEXT
    // ============================================================

    return accountRole, "", "", ""
}

// getAccountRole gets the user's account-level role from auth repository
// Returns: role, error
func (s *service) getAccountRole(ctx context.Context, userID string) (string, error) {
    members, err := s.repo.GetAccountMembersByUser(ctx, userID)
    if err != nil || len(members) == 0 {
        return "", nil
    }

    member := members[0]
    return member.Role, nil
}

// GetTokenContext implements Service.
func (s *service) GetTokenContext(ctx context.Context, user *authdomain.User) (*authdomain.TokenContext, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}

	role, teamTypeSlug, teamID, accountID := s.determineRoleAndTeamType(ctx, user)

	accountType, err := s.repo.GetAccountTypeByID(ctx, user.AccountTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account type: %w", err)
	}
	if accountType == nil {
		return nil, fmt.Errorf("account type not found")
	}

	return &authdomain.TokenContext{
		UserID:          user.ID,
		Email:           user.Email,
		DisplayName:     user.DisplayName,
		Role:            role,
		AccountID:       accountID,
		AccountTypeSlug: accountType.Slug,
		TeamID:          teamID,
		TeamTypeSlug:    teamTypeSlug,
		IsVerified:      user.EmailVerified,
		IsActive:        user.IsActive,
	}, nil
}

// getAccountRoleForUserInAccount gets a user's role in a specific account
func (s *service) getAccountRoleForUserInAccount(ctx context.Context, userID, accountID string) (string, error) {
	members, err := s.repo.GetAccountMembersByUser(ctx, userID)
	if err != nil {
		return "", err
	}

	for _, member := range members {
		if member.AccountID == accountID {
			return member.Role, nil
		}
	}

	return "", nil
}

