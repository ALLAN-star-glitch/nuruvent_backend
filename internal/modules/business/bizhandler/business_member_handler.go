package bizhandler

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/business/bizservice"
	"github.com/ALLAN_star_glitch/nuruvent-backend/pkg/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type MemberHandler struct {
	service *bizservice.MemberService
}

func NewMemberHandler(service *bizservice.MemberService) *MemberHandler {
	return &MemberHandler{service: service}
}

// ================================================
// REQUEST MODELS
// ================================================

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role   string `json:"role" binding:"required" enum:"host,event_manager,member" example:"event_manager"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" binding:"required" enum:"host,event_manager,member" example:"host"`
}

// ================================================
// HANDLERS
// ================================================

// GetBusinessMembers godoc
// @Summary Get business members
// @Description Get all members of a business
// @Tags Business Members
// @Produce json
// @Security BearerAuth
// @Param id path string true "Business ID"
// @Success 200 {object} response.BaseResponse{data=[]models.BusinessMember}
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{id}/members [get]
func (h *MemberHandler) GetBusinessMembers(c fiber.Ctx) error {
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

	members, err := h.service.GetBusinessMembers(c.Context(), uid, bizUID)
	if err != nil {
		if err.Error() == "insufficient permissions to view members" {
			return response.Forbidden(c, "You don't have permission to view members of this business", nil)
		}
		return response.InternalError(c, "Failed to get members", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Members retrieved successfully", members)
}

// AddMember godoc
// @Summary Add a member to a business
// @Description Add a user as a member of a business (requires host or admin role)
// @Tags Business Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Business ID"
// @Param request body AddMemberRequest true "Member details"
// @Success 201 {object} response.BaseResponse{data=models.BusinessMember}
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{id}/members [post]
func (h *MemberHandler) AddMember(c fiber.Ctx) error {
	adminID := c.Locals("user_id")
	if adminID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	adminIDStr, ok := adminID.(string)
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

	var req AddMemberRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	if req.UserID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}
	if req.Role == "" {
		return response.BadRequest(c, "Role is required", nil)
	}

	userUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return response.BadRequest(c, "Invalid user ID", nil)
	}

	adminUID, err := uuid.Parse(adminIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	member, err := h.service.AddMember(c.Context(), adminUID, bizUID, userUID, req.Role)
	if err != nil {
		if err.Error() == "insufficient permissions to add members" {
			return response.Forbidden(c, "You don't have permission to add members to this business", nil)
		}
		if err.Error() == "user is already a member of this business" {
			return response.BadRequest(c, "User is already a member of this business", nil)
		}
		return response.InternalError(c, "Failed to add member", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Member added successfully", member)
}

// RemoveMember godoc
// @Summary Remove a member from a business
// @Description Remove a member from a business (requires host or admin role)
// @Tags Business Members
// @Produce json
// @Security BearerAuth
// @Param id path string true "Business ID"
// @Param memberId path string true "Member User ID"
// @Success 200 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{id}/members/{memberId} [delete]
func (h *MemberHandler) RemoveMember(c fiber.Ctx) error {
	adminID := c.Locals("user_id")
	if adminID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	adminIDStr, ok := adminID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	businessID := c.Params("id")
	if businessID == "" {
		return response.BadRequest(c, "Business ID is required", nil)
	}

	memberID := c.Params("memberId")
	if memberID == "" {
		return response.BadRequest(c, "Member ID is required", nil)
	}

	bizUID, err := uuid.Parse(businessID)
	if err != nil {
		return response.BadRequest(c, "Invalid business ID", nil)
	}

	memberUID, err := uuid.Parse(memberID)
	if err != nil {
		return response.BadRequest(c, "Invalid member ID", nil)
	}

	adminUID, err := uuid.Parse(adminIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	err = h.service.RemoveMember(c.Context(), adminUID, bizUID, memberUID)
	if err != nil {
		if err.Error() == "insufficient permissions to remove members" {
			return response.Forbidden(c, "You don't have permission to remove members from this business", nil)
		}
		if err.Error() == "member not found" {
			return response.NotFound(c, "Member not found", nil)
		}
		return response.InternalError(c, "Failed to remove member", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Member removed successfully", nil)
}

// UpdateMemberRole godoc
// @Summary Update a member's role
// @Description Update a member's role in a business (requires host or admin role)
// @Tags Business Members
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Business ID"
// @Param memberId path string true "Member User ID"
// @Param request body UpdateMemberRoleRequest true "New role"
// @Success 200 {object} response.BaseResponse
// @Failure 400 {object} response.BaseResponse
// @Failure 401 {object} response.BaseResponse
// @Failure 403 {object} response.BaseResponse
// @Failure 404 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{id}/members/{memberId}/role [put]
func (h *MemberHandler) UpdateMemberRole(c fiber.Ctx) error {
	adminID := c.Locals("user_id")
	if adminID == nil {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	adminIDStr, ok := adminID.(string)
	if !ok {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	businessID := c.Params("id")
	if businessID == "" {
		return response.BadRequest(c, "Business ID is required", nil)
	}

	memberID := c.Params("memberId")
	if memberID == "" {
		return response.BadRequest(c, "Member ID is required", nil)
	}

	var req UpdateMemberRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request", fiber.Map{
			"error": err.Error(),
		})
	}

	if req.Role == "" {
		return response.BadRequest(c, "Role is required", nil)
	}

	bizUID, err := uuid.Parse(businessID)
	if err != nil {
		return response.BadRequest(c, "Invalid business ID", nil)
	}

	memberUID, err := uuid.Parse(memberID)
	if err != nil {
		return response.BadRequest(c, "Invalid member ID", nil)
	}

	adminUID, err := uuid.Parse(adminIDStr)
	if err != nil {
		return response.Unauthorized(c, "Invalid user ID", nil)
	}

	err = h.service.UpdateMemberRole(c.Context(), adminUID, bizUID, memberUID, req.Role)
	if err != nil {
		if err.Error() == "insufficient permissions to update member role" {
			return response.Forbidden(c, "You don't have permission to update member roles in this business", nil)
		}
		if err.Error() == "member not found" {
			return response.NotFound(c, "Member not found", nil)
		}
		return response.InternalError(c, "Failed to update member role", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Member role updated successfully", nil)
}

// CheckMembership godoc
// @Summary Check membership
// @Description Check if the current user is a member of a business
// @Tags Business Members
// @Produce json
// @Security BearerAuth
// @Param id path string true "Business ID"
// @Success 200 {object} response.BaseResponse{data=map[string]interface{}}
// @Failure 401 {object} response.BaseResponse
// @Failure 500 {object} response.BaseResponse
// @Router /api/v1/businesses/{id}/members/check [get]
func (h *MemberHandler) CheckMembership(c fiber.Ctx) error {
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

	isMember, role, err := h.service.CheckMembership(c.Context(), uid, bizUID)
	if err != nil {
		return response.InternalError(c, "Failed to check membership", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Membership checked successfully", fiber.Map{
		"is_member": isMember,
		"role":      role,
	})
}