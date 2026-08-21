// internal/modules/auth/service/password_reset.go

package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================
// PASSWORD RESET METHODS
// ============================================================

func (s *service) InitiatePasswordReset(ctx context.Context, email, newPassword string) error {
	log.Printf("🔐 [Auth] InitiatePasswordReset called for: %s", email)

	account, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		log.Printf("❌ [Auth] Error getting account: %v", err)
		return err
	}
	if account == nil {
		log.Printf("⚠️ [Auth] Account not found for: %s", email)
		return nil // Don't reveal if email exists for security
	}

	log.Printf("✅ [Auth] Account found: %s (ID: %s)", account.Email, account.ID)

	// ✅ Generate OTP
	otp := s.GenerateOTP()
	log.Printf("🔑 [Auth] Generated OTP: %s for %s", otp, email)

	// ✅ Store OTP with purpose - THIS IS THE KEY FIX
	if err := s.StoreOTP(ctx, email, otp, "password_reset"); err != nil {
		log.Printf("❌ [Auth] Failed to store OTP: %v", err)
		return fmt.Errorf("failed to store OTP: %w", err)
	}
	log.Printf("✅ [Auth] OTP stored in Redis for: %s", email)

	// Store reset data with OTP and new password
	if err := s.StoreResetData(ctx, email, otp, newPassword); err != nil {
		log.Printf("❌ [Auth] Failed to store reset data: %v", err)
		return fmt.Errorf("failed to store reset data: %w", err)
	}
	log.Printf("✅ [Auth] Reset data stored in Redis")

	// Send password reset OTP via unified SendOTP
	if s.notifSvc == nil {
		log.Printf("❌ [Auth] Notification service is NIL!")
		return fmt.Errorf("notification service not available")
	}

	if err := s.notifSvc.SendOTP(ctx, authdomain.SendOTPRequest{
		To:      account.Email,
		Name:    account.Name,
		OTP:     otp,
		Expires: "15 minutes",
		Purpose: "password_reset",
		Meta:    nil,
	}); err != nil {
		log.Printf("❌ [Auth] Failed to send password reset OTP: %v", err)
		// Don't fail the request - OTP is stored, user can retry
	} else {
		log.Printf("✅ [Auth] Password reset OTP sent to: %s", email)
	}

	return nil
}

func (s *service) VerifyResetOTPAndResetPassword(ctx context.Context, email, otp string) error {
	log.Printf("🔐 [Auth] VerifyResetOTPAndResetPassword called for: %s", email)

	// ✅ Verify OTP using unified VerifyOTP
	if err := s.VerifyOTP(ctx, email, otp, "password_reset"); err != nil {
		log.Printf("❌ [Auth] Invalid OTP for %s: %v", email, err)
		return err
	}
	log.Printf("✅ [Auth] OTP verified successfully for: %s", email)

	// Get reset data
	resetData, err := s.GetResetData(ctx, email)
	if err != nil {
		log.Printf("❌ [Auth] Failed to get reset data: %v", err)
		return fmt.Errorf("reset data not found")
	}

	newPassword, ok := resetData["new_password"]
	if !ok || newPassword == "" {
		log.Printf("❌ [Auth] Invalid reset data for %s", email)
		return fmt.Errorf("invalid reset data")
	}

	// Get account
	account, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		log.Printf("❌ [Auth] Error getting account: %v", err)
		return err
	}
	if account == nil {
		log.Printf("❌ [Auth] Account not found for %s", email)
		return authdomain.ErrAccountNotFound
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ [Auth] Failed to hash password: %v", err)
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	account.UpdatePassword(string(hashedPassword))
	if err := s.repo.UpdateAccount(account); err != nil {
		log.Printf("❌ [Auth] Failed to update account: %v", err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Clean up
	if err := s.DeleteOTP(ctx, email, "password_reset"); err != nil {
		log.Printf("⚠️ [Auth] Failed to delete OTP: %v", err)
	}
	if err := s.DeleteResetData(ctx, email); err != nil {
		log.Printf("⚠️ [Auth] Failed to delete reset data: %v", err)
	}

	log.Printf("✅ [Auth] Password reset successfully for %s", email)

	// Send password reset confirmation
	if err := s.notifSvc.SendPasswordResetConfirm(ctx, authdomain.SendPasswordResetConfirmRequest{
		To:   account.Email,
		Name: account.Name,
	}); err != nil {
		log.Printf("⚠️ [Auth] Failed to send password reset confirmation: %v", err)
	}

	return nil
}