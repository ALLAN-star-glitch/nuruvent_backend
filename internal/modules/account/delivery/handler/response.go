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
	Type        string `json:"type"` // "personal" or "institution"
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

// MemberResponse represents an account member in API responses
type MemberResponse struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	JoinedAt  string `json:"joined_at"`
}

// NewMemberResponse creates a new MemberResponse from domain AccountMember
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

// AccountTypeResponse represents an account type in API responses
type AccountTypeResponse struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// NewAccountTypeResponse creates a new AccountTypeResponse from domain AccountType
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