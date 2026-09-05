// internal/modules/profile/domain/notification.go

package domain

import "context"

// ============================================================
// NOTIFICATION SERVICE - Outbound Port for Profile Module
// ============================================================

// NotificationService defines the notification operations that the profile module requires
type NotificationService interface {
    // SendTeamInvite sends a team invitation email to an existing user
    SendTeamInvite(ctx context.Context, req SendTeamInviteRequest) error

    // SendTeamInviteRegistration sends an OTP for team invite registration (new users)
    SendTeamInviteRegistration(ctx context.Context, req SendTeamInviteRegistrationRequest) error

    // SendTeamInviteAccepted sends notification to admin when a user accepts an invitation
    SendTeamInviteAccepted(ctx context.Context, req SendTeamInviteAcceptedRequest) error

    // SendTeamInviteDeclined sends notification to admin when a user declines an invitation
    SendTeamInviteDeclined(ctx context.Context, req SendTeamInviteDeclinedRequest) error
}

// ============================================================
// NOTIFICATION COMMANDS
// ============================================================

// SendTeamInviteRequest - Team invitation email for existing users
type SendTeamInviteRequest struct {
    To              string // User's email
    UserName        string // User's name (if known)
    InvitedBy       string // Name of person who invited them
    InstitutionName string // Name of the institution
    InstitutionID   string // ID of the institution
    Role            string // Role they're being invited to
    InviteLink      string // Link to accept the invitation
    ExpiresIn       string // Human-readable expiry (e.g., "7 days")
}

// SendTeamInviteRegistrationRequest - OTP for team invite registration (new users)
type SendTeamInviteRegistrationRequest struct {
    To              string // User's email
    Name            string // User's name
    OTP             string // Generated OTP
    Expires         string // Human-readable expiry
    InstitutionName string // Name of the institution
    InvitedBy       string // Name of person who invited them
}

// SendTeamInviteAcceptedRequest - Notification to admin when invitation is accepted
type SendTeamInviteAcceptedRequest struct {
    To              string // Admin's email
    AdminName       string // Admin's name
    UserName        string // Name of user who accepted
    UserEmail       string // Email of user who accepted
    InstitutionName string // Name of the institution
}

// SendTeamInviteDeclinedRequest - Notification to admin when invitation is declined
type SendTeamInviteDeclinedRequest struct {
    To              string // Admin's email
    AdminName       string // Admin's name
    UserName        string // Name of user who declined
    UserEmail       string // Email of user who declined
    InstitutionName string // Name of the institution
}