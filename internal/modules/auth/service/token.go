// internal/modules/auth/service/token.go

package service

import (
	"context"
	"fmt"
	"time"

	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// ============================================================
// TOKEN GENERATION
// ============================================================

// GenerateTokens generates both access and refresh tokens for a user.
func (s *service) GenerateTokens(
	ctx context.Context,
	user *authdomain.User,
) (string, string, error) {
	role, teamTypeSlug, teamID, accountID := s.determineRoleAndTeamType(ctx, user)

	accountType, err := s.repo.GetAccountTypeByID(ctx, user.AccountTypeID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get account type: %w", err)
	}
	if accountType == nil {
		return "", "", fmt.Errorf("account type not found")
	}

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

	accessToken, err := s.tokenSvc.GenerateAccessToken(tokenCtx)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.tokenSvc.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

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

// GenerateTokensWithContext generates tokens with user agent and IP for tracking.
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

// RefreshTokens refreshes an expired access token using a refresh token.
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

// RevokeToken revokes a refresh token.
func (s *service) RevokeToken(ctx context.Context, refreshToken string) error {
	return s.repo.RevokeRefreshToken(ctx, refreshToken)
}

// RevokeAllUserTokens revokes all refresh tokens for a user.
func (s *service) RevokeAllUserTokens(ctx context.Context, userID string) error {
	return s.repo.RevokeAllRefreshTokensForUser(ctx, userID)
}

// ============================================================
// TOKEN CONTEXT RESOLUTION
// ============================================================

// determineRoleAndTeamType determines the user's role, team type, team ID,
// and account ID using account-level roles only.
//
// Returns: (role, teamTypeSlug, teamID, accountID)
//
// POST-REVAMP: the account ID comes from the user's ACCOUNT MEMBERSHIP,
// which is the source of truth and is set at registration. The team ID
// comes from a team lookup and may be empty if the user has no team in
// this account (e.g. team creation failed or is deferred).
//
// The account ID is populated even when no team is found, so tokens
// always carry the caller's account context.
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
	// 2. GET ACCOUNT MEMBERSHIP (source of truth for account + role)
	// ============================================================

	members, err := s.repo.GetAccountMembersByUser(ctx, user.ID)
	if err != nil || len(members) == 0 {
		// No account membership — treat as guest with no scope.
		return authdomain.RoleGuest.String(), "", "", ""
	}

	// Use the first membership. If the app later supports "active account"
	// switching, this is where the selection logic would live.
	activeMembership := members[0]
	accountRole := activeMembership.Role
	accountID := activeMembership.AccountID

	// ============================================================
	// 3. FIND USER'S TEAM WITHIN THE ACCOUNT
	// ============================================================

	// Personal team first (priority)
	personalTeam, err := s.teamSvc.GetPersonalTeamByUserID(ctx, user.ID)
	if err == nil && personalTeam != nil {
		return accountRole, "personal", personalTeam.ID, accountID
	}

	// Fall back to any institution team in the same account
	institutionTeams, err := s.teamSvc.GetUserInstitutionTeamIDs(ctx, user.ID)
	if err == nil && len(institutionTeams) > 0 {
		team, err := s.teamSvc.GetTeamByID(ctx, institutionTeams[0])
		if err == nil && team != nil {
			return accountRole, "institution", team.ID, accountID
		}
	}

	// ============================================================
	// 4. NO TEAM FOUND — return account context without team
	// ============================================================

	return accountRole, "", "", accountID
}

// GetTokenContext builds a TokenContext for a user without generating
// tokens. Useful for the auth middleware when it needs to refresh context
// on the fly.
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
