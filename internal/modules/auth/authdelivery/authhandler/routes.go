package authhandler

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers all auth routes
func (h *AuthHandler) RegisterRoutes(
	router fiber.Router,
	authMiddleware fiber.Handler,
) {
	// ============================================================
	// PUBLIC ROUTES (No auth required)
	// ============================================================
	public := router.Group("/auth")
	{
		public.Post("/register", h.Register)
		public.Post("/verify-otp", h.VerifyOTP)
		public.Post("/resend-otp", h.ResendOTP)
		public.Post("/login", h.Login)
		public.Post("/verify-2fa", h.VerifyTwoFactorOTP)
		public.Post("/refresh", h.RefreshToken)
		public.Post("/forgot-password", h.ForgotPassword)
		public.Post("/verify-reset-otp", h.VerifyResetOTP)
	}

	// ============================================================
	// PROTECTED ROUTES (Auth required)
	// ============================================================
	protected := router.Group("/auth")
	protected.Use(authMiddleware)
	{
		protected.Post("/logout", h.Logout)
	}
}