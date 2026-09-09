// internal/modules/team/service/notification_service.go

package service

import "context"

// ============================================================
// AI-GENERATED CONTENT (Future)
// ============================================================

// PersonalizedInvitationContent represents AI-generated email content
// ✅ This is where AI-generated content will go in the future
type PersonalizedInvitationContent struct {
	Subject      string   // Personalized subject line
	Greeting     string   // "Hi John,"
	Intro        string   // Personalized introduction
	Body         string   // Main message
	Benefits     []string // What they'll gain
	CallToAction string   // Button text
	Closing      string   // Sign-off
	PSS          string   // P.S. message
}

// ============================================================
// NOTIFICATION SERVICE INTERFACE
// ============================================================

// NotificationService handles all email notifications for the team module
type NotificationService interface {
	// SendTeamInviteExistingUser sends invitation to existing users (has account)
	// ✅ NO ROLE - Roles are inherited from account level
	// ✅ NO OTP - User clicks accept link to join
	// ✅ AI-READY - PersonalizedContent field for future AI integration
	SendTeamInviteExistingUser(ctx context.Context, req SendTeamInviteExistingUserRequest) error
	
	// SendTeamInviteRegistration sends invitation to new users (no account yet)
	// ✅ NO ROLE - Roles are inherited from account level
	// ✅ NO OTP - User clicks registration link with token embedded
	// ✅ AI-READY - PersonalizedContent field for future AI integration
	SendTeamInviteRegistration(ctx context.Context, req SendTeamInviteRegistrationRequest) error
	
	// SendTeamInviteAccepted sends notification to admin when user accepts
	// ✅ AI-READY - PersonalizedContent field for future AI integration
	SendTeamInviteAccepted(ctx context.Context, req SendTeamInviteAcceptedRequest) error
	
	// SendTeamInviteDeclined sends notification to admin when user declines
	// ✅ AI-READY - PersonalizedContent field for future AI integration
	SendTeamInviteDeclined(ctx context.Context, req SendTeamInviteDeclinedRequest) error
}

// ============================================================
// REQUEST STRUCTS (AI-Ready)
// ============================================================

// SendTeamInviteExistingUserRequest - Existing users (has account)
type SendTeamInviteExistingUserRequest struct {
	To         string // Recipient email
	UserName   string // Recipient name
	InvitedBy  string // Name of person who invited
	TeamName   string // Name of the team
	TeamID     string // Team ID
	AcceptLink string // https://nuruvent.com/invitations/accept?token=xxx
	ExpiresIn  string // "7 days"
	
	// ✅ AI-Ready: Personalized content (optional, for future use)
	PersonalizedContent *PersonalizedInvitationContent
}

// SendTeamInviteRegistrationRequest - New users (no account)
type SendTeamInviteRegistrationRequest struct {
	To               string // Recipient email
	Name             string // Recipient name (or email if unknown)
	InvitedBy        string // Name of person who invited
	TeamName         string // Name of the team
	TeamID           string // Team ID
	RegistrationLink string // https://nuruvent.com/register?token=xxx
	ExpiresIn        string // "7 days"
	
	// ✅ AI-Ready: Personalized content (optional, for future use)
	PersonalizedContent *PersonalizedInvitationContent
}

// SendTeamInviteAcceptedRequest - Admin notification
type SendTeamInviteAcceptedRequest struct {
	To         string // Admin email
	AdminName  string // Admin name
	UserName   string // User who accepted
	UserEmail  string // User who accepted email
	TeamName   string // Team name
	TeamID     string // Team ID
	
	// ✅ AI-Ready: Personalized content (optional, for future use)
	PersonalizedContent *PersonalizedInvitationContent
}

// SendTeamInviteDeclinedRequest - Admin notification
type SendTeamInviteDeclinedRequest struct {
	To         string // Admin email
	AdminName  string // Admin name
	UserName   string // User who declined
	UserEmail  string // User who declined email
	TeamName   string // Team name
	TeamID     string // Team ID
	
	// ✅ AI-Ready: Personalized content (optional, for future use)
	PersonalizedContent *PersonalizedInvitationContent
}