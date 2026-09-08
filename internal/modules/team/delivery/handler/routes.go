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
		// Team operations
		protected.Post("/", authzMiddleware, h.CreatePersonalTeam)
		protected.Get("/", authzMiddleware, h.GetUserTeams)
		protected.Get("/:id", authzMiddleware, h.GetTeam)
		protected.Patch("/:id", authzMiddleware, h.UpdateTeam)
		protected.Delete("/:id", authzMiddleware, h.DeleteTeam)

		// Member operations
		protected.Get("/:id/members", authzMiddleware, h.GetTeamMembers)
		protected.Post("/:id/members", authzMiddleware, h.AddMember)
		protected.Delete("/:id/members/:userId", authzMiddleware, h.RemoveMember)
		protected.Post("/:id/leave", authzMiddleware, h.LeaveTeam)

		// Invitation operations
		protected.Post("/:id/invite", authzMiddleware, h.InviteMember)  // ✅ Updated: team ID in URL
		protected.Get("/:id/invitations", authzMiddleware, h.GetTeamInvitations)
		protected.Post("/:id/invitations/resend", authzMiddleware, h.ResendInvitation)
	}
}