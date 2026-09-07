// internal/modules/team/delivery/handler/team_handler.go

package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/response"
)

type TeamHandler struct {
	service service.Service
}

func NewTeamHandler(service service.Service) *TeamHandler {
	return &TeamHandler{service: service}
}

// ============================================================
// PUBLIC HANDLERS
// ============================================================

// ValidateInvitation validates an invitation token
func (h *TeamHandler) ValidateInvitation(c fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return response.BadRequest(c, "token is required", nil)
	}

	invitation, err := h.service.ValidateInvitationToken(c.Context(), token)
	if err != nil {
		if err == teamdomain.ErrInvitationNotFound {
			return response.NotFound(c, "Invitation not found", nil)
		}
		if err == teamdomain.ErrInvitationExpired {
			return response.BadRequest(c, "Invitation has expired", nil)
		}
		if err == teamdomain.ErrInvitationAlreadyAccepted {
			return response.BadRequest(c, "Invitation already accepted", nil)
		}
		return response.InternalError(c, "Failed to validate invitation", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Invitation validated successfully", fiber.Map{
		"valid":         true,
		"email":         invitation.Email,
		"role":          string(invitation.Role),
		"expires_at":    invitation.ExpiresAt,
		"invitation_id": invitation.ID,
	})
}


// ============================================================
// PROTECTED HANDLERS - TEAM OPERATIONS
// ============================================================

// GetUserTeams returns all teams for the authenticated user
func (h *TeamHandler) GetUserTeams(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	teams, err := h.service.GetUserTeams(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, "Failed to get user teams", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Teams retrieved successfully", fiber.Map{
		"teams": teams,
		"count": len(teams),
	})
}

// GetTeam returns a team by ID
func (h *TeamHandler) GetTeam(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	team, err := h.service.GetTeamByID(c.Context(), teamID)
	if err != nil {
		if err == teamdomain.ErrTeamNotFound {
			return response.NotFound(c, "Team not found", nil)
		}
		return response.InternalError(c, "Failed to get team", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Team retrieved successfully", fiber.Map{
		"team": team,
	})
}

// CreatePersonalTeam creates a personal team for the authenticated user
func (h *TeamHandler) CreatePersonalTeam(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	userName := c.Locals("user_name").(string)
	if userName == "" {
		userName = req.Name
	}

	team, err := h.service.CreatePersonalTeam(c.Context(), userID, userName)
	if err != nil {
		if err == teamdomain.ErrTeamAlreadyExists {
			return response.Conflict(c, "Personal team already exists", nil)
		}
		return response.InternalError(c, "Failed to create personal team", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Personal team created successfully", fiber.Map{
		"team": team,
	})
}

// UpdateTeam updates a team
func (h *TeamHandler) UpdateTeam(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	var updates map[string]interface{}
	if err := c.Bind().Body(&updates); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	team, err := h.service.UpdateTeam(c.Context(), teamID, updates)
	if err != nil {
		if err == teamdomain.ErrTeamNotFound {
			return response.NotFound(c, "Team not found", nil)
		}
		if err == teamdomain.ErrPermissionDenied {
			return response.Forbidden(c, "Permission denied", nil)
		}
		return response.InternalError(c, "Failed to update team", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Team updated successfully", fiber.Map{
		"team": team,
	})
}

// DeleteTeam deletes a team
func (h *TeamHandler) DeleteTeam(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	err := h.service.DeleteTeam(c.Context(), teamID)
	if err != nil {
		if err == teamdomain.ErrTeamNotFound {
			return response.NotFound(c, "Team not found", nil)
		}
		if err == teamdomain.ErrPermissionDenied {
			return response.Forbidden(c, "Permission denied", nil)
		}
		return response.InternalError(c, "Failed to delete team", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Team deleted successfully", nil)
}

// ============================================================
// PROTECTED HANDLERS - MEMBER OPERATIONS
// ============================================================

// GetTeamMembers returns all members of a team
func (h *TeamHandler) GetTeamMembers(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	filters := teamdomain.ListMembersFilters{
		Limit:  limit,
		Offset: offset,
		Search: c.Query("search"),
		Role:   teamdomain.MemberRole(c.Query("role")),
	}

	members, total, err := h.service.GetTeamMembers(c.Context(), teamID, filters)
	if err != nil {
		if err == teamdomain.ErrTeamNotFound {
			return response.NotFound(c, "Team not found", nil)
		}
		return response.InternalError(c, "Failed to get team members", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Team members retrieved successfully", fiber.Map{
		"members": members,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// AddMember adds a member to a team
func (h *TeamHandler) AddMember(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if req.UserID == "" {
		return response.BadRequest(c, "user_id is required", nil)
	}
	if req.Role == "" {
		return response.BadRequest(c, "role is required", nil)
	}
	if !teamdomain.IsValidRole(req.Role) {
		return response.BadRequest(c, "Invalid role. Valid roles: account_admin, trainer", nil)
	}

	addedBy := c.Locals("user_id").(string)

	member, err := h.service.AddMember(c.Context(), teamID, req.UserID, teamdomain.MemberRole(req.Role), addedBy)
	if err != nil {
		if err == teamdomain.ErrTeamNotFound {
			return response.NotFound(c, "Team not found", nil)
		}
		if err == teamdomain.ErrMemberAlreadyExists {
			return response.Conflict(c, "User is already a member of this team", nil)
		}
		if err == teamdomain.ErrPermissionDenied {
			return response.Forbidden(c, "Permission denied", nil)
		}
		return response.InternalError(c, "Failed to add member", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Member added successfully", fiber.Map{
		"member": member,
	})
}

// UpdateMemberRole updates a member's role
func (h *TeamHandler) UpdateMemberRole(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	userID := c.Params("userId")
	if userID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if req.Role == "" {
		return response.BadRequest(c, "role is required", nil)
	}
	if !teamdomain.IsValidRole(req.Role) {
		return response.BadRequest(c, "Invalid role. Valid roles: account_admin, trainer", nil)
	}

	updatedBy := c.Locals("user_id").(string)

	member, err := h.service.UpdateMemberRole(c.Context(), teamID, userID, teamdomain.MemberRole(req.Role), updatedBy)
	if err != nil {
		if err == teamdomain.ErrTeamNotFound || err == teamdomain.ErrMemberNotFound {
			return response.NotFound(c, "Member not found", nil)
		}
		if err == teamdomain.ErrPermissionDenied {
			return response.Forbidden(c, "Permission denied", nil)
		}
		if err == teamdomain.ErrCannotChangeOwnRole {
			return response.BadRequest(c, "Cannot change your own role", nil)
		}
		return response.InternalError(c, "Failed to update member role", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Member role updated successfully", fiber.Map{
		"member": member,
	})
}

// RemoveMember removes a member from a team
func (h *TeamHandler) RemoveMember(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	userID := c.Params("userId")
	if userID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	removedBy := c.Locals("user_id").(string)

	err := h.service.RemoveMember(c.Context(), teamID, userID, removedBy)
	if err != nil {
		if err == teamdomain.ErrTeamNotFound || err == teamdomain.ErrMemberNotFound {
			return response.NotFound(c, "Member not found", nil)
		}
		if err == teamdomain.ErrPermissionDenied {
			return response.Forbidden(c, "Permission denied", nil)
		}
		if err == teamdomain.ErrCannotRemoveSelf {
			return response.BadRequest(c, "Cannot remove yourself from the team", nil)
		}
		if err == teamdomain.ErrLastAdminCannotLeave {
			return response.BadRequest(c, "Cannot remove the last admin of the team", nil)
		}
		return response.InternalError(c, "Failed to remove member", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Member removed successfully", nil)
}

// LeaveTeam allows a user to leave a team
func (h *TeamHandler) LeaveTeam(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	userID := c.Locals("user_id").(string)

	err := h.service.LeaveTeam(c.Context(), teamID, userID)
	if err != nil {
		if err == teamdomain.ErrTeamNotFound || err == teamdomain.ErrMemberNotFound {
			return response.NotFound(c, "Team or member not found", nil)
		}
		if err == teamdomain.ErrLastAdminCannotLeave {
			return response.BadRequest(c, "Cannot leave as the last admin of the team", nil)
		}
		return response.InternalError(c, "Failed to leave team", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "You have left the team successfully", nil)
}

// ============================================================
// PROTECTED HANDLERS - INVITATION OPERATIONS
// ============================================================

// InviteMember invites a user to join a team
func (h *TeamHandler) InviteMember(c fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if req.Email == "" {
		return response.BadRequest(c, "email is required", nil)
	}
	if req.Role == "" {
		return response.BadRequest(c, "role is required", nil)
	}
	if !teamdomain.IsValidRole(req.Role) {
		return response.BadRequest(c, "Invalid role. Valid roles: account_admin, trainer", nil)
	}

	invitation, err := h.service.InviteMember(c.Context(), service.InviteMemberCommand{
		TeamID:    teamID,
		Email:     req.Email,
		Role:      req.Role,
		InvitedBy: userID,
	})
	if err != nil {
		if err == teamdomain.ErrTeamNotFound {
			return response.NotFound(c, "Team not found", nil)
		}
		if err == teamdomain.ErrMemberAlreadyExists {
			return response.Conflict(c, "User is already a member of this team", nil)
		}
		if err == teamdomain.ErrPermissionDenied {
			return response.Forbidden(c, "Permission denied", nil)
		}
		return response.InternalError(c, "Failed to invite member", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Invitation sent successfully", fiber.Map{
		"invitation": invitation,
	})
}

// GetTeamInvitations retrieves all invitations for a team
func (h *TeamHandler) GetTeamInvitations(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	filters := teamdomain.ListInvitationsFilters{
		Limit:  limit,
		Offset: offset,
		Email:  c.Query("email"),
		Status: teamdomain.InvitationStatus(c.Query("status")),
	}

	invitations, total, err := h.service.GetTeamInvitations(c.Context(), teamID, filters)
	if err != nil {
		if err == teamdomain.ErrTeamNotFound {
			return response.NotFound(c, "Team not found", nil)
		}
		return response.InternalError(c, "Failed to get team invitations", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Team invitations retrieved successfully", fiber.Map{
		"invitations": invitations,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// ResendInvitation resends an invitation
func (h *TeamHandler) ResendInvitation(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	var req struct {
		InvitationID string `json:"invitation_id"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if req.InvitationID == "" {
		return response.BadRequest(c, "invitation_id is required", nil)
	}

	invitation, err := h.service.ResendInvitation(c.Context(), req.InvitationID)
	if err != nil {
		if err == teamdomain.ErrInvitationNotFound {
			return response.NotFound(c, "Invitation not found", nil)
		}
		if err == teamdomain.ErrPermissionDenied {
			return response.Forbidden(c, "Permission denied", nil)
		}
		return response.InternalError(c, "Failed to resend invitation", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Invitation resent successfully", fiber.Map{
		"invitation": invitation,
	})
}