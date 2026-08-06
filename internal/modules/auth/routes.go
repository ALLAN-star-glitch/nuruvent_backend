// internal/modules/auth/routes.go

package auth

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/auth/handler"
	"github.com/gofiber/fiber/v3"
)

// RegisterAuthRoutes registers public auth routes (no authentication required)
func RegisterAuthRoutes(router fiber.Router, h *handler.Handler) {
	auth := router.Group("/auth")
	{
		// ================================================
		// REGISTRATION FLOW
		// ================================================
		auth.Post("/register", h.Register)
		auth.Post("/verify-otp", h.VerifyOTP)
		auth.Post("/resend-otp", h.ResendOTP)

		// ================================================
		// AUTHENTICATION FLOW
		// ================================================
		auth.Post("/login", h.Login)
		auth.Post("/verify-2fa", h.VerifyTwoFactorOTP)

		// ================================================
		// TOKEN MANAGEMENT
		// ================================================
		auth.Post("/refresh", h.RefreshToken)

		// ================================================
		// PASSWORD RESET FLOW
		// ================================================
		auth.Post("/forgot-password", h.ForgotPassword)
		auth.Post("/verify-reset-otp", h.VerifyResetOTP)
	}
}

// RegisterProtectedRoutes registers protected auth routes (authentication required)
func RegisterProtectedRoutes(router fiber.Router, h *handler.Handler) {
	// ================================================
	// SESSION MANAGEMENT
	// ================================================
	router.Post("/auth/logout", h.Logout)
}