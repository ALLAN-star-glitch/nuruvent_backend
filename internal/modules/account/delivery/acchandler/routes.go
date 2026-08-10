package acchandler

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers all account routes
func (h *AccountHandler) RegisterRoutes(
	router fiber.Router,
	authMiddleware fiber.Handler,      // Authentication middleware (JWT)
	authzMiddleware fiber.Handler,     // Authorization middleware (Casbin)
) {
	// ============================================================
	// PUBLIC ROUTES (No auth required)
	// ============================================================
	public := router.Group("/accounts")
	{
		public.Get("/:id", h.GetAccountByID)           // Get account by ID (public)
		public.Get("/email/:email", h.GetAccountByEmail) // Get account by email (public)
	}

	// ============================================================
	// PROTECTED ROUTES (Auth required)
	// ============================================================
	protected := router.Group("/accounts")
	protected.Use(authMiddleware)        // ✅ Validate JWT first
	protected.Use(authzMiddleware)       // ✅ Then check permissions
	{
		protected.Put("/:id", h.UpdateAccount)                    // Update account
		protected.Delete("/:id", h.DeleteAccount)                 // Delete account
		protected.Put("/:id/profile", h.UpdateProfile)            // Update profile
		protected.Put("/:id/professional-type", h.UpdateProfessionalType) // Update professional type
	}

	// ============================================================
	// CURRENT USER ROUTES (Auth required)
	// ============================================================
	me := router.Group("/accounts/me")
	me.Use(authMiddleware)        // ✅ Validate JWT first
	me.Use(authzMiddleware)       // ✅ Then check permissions
	{
		me.Get("/", h.GetCurrentAccount)          // Get current user's account
		me.Put("/", h.UpdateCurrentAccount)       // Update current user's account
		me.Put("/profile", h.UpdateCurrentProfile) // Update current user's profile
	}
}