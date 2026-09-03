// internal/modules/profile/delivery/handler/handler.go

package handler

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

// ============================================================
// PROFILE HANDLER
// ============================================================

type ProfileHandler struct {
	svc domain.Service
}

func NewProfileHandler(svc domain.Service) *ProfileHandler {
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

// parseScope parses a scope string into a domain.Scope
func parseScope(scopeStr string) (domain.Scope, error) {
	parts := strings.SplitN(scopeStr, ":", 2)
	if len(parts) != 2 {
		return domain.Scope{}, errors.New("invalid scope format. Use 'personal:user_id' or 'institution:institution_id'")
	}

	switch parts[0] {
	case "personal":
		return domain.NewPersonalTeamScope(parts[1]), nil
	case "institution":
		return domain.NewInstitutionTeamScope(parts[1]), nil
	default:
		return domain.Scope{}, errors.New("invalid scope type. Must be 'personal' or 'institution'")
	}
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
// @Param team_id query string false "Team ID (user_id or institution_id)"
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
// INSTITUTION PROFILE HANDLERS
// ============================================================

// GetInstitutionProfile godoc
// @Summary Get institution profile
// @Description Get an institution's profile (public - basic info only)
// @Tags Profile
// @Produce json
// @Param id path string true "Institution ID"
// @Success 200 {object} response.BaseResponse{data=InstitutionProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/profile/institutions/{id} [get]
func (h *ProfileHandler) GetInstitutionProfile(c fiber.Ctx) error {
	institutionID := c.Params("id")
	if institutionID == "" {
		institutionID = c.Params("institutionId")
	}
	if institutionID == "" {
		return response.BadRequest(c, "Institution ID is required", nil)
	}

	viewerID := getUserIDOptional(c)
	ctx := context.WithValue(c.Context(), "user_id", viewerID)

	profile, err := h.svc.GetInstitutionProfile(ctx, institutionID)
	if err != nil {
		if errors.Is(err, domain.ErrInstitutionNotFound) {
			return response.NotFound(c, "Institution profile not found", nil)
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to view this institution profile", nil)
		}
		return response.InternalError(c, "Failed to get institution profile", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Institution profile retrieved successfully", NewInstitutionProfileResponse(profile))
}

// GetInstitutionProfiles godoc
// @Summary Get multiple institution profiles
// @Description Get profiles for multiple institutions (basic info only)
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Param ids query string true "Comma-separated institution IDs"
// @Success 200 {object} response.BaseResponse{data=[]InstitutionProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/profile/institutions [get]
func (h *ProfileHandler) GetInstitutionProfiles(c fiber.Ctx) error {
	idsParam := getQueryString(c, "ids", "")
	if idsParam == "" {
		return response.BadRequest(c, "Institution IDs are required", nil)
	}

	ids := splitIDs(idsParam)
	if len(ids) == 0 {
		return response.BadRequest(c, "At least one institution ID is required", nil)
	}

	currentUserID := getUserIDOptional(c)
	ctx := context.WithValue(c.Context(), "user_id", currentUserID)

	profiles, err := h.svc.GetInstitutionProfiles(ctx, ids)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to view these institution profiles", nil)
		}
		return response.InternalError(c, "Failed to get institution profiles", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]InstitutionProfileResponse, len(profiles))
	for i, profile := range profiles {
		responses[i] = NewInstitutionProfileResponse(profile)
	}

	return response.Success(c, "Institution profiles retrieved successfully", responses)
}

// ListInstitutions godoc
// @Summary List institutions
// @Description List institutions with filters (requires auth)
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Param team_id query string false "Team ID (institution_id)"
// @Param team_type query string false "Team Type (institution)"
// @Param institution_id query string false "Filter by institution ID"
// @Param search query string false "Search by name or email"
// @Param include_deleted query bool false "Include soft-deleted institutions"
// @Param only_deleted query bool false "Show ONLY soft-deleted institutions"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/profile/institutions/list [get]
func (h *ProfileHandler) ListInstitutions(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req ListInstitutionsRequest
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
	filters := domain.ListInstitutionsFilters{
		Team:           team,
		InstitutionID:  req.InstitutionID,
		Search:         req.Search,
		IncludeDeleted: req.IncludeDeleted,
		OnlyDeleted:    req.OnlyDeleted,
		Limit:          req.Limit,
		Offset:         req.Offset,
		SortBy:         req.SortBy,
		SortOrder:      req.SortOrder,
	}

	ctx := context.WithValue(c.Context(), "user_id", userID)

	institutions, total, err := h.svc.ListInstitutions(ctx, filters)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to list institutions", nil)
		}
		return response.InternalError(c, "Failed to list institutions", fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]InstitutionProfileResponse, len(institutions))
	for i, institution := range institutions {
		responses[i] = NewInstitutionProfileResponse(institution)
	}

	return response.Success(c, "Institutions retrieved successfully", fiber.Map{
		"data":        responses,
		"total":       total,
		"limit":       req.Limit,
		"offset":      req.Offset,
		"sort_by":     req.SortBy,
		"sort_order":  req.SortOrder,
	})
}

// UpdateInstitutionProfile godoc
// @Summary Update institution profile
// @Description Update an institution's profile (requires admin access)
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param institutionId path string true "Institution ID"
// @Param request body UpdateInstitutionRequest true "Institution update data"
// @Success 200 {object} response.BaseResponse{data=InstitutionProfileResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/institutions/{institutionId}/profile [put]
func (h *ProfileHandler) UpdateInstitutionProfile(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	institutionID := c.Params("institutionId")
	if institutionID == "" {
		institutionID = c.Params("id")
	}
	if institutionID == "" {
		return response.BadRequest(c, "Institution ID is required", nil)
	}

	var req UpdateInstitutionRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	ctx := context.WithValue(c.Context(), "user_id", userID)

	updates := req.ToMap()

	profile, err := h.svc.UpdateInstitutionProfile(ctx, institutionID, updates)
	if err != nil {
		if errors.Is(err, domain.ErrInstitutionNotFound) {
			return response.NotFound(c, "Institution profile not found", nil)
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			return response.Forbidden(c, "You don't have permission to update this institution profile", nil)
		}
		return response.InternalError(c, "Failed to update institution profile", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Institution profile updated successfully", NewInstitutionProfileResponse(profile))
}

// ============================================================
// ORGANIZER INFO HANDLER
// ============================================================

// GetOrganizerInfo godoc
// @Summary Get organizer info for events
// @Description Get public-facing organizer info based on scope
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Param scope query string true "Scope (personal:user_id or institution:institution_id)"
// @Success 200 {object} response.BaseResponse{data=OrganizerInfoResponse}
// @Failure 400 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/profile/organizer [get]
func (h *ProfileHandler) GetOrganizerInfo(c fiber.Ctx) error {
	scopeParam := getQueryString(c, "scope", "")
	if scopeParam == "" {
		return response.BadRequest(c, "Scope is required (personal:user_id or institution:institution_id)", nil)
	}

	// Parse scope from string
	scope, err := parseScope(scopeParam)
	if err != nil {
		return response.BadRequest(c, "Invalid scope format. Use 'personal:user_id' or 'institution:institution_id'", nil)
	}

	currentUserID := getUserIDOptional(c)
	ctx := context.WithValue(c.Context(), "user_id", currentUserID)

	organizer, err := h.svc.GetOrganizerInfo(ctx, scope)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrInstitutionNotFound) {
			return response.NotFound(c, "Organizer not found", nil)
		}
		if errors.Is(err, domain.ErrInvalidScope) {
			return response.BadRequest(c, "Invalid scope", nil)
		}
		return response.InternalError(c, "Failed to get organizer info", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Organizer info retrieved successfully", NewOrganizerInfoResponse(organizer))
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

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