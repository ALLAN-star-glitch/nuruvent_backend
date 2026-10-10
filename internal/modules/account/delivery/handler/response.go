// internal/modules/account/delivery/handler/response.go

package handler

import (
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
)

// ============================================================
// RESPONSE DTOS
// ============================================================

// AccountResponse represents an account in API responses
type AccountResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Slug        string `json:"slug"`
	Type        string `json:"type"`
	Email       string `json:"email"`
	Phone       string `json:"phone,omitempty"`
	Website     string `json:"website,omitempty"`
	Description string `json:"description,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	Country     string `json:"country,omitempty"`
	Status      string `json:"status"`
	KYCStatus   string `json:"kyc_status,omitempty"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// NewAccountResponse creates a new AccountResponse from domain Account
func NewAccountResponse(account *accountdomain.Account) AccountResponse {
	if account == nil {
		return AccountResponse{}
	}
	return AccountResponse{
		ID:          account.ID,
		Name:        account.Name,
		DisplayName: account.DisplayName,
		Slug:        account.Slug,
		Type:        account.Type,
		Email:       account.Email,
		Phone:       account.Phone,
		Website:     account.Website,
		Description: account.Description,
		LogoURL:     account.LogoURL,
		Address:     account.Address,
		City:        account.City,
		Country:     account.Country,
		Status:      account.Status,
		KYCStatus:   account.KYCStatus,
		IsActive:    account.IsActive,
		CreatedAt:   account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   account.UpdatedAt.Format(time.RFC3339),
	}
}

// ============================================================
// MEMBER RESPONSE
// ============================================================
//
// Two constructors:
//
//   NewMemberResponse(member)          → metadata only (member.id, role, ...)
//   NewMemberResponseWithUser(m, user) → metadata + user identity fields
//
// The "with user" variant is used by GetAccountMembers so the response
// carries name/email/avatar for the UI's member pickers. The plain
// variant is used by mutations (AddMember, UpdateMemberRole) where the
// UI doesn't need identity info.

type MemberResponse struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	JoinedAt  string `json:"joined_at"`

	// User identity — only populated by GetAccountMembers. Empty on
	// mutation responses (AddMember, UpdateMemberRole).
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// NewMemberResponse builds a member response without user identity.
// Used by mutations where the caller already knows who they acted on.
func NewMemberResponse(member *accountdomain.AccountMember) MemberResponse {
	if member == nil {
		return MemberResponse{}
	}
	return MemberResponse{
		ID:        member.ID,
		AccountID: member.AccountID,
		UserID:    member.UserID,
		Role:      string(member.Role),
		IsActive:  member.IsActive,
		JoinedAt:  member.JoinedAt.Format(time.RFC3339),
	}
}

// NewMemberResponseWithUser builds a member response enriched with the
// user's display fields. Used by GetAccountMembers.
//
// user may be nil — for example if a user row was soft-deleted but the
// account_members row still exists. In that case only the member
// metadata is returned, and the frontend falls back to the user ID.
func NewMemberResponseWithUser(
	member *accountdomain.AccountMember,
	user *accountdomain.UserInfo,
) MemberResponse {
	if member == nil {
		return MemberResponse{}
	}

	resp := MemberResponse{
		ID:        member.ID,
		AccountID: member.AccountID,
		UserID:    member.UserID,
		Role:      string(member.Role),
		IsActive:  member.IsActive,
		JoinedAt:  member.JoinedAt.Format(time.RFC3339),
	}

	if user != nil {
		resp.Name = user.Name
		resp.DisplayName = user.DisplayName
		resp.Email = user.Email
		resp.AvatarURL = user.AvatarURL
	}

	return resp
}

// ============================================================
// ACCOUNT TYPE
// ============================================================

type AccountTypeResponse struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func NewAccountTypeResponse(accountType *accountdomain.AccountType) AccountTypeResponse {
	if accountType == nil {
		return AccountTypeResponse{}
	}
	return AccountTypeResponse{
		ID:          accountType.ID,
		Slug:        accountType.Slug,
		Name:        accountType.Name,
		DisplayName: accountType.DisplayName,
		Description: accountType.Description,
		IsActive:    accountType.IsActive,
	}
}