// internal/modules/team/service/notification_service.go

package service

import "context"

// ============================================================
// NOTIFICATION SERVICE INTERFACE (Outbound Port)
// ============================================================

// NotificationService defines the notification operations needed by the team module
type NotificationService interface {
    // SendTeamInvite sends a team invitation email to an existing user
    SendTeamInvite(ctx context.Context, req SendTeamInviteRequest) error

    // SendTeamInviteRegistration sends a team invitation email to a new user (NO OTP)
    SendTeamInviteRegistration(ctx context.Context, req SendTeamInviteRegistrationRequest) error

    // SendTeamInviteAccepted sends a notification when a user accepts an invitation
    SendTeamInviteAccepted(ctx context.Context, req SendTeamInviteAcceptedRequest) error

    // SendTeamInviteDeclined sends a notification when a user declines an invitation
    SendTeamInviteDeclined(ctx context.Context, req SendTeamInviteDeclinedRequest) error
}

// ============================================================
// NOTIFICATION REQUESTS
// ============================================================

// SendTeamInviteRequest - Team invitation email for existing users
type SendTeamInviteRequest struct {
    To         string // User's email
    UserName   string // User's name (if known)
    InvitedBy  string // Name of person who invited them
    TeamName   string // Name of the team
    TeamID     string // ID of the team
    Role       string // Role: account_admin, event_manager, team_member
    InviteLink string // Link to accept the invitation
    ExpiresIn  string // Human-readable expiry (e.g., "7 days", or "N/A" for existing users)
}

// SendTeamInviteRegistrationRequest - Invitation email for new users (NO OTP)
type SendTeamInviteRegistrationRequest struct {
    To               string // User's email
    Name             string // User's name
    InvitedBy        string // Name of person who invited them
    TeamName         string // Name of the team
    Role             string // Role: account_admin, event_manager, team_member
    RegistrationLink string // Link to registration page with token
    ExpiresIn        string // Human-readable expiry (e.g., "7 days")
}

// SendTeamInviteAcceptedRequest - Notification to admin when user accepts invitation
type SendTeamInviteAcceptedRequest struct {
    To         string // Admin's email
    AdminName  string // Admin's name
    UserName   string // Name of user who accepted
    UserEmail  string // Email of user who accepted
    TeamName   string // Name of the team
}

// SendTeamInviteDeclinedRequest - Notification to admin when user declines invitation
type SendTeamInviteDeclinedRequest struct {
    To         string // Admin's email
    AdminName  string // Admin's name
    UserName   string // Name of user who declined
    UserEmail  string // Email of user who declined
    TeamName   string // Name of the team
}