// internal/modules/team/delivery/handler/team_handler.go

package handler

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v3"

	authDomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
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
// CONTEXT HELPERS
// ============================================================

// authenticatedUserID returns the user ID from the Fiber request context,
// or "" if unauthenticated.
func authenticatedUserID(c fiber.Ctx) string {
	id, _ := c.Locals(authDomain.ContextKeyUserID).(string)
	return id
}

// requestContext returns a context.Context with the authenticated user ID
// set. Service methods that authorize the actor (UpdateTeam, DeleteTeam)
// read it from here.
//
// If the user is unauthenticated, the context is returned unchanged and
// the service will reject with ErrPermissionDenied.
func requestContext(c fiber.Ctx) context.Context {
	ctx := c.Context()
	if userID := authenticatedUserID(c); userID != "" {
		ctx = service.WithActor(ctx, userID)
	}
	return ctx
}



// ============================================================
// PUBLIC HANDLERS
// ============================================================

// ValidateInvitation validates an invitation token.
func (h *TeamHandler) ValidateInvitation(c fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return response.BadRequest(c, "token is required", nil)
	}

	invitation, err := h.service.ValidateInvitationToken(c.Context(), token)
	if err != nil {
		switch err {
		case teamdomain.ErrInvitationNotFound:
			return response.NotFound(c, "Invitation not found", nil)
		case teamdomain.ErrInvitationExpired:
			return response.BadRequest(c, "Invitation has expired", nil)
		case teamdomain.ErrInvitationAlreadyAccepted:
			return response.BadRequest(c, "Invitation already accepted", nil)
		}
		return response.InternalError(c, "Failed to validate invitation", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Invitation validated successfully", fiber.Map{
		"valid":         true,
		"email":         invitation.Email,
		"expires_at":    invitation.ExpiresAt,
		"invitation_id": invitation.ID,
	})
}

// ============================================================
// PROTECTED HANDLERS - TEAM OPERATIONS
// ============================================================

// GetUserTeams returns all teams for the authenticated user.
func (h *TeamHandler) GetUserTeams(c fiber.Ctx) error {
	userID := authenticatedUserID(c)
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

// GetTeam returns a team by ID.
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

// CreatePersonalTeam creates a personal team for the authenticated user.
//
// No permission check is required — a user always has access to their own
// personal account. The service accepts a `role` argument for historical
// reasons but ignores it.
func (h *TeamHandler) CreatePersonalTeam(c fiber.Ctx) error {
	userID := authenticatedUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	userName, _ := c.Locals(authDomain.ContextKeyUserName).(string)
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

// UpdateTeam updates a team.
//
// The authenticated user is passed to the service via requestContext so
// the service can authorize against the team's parent account domain.
func (h *TeamHandler) UpdateTeam(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	if authenticatedUserID(c) == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var updates map[string]interface{}
	if err := c.Bind().Body(&updates); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	team, err := h.service.UpdateTeam(requestContext(c), teamID, updates)
	if err != nil {
		switch err {
		case teamdomain.ErrTeamNotFound:
			return response.NotFound(c, "Team not found", nil)
		case teamdomain.ErrPermissionDenied:
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

// DeleteTeam deletes a team.
func (h *TeamHandler) DeleteTeam(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	if authenticatedUserID(c) == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	err := h.service.DeleteTeam(requestContext(c), teamID)
	if err != nil {
		switch err {
		case teamdomain.ErrTeamNotFound:
			return response.NotFound(c, "Team not found", nil)
		case teamdomain.ErrPermissionDenied:
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

// GetTeamMembers returns all members of a team.
//
// The authenticated user is passed as the viewer so the service can
// authorize the read against the team's parent account domain.
func (h *TeamHandler) GetTeamMembers(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	viewerUserID := authenticatedUserID(c)
	if viewerUserID == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	filters := teamdomain.ListMembersFilters{
		Limit:  limit,
		Offset: offset,
		Search: c.Query("search"),
	}

	members, total, err := h.service.GetTeamMembers(c.Context(), viewerUserID, teamID, filters)
	if err != nil {
		switch err {
		case teamdomain.ErrTeamNotFound:
			return response.NotFound(c, "Team not found", nil)
		case teamdomain.ErrPermissionDenied:
			return response.Forbidden(c, "Permission denied", nil)
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

// AddMember adds a member to a team.
func (h *TeamHandler) AddMember(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	if req.UserID == "" {
		return response.BadRequest(c, "user_id is required", nil)
	}

	addedBy := authenticatedUserID(c)
	if addedBy == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	member, err := h.service.AddMember(c.Context(), teamID, req.UserID, addedBy)
	if err != nil {
		switch err {
		case teamdomain.ErrTeamNotFound:
			return response.NotFound(c, "Team not found", nil)
		case teamdomain.ErrMemberAlreadyExists:
			return response.Conflict(c, "User is already a member of this team", nil)
		case teamdomain.ErrPermissionDenied:
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

// RemoveMember removes a member from a team.
func (h *TeamHandler) RemoveMember(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	userID := c.Params("userId")
	if userID == "" {
		return response.BadRequest(c, "User ID is required", nil)
	}

	removedBy := authenticatedUserID(c)
	if removedBy == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	err := h.service.RemoveMember(c.Context(), teamID, userID, removedBy)
	if err != nil {
		switch err {
		case teamdomain.ErrTeamNotFound, teamdomain.ErrMemberNotFound:
			return response.NotFound(c, "Member not found", nil)
		case teamdomain.ErrPermissionDenied:
			return response.Forbidden(c, "Permission denied", nil)
		case teamdomain.ErrCannotRemoveSelf:
			return response.BadRequest(c, "Cannot remove yourself from the team", nil)
		}
		return response.InternalError(c, "Failed to remove member", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Member removed successfully", nil)
}

// LeaveTeam allows a user to leave a team.
func (h *TeamHandler) LeaveTeam(c fiber.Ctx) error {
	teamID := c.Params("id")
	if teamID == "" {
		return response.BadRequest(c, "Team ID is required", nil)
	}

	userID := authenticatedUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	err := h.service.LeaveTeam(c.Context(), teamID, userID)
	if err != nil {
		switch err {
		case teamdomain.ErrTeamNotFound, teamdomain.ErrMemberNotFound:
			return response.NotFound(c, "Team or member not found", nil)
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

// InviteMember invites a user to join a team.
func (h *TeamHandler) InviteMember(c fiber.Ctx) error {
	userID := authenticatedUserID(c)
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
		return response.BadRequest(c, "Invalid request body", fiber.Map{
			"error": err.Error(),
		})
	}

	if req.Email == "" {
		return response.BadRequest(c, "email is required", nil)
	}
	if req.Role == "" {
		return response.BadRequest(c, "role is required", nil)
	}
	if req.Role != "account_admin" && req.Role != "trainer" {
		return response.BadRequest(c, "invalid role; must be account_admin or trainer", nil)
	}

	invitation, err := h.service.InviteMember(c.Context(), service.InviteMemberCommand{
		TeamID:    teamID,
		Email:     req.Email,
		Role:      req.Role,
		InvitedBy: userID,
	})
	if err != nil {
		switch err {
		case teamdomain.ErrTeamNotFound:
			return response.NotFound(c, "Team not found", nil)
		case teamdomain.ErrMemberAlreadyExists:
			return response.Conflict(c, "User is already a member of this team", nil)
		case teamdomain.ErrInvitationPending:
			return response.Conflict(c, "An invitation is already pending for this email", nil)
		case teamdomain.ErrPermissionDenied:
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

// GetTeamInvitations retrieves all invitations for a team.
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

// ResendInvitation resends an invitation.
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
		switch err {
		case teamdomain.ErrInvitationNotFound:
			return response.NotFound(c, "Invitation not found", nil)
		case teamdomain.ErrPermissionDenied:
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