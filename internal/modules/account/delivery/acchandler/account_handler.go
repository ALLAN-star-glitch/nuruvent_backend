package acchandler

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// AccountHandler handles HTTP requests for account operations
type AccountHandler struct {
	service service.Service
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(svc service.Service) *AccountHandler {
	return &AccountHandler{
		service: svc,
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

// ============================================================
// ACCOUNT CRUD HANDLERS
// ============================================================

// GetAccountByID handles GET /api/v1/accounts/:id
// @Summary Get account by ID
// @Description Get account details by ID
// @Tags Accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Security BearerAuth
// @Router /api/v1/accounts/{id} [get]
func (h *AccountHandler) GetAccountByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	account, err := h.service.GetAccountByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		return response.InternalError(c, "Failed to get account", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account retrieved successfully", NewAccountResponse(account))
}

// GetAccountByEmail handles GET /api/v1/accounts/email/:email
// @Summary Get account by email
// @Description Get account details by email address
// @Tags Accounts
// @Accept json
// @Produce json
// @Param email path string true "Email address"
// @Success 200 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Security BearerAuth
// @Router /api/v1/accounts/email/{email} [get]
func (h *AccountHandler) GetAccountByEmail(c fiber.Ctx) error {
	email := c.Params("email")
	if email == "" {
		return response.BadRequest(c, "Email is required", nil)
	}

	account, err := h.service.GetAccountByEmail(c.Context(), email)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		return response.InternalError(c, "Failed to get account", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account retrieved successfully", NewAccountResponse(account))
}

// UpdateAccount handles PUT /api/v1/accounts/:id
// @Summary Update account
// @Description Update account details
// @Tags Accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param request body UpdateAccountRequest true "Update account request"
// @Success 200 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Security BearerAuth
// @Router /api/v1/accounts/{id} [put]
func (h *AccountHandler) UpdateAccount(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	var req UpdateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	// Get existing account
	account, err := h.service.GetAccountByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		return response.InternalError(c, "Failed to get account", fiber.Map{
			"error": err.Error(),
		})
	}

	// Update fields
	if req.Name != "" {
		account.Name = req.Name
	}
	if req.DisplayName != "" {
		account.DisplayName = req.DisplayName
	}
	if req.Phone != "" {
		account.Phone = req.Phone
	}

	if err := h.service.UpdateAccount(c.Context(), account); err != nil {
		return response.InternalError(c, "Failed to update account", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account updated successfully", NewAccountResponse(account))
}

// DeleteAccount handles DELETE /api/v1/accounts/:id
// @Summary Delete account
// @Description Soft delete an account
// @Tags Accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Security BearerAuth
// @Router /api/v1/accounts/{id} [delete]
func (h *AccountHandler) DeleteAccount(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	if err := h.service.DeleteAccount(c.Context(), id); err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		return response.InternalError(c, "Failed to delete account", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account deleted successfully", nil)
}

// ============================================================
// PROFILE HANDLERS
// ============================================================

// UpdateProfile handles PUT /api/v1/accounts/:id/profile
// @Summary Update profile
// @Description Update account profile (name, phone, display name)
// @Tags Accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param request body UpdateProfileRequest true "Update profile request"
// @Success 200 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Security BearerAuth
// @Router /api/v1/accounts/{id}/profile [put]
func (h *AccountHandler) UpdateProfile(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	var req UpdateProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	request := service.UpdateProfileRequest{
		Name:        req.Name,
		Phone:       req.Phone,
		DisplayName: req.DisplayName,
	}

	account, err := h.service.UpdateProfile(c.Context(), id, request)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		return response.InternalError(c, "Failed to update profile", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Profile updated successfully", NewAccountResponse(account))
}

// UpdateProfessionalType handles PUT /api/v1/accounts/:id/professional-type
// @Summary Update professional type
// @Description Update account professional type
// @Tags Accounts
// @Accept json
// @Produce json
// @Param id path string true "Account ID"
// @Param request body UpdateProfessionalTypeRequest true "Update professional type request"
// @Success 200 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Security BearerAuth
// @Router /api/v1/accounts/{id}/professional-type [put]
func (h *AccountHandler) UpdateProfessionalType(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	var req UpdateProfessionalTypeRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	account, err := h.service.UpdateProfessionalType(c.Context(), id, req.ProfessionalTypeID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		return response.InternalError(c, "Failed to update professional type", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Professional type updated successfully", NewAccountResponse(account))
}

// ============================================================
// CURRENT USER HANDLERS
// ============================================================

// GetCurrentAccount handles GET /api/v1/accounts/me
// @Summary Get current account
// @Description Get the currently authenticated user's account
// @Tags Accounts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse{data=AccountResponse}
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/me [get]
func (h *AccountHandler) GetCurrentAccount(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	account, err := h.service.GetAccountByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		return response.InternalError(c, "Failed to get account", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account retrieved successfully", NewAccountResponse(account))
}

// UpdateCurrentAccount handles PUT /api/v1/accounts/me
// @Summary Update current account
// @Description Update the currently authenticated user's account
// @Tags Accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateAccountRequest true "Update account request"
// @Success 200 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/me [put]
func (h *AccountHandler) UpdateCurrentAccount(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req UpdateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	account, err := h.service.GetAccountByID(c.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		return response.InternalError(c, "Failed to get account", fiber.Map{
			"error": err.Error(),
		})
	}

	if req.Name != "" {
		account.Name = req.Name
	}
	if req.DisplayName != "" {
		account.DisplayName = req.DisplayName
	}
	if req.Phone != "" {
		account.Phone = req.Phone
	}

	if err := h.service.UpdateAccount(c.Context(), account); err != nil {
		return response.InternalError(c, "Failed to update account", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account updated successfully", NewAccountResponse(account))
}

// UpdateCurrentProfile handles PUT /api/v1/accounts/me/profile
// @Summary Update current profile
// @Description Update the currently authenticated user's profile
// @Tags Accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Update profile request"
// @Success 200 {object} response.BaseResponse{data=AccountResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/me/profile [put]
func (h *AccountHandler) UpdateCurrentProfile(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req UpdateProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	request := service.UpdateProfileRequest{
		Name:        req.Name,
		Phone:       req.Phone,
		DisplayName: req.DisplayName,
	}

	account, err := h.service.UpdateProfile(c.Context(), userID, request)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		return response.InternalError(c, "Failed to update profile", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Profile updated successfully", NewAccountResponse(account))
}