package acchandler

import (
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/domain"
)

// ============================================================
// REQUEST DTOS
// ============================================================

// UpdateAccountRequest represents the request to update an account
type UpdateAccountRequest struct {
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

// UpdateProfileRequest represents the request to update a profile
type UpdateProfileRequest struct {
	Name        string `json:"name,omitempty"`
	Phone       string `json:"phone,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

// UpdateProfessionalTypeRequest represents the request to update professional type
type UpdateProfessionalTypeRequest struct {
	ProfessionalTypeID *string `json:"professional_type_id,omitempty"`
}

// ============================================================
// RESPONSE DTOS
// ============================================================

// AccountResponse represents the account response
type AccountResponse struct {
	ID             string     `json:"id"`
	Slug           string     `json:"slug"`
	Name           string     `json:"name"`
	DisplayName    string     `json:"display_name,omitempty"`
	Email          string     `json:"email"`
	Phone          string     `json:"phone"`
	AccountTypeID  string     `json:"account_type_id"`
	ProfessionalTypeID *string `json:"professional_type_id,omitempty"`
	InstitutionID  *string    `json:"institution_id,omitempty"`
	EmailVerified  bool       `json:"email_verified"`
	IdentityVerified bool     `json:"identity_verified"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ============================================================
// RESPONSE BUILDERS
// ============================================================

// NewAccountResponse converts domain.Account to AccountResponse
func NewAccountResponse(account *domain.Account) AccountResponse {
	if account == nil {
		return AccountResponse{}
	}

	return AccountResponse{
		ID:              account.ID,
		Slug:            account.Slug,
		Name:            account.Name,
		DisplayName:     account.DisplayName,
		Email:           account.Email,
		Phone:           account.Phone,
		AccountTypeID:   account.AccountTypeID,
		ProfessionalTypeID: account.ProfessionalTypeID,
		InstitutionID:   account.InstitutionID,
		EmailVerified:   account.EmailVerified,
		IdentityVerified: account.IdentityVerified,
		IsActive:        account.IsActive,
		CreatedAt:       account.CreatedAt,
		UpdatedAt:       account.UpdatedAt,
	}
}