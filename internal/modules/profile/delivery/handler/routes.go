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
	// 1. PUBLIC ROUTES (No authentication)
	// ============================================================
	public := router.Group("/profile")
	{
		// Get public user profile
		public.Get("/users/:id", h.GetUserProfile)

		// Get public institution profile
		public.Get("/institutions/:id", h.GetInstitutionProfile)
	}

	// ============================================================
	// 2. AUTHENTICATED ROUTES
	// ============================================================
	auth := router.Group("/profile")
	auth.Use(authMiddleware)
	auth.Use(authzMiddleware)
	{
		// ---- USER ----
		// List users with filters (pagination, search, etc.)
		auth.Get("/users", h.ListUsers)

		// Get multiple users by IDs (bulk)
		auth.Get("/users/bulk", h.GetUserProfiles)

		// ---- INSTITUTION ----
		// List institutions with filters (pagination, search, etc.)
		auth.Get("/institutions", h.ListInstitutions)

		// Get multiple institutions by IDs (bulk)
		auth.Get("/institutions/bulk", h.GetInstitutionProfiles)

		// ---- ORGANIZER ----
		// Get organizer info for events module
		auth.Get("/organizer", h.GetOrganizerInfo)
	}

	// ============================================================
	// 3. PERSONAL TEAM ROUTES
	// ============================================================
	personal := router.Group("/users/me")
	personal.Use(authMiddleware)
	personal.Use(authzMiddleware)
	{
		personal.Get("/profile", h.GetMyProfile)
		personal.Put("/profile", h.UpdateMyProfile)
		personal.Post("/avatar", h.UploadUserAvatar)
		personal.Delete("/avatar", h.DeleteUserAvatar)
	}

	// ============================================================
	// 4. INSTITUTION TEAM ROUTES
	// ============================================================
	institution := router.Group("/institutions/:institutionId")
	institution.Use(authMiddleware)
	institution.Use(authzMiddleware)
	{
		institution.Get("/profile", h.GetInstitutionProfile)
		institution.Put("/profile", h.UpdateInstitutionProfile)
		institution.Post("/logo", h.UploadInstitutionLogo)
		institution.Delete("/logo", h.DeleteInstitutionLogo)
	}
}