package bizhandler

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/business/bizservice"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type BusinessHandler struct {
	service *bizservice.BusinessService
}

func NewBusinessHandler(service *bizservice.BusinessService) *BusinessHandler {
	return &BusinessHandler{service: service}
}

// ================================================
// REQUEST MODELS
// ================================================

type CreateBusinessRequest struct {
	Name            string `json:"name" binding:"required" example:"Nuruvent Training Institute"`
	BusinessTypeID  string `json:"business_type_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email           string `json:"email" example:"info@nuruvent.com"`
	Phone           string `json:"phone" example:"+254700000000"`
	Address         string `json:"address" example:"Nairobi, Kenya"`
	Logo            string `json:"logo" example:"https://example.com/logo.png"`
	Website         string `json:"website" example:"https://nuruvent.com"`
	Description     string `json:"description" example:"Leading training provider in Kenya"`
}

type UpdateBusinessRequest struct {
	Name            string `json:"name" example:"Nuruvent Training Institute"`
	BusinessTypeID  string `json:"business_type_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email           string `json:"email" example:"info@nuruvent.com"`
	Phone           string `json:"phone" example:"+254700000000"`
	Address         string `json:"address" example:"Nairobi, Kenya"`
	Logo            string `json:"logo" example:"https://example.com/logo.png"`
	Website         string `json:"website" example:"https://nuruvent.com"`
	Description     string `json:"description" example:"Leading training provider in Kenya"`
}

type SearchBusinessRequest struct {
	Query    string `query:"q" example:"training"`
	Page     int    `query:"page" default:"1" example:"1"`
	PageSize int    `query:"page_size" default:"20" example:"20"`
}

// ================================================
// PUBLIC HANDLERS
// ================================================

// GetBusiness godoc
// @Summary Get business by ID
// @Description Get business details by ID (public)
// @Tags Businesses
// @Produce json
// @Param id path string true "Business ID"
// @Success 200 {object} response.BaseResponse{data=models.Business}
// @Failure 400 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{id} [get]
func (h *BusinessHandler) GetBusiness(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.BadRequest(c, "Business ID is required", nil)
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return response.BadRequest(c, "Invalid business ID", nil)
	}

	business, err := h.service.GetBusinessByIDPublic(c.Context(), uid)
	if err != nil {
		return response.InternalError(c, "Failed to get business", fiber.Map{
			"error": err.Error(),
		})
	}

	if business == nil {
		return response.NotFound(c, "Business not found", nil)
	}

	return response.Success(c, "Business retrieved successfully", business)
}

// GetBusinessTypes godoc
// @Summary Get all business types
// @Description Get list of all business types
// @Tags Businesses
// @Produce json
// @Success 200 {object} response.BaseResponse{data=[]models.BusinessType}
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/types [get]
func (h *BusinessHandler) GetBusinessTypes(c fiber.Ctx) error {
	types, err := h.service.GetBusinessTypes(c.Context())
	if err != nil {
		return response.InternalError(c, "Failed to get business types", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Business types retrieved successfully", types)
}

// SearchBusinesses godoc
// @Summary Search businesses
// @Description Search businesses by name or description
// @Tags Businesses
// @Produce json
// @Param q query string false "Search query" example:"training"
// @Param page query int false "Page number" default(1) example:"1"
// @Param page_size query int false "Page size" default(20) example:"20"
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/search [get]
func (h *BusinessHandler) SearchBusinesses(c fiber.Ctx) error {
	var req SearchBusinessRequest
	if err := c.Bind().Query(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	businesses, total, err := h.service.SearchBusinesses(c.Context(), req.Query, req.Page, req.PageSize)
	if err != nil {
		return response.InternalError(c, "Failed to search businesses", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Businesses retrieved successfully", fiber.Map{
		"data":        businesses,
		"total":       total,
		"page":        req.Page,
		"page_size":   req.PageSize,
		"total_pages": (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	})
}

// ================================================
// PROTECTED HANDLERS
// ================================================

// CreateBusiness godoc
// @Summary Create a new business
// @Description Create a new business (requires authentication)
// @Tags Businesses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateBusinessRequest true "Business details"
// @Success 201 {object} response.BaseResponse{data=models.Business}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses [post]
func (h *BusinessHandler) CreateBusiness(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	var req CreateBusinessRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	if req.Name == "" {
		return response.BadRequest(c, "Business name is required", nil)
	}
	if req.BusinessTypeID == "" {
		return response.BadRequest(c, "Business type ID is required", nil)
	}

	businessTypeID, err := uuid.Parse(req.BusinessTypeID)
	if err != nil {
		return response.BadRequest(c, "Invalid business type ID", nil)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	business := &models.Business{
		Name:           req.Name,
		BusinessTypeID: businessTypeID,
		Email:          req.Email,
		Phone:          req.Phone,
		Address:        req.Address,
		Logo:           req.Logo,
		Website:        req.Website,
		Description:    req.Description,
		IsActive:       true,
	}

	created, err := h.service.CreateBusinessWithAdmin(c.Context(), uid, business)
	if err != nil {
		return response.InternalError(c, "Failed to create business", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Business created successfully", created)
}

// UpdateBusiness godoc
// @Summary Update a business
// @Description Update business details (requires business_admin or admin role)
// @Tags Businesses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Business ID"
// @Param request body UpdateBusinessRequest true "Business update details"
// @Success 200 {object} response.BaseResponse{data=models.Business}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{id} [put]
func (h *BusinessHandler) UpdateBusiness(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	businessID := c.Params("id")
	if businessID == "" {
		return response.BadRequest(c, "Business ID is required", nil)
	}

	bizUID, err := uuid.Parse(businessID)
	if err != nil {
		return response.BadRequest(c, "Invalid business ID", nil)
	}

	var req UpdateBusinessRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Logo != "" {
		updates["logo"] = req.Logo
	}
	if req.Website != "" {
		updates["website"] = req.Website
	}
	if req.BusinessTypeID != "" {
		updates["business_type_id"] = req.BusinessTypeID
	}

	if len(updates) == 0 {
		return response.BadRequest(c, "No fields to update", nil)
	}

	updated, err := h.service.UpdateBusiness(c.Context(), uid, bizUID, updates)
	if err != nil {
		if err.Error() == "insufficient permissions to update business" {
			return response.Forbidden(c, "You don't have permission to update this business", nil)
		}
		if err.Error() == "business not found" {
			return response.NotFound(c, "Business not found", nil)
		}
		return response.InternalError(c, "Failed to update business", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Business updated successfully", updated)
}

// DeleteBusiness godoc
// @Summary Delete a business
// @Description Delete a business (requires business_admin or admin role)
// @Tags Businesses
// @Produce json
// @Security BearerAuth
// @Param id path string true "Business ID"
// @Success 200 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{id} [delete]
func (h *BusinessHandler) DeleteBusiness(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	businessID := c.Params("id")
	if businessID == "" {
		return response.BadRequest(c, "Business ID is required", nil)
	}

	bizUID, err := uuid.Parse(businessID)
	if err != nil {
		return response.BadRequest(c, "Invalid business ID", nil)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	err = h.service.DeleteBusiness(c.Context(), uid, bizUID)
	if err != nil {
		if err.Error() == "insufficient permissions to delete business" {
			return response.Forbidden(c, "You don't have permission to delete this business", nil)
		}
		if err.Error() == "business not found" {
			return response.NotFound(c, "Business not found", nil)
		}
		return response.InternalError(c, "Failed to delete business", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Business deleted successfully", nil)
}

// GetMyBusiness godoc
// @Summary Get my primary business
// @Description Get the primary business of the authenticated user
// @Tags Businesses
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse{data=[]models.Business}
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/me [get]
func (h *BusinessHandler) GetMyBusiness(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	businesses, err := h.service.GetMyBusinesses(c.Context(), uid)
	if err != nil {
		return response.InternalError(c, "Failed to get businesses", fiber.Map{
			"error": err.Error(),
		})
	}

	if len(businesses) == 0 {
		return response.NotFound(c, "No businesses found for user", nil)
	}

	// Return the first business (or you can implement logic to get the primary one)
	return response.Success(c, "Business retrieved successfully", businesses[0])
}

// GetMyBusinesses godoc
// @Summary Get all my businesses
// @Description Get all businesses the authenticated user belongs to
// @Tags Businesses
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse{data=[]models.Business}
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/my [get]
func (h *BusinessHandler) GetMyBusinesses(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	businesses, err := h.service.GetMyBusinesses(c.Context(), uid)
	if err != nil {
		return response.InternalError(c, "Failed to get businesses", fiber.Map{
			"error": err.Error(),
		})
	}

	if len(businesses) == 0 {
		return response.Success(c, "No businesses found", []interface{}{})
	}

	return response.Success(c, "Businesses retrieved successfully", businesses)
}

// GetBusinessStats godoc
// @Summary Get business statistics
// @Description Get statistics for a business (events, attendees, revenue, members)
// @Tags Businesses
// @Produce json
// @Security BearerAuth
// @Param id path string true "Business ID"
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{id}/stats [get]
func (h *BusinessHandler) GetBusinessStats(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	businessID := c.Params("id")
	if businessID == "" {
		return response.BadRequest(c, "Business ID is required", nil)
	}

	bizUID, err := uuid.Parse(businessID)
	if err != nil {
		return response.BadRequest(c, "Invalid business ID", nil)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	stats, err := h.service.GetBusinessStats(c.Context(), uid, bizUID)
	if err != nil {
		if err.Error() == "insufficient permissions to view business stats" {
			return response.Forbidden(c, "You don't have permission to view this business stats", nil)
		}
		return response.InternalError(c, "Failed to get business stats", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Business stats retrieved successfully", stats)
}