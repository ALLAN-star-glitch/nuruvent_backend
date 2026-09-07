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

		// Get public account profile
		public.Get("/accounts/:id", h.GetAccountProfile)
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

		// ---- ACCOUNT ----
		// List accounts with filters (pagination, search, etc.)
		auth.Get("/accounts", h.ListAccounts)

		// Get multiple accounts by IDs (bulk)
		auth.Get("/accounts/bulk", h.GetAccountProfiles)

		// ---- ORGANIZER ----
		// Get organizer info for events module
		auth.Get("/organizer", h.GetOrganizerInfo)
	}

	// ============================================================
	// 3. PERSONAL USER ROUTES
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
	// 4. ACCOUNT ROUTES (replaces Institution routes)
	// ============================================================
	account := router.Group("/accounts/:accountId")
	account.Use(authMiddleware)
	account.Use(authzMiddleware)
	{
		account.Get("/profile", h.GetAccountProfile)
		account.Put("/profile", h.UpdateAccountProfile)
		account.Post("/logo", h.UploadAccountLogo)
		account.Delete("/logo", h.DeleteAccountLogo)
	}
}