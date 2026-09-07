// internal/modules/account/delivery/handler/routes.go

package handler

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers all account routes
func (h *AccountHandler) RegisterRoutes(
	router fiber.Router,
	authMiddleware fiber.Handler,
	authzMiddleware fiber.Handler,
) {
	// ============================================================
	// 1. PUBLIC ROUTES (No authentication)
	// ============================================================
	public := router.Group("/account-types")
	{
		public.Get("/", h.GetAccountTypes)
		public.Get("/:id", h.GetAccountTypeByID)
		public.Get("/slug/:slug", h.GetAccountTypeBySlug)
	}

	// ============================================================
	// 2. ACCOUNT ROUTES
	// ============================================================
	account := router.Group("/accounts")
	account.Use(authMiddleware)
	account.Use(authzMiddleware)
	{
		// ---- CREATE ----
		account.Post("/personal", h.CreatePersonalAccount)
		account.Post("/institution", h.CreateInstitutionAccount)

		// ---- READ ----
		account.Get("/", h.GetMyAccounts)
		account.Get("/:id", h.GetAccountByID)
		account.Get("/slug/:slug", h.GetAccountBySlug)

		// ---- UPDATE ----
		account.Put("/:id", h.UpdateAccount)

		// ---- DELETE ----
		account.Delete("/:id", h.DeleteAccount)

		// ---- LEAVE ----
		account.Post("/:id/leave", h.LeaveAccount)
	}

	// ============================================================
	// 3. ACCOUNT MEMBER ROUTES
	// ============================================================
	members := router.Group("/accounts/:id/members")
	members.Use(authMiddleware)
	members.Use(authzMiddleware)
	{
		// ---- READ ----
		members.Get("/", h.GetAccountMembers)

		// ---- CREATE ----
		members.Post("/", h.AddMember)

		// ---- UPDATE ----
		members.Put("/:userId/role", h.UpdateMemberRole)

		// ---- DELETE ----
		members.Delete("/:userId", h.RemoveMember)
	}

	// ============================================================
	// 4. USER'S ACCOUNTS ROUTES
	// ============================================================
	user := router.Group("/users/me/accounts")
	user.Use(authMiddleware)
	user.Use(authzMiddleware)
	{
		user.Get("/", h.GetMyAccounts)
	}
}