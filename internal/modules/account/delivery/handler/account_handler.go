// internal/modules/account/delivery/handler/account_handler.go

package handler

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/handlerhelper"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// AccountHandler handles account HTTP requests
type AccountHandler struct {
	svc service.Service
}



// NewAccountHandler creates a new account handler
func NewAccountHandler(svc service.Service) *AccountHandler {
	return &AccountHandler{
		svc: svc,
	}
}

// ============================================================
// ACCOUNT TYPE HANDLERS
// ============================================================

// GetAccountTypes godoc
// @Summary Get all account types
// @Description Get list of all account types (personal, institution, etc.)
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse{data=[]accountdomain.AccountType}
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/account-types [get]
func (h *AccountHandler) GetAccountTypes(c fiber.Ctx) error {
	ctx := handlerhelper.EnrichUserContext(c)

	types, err := h.svc.GetAccountTypes(ctx)
	if err != nil {
		return response.InternalError(c, "Failed to get account types", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account types retrieved successfully", types)
}

// GetAccountTypeByID godoc
// @Summary Get account type by ID
// @Description Get account type details by ID
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account Type ID"
// @Success 200 {object} response.BaseResponse{data=accountdomain.AccountType}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/account-types/{id} [get]
func (h *AccountHandler) GetAccountTypeByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Account type ID is required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	accountType, err := h.svc.GetAccountTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountTypeNotFound) {
			return response.NotFound(c, "Account type not found", nil)
		}
		return response.InternalError(c, "Failed to get account type", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account type retrieved successfully", accountType)
}

// GetAccountTypeBySlug godoc
// @Summary Get account type by slug
// @Description Get account type details by slug
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Account Type Slug"
// @Success 200 {object} response.BaseResponse{data=accountdomain.AccountType}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/account-types/slug/{slug} [get]
func (h *AccountHandler) GetAccountTypeBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return response.BadRequest(c, "Account type slug is required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	accountType, err := h.svc.GetAccountTypeBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountTypeNotFound) {
			return response.NotFound(c, "Account type not found", nil)
		}
		return response.InternalError(c, "Failed to get account type", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account type retrieved successfully", accountType)
}

// ============================================================
// ACCOUNT HANDLERS
// ============================================================

// CreatePersonalAccount godoc
// @Summary Create a personal account
// @Description Create a new personal account
// @Tags Accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreatePersonalAccountRequest true "Personal account details"
// @Success 201 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 409 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/personal [post]
func (h *AccountHandler) CreatePersonalAccount(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req CreatePersonalAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := handlerhelper.EnrichUserContext(c)

	cmd := service.CreatePersonalAccountCommand{
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		CreatedBy: userID,
	}

	account, err := h.svc.CreatePersonalAccount(ctx, cmd)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountAlreadyExists) {
			return response.Conflict(c, "Account with this email already exists", nil)
		}
		return response.InternalError(c, "Failed to create personal account", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Personal account created successfully", NewAccountResponse(account))
}

// CreateInstitutionAccount godoc
// @Summary Create an institution account
// @Description Create a new institution account
// @Tags Accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateInstitutionAccountRequest true "Institution account details"
// @Success 201 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 409 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/institution [post]
func (h *AccountHandler) CreateInstitutionAccount(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req CreateInstitutionAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := handlerhelper.EnrichUserContext(c)

	cmd := service.CreateInstitutionAccountCommand{
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		Website:     req.Website,
		Description: req.Description,
		CreatedBy:   userID,
	}

	account, err := h.svc.CreateInstitutionAccount(ctx, cmd)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountAlreadyExists) {
			return response.Conflict(c, "Account with this email already exists", nil)
		}
		return response.InternalError(c, "Failed to create institution account", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Institution account created successfully", NewAccountResponse(account))
}

// GetAccountByID godoc
// @Summary Get account by ID
// @Description Get account details by ID
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{id} [get]
func (h *AccountHandler) GetAccountByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	account, err := h.svc.GetAccountByID(ctx, id)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, accountdomain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to view this account", nil)
		}
		return response.InternalError(c, "Failed to get account", fiber.Map{
			"error": err.Error(),
		})
	}

	if account == nil {
		return response.NotFound(c, "Account not found", nil)
	}

	return response.Success(c, "Account retrieved successfully", NewAccountResponse(account))
}

// GetAccountBySlug godoc
// @Summary Get account by slug
// @Description Get account details by slug
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Account Slug"
// @Success 200 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/slug/{slug} [get]
func (h *AccountHandler) GetAccountBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return response.BadRequest(c, "Account slug is required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	account, err := h.svc.GetAccountBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, accountdomain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to view this account", nil)
		}
		return response.InternalError(c, "Failed to get account", fiber.Map{
			"error": err.Error(),
		})
	}

	if account == nil {
		return response.NotFound(c, "Account not found", nil)
	}

	return response.Success(c, "Account retrieved successfully", NewAccountResponse(account))
}

// GetMyAccounts godoc
// @Summary Get my accounts
// @Description Get all accounts the authenticated user belongs to
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse{data=[]AccountResponse}
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/accounts [get]
func (h *AccountHandler) GetMyAccounts(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	accounts, err := h.svc.GetUserAccounts(ctx, userID)
	if err != nil {
		return response.InternalError(c, "Failed to get accounts", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]AccountResponse, len(accounts))
	for i, account := range accounts {
		responses[i] = NewAccountResponse(account)
	}

	return response.Success(c, "Accounts retrieved successfully", responses)
}

// UpdateAccount godoc
// @Summary Update an account
// @Description Update account details
// @Tags Accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param request body UpdateAccountRequest true "Account update details"
// @Success 200 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{id} [put]
func (h *AccountHandler) UpdateAccount(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	if _, err := handlerhelper.GetUserID(c); err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req UpdateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := handlerhelper.EnrichUserContext(c)

	updates := req.ToMap()

	account, err := h.svc.UpdateAccount(ctx, id, updates)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, accountdomain.ErrForbidden) {
			return response.Forbidden(c, "You don't have permission to update this account", nil)
		}
		return response.InternalError(c, "Failed to update account", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account updated successfully", NewAccountResponse(account))
}

// DeleteAccount godoc
// @Summary Delete an account
// @Description Soft delete an account
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{id} [delete]
func (h *AccountHandler) DeleteAccount(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	if err := h.svc.DeleteAccount(ctx, id); err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, accountdomain.ErrForbidden) {
			return response.Forbidden(c, "You don't have permission to delete this account", nil)
		}
		return response.InternalError(c, "Failed to delete account", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account deleted successfully", nil)
}

// ============================================================
// ACCOUNT MEMBER HANDLERS
// ============================================================

// AddMember godoc
// @Summary Add member to account
// @Description Add a user to an account with a specific role
// @Tags Accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param request body AddMemberRequest true "Member details"
// @Success 200 {object} response.BaseResponse{data=MemberResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 409 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{id}/members [post]
func (h *AccountHandler) AddMember(c fiber.Ctx) error {
	accountID := c.Params("id")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req AddMemberRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := handlerhelper.EnrichUserContext(c)

	cmd := service.AddMemberCommand{
		AccountID: accountID,
		UserID:    req.UserID,
		Role:      req.Role,
		InvitedBy: userID,
	}

	member, err := h.svc.AddMember(ctx, cmd)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, accountdomain.ErrAccountMemberAlreadyExists) {
			return response.Conflict(c, "User is already a member of this account", nil)
		}
		if errors.Is(err, accountdomain.ErrForbidden) {
			return response.Forbidden(c, "You don't have permission to add members to this account", nil)
		}
		return response.InternalError(c, "Failed to add member", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Member added successfully", NewMemberResponse(member))
}

// RemoveMember godoc
// @Summary Remove member from account
// @Description Remove a user from an account
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param userId path string true "User ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{id}/members/{userId} [delete]
func (h *AccountHandler) RemoveMember(c fiber.Ctx) error {
	accountID := c.Params("id")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	memberUserID := c.Params("userId")
	if memberUserID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	if err := h.svc.RemoveMember(ctx, accountID, memberUserID, userID); err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, accountdomain.ErrAccountMemberNotFound) {
			return response.NotFound(c, "Member not found", nil)
		}
		if errors.Is(err, accountdomain.ErrCannotRemoveSelf) {
			return response.BadRequest(c, "Cannot remove yourself from an account", nil)
		}
		if errors.Is(err, accountdomain.ErrLastAdminCannotLeave) {
			return response.BadRequest(c, "Cannot remove the last admin from an account", nil)
		}
		if errors.Is(err, accountdomain.ErrForbidden) {
			return response.Forbidden(c, "You don't have permission to remove members from this account", nil)
		}
		return response.InternalError(c, "Failed to remove member", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Member removed successfully", nil)
}

// UpdateMemberRole godoc
// @Summary Update member role
// @Description Update a member's role in an account
// @Tags Accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param userId path string true "User ID"
// @Param request body UpdateMemberRoleRequest true "New role"
// @Success 200 {object} response.BaseResponse{data=MemberResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{id}/members/{userId}/role [put]
func (h *AccountHandler) UpdateMemberRole(c fiber.Ctx) error {
	accountID := c.Params("id")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	memberUserID := c.Params("userId")
	if memberUserID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req UpdateMemberRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := handlerhelper.EnrichUserContext(c)

	member, err := h.svc.UpdateMemberRole(ctx, accountID, memberUserID, req.Role, userID)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, accountdomain.ErrAccountMemberNotFound) {
			return response.NotFound(c, "Member not found", nil)
		}
		if errors.Is(err, accountdomain.ErrCannotChangeOwnRole) {
			return response.BadRequest(c, "Cannot change your own role", nil)
		}
		if errors.Is(err, accountdomain.ErrInvalidRole) {
			return response.BadRequest(c, "Invalid role", nil)
		}
		if errors.Is(err, accountdomain.ErrForbidden) {
			return response.Forbidden(c, "You don't have permission to update member roles", nil)
		}
		return response.InternalError(c, "Failed to update member role", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Member role updated successfully", NewMemberResponse(member))
}

// GetAccountMembers godoc
// @Summary Get account members
// @Description Get all members of an account
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.BaseResponse{data=[]MemberResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{id}/members [get]
func (h *AccountHandler) GetAccountMembers(c fiber.Ctx) error {
	accountID := c.Params("id")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	members, err := h.svc.GetAccountMembers(ctx, accountID)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, accountdomain.ErrForbidden) {
			return response.Forbidden(c, "You don't have permission to view this account's members", nil)
		}
		return response.InternalError(c, "Failed to get account members", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]MemberResponse, len(members))
	for i, member := range members {
		responses[i] = NewMemberResponse(member)
	}

	return response.Success(c, "Account members retrieved successfully", responses)
}

// LeaveAccount godoc
// @Summary Leave an account
// @Description Leave an account you are a member of
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{id}/leave [post]
func (h *AccountHandler) LeaveAccount(c fiber.Ctx) error {
	accountID := c.Params("id")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	if err := h.svc.LeaveAccount(ctx, accountID, userID); err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, accountdomain.ErrAccountMemberNotFound) {
			return response.NotFound(c, "You are not a member of this account", nil)
		}
		if errors.Is(err, accountdomain.ErrLastAdminCannotLeave) {
			return response.BadRequest(c, "Cannot leave as the last admin of an account", nil)
		}
		return response.InternalError(c, "Failed to leave account", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "You have left the account successfully", nil)
}


// ============================================================
// USER AVATAR HANDLERS
// ============================================================

// UploadMyAvatar godoc
// @Summary Upload my avatar
// @Description Upload or replace the authenticated user's avatar
// @Tags Accounts
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Avatar image"
// @Success 200 {object} response.BaseResponse{data=UserInfoResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/avatar [post]
func (h *AccountHandler) UploadMyAvatar(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return response.BadRequest(c, "Avatar file is required", nil)
	}

	data, contentType, err := handlerhelper.ReadUploadedImage(file)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	user, err := h.svc.UploadUserAvatar(ctx, userID, data, file.Filename, contentType)
	if err != nil {
		if errors.Is(err, accountdomain.ErrForbidden) {
			return response.Forbidden(c, "You cannot modify this avatar", nil)
		}
		return response.InternalError(c, "Failed to upload avatar", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Avatar uploaded successfully", NewUserInfoResponse(user))
}

// DeleteMyAvatar godoc
// @Summary Delete my avatar
// @Description Remove the authenticated user's avatar
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/avatar [delete]
func (h *AccountHandler) DeleteMyAvatar(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	if err := h.svc.DeleteUserAvatar(ctx, userID); err != nil {
		if errors.Is(err, accountdomain.ErrForbidden) {
			return response.Forbidden(c, "You cannot delete this avatar", nil)
		}
		return response.InternalError(c, "Failed to delete avatar", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Avatar deleted successfully", nil)
}

// ============================================================
// ACCOUNT LOGO HANDLERS
// ============================================================

// UploadAccountLogo godoc
// @Summary Upload account logo
// @Description Upload or replace the logo for an account
// @Tags Accounts
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param logo formData file true "Logo image"
// @Success 200 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{id}/logo [post]
func (h *AccountHandler) UploadAccountLogo(c fiber.Ctx) error {
	accountID := c.Params("id")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	if _, err := handlerhelper.GetUserID(c); err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	file, err := c.FormFile("logo")
	if err != nil {
		return response.BadRequest(c, "Logo file is required", nil)
	}

	data, contentType, err := handlerhelper.ReadUploadedImage(file)
	if err != nil {
		return response.BadRequest(c, err.Error(), nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	account, err := h.svc.UploadAccountLogo(ctx, accountID, data, file.Filename, contentType)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, accountdomain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to modify this account", nil)
		}
		return response.InternalError(c, "Failed to upload logo", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Logo uploaded successfully", NewAccountResponse(account))
}

// DeleteAccountLogo godoc
// @Summary Delete account logo
// @Description Remove the logo for an account
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{id}/logo [delete]
func (h *AccountHandler) DeleteAccountLogo(c fiber.Ctx) error {
	accountID := c.Params("id")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	if _, err := handlerhelper.GetUserID(c); err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	if err := h.svc.DeleteAccountLogo(ctx, accountID); err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, accountdomain.ErrForbidden) {
			return response.Forbidden(c, "You do not have permission to modify this account", nil)
		}
		return response.InternalError(c, "Failed to delete logo", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Logo deleted successfully", nil)
}

// ============================================================
// USER PROFILE HANDLERS
// ============================================================

// GetMyProfile godoc
// @Summary Get my profile
// @Description Get the authenticated user's full profile
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse{data=ProfileResponse}
// @Failure 401 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/profile [get]
func (h *AccountHandler) GetMyProfile(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	user, err := h.svc.GetMyProfile(ctx, userID)
	if err != nil {
		return response.InternalError(c, "Failed to load profile", fiber.Map{
			"error": err.Error(),
		})
	}
	if user == nil {
		return response.NotFound(c, "Profile not found", nil)
	}

	return response.Success(c, "Profile retrieved successfully", NewProfileResponse(user))
}

// UpdateMyProfile godoc
// @Summary Update my profile
// @Description Update the authenticated user's profile. Partial update — only
//              fields present in the body are changed.
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Profile updates"
// @Success 200 {object} response.BaseResponse{data=ProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/profile [put]
func (h *AccountHandler) UpdateMyProfile(c fiber.Ctx) error {
	userID, err := handlerhelper.GetUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req UpdateProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := handlerhelper.EnrichUserContext(c)

	user, err := h.svc.UpdateMyProfile(ctx, userID, req.ToService())
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "cannot be empty") ||
			strings.Contains(msg, "must be at most") ||
			strings.Contains(msg, "must be a valid URL") ||
			strings.Contains(msg, "at most") ||
			strings.Contains(msg, "invalid after sanitization") {
			return response.BadRequest(c, msg, nil)
		}
		return response.InternalError(c, "Failed to update profile", fiber.Map{
			"error": msg,
		})
	}

	return response.Success(c, "Profile updated successfully", NewProfileResponse(user))
}

// GetPublicProfile godoc
// @Summary Get a public user profile
// @Description Get a user's public profile (safe fields only). No auth required.
// @Tags Profile
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.BaseResponse{data=PublicProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/{id}/profile [get]
func (h *AccountHandler) GetPublicProfile(c fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	profile, err := h.svc.GetPublicProfile(ctx, userID)
	if err != nil {
		return response.InternalError(c, "Failed to load profile", fiber.Map{
			"error": err.Error(),
		})
	}
	if profile == nil {
		return response.NotFound(c, "Profile not found", nil)
	}

	return response.Success(c, "Profile retrieved successfully", NewPublicProfileResponse(profile))
}

// GetPublicProfileBySlug godoc
// @Summary Get a public user profile by slug
// @Description Get a user's public profile by slug. No auth required.
// @Tags Profile
// @Produce json
// @Param slug path string true "User Slug"
// @Success 200 {object} response.BaseResponse{data=PublicProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/slug/{slug}/profile [get]
func (h *AccountHandler) GetPublicProfileBySlug(c fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return response.BadRequest(c, "Slug is required", nil)
	}

	ctx := handlerhelper.EnrichUserContext(c)

	profile, err := h.svc.GetPublicProfileBySlug(ctx, slug)
	if err != nil {
		return response.InternalError(c, "Failed to load profile", fiber.Map{
			"error": err.Error(),
		})
	}
	if profile == nil {
		return response.NotFound(c, "Profile not found", nil)
	}

	return response.Success(c, "Profile retrieved successfully", NewPublicProfileResponse(profile))
}