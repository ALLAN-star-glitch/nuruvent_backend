// internal/modules/account/delivery/handler/routes.go

package handler

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers all account routes.
//
// IMPORTANT — Route ordering rule:
//   Fiber matches routes in the order they are registered. Any route with a
//   literal segment that could also be captured by a parametrised sibling
//   (e.g. "/users/me/profile" vs "/users/:id/profile") MUST be registered
//   first, otherwise the parametrised route wins and the literal segment is
//   treated as an ID (which then fails UUID validation downstream).
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
		// Static segment "slug" before parametrised ":id" so "/slug/:slug"
		// is not shadowed by "/:id" (which would capture "slug" as the id).
		public.Get("/slug/:slug", h.GetAccountTypeBySlug)
		public.Get("/:id", h.GetAccountTypeByID)
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
		// Static "slug" segment before parametrised ":id".
		account.Get("/slug/:slug", h.GetAccountBySlug)
		account.Get("/:id", h.GetAccountByID)

		// ---- UPDATE ----
		account.Put("/:id", h.UpdateAccount)

		// ---- DELETE ----
		account.Delete("/:id", h.DeleteAccount)

		// ---- LEAVE ----
		account.Post("/:id/leave", h.LeaveAccount)

		// ---- LOGO ----
		account.Post("/:id/logo", h.UploadAccountLogo)
		account.Delete("/:id/logo", h.DeleteAccountLogo)
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

	// ============================================================
	// 5. USER AVATAR ROUTES
	// ============================================================
	avatar := router.Group("/users/me/avatar")
	avatar.Use(authMiddleware)
	avatar.Use(authzMiddleware)
	{
		avatar.Post("/", h.UploadMyAvatar)
		avatar.Delete("/", h.DeleteMyAvatar)
	}

	// ============================================================
	// 6. USER'S OWN PROFILE ROUTES (Auth required)
	//
	// ⚠️ MUST be registered BEFORE the public "/users/:id/profile"
	//    route below. Fiber matches in registration order, so if the
	//    parametrised route comes first, "me" is captured as :id and
	//    passed to the repository, producing:
	//    "invalid input syntax for type uuid: \"me\"".
	// ============================================================
	profile := router.Group("/users/me/profile")
	profile.Use(authMiddleware)
	profile.Use(authzMiddleware)
	{
		profile.Get("/", h.GetMyProfile)
		profile.Put("/", h.UpdateMyProfile)
	}

	
	// ============================================================
	// 7. PUBLIC USER PROFILE ROUTES (No auth)
	// ============================================================
	profilePublic := router.Group("/users")
	{
		profilePublic.Get("/slug/:slug/profile", h.GetPublicProfileBySlug)
		profilePublic.Get("/:id/profile", h.GetPublicProfile)
	}
}