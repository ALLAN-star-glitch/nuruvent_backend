// internal/modules/team/delivery/handler/routes.go

package handler

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers all team routes
func (h *TeamHandler) RegisterRoutes(
	router fiber.Router,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
) {
	// ============================================================
	// PUBLIC ROUTES (No auth required)
	// ============================================================
	public := router.Group("/teams")
	{
		public.Get("/invitations/validate", h.ValidateInvitation)
	}

	// ============================================================
	// PROTECTED ROUTES (Auth required)
	// ============================================================
	protected := router.Group("/teams")
	protected.Use(authMiddleware)
	{
		// ============================================================
		// INVITATION LIFECYCLE — invitee-side (accept/decline)
		//
		// These require only authMiddleware: the invitee has no
		// permissions on the team yet, so authzMiddleware has nothing to
		// check. The service validates the token, matches the invitation
		// email against the authenticated user, and writes the memberships.
		//
		// Registered before the /:id/... routes so the literal
		// "invitations" segment is never captured as a team ID.
		// ============================================================
		protected.Post("/invitations/accept", h.AcceptInvitation)
		protected.Post("/invitations/decline", h.DeclineInvitation)

		// ============================================================
		// TEAM OPERATIONS
		// ============================================================
		protected.Post("/", authzMiddleware, h.CreatePersonalTeam)
		protected.Get("/", authzMiddleware, h.GetUserTeams)
		protected.Get("/:id", authzMiddleware, h.GetTeam)
		protected.Patch("/:id", authzMiddleware, h.UpdateTeam)
		protected.Delete("/:id", authzMiddleware, h.DeleteTeam)

		// ============================================================
		// MEMBER OPERATIONS
		// ✅ REMOVED: UpdateMemberRole - roles are managed at account level
		// ============================================================
		protected.Get("/:id/members", authzMiddleware, h.GetTeamMembers)
		protected.Post("/:id/members", authzMiddleware, h.AddMember)
		protected.Delete("/:id/members/:userId", authzMiddleware, h.RemoveMember)
		protected.Post("/:id/leave", authzMiddleware, h.LeaveTeam)

		// ============================================================
		// INVITATION OPERATIONS — admin-side (invite/list/resend)
		// ✅ REMOVED: Role from invitations - roles are inherited from account
		// ============================================================
		protected.Post("/:id/invitations", authzMiddleware, h.InviteMember)
		protected.Get("/:id/invitations", authzMiddleware, h.GetTeamInvitations)
		protected.Post("/:id/invitations/resend", authzMiddleware, h.ResendInvitation)
	}
}