// internal/modules/account/delivery/handler/account_handler.go

package handler

import (
	"context"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
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
// HELPERS
// ============================================================

func getUserID(c fiber.Ctx) (string, error) {
	userID := c.Locals("user_id")
	if userID == nil {
		return "", errors.New("user not authenticated")
	}
	userIDStr, ok := userID.(string)
	if !ok {
		return "", errors.New("invalid user ID")
	}
	return userIDStr, nil
}

func getUserIDOptional(c fiber.Ctx) string {
	if user := c.Locals("user_id"); user != nil {
		if id, ok := user.(string); ok {
			return id
		}
	}
	return ""
}

func getQueryInt(c fiber.Ctx, key string, defaultValue int) int {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVal
}

func getQueryString(c fiber.Ctx, key string, defaultValue string) string {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// withContext adds user ID to context if available
func (h *AccountHandler) withContext(c fiber.Ctx) context.Context {
	ctx := c.Context()
	if userID := getUserIDOptional(c); userID != "" {
		ctx = context.WithValue(ctx, "user_id", userID)
	}
	return ctx
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
	ctx := h.withContext(c)

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

	ctx := h.withContext(c)

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

	ctx := h.withContext(c)

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
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req CreatePersonalAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := h.withContext(c)

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
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req CreateInstitutionAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := h.withContext(c)

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

	ctx := h.withContext(c)

	account, err := h.svc.GetAccountByID(ctx, id)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		return response.InternalError(c, "Failed to get account", fiber.Map{
			"error": err.Error(),
		})
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

	ctx := h.withContext(c)

	account, err := h.svc.GetAccountBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		return response.InternalError(c, "Failed to get account", fiber.Map{
			"error": err.Error(),
		})
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
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := h.withContext(c)

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

	// Get user ID for context and permission checks
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req UpdateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	// Create context with user ID
	ctx := context.WithValue(c.Context(), "user_id", userID)

	updates := req.ToMap()

	account, err := h.svc.UpdateAccount(ctx, id, updates)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		// Check for permission errors
		if errors.Is(err, accountdomain.ErrPermissionDenied) {
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

	ctx := h.withContext(c)

	if err := h.svc.DeleteAccount(ctx, id); err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
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

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req AddMemberRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := h.withContext(c)

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

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := h.withContext(c)

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

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req UpdateMemberRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := h.withContext(c)

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

	ctx := h.withContext(c)

	members, err := h.svc.GetAccountMembers(ctx, accountID)
	if err != nil {
		if errors.Is(err, accountdomain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
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

	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := h.withContext(c)

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