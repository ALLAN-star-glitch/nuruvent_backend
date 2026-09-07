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


	// 4. Get user
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

// determineRoleAndTeamType determines the user's role, team type, and IDs using Casbin
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
	// 2. CHECK ACCOUNT-LEVEL ROLE (DEFAULT)
	// ============================================================

	// Get user's account role from account_members (Auth repo)
	accountRole, accountID, err := s.getAccountRole(ctx, user.ID)
	if err == nil && accountRole != "" && accountID != "" {

		// Check if user has team-level override using TeamService port
		teamOverride, teamID, teamType := s.getTeamOverride(ctx, user.ID)

		if teamOverride != "" && teamID != "" {
			// Team override takes precedence
			return teamOverride, teamType, teamID, accountID
		}

		// If no team override, use account role with personal team
		personalTeam, err := s.teamSvc.GetPersonalTeamByUserID(ctx, user.ID)
		if err == nil && personalTeam != nil {
			return accountRole, "personal-team", personalTeam.ID, accountID
		}

		// If no personal team, check if user has any institution team
		institutionTeams, err := s.teamSvc.GetUserInstitutionTeamIDs(ctx, user.ID)
		if err == nil && len(institutionTeams) > 0 {
			team, err := s.teamSvc.GetTeamByID(ctx, institutionTeams[0])
			if err == nil && team != nil {
				return accountRole, "institution-team", team.ID, accountID
			}
		}

		return accountRole, "", "", accountID
	}

	// ============================================================
	// 3. CHECK INSTITUTION TEAM ROLES (if no account role)
	// ============================================================

	institutionTeams, err := s.teamSvc.GetUserInstitutionTeamIDs(ctx, user.ID)
	if err == nil && len(institutionTeams) > 0 {
		for _, teamID := range institutionTeams {
			team, err := s.teamSvc.GetTeamByID(ctx, teamID)
			if err != nil || team == nil {
				continue
			}

			domain := authdomain.InstitutionTeamDomain(teamID)

			// Check account_admin role in institution team
			isAdmin, _ := s.permChecker.IsAccountAdmin(ctx, user.ID, domain)
			if isAdmin {
				account, _ := s.teamSvc.GetAccountByTeamID(ctx, teamID)
				if account != nil {
					return authdomain.RoleAccountAdmin.String(), "institution-team", teamID, account.ID
				}
				return authdomain.RoleAccountAdmin.String(), "institution-team", teamID, ""
			}

			// Check trainer role in institution team
			isTrainer, _ := s.permChecker.IsTrainer(ctx, user.ID, domain)
			if isTrainer {
				account, _ := s.teamSvc.GetAccountByTeamID(ctx, teamID)
				if account != nil {
					return authdomain.RoleTrainer.String(), "institution-team", teamID, account.ID
				}
				return authdomain.RoleTrainer.String(), "institution-team", teamID, ""
			}
		}
	}

	// ============================================================
	// 4. CHECK PERSONAL TEAM (if no account role)
	// ============================================================

	personalTeam, err := s.teamSvc.GetPersonalTeamByUserID(ctx, user.ID)
	if err == nil && personalTeam != nil {
		domain := authdomain.PersonalTeamDomain(personalTeam.ID)

		isAdmin, _ := s.permChecker.IsAccountAdmin(ctx, user.ID, domain)
		if isAdmin {
			return authdomain.RoleAccountAdmin.String(), "personal-team", personalTeam.ID, user.ID
		}

		isTrainer, _ := s.permChecker.IsTrainer(ctx, user.ID, domain)
		if isTrainer {
			return authdomain.RoleTrainer.String(), "personal-team", personalTeam.ID, user.ID
		}
	}

	// ============================================================
	// 5. DEFAULT ROLE
	// ============================================================

	return authdomain.RoleGuest.String(), "", "", ""
}

// GetTokenContext implements Service.
func (s *service) GetTokenContext(ctx context.Context, user *authdomain.User) (*authdomain.TokenContext, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}

	// Determine role and team type using Casbin
	role, teamTypeSlug, teamID, accountID := s.determineRoleAndTeamType(ctx, user)

	// Get account type for account type slug
	accountType, err := s.repo.GetAccountTypeByID(ctx, user.AccountTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account type: %w", err)
	}
	if accountType == nil {
		return nil, fmt.Errorf("account type not found")
	}

	// Build TokenContext with all needed fields
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

// getAccountRole gets the user's account-level role from auth repository
// Returns: role, accountID, error
func (s *service) getAccountRole(ctx context.Context, userID string) (string, string, error) {
	// Get account memberships for user from auth repository
	members, err := s.repo.GetAccountMembersByUser(ctx, userID)
	if err != nil || len(members) == 0 {
		return "", "", nil
	}

	// Use the first account membership (primary account)
	member := members[0]
	return member.Role, member.AccountID, nil
}

// getTeamOverride checks if user has a team-level role override using TeamService port
// Returns: role, teamID, teamType
func (s *service) getTeamOverride(ctx context.Context, userID string) (string, string, string) {
	// Get user's team memberships from TeamService port
	members, err := s.teamSvc.GetUserTeamMemberships(ctx, userID)
	if err != nil || len(members) == 0 {
		return "", "", ""
	}

	// Check if any team membership role differs from account role
	// For now, return the first team membership
	// This will be enhanced with proper override logic
	for _, member := range members {
		team, err := s.teamSvc.GetTeamByID(ctx, member.TeamID)
		if err != nil || team == nil {
			continue
		}
		return member.Role, member.TeamID, team.Type
	}

	return "", "", ""
}

// GetUserRolesForDomain gets all roles for a user in a specific domain
func (s *service) GetUserRolesForDomain(ctx context.Context, userID, domain string) ([]string, error) {
	return s.permChecker.GetUserRoles(ctx, userID, domain)
}

// GetUserTeamRoles returns all roles for a user in a specific team
func (s *service) GetUserTeamRoles(ctx context.Context, userID, teamID string) ([]string, error) {
	// Try personal team domain first
	domain := authdomain.PersonalTeamDomain(teamID)
	roles, err := s.permChecker.GetUserRoles(ctx, userID, domain)
	if err == nil && len(roles) > 0 {
		return roles, nil
	}

	// Try institution team domain
	domain = authdomain.InstitutionTeamDomain(teamID)
	return s.permChecker.GetUserRoles(ctx, userID, domain)
}