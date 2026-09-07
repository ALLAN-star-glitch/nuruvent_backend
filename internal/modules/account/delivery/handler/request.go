// internal/modules/account/delivery/handler/request.go

package handler

// ============================================================
// REQUEST DTOS
// ============================================================

// CreatePersonalAccountRequest represents a request to create a personal account
type CreatePersonalAccountRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Phone string `json:"phone,omitempty"`
}

// CreateInstitutionAccountRequest represents a request to create an institution account
type CreateInstitutionAccountRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Phone       string `json:"phone,omitempty"`
	Website     string `json:"website,omitempty"`
	Description string `json:"description,omitempty"`
}

// UpdateAccountRequest represents a request to update an account
type UpdateAccountRequest struct {
	Name        string `json:"name,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Website     string `json:"website,omitempty"`
	Description string `json:"description,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
}

// ToMap converts UpdateAccountRequest to a map of updates
func (r *UpdateAccountRequest) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})
	if r.Name != "" {
		updates["name"] = r.Name
	}
	if r.Email != "" {
		updates["email"] = r.Email
	}
	if r.Phone != "" {
		updates["phone"] = r.Phone
	}
	if r.Website != "" {
		updates["website"] = r.Website
	}
	if r.Description != "" {
		updates["description"] = r.Description
	}
	if r.LogoURL != "" {
		updates["logo_url"] = r.LogoURL
	}
	return updates
}

// AddMemberRequest represents a request to add a member to an account
type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=account_admin trainer"`
}

// UpdateMemberRoleRequest represents a request to update a member's role
type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=account_admin trainer"`
}

// ============================================================
// LIST REQUEST DTOS
// ============================================================

// ListAccountsRequest represents a request to list accounts
type ListAccountsRequest struct {
	Type   string `json:"type" query:"type"` // "personal" or "institution"
	Search string `json:"search" query:"search"`
	Status string `json:"status" query:"status"`
	Limit  int    `json:"limit" query:"limit"`
	Offset int    `json:"offset" query:"offset"`
}