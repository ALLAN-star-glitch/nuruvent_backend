package business

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/business/bizhandler"
	"github.com/gofiber/fiber/v3"
)

func RegisterBusinessRoutes(
	router fiber.Router,
	businessHandler *bizhandler.BusinessHandler,
	memberHandler *bizhandler.MemberHandler,
	enforcer *authorization.Enforcer,
) {
	// ================================================
	// PUBLIC REFERENCE ROUTES (No auth required)
	// ================================================
	public := router.Group("/businesses")
	{
		public.Get("/types", businessHandler.GetBusinessTypes)
		public.Get("/search", businessHandler.SearchBusinesses)
		public.Get("/:id", businessHandler.GetBusiness)
	}

	// ================================================
	// PROTECTED BUSINESS ROUTES
	// ================================================
	protected := router.Group("/businesses")
	protected.Use(authorization.AuthorizationMiddleware(enforcer))
	{
		// "ME" endpoints - get user's businesses
		protected.Get("/me", businessHandler.GetMyBusiness)
		protected.Get("/my", businessHandler.GetMyBusinesses)

		// Business CRUD
		protected.Post("/", businessHandler.CreateBusiness)
		protected.Put("/:id", businessHandler.UpdateBusiness)
		protected.Delete("/:id", businessHandler.DeleteBusiness)

		// Business stats
		protected.Get("/:id/stats", businessHandler.GetBusinessStats)
	}

	// ================================================
	// MEMBER ROUTES (Protected + Business Access)
	// ================================================
	memberRoutes := router.Group("/businesses")
	memberRoutes.Use(authorization.AuthorizationMiddleware(enforcer))
	memberRoutes.Use(authorization.RequireBusinessRole(enforcer))
	{
		// Get all members
		memberRoutes.Get("/:id/members", memberHandler.GetBusinessMembers)
		
		// Add/remove members
		memberRoutes.Post("/:id/members", memberHandler.AddMember)
		memberRoutes.Delete("/:id/members/:memberId", memberHandler.RemoveMember)
		
		// Update member role
		memberRoutes.Put("/:id/members/:memberId/role", memberHandler.UpdateMemberRole)
		
		// Check membership
		memberRoutes.Get("/:id/members/check", memberHandler.CheckMembership)
	}
}