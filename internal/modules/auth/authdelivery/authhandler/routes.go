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
		// ---- SELF-SERVICE SIGNUP (OTP-based) ----
		public.Post("/register", h.Register)
		public.Post("/verify-otp", h.VerifyOTP)
		public.Post("/resend-otp", h.ResendOTP)

		// ---- INVITATION SIGNUP (no OTP) ----
		// The invitee's email is read from the invitation record; the
		// request body is {token, name, password}. On success the user
		// is created, the invitation is accepted, and auth tokens are
		// returned — all in one call.
		public.Post("/register-with-invitation", h.RegisterWithInvitation)

		// ---- LOGIN ----
		public.Post("/login", h.Login)
		public.Post("/verify-2fa", h.VerifyTwoFactorOTP)
		public.Post("/refresh", h.RefreshToken)

		// ---- PASSWORD RESET ----
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