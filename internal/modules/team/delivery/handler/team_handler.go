// internal/modules/team/delivery/handler/team_handler.go

package handler

import (
    "github.com/gofiber/fiber/v3"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/service"
    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

type TeamHandler struct {
    service service.Service
}

func NewTeamHandler(service service.Service) *TeamHandler {
    return &TeamHandler{service: service}
}

// RegisterRoutes registers all team routes
func (h *TeamHandler) RegisterRoutes(app fiber.Router) {
    // Public routes (no auth required)
    app.Get("/teams/invitations/validate", h.ValidateInvitation)
    app.Post("/teams/register", h.RegisterAndAcceptInvitation)

    // Protected routes (auth required)
    team := app.Group("/teams")
    team.Post("/invite", h.InviteMember)
    team.Get("/", h.GetUserTeams)
    team.Get("/:id", h.GetTeam)
    team.Patch("/:id", h.UpdateTeam)
    team.Get("/:id/members", h.GetTeamMembers)
    team.Patch("/:id/members/:userId/role", h.UpdateMemberRole)
    team.Delete("/:id/members/:userId", h.RemoveMember)
    team.Post("/:id/leave", h.LeaveTeam)
    team.Get("/:id/invitations", h.GetTeamInvitations)
    team.Post("/:id/invitations/resend", h.ResendInvitation)
}

// ============================================================
// HANDLER METHODS
// ============================================================

// InviteMember handles inviting a member to a team
func (h *TeamHandler) InviteMember(c fiber.Ctx) error {
    var req struct {
        Email string `json:"email"`
        Role  string `json:"role"`
    }
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    // Get team ID from path
    teamID := c.Params("id")
    if teamID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "team ID is required",
        })
    }

    // Get current user ID from context (from auth middleware)
    userID := c.Locals("user_id").(string)

    // Validate role
    if !teamdomain.IsValidRole(req.Role) {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid role. Valid roles: account_admin, event_manager, team_member",
        })
    }

    // Call service
    invitation, err := h.service.InviteMember(c.Context(), service.InviteMemberCommand{
        TeamID:    teamID,
        Email:     req.Email,
        Role:      teamdomain.MemberRole(req.Role),
        InvitedBy: userID,
    })
    if err != nil {
        status := fiber.StatusInternalServerError
        switch err {
        case teamdomain.ErrTeamNotFound:
            status = fiber.StatusNotFound
        case teamdomain.ErrMemberAlreadyExists:
            status = fiber.StatusConflict
        }
        return c.Status(status).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message":    "Invitation sent successfully",
        "invitation": invitation,
    })
}

// ValidateInvitation validates an invitation token
func (h *TeamHandler) ValidateInvitation(c fiber.Ctx) error {
    token := c.Query("token")
    if token == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "token is required",
        })
    }

    invitation, err := h.service.ValidateInvitationToken(c.Context(), token)
    if err != nil {
        status := fiber.StatusBadRequest
        switch err {
        case teamdomain.ErrInvitationNotFound:
            status = fiber.StatusNotFound
        case teamdomain.ErrInvitationExpired:
            status = fiber.StatusGone
        case teamdomain.ErrInvitationAlreadyAccepted:
            status = fiber.StatusConflict
        }
        return c.Status(status).JSON(fiber.Map{
            "valid":   false,
            "message": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "valid":           true,
        "email":           invitation.Email,
        "role":            string(invitation.Role),
        "expires_at":      invitation.ExpiresAt,
        "invitation_id":   invitation.ID,
    })
}

// RegisterAndAcceptInvitation handles registration and invitation acceptance
func (h *TeamHandler) RegisterAndAcceptInvitation(c fiber.Ctx) error {
    var req struct {
        Token    string `json:"token"`
        Name     string `json:"name"`
        Password string `json:"password"`
        Phone    string `json:"phone"`
    }
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    // Validate
    if req.Token == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "token is required",
        })
    }
    if req.Name == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "name is required",
        })
    }
    if req.Password == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "password is required",
        })
    }

    // TODO: This should create user via auth service and accept invitation
    // For now, we need to get user ID after creation

    return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
        "error": "Registration flow coming soon",
    })
}

// GetUserTeams retrieves all teams for the current user
func (h *TeamHandler) GetUserTeams(c fiber.Ctx) error {
    userID := c.Locals("user_id").(string)

    teams, err := h.service.GetUserTeams(c.Context(), userID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "teams": teams,
    })
}

// GetTeam retrieves a team by ID
func (h *TeamHandler) GetTeam(c fiber.Ctx) error {
    teamID := c.Params("id")
    if teamID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "team ID is required",
        })
    }

    team, err := h.service.GetTeamByID(c.Context(), teamID)
    if err != nil {
        status := fiber.StatusInternalServerError
        if err == teamdomain.ErrTeamNotFound {
            status = fiber.StatusNotFound
        }
        return c.Status(status).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "team": team,
    })
}

// UpdateTeam updates a team
func (h *TeamHandler) UpdateTeam(c fiber.Ctx) error {
    teamID := c.Params("id")
    if teamID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "team ID is required",
        })
    }

    var updates map[string]interface{}
    if err := c.BodyParser(&updates); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    team, err := h.service.UpdateTeam(c.Context(), teamID, updates)
    if err != nil {
        status := fiber.StatusInternalServerError
        if err == teamdomain.ErrTeamNotFound {
            status = fiber.StatusNotFound
        }
        return c.Status(status).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Team updated successfully",
        "team":    team,
    })
}

// GetTeamMembers retrieves all members of a team
func (h *TeamHandler) GetTeamMembers(c fiber.Ctx) error {
    teamID := c.Params("id")
    if teamID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "team ID is required",
        })
    }

    filters := teamdomain.ListMembersFilters{
        Limit:  c.QueryInt("limit", 20),
        Offset: c.QueryInt("offset", 0),
        Search: c.Query("search"),
        Role:   teamdomain.MemberRole(c.Query("role")),
    }

    members, total, err := h.service.GetTeamMembers(c.Context(), teamID, filters)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "members": members,
        "total":   total,
        "limit":   filters.Limit,
        "offset":  filters.Offset,
    })
}

// UpdateMemberRole updates a member's role
func (h *TeamHandler) UpdateMemberRole(c fiber.Ctx) error {
    teamID := c.Params("id")
    if teamID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "team ID is required",
        })
    }

    userID := c.Params("userId")
    if userID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "user ID is required",
        })
    }

    var req struct {
        Role string `json:"role"`
    }
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    if !teamdomain.IsValidRole(req.Role) {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid role. Valid roles: account_admin, event_manager, team_member",
        })
    }

    updatedBy := c.Locals("user_id").(string)

    member, err := h.service.UpdateMemberRole(c.Context(), teamID, userID, teamdomain.MemberRole(req.Role), updatedBy)
    if err != nil {
        status := fiber.StatusInternalServerError
        switch err {
        case teamdomain.ErrTeamNotFound, teamdomain.ErrMemberNotFound:
            status = fiber.StatusNotFound
        case teamdomain.ErrCannotChangeOwnRole:
            status = fiber.StatusForbidden
        }
        return c.Status(status).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Role updated successfully",
        "member":  member,
    })
}

// RemoveMember removes a member from a team
func (h *TeamHandler) RemoveMember(c fiber.Ctx) error {
    teamID := c.Params("id")
    if teamID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "team ID is required",
        })
    }

    userID := c.Params("userId")
    if userID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "user ID is required",
        })
    }

    removedBy := c.Locals("user_id").(string)

    err := h.service.RemoveMember(c.Context(), teamID, userID, removedBy)
    if err != nil {
        status := fiber.StatusInternalServerError
        switch err {
        case teamdomain.ErrTeamNotFound, teamdomain.ErrMemberNotFound:
            status = fiber.StatusNotFound
        case teamdomain.ErrCannotRemoveSelf:
            status = fiber.StatusForbidden
        case teamdomain.ErrLastAdminCannotLeave:
            status = fiber.StatusBadRequest
        }
        return c.Status(status).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Member removed successfully",
    })
}

// LeaveTeam allows a user to leave a team
func (h *TeamHandler) LeaveTeam(c fiber.Ctx) error {
    teamID := c.Params("id")
    if teamID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "team ID is required",
        })
    }

    userID := c.Locals("user_id").(string)

    err := h.service.LeaveTeam(c.Context(), teamID, userID)
    if err != nil {
        status := fiber.StatusInternalServerError
        switch err {
        case teamdomain.ErrTeamNotFound, teamdomain.ErrMemberNotFound:
            status = fiber.StatusNotFound
        case teamdomain.ErrLastAdminCannotLeave:
            status = fiber.StatusBadRequest
        }
        return c.Status(status).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "You have left the team",
    })
}

// GetTeamInvitations retrieves all invitations for a team
func (h *TeamHandler) GetTeamInvitations(c fiber.Ctx) error {
    teamID := c.Params("id")
    if teamID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "team ID is required",
        })
    }

    filters := teamdomain.ListInvitationsFilters{
        Limit:  c.QueryInt("limit", 20),
        Offset: c.QueryInt("offset", 0),
        Email:  c.Query("email"),
        Status: teamdomain.InvitationStatus(c.Query("status")),
    }

    invitations, total, err := h.service.GetTeamInvitations(c.Context(), teamID, filters)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "invitations": invitations,
        "total":       total,
        "limit":       filters.Limit,
        "offset":      filters.Offset,
    })
}

// ResendInvitation resends an invitation
func (h *TeamHandler) ResendInvitation(c fiber.Ctx) error {
    var req struct {
        InvitationID string `json:"invitation_id"`
    }
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    if req.InvitationID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "invitation_id is required",
        })
    }

    invitation, err := h.service.ResendInvitation(c.Context(), req.InvitationID)
    if err != nil {
        status := fiber.StatusInternalServerError
        if err == teamdomain.ErrInvitationNotFound {
            status = fiber.StatusNotFound
        }
        return c.Status(status).JSON(fiber.Map{
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message":    "Invitation resent successfully",
        "invitation": invitation,
    })
}