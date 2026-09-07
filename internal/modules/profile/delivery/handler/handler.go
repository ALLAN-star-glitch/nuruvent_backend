// internal/modules/profile/delivery/handler/handler.go

package handler

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// ============================================================
// PROFILE HANDLER
// ============================================================

type ProfileHandler struct {
	svc service.Service
}

func NewProfileHandler(svc service.Service) *ProfileHandler {
	return &ProfileHandler{
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

func getQueryBool(c fiber.Ctx, key string, defaultValue bool) bool {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val == "true" || val == "1"
}

// splitIDs splits a comma-separated string into a slice of IDs
func splitIDs(idsParam string) []string {
	if idsParam == "" {
		return []string{}
	}
	parts := []string{}
	for _, p := range splitString(idsParam, ",") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

// splitString is a simple string split helper
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

// detectImageMimeType detects MIME type from file content and filename
func detectImageMimeType(data []byte, filename string) string {
	// Check by magic bytes first (most reliable)
	if len(data) >= 4 {
		// PNG: 89 50 4E 47
		if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
			return "image/png"
		}
		// JPEG: FF D8 FF
		if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
			return "image/jpeg"
		}
		// GIF: 47 49 46
		if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
			return "image/gif"
		}
		// WEBP: 52 49 46 46 ... 57 45 42 50
		if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
			data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
			return "image/webp"
		}
		// BMP: 42 4D
		if data[0] == 0x42 && data[1] == 0x4D {
			return "image/bmp"
		}
	}

	// Check by file extension as fallback
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".bmp":
		return "image/bmp"
	}

	return "application/octet-stream"
}

// ============================================================
// USER PROFILE HANDLERS
// ============================================================

// GetMyProfile godoc
// @Summary Get my profile
// @Description Get the authenticated user's profile with full details
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse{data=UserProfileResponse}
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/profile [get]
func (h *ProfileHandler) GetMyProfile(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := context.WithValue(c.Context(), "user_id", userID)

	profile, err := h.svc.GetUserProfileWithDetails(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return response.NotFound(c, "User profile not found", nil)
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to view this profile", nil)
		}
		return response.InternalError(c, "Failed to get profile", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Profile retrieved successfully", NewUserProfileResponse(profile))
}

// GetUserProfile godoc
// @Summary Get user profile by ID
// @Description Get a user's profile by ID (public - basic info only)
// @Tags Profile
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.BaseResponse{data=UserProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/profile/users/{id} [get]
func (h *ProfileHandler) GetUserProfile(c fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	currentUserID := getUserIDOptional(c)
	ctx := context.WithValue(c.Context(), "user_id", currentUserID)

	// If requesting own profile, get with details
	if currentUserID == userID {
		profile, err := h.svc.GetUserProfileWithDetails(ctx, userID)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				return response.NotFound(c, "User profile not found", nil)
			}
			if errors.Is(err, domain.ErrPermissionDenied) {
				return response.Forbidden(c, "You don't have permission to view this profile", nil)
			}
			return response.InternalError(c, "Failed to get profile", fiber.Map{
				"error": err.Error(),
			})
		}
		return response.Success(c, "Profile retrieved successfully", NewUserProfileResponse(profile))
	}

	// Public endpoint - only basic info
	profile, err := h.svc.GetUserProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return response.NotFound(c, "User profile not found", nil)
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to view this profile", nil)
		}
		return response.InternalError(c, "Failed to get profile", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Profile retrieved successfully", NewUserProfileResponse(profile))
}

// GetUserProfiles godoc
// @Summary Get multiple user profiles
// @Description Get profiles for multiple users (basic info only)
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Param ids query string true "Comma-separated user IDs"
// @Success 200 {object} response.BaseResponse{data=[]UserProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/profile/users [get]
func (h *ProfileHandler) GetUserProfiles(c fiber.Ctx) error {
	idsParam := getQueryString(c, "ids", "")
	if idsParam == "" {
		return response.BadRequest(c, "User IDs are required", nil)
	}

	ids := splitIDs(idsParam)
	if len(ids) == 0 {
		return response.BadRequest(c, "At least one user ID is required", nil)
	}

	currentUserID := getUserIDOptional(c)
	ctx := context.WithValue(c.Context(), "user_id", currentUserID)

	profiles, err := h.svc.GetUserProfiles(ctx, ids)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to view these profiles", nil)
		}
		return response.InternalError(c, "Failed to get profiles", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]UserProfileResponse, len(profiles))
	for i, profile := range profiles {
		responses[i] = NewUserProfileResponse(profile)
	}

	return response.Success(c, "Profiles retrieved successfully", responses)
}

// ListUsers godoc
// @Summary List users
// @Description List users with filters (requires auth)
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Param team_id query string false "Team ID (user_id or account_id)"
// @Param team_type query string false "Team Type (personal or institution)"
// @Param user_id query string false "Filter by user ID"
// @Param search query string false "Search by name or email"
// @Param include_deleted query bool false "Include soft-deleted users"
// @Param only_deleted query bool false "Show ONLY soft-deleted users"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/profile/users/list [get]
func (h *ProfileHandler) ListUsers(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req ListUsersRequest
	if err := c.Bind().Query(&req); err != nil {
		return response.BadRequest(c, "Invalid query parameters", nil)
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Build team filter
	team := domain.TeamFilter{}
	if req.TeamID != "" && req.TeamType != "" {
		team = domain.TeamFilter{
			ID:   req.TeamID,
			Type: req.TeamType,
		}
	}

	// Build filters
	filters := domain.ListUsersFilters{
		Team:           team,
		UserID:         req.UserID,
		Search:         req.Search,
		IncludeDeleted: req.IncludeDeleted,
		OnlyDeleted:    req.OnlyDeleted,
		Limit:          req.Limit,
		Offset:         req.Offset,
		SortBy:         req.SortBy,
		SortOrder:      req.SortOrder,
	}

	ctx := context.WithValue(c.Context(), "user_id", userID)

	users, total, err := h.svc.ListUsers(ctx, filters)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to list users", nil)
		}
		return response.InternalError(c, "Failed to list users", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]UserProfileResponse, len(users))
	for i, user := range users {
		responses[i] = NewUserProfileResponse(user)
	}

	return response.Success(c, "Users retrieved successfully", fiber.Map{
		"data":        responses,
		"total":       total,
		"limit":       req.Limit,
		"offset":      req.Offset,
		"sort_by":     req.SortBy,
		"sort_order":  req.SortOrder,
	})
}

// UpdateMyProfile godoc
// @Summary Update my profile
// @Description Update the authenticated user's profile
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Profile update data"
// @Success 200 {object} response.BaseResponse{data=UserProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/profile [put]
func (h *ProfileHandler) UpdateMyProfile(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req UpdateProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := context.WithValue(c.Context(), "user_id", userID)

	updates := req.ToMap()

	profile, err := h.svc.UpdateUserProfile(ctx, userID, updates)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return response.NotFound(c, "User profile not found", nil)
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to update this profile", nil)
		}
		return response.InternalError(c, "Failed to update profile", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Profile updated successfully", NewUserProfileResponse(profile))
}

// ============================================================
// ACCOUNT PROFILE HANDLERS (replaces Institution)
// ============================================================

// GetAccountProfile godoc
// @Summary Get account profile
// @Description Get an account's profile (public - basic info only)
// @Tags Profile
// @Produce json
// @Param id path string true "Account ID"
// @Success 200 {object} response.BaseResponse{data=AccountProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/profile/accounts/{id} [get]
func (h *ProfileHandler) GetAccountProfile(c fiber.Ctx) error {
	accountID := c.Params("id")
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	viewerID := getUserIDOptional(c)
	ctx := context.WithValue(c.Context(), "user_id", viewerID)

	profile, err := h.svc.GetAccountProfile(ctx, accountID)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account profile not found", nil)
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to view this account profile", nil)
		}
		return response.InternalError(c, "Failed to get account profile", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account profile retrieved successfully", NewAccountProfileResponse(profile))
}

// GetAccountProfiles godoc
// @Summary Get multiple account profiles
// @Description Get profiles for multiple accounts (basic info only)
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Param ids query string true "Comma-separated account IDs"
// @Success 200 {object} response.BaseResponse{data=[]AccountProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/profile/accounts [get]
func (h *ProfileHandler) GetAccountProfiles(c fiber.Ctx) error {
	idsParam := getQueryString(c, "ids", "")
	if idsParam == "" {
		return response.BadRequest(c, "Account IDs are required", nil)
	}

	ids := splitIDs(idsParam)
	if len(ids) == 0 {
		return response.BadRequest(c, "At least one account ID is required", nil)
	}

	currentUserID := getUserIDOptional(c)
	ctx := context.WithValue(c.Context(), "user_id", currentUserID)

	profiles, err := h.svc.GetAccountProfiles(ctx, ids)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to view these account profiles", nil)
		}
		return response.InternalError(c, "Failed to get account profiles", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]AccountProfileResponse, len(profiles))
	for i, profile := range profiles {
		responses[i] = NewAccountProfileResponse(profile)
	}

	return response.Success(c, "Account profiles retrieved successfully", responses)
}

// ListAccounts godoc
// @Summary List accounts
// @Description List accounts with filters (requires auth)
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Param type query string false "Account type (personal or institution)"
// @Param search query string false "Search by name or email"
// @Param status query string false "Filter by account status"
// @Param kyc_status query string false "Filter by KYC status"
// @Param include_deleted query bool false "Include soft-deleted accounts"
// @Param only_deleted query bool false "Show ONLY soft-deleted accounts"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/profile/accounts/list [get]
func (h *ProfileHandler) ListAccounts(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req ListAccountsRequest
	if err := c.Bind().Query(&req); err != nil {
		return response.BadRequest(c, "Invalid query parameters", nil)
	}

	// Set defaults
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Build filters
	filters := domain.ListAccountsFilters{
		Type:           req.Type,
		AccountID:      req.AccountID,
		Search:         req.Search,
		Status:         req.Status,
		KYCStatus:      req.KYCStatus,
		IncludeDeleted: req.IncludeDeleted,
		OnlyDeleted:    req.OnlyDeleted,
		Limit:          req.Limit,
		Offset:         req.Offset,
		SortBy:         req.SortBy,
		SortOrder:      req.SortOrder,
	}

	ctx := context.WithValue(c.Context(), "user_id", userID)

	accounts, total, err := h.svc.ListAccounts(ctx, filters)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to list accounts", nil)
		}
		return response.InternalError(c, "Failed to list accounts", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]AccountProfileResponse, len(accounts))
	for i, account := range accounts {
		responses[i] = NewAccountProfileResponse(account)
	}

	return response.Success(c, "Accounts retrieved successfully", fiber.Map{
		"data":        responses,
		"total":       total,
		"limit":       req.Limit,
		"offset":      req.Offset,
		"sort_by":     req.SortBy,
		"sort_order":  req.SortOrder,
	})
}

// UpdateAccountProfile godoc
// @Summary Update account profile
// @Description Update an account's profile (requires admin access)
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param request body UpdateAccountRequest true "Account update data"
// @Success 200 {object} response.BaseResponse{data=AccountProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/profile [put]
func (h *ProfileHandler) UpdateAccountProfile(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	accountID := c.Params("accountId")
	if accountID == "" {
		accountID = c.Params("id")
	}
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	var req UpdateAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := context.WithValue(c.Context(), "user_id", userID)

	updates := req.ToMap()

	profile, err := h.svc.UpdateAccountProfile(ctx, accountID, updates)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account profile not found", nil)
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to update this account profile", nil)
		}
		return response.InternalError(c, "Failed to update account profile", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Account profile updated successfully", NewAccountProfileResponse(profile))
}

// ============================================================
// ORGANIZER INFO HANDLER
// ============================================================

// GetOrganizerInfo godoc
// @Summary Get organizer info for events
// @Description Get public-facing organizer info (used by events module)
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Param type query string true "Organizer type (personal or institution)"
// @Param id query string true "Organizer ID"
// @Success 200 {object} response.BaseResponse{data=OrganizerInfoResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/profile/organizer [get]
func (h *ProfileHandler) GetOrganizerInfo(c fiber.Ctx) error {
	organizerType := getQueryString(c, "type", "")
	organizerID := getQueryString(c, "id", "")

	if organizerType == "" {
		return response.BadRequest(c, "Organizer type is required (personal or institution)", nil)
	}
	if organizerID == "" {
		return response.BadRequest(c, "Organizer ID is required", nil)
	}

	if organizerType != "personal" && organizerType != "institution" {
		return response.BadRequest(c, "Invalid organizer type. Must be 'personal' or 'institution'", nil)
	}

	currentUserID := getUserIDOptional(c)
	ctx := context.WithValue(c.Context(), "user_id", currentUserID)

	organizer, err := h.svc.GetOrganizerInfo(ctx, organizerType, organizerID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Organizer not found", nil)
		}
		if errors.Is(err, domain.ErrInvalidOrganizerType) {
			return response.BadRequest(c, "Invalid organizer type", nil)
		}
		return response.InternalError(c, "Failed to get organizer info", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Organizer info retrieved successfully", NewOrganizerInfoResponse(organizer))
}

// ============================================================
// MEDIA UPLOAD HANDLERS
// ============================================================

// UploadUserAvatar godoc
// @Summary Upload user avatar
// @Description Upload an avatar for the authenticated user
// @Tags Profile
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Avatar image (JPEG, PNG, GIF, WEBP, SVG)"
// @Success 200 {object} response.BaseResponse{data=UserProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/avatar [post]
func (h *ProfileHandler) UploadUserAvatar(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	// Get file from form
	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		return response.BadRequest(c, "Avatar file is required", nil)
	}

	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		return response.InternalError(c, "Failed to open file", nil)
	}
	defer file.Close()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return response.InternalError(c, "Failed to read file", nil)
	}

	// Detect MIME type from file content
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "application/octet-stream" || contentType == "" {
		contentType = detectImageMimeType(fileContent, fileHeader.Filename)
	}

	ctx := context.WithValue(c.Context(), "user_id", userID)

	profile, err := h.svc.UploadUserAvatar(ctx, userID, fileContent, fileHeader.Filename, contentType)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return response.NotFound(c, "User not found", nil)
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to upload avatar", nil)
		}
		return response.InternalError(c, "Failed to upload avatar", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Avatar uploaded successfully", NewUserProfileResponse(profile))
}

// UploadAccountLogo godoc
// @Summary Upload account logo
// @Description Upload a logo for an account (admin only)
// @Tags Profile
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Param logo formData file true "Logo image (JPEG, PNG, GIF, WEBP, SVG)"
// @Success 200 {object} response.BaseResponse{data=AccountProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/logo [post]
func (h *ProfileHandler) UploadAccountLogo(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	accountID := c.Params("accountId")
	if accountID == "" {
		accountID = c.Params("id")
	}
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	// Get file from form
	fileHeader, err := c.FormFile("logo")
	if err != nil {
		return response.BadRequest(c, "Logo file is required", nil)
	}

	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		return response.InternalError(c, "Failed to open file", nil)
	}
	defer file.Close()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return response.InternalError(c, "Failed to read file", nil)
	}

	// Detect MIME type from file content
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "application/octet-stream" || contentType == "" {
		contentType = detectImageMimeType(fileContent, fileHeader.Filename)
	}

	ctx := context.WithValue(c.Context(), "user_id", userID)

	profile, err := h.svc.UploadAccountLogo(ctx, accountID, fileContent, fileHeader.Filename, contentType)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to upload account logo", nil)
		}
		return response.InternalError(c, "Failed to upload logo", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Logo uploaded successfully", NewAccountProfileResponse(profile))
}

// DeleteUserAvatar godoc
// @Summary Delete user avatar
// @Description Delete the authenticated user's avatar
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/users/me/avatar [delete]
func (h *ProfileHandler) DeleteUserAvatar(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	ctx := context.WithValue(c.Context(), "user_id", userID)

	if err := h.svc.DeleteUserAvatar(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return response.NotFound(c, "User not found", nil)
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to delete avatar", nil)
		}
		return response.InternalError(c, "Failed to delete avatar", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Avatar deleted successfully", nil)
}

// DeleteAccountLogo godoc
// @Summary Delete account logo
// @Description Delete an account's logo (admin only)
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Param accountId path string true "Account ID"
// @Success 200 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/accounts/{accountId}/logo [delete]
func (h *ProfileHandler) DeleteAccountLogo(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	accountID := c.Params("accountId")
	if accountID == "" {
		accountID = c.Params("id")
	}
	if accountID == "" {
		return response.BadRequest(c, "Account ID is required", nil)
	}

	ctx := context.WithValue(c.Context(), "user_id", userID)

	if err := h.svc.DeleteAccountLogo(ctx, accountID); err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return response.NotFound(c, "Account not found", nil)
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to delete account logo", nil)
		}
		return response.InternalError(c, "Failed to delete logo", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Logo deleted successfully", nil)
}