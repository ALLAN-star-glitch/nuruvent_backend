// internal/modules/team/delivery/handler/routes.go

package handler

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers all team routes.
func (h *TeamHandler) RegisterRoutes(
	router fiber.Router,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
	optionalAuth fiber.Handler,
) {
	// ============================================================
	// PUBLIC / OPTIONAL-AUTH ROUTES
	//
	// Invitation accept/decline must accept both anonymous and
	// authenticated callers:
	//
	//   - Existing user, already signed in → optionalAuth populates the
	//     user ID; the service matches it against the invitation email.
	//   - New user (never signed up) → optionalAuth finds no token, the
	//     request proceeds anonymous; the handler/service creates the
	//     user from the request body using the invitation's email.
	//
	// Static "invitations/..." segments MUST be registered before any
	// dynamic /:id route in this file, otherwise the router would
	// capture "invitations" as the team ID.
	// ============================================================
	public := router.Group("/teams")
	{
		public.Get("/invitations/validate", h.ValidateInvitation)
		public.Post("/invitations/accept", optionalAuth, h.AcceptInvitation)
		public.Post("/invitations/decline", optionalAuth, h.DeclineInvitation)
	}

	// ============================================================
	// PROTECTED ROUTES (auth required)
	// ============================================================
	protected := router.Group("/teams")
	protected.Use(authMiddleware)
	{
		// ---- Personal / institution team creation ----
		protected.Post("/personal", authzMiddleware, h.CreatePersonalTeam)
		protected.Post("/institution", authzMiddleware, h.CreateInstitutionTeam)

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
		// ============================================================
		protected.Get("/:id/members", authzMiddleware, h.GetTeamMembers)
		protected.Post("/:id/members", authzMiddleware, h.AddMember)
		protected.Delete("/:id/members/:userId", authzMiddleware, h.RemoveMember)
		protected.Post("/:id/leave", authzMiddleware, h.LeaveTeam)

		// ============================================================
		// INVITATION OPERATIONS — admin side
		// ============================================================
		protected.Post("/:id/invitations", authzMiddleware, h.InviteMember)
		protected.Get("/:id/invitations", authzMiddleware, h.GetTeamInvitations)
		protected.Post("/:id/invitations/resend", authzMiddleware, h.ResendInvitation)
	}
}