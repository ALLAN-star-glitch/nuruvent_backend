package auth

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterAuthRoutes registers public auth routes (no authentication required)
func RegisterAuthRoutes(router fiber.Router, handler *Handler) {
	auth := router.Group("/auth")
	{
		// ================================================
		// REGISTRATION FLOW
		// ================================================
		auth.Post("/register", handler.Register)
		auth.Post("/verify-otp", handler.VerifyOTP)
		auth.Post("/resend-otp", handler.ResendOTP)

		// ================================================
		// AUTHENTICATION FLOW
		// ================================================
		auth.Post("/login", handler.Login)
		auth.Post("/verify-2fa", handler.VerifyTwoFactorOTP)

		// ================================================
		// TOKEN MANAGEMENT
		// ================================================
		auth.Post("/refresh", handler.RefreshToken)

		// ================================================
		// PASSWORD RESET FLOW
		// ================================================
		auth.Post("/forgot-password", handler.ForgotPassword)
		auth.Post("/verify-reset-otp", handler.VerifyResetOTP)
	}
}

// RegisterProtectedRoutes registers protected auth routes (authentication required)
func RegisterProtectedRoutes(router fiber.Router, handler *Handler) {
	// ================================================
	// SESSION MANAGEMENT
	// ================================================
	router.Post("/auth/logout", handler.Logout)
}