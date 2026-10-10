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
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/validation"
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

func authenticatedUserID(c fiber.Ctx) string {
	id, _ := c.Locals(authDomain.ContextKeyUserID).(string)
	return id
}

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
		"expires_at":    invitation.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
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

	return response.Success(c, "Teams retrieved successfully", NewTeamsListResponse(teams))
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
		"team": NewTeamResponse(team),
	})
}

// CreatePersonalTeam creates a personal team under the given account.
func (h *TeamHandler) CreatePersonalTeam(c fiber.Ctx) error {
	userID := authenticatedUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req struct {
		Name      string `json:"name"`
		AccountID string `json:"account_id"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}
	if req.AccountID == "" {
		return response.BadRequest(c, "account_id is required", nil)
	}

	sanitizer := validation.Sanitize{}
	displayName := sanitizer.DisplayName(req.Name)

	userName, _ := c.Locals(authDomain.ContextKeyUserName).(string)
	if userName == "" {
		userName = "User"
	}

	team, err := h.service.CreatePersonalTeam(
		c.Context(),
		userID,
		userName,
		req.AccountID,
		displayName,
	)
	if err != nil {
		switch err {
		case teamdomain.ErrPermissionDenied:
			return response.Forbidden(c, "You don't have access to create teams in this account", nil)
		}
		return response.InternalError(c, "Failed to create personal team", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Personal team created successfully", fiber.Map{
		"team": NewTeamResponse(team),
	})
}

// UpdateTeam updates a team.
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
		"team": NewTeamResponse(team),
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
		"members": NewMembersListResponse(members),
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

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
		"member": NewMemberResponse(member),
	})
}

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
	if !teamdomain.IsValidAccountRole(req.Role) {
		return response.BadRequest(c, "invalid role; must be account_admin, trainer, or learner", nil)
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
		case teamdomain.ErrInviteeAlreadyAccountMember:
			return response.Conflict(c, "This user is already a member of this account; update their role from the members list instead", nil)
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
		"invitation": NewInvitationResponse(invitation),
	})
}

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
		"invitations": NewInvitationsListResponse(invitations),
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

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
		"invitation": NewInvitationResponse(invitation),
	})
}

func (h *TeamHandler) AcceptInvitation(c fiber.Ctx) error {
	userID := authenticatedUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	token := c.Query("token")
	if token == "" {
		return response.BadRequest(c, "token is required", nil)
	}

	member, err := h.service.AcceptInvitation(c.Context(), token, userID)
	if err != nil {
		switch err {
		case teamdomain.ErrInvitationNotFound:
			return response.NotFound(c, "Invitation not found", nil)
		case teamdomain.ErrInvitationExpired:
			return response.BadRequest(c, "Invitation has expired", nil)
		case teamdomain.ErrInvitationAlreadyAccepted:
			return response.BadRequest(c, "Invitation already accepted", nil)
		case teamdomain.ErrInvitationEmailMismatch:
			return response.Forbidden(c, "This invitation is for a different email address", nil)
		case teamdomain.ErrTeamNotFound:
			return response.NotFound(c, "Team not found", nil)
		}
		return response.InternalError(c, "Failed to accept invitation", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Invitation accepted successfully", fiber.Map{
		"member": NewMemberResponse(member),
	})
}

func (h *TeamHandler) DeclineInvitation(c fiber.Ctx) error {
	userID := authenticatedUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	token := c.Query("token")
	if token == "" {
		return response.BadRequest(c, "token is required", nil)
	}

	if err := h.service.DeclineInvitation(c.Context(), token, userID); err != nil {
		switch err {
		case teamdomain.ErrInvitationNotFound:
			return response.NotFound(c, "Invitation not found", nil)
		case teamdomain.ErrInvitationExpired:
			return response.BadRequest(c, "Invitation has expired", nil)
		case teamdomain.ErrInvitationAlreadyAccepted:
			return response.BadRequest(c, "Invitation already accepted", nil)
		case teamdomain.ErrInvitationEmailMismatch:
			return response.Forbidden(c, "This invitation is for a different email address", nil)
		}
		return response.InternalError(c, "Failed to decline invitation", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Success(c, "Invitation declined successfully", nil)
}

func (h *TeamHandler) CreateInstitutionTeam(c fiber.Ctx) error {
	userID := authenticatedUserID(c)
	if userID == "" {
		return response.Unauthorized(c, "User not authenticated", nil)
	}

	var req struct {
		AccountID   string `json:"account_id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Slug        string `json:"slug"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return response.BadRequest(c, "Invalid request body", nil)
	}

	team, err := h.service.CreateInstitutionTeam(
		c.Context(),
		req.AccountID,
		req.Name,
		req.DisplayName,
		req.Slug,
		userID,
	)
	if err != nil {
		switch err {
		case teamdomain.ErrTeamNotFound:
			return response.NotFound(c, "Account not found", nil)
		case teamdomain.ErrPermissionDenied:
			return response.Forbidden(c, "Permission denied", nil)
		case teamdomain.ErrTeamAlreadyExists:
			return response.Conflict(c, "A team with this slug already exists", nil)
		case teamdomain.ErrTeamNameRequired:
			return response.BadRequest(c, "Team name is required", nil)
		case teamdomain.ErrTeamSlugRequired:
			return response.BadRequest(c, "Team slug is required", nil)
		}
		return response.InternalError(c, "Failed to create institution team", fiber.Map{
			"error": err.Error(),
		})
	}

	return response.Created(c, "Institution team created successfully", fiber.Map{
		"team": NewTeamResponse(team),
	})
}