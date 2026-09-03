// internal/modules/profile/delivery/handler/routes.go

package handler

import (
	"github.com/gofiber/fiber/v3"
)

func (h *ProfileHandler) RegisterRoutes(
	router fiber.Router,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
) {
	// ============================================================
	// 1. PUBLIC ROUTES - Accessible without authentication
	//    Purpose: Public profile viewing
	// ============================================================
	public := router.Group("/profile")
	{
		// Get public user profile (basic info - name, avatar, etc.)
		public.Get("/users/:id", h.GetUserProfile)

		// Get public institution profile (basic info)
		public.Get("/institutions/:id", h.GetInstitutionProfile)
	}

	// ============================================================
	// 2. PERSONAL TEAM ROUTES - User's own profile
	//    Purpose: View and update own profile
	//    Scope: personal:team:{user_id}
	// ============================================================
	personal := router.Group("/users/me/profile")
	personal.Use(authMiddleware)
	personal.Use(authzMiddleware)
	{
		// View my full profile (with email, phone, etc.)
		personal.Get("/", h.GetMyProfile)

		// Update my profile
		personal.Put("/", h.UpdateMyProfile)
	}

	// ============================================================
	// 3. INSTITUTION TEAM ROUTES - Institution profile
	//    Purpose: View and update institution profile
	//    Scope: institution:team:{institution_id}
	// ============================================================
	institution := router.Group("/institutions/:institutionId/profile")
	institution.Use(authMiddleware)
	institution.Use(authzMiddleware)
	{
		// View institution profile (with details if authorized)
		institution.Get("/", h.GetInstitutionProfile)

		// Update institution profile (Account Admin only)
		institution.Put("/", h.UpdateInstitutionProfile)
	}

	// ============================================================
	// 4. CROSS-TEAM PROFILE OPERATIONS
	//    Purpose: View profiles across teams (for admins/managers)
	// ============================================================
	cross := router.Group("/profile")
	cross.Use(authMiddleware)
	cross.Use(authzMiddleware)
	{
		// Get multiple user profiles (bulk)
		cross.Get("/users", h.GetUserProfiles)

		// Get multiple institution profiles (bulk)
		cross.Get("/institutions", h.GetInstitutionProfiles)

		// List users with filters (pagination, search, etc.)
		cross.Get("/users/list", h.ListUsers)

		// List institutions with filters (pagination, search, etc.)
		cross.Get("/institutions/list", h.ListInstitutions)

		// Get organizer info for events module
		cross.Get("/organizer", h.GetOrganizerInfo)
	}
}
