// internal/modules/auth/service/password_reset.go

package service

import (
	"context"
	"errors"
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
		return nil
	}

	log.Printf("✅ [Auth] Account found: %s (ID: %d)", account.Email, account.ID)

	otp := s.GenerateOTP()
	log.Printf("🔑 [Auth] Generated OTP: %s for %s", otp, email)

	if err := s.StoreResetData(email, otp, newPassword); err != nil {
		log.Printf("❌ [Auth] Failed to store reset data: %v", err)
		return fmt.Errorf("failed to store reset data: %w", err)
	}
	log.Printf("✅ [Auth] Reset data stored in Redis")

	// Send password reset OTP via notification service
	if s.notifSvc == nil {
		log.Printf("❌ [Auth] Notification service is NIL!")
		return errors.New("notification service not available")
	}

	if err := s.sendPasswordResetOTP(email, account.Name, otp); err != nil {
		log.Printf("❌ [Auth] Failed to send password reset OTP: %v", err)
		// Don't fail the request
	} else {
		log.Printf("✅ [Auth] Password reset OTP sent to: %s", email)
	}

	return nil
}

func (s *service) VerifyResetOTPAndResetPassword(ctx context.Context, email, otp string) error {
	log.Printf("🔐 [Auth] VerifyResetOTPAndResetPassword called for: %s", email)

	data, err := s.GetResetData(email)
	if err != nil {
		log.Printf("❌ [Auth] Failed to get reset data: %v", err)
		return authdomain.ErrInvalidOTP
	}

	storedOTP, ok := data["otp"]
	if !ok || otp != storedOTP {
		log.Printf("❌ [Auth] Invalid OTP for %s", email)
		return authdomain.ErrInvalidOTP
	}

	newPassword, ok := data["new_password"]
	if !ok {
		log.Printf("❌ [Auth] Invalid reset data for %s", email)
		return errors.New("invalid reset data")
	}

	account, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		log.Printf("❌ [Auth] Error getting account: %v", err)
		return err
	}
	if account == nil {
		log.Printf("❌ [Auth] Account not found for %s", email)
		return authdomain.ErrAccountNotFound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("❌ [Auth] Failed to hash password: %v", err)
		return fmt.Errorf("failed to hash password: %w", err)
	}

	account.UpdatePassword(string(hashedPassword))

	if err := s.repo.UpdateAccount(account); err != nil {
		log.Printf("❌ [Auth] Failed to update account: %v", err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err := s.DeleteResetData(email); err != nil {
		log.Printf("⚠️ [Auth] Failed to delete reset data: %v", err)
	}

	log.Printf("✅ [Auth] Password reset successfully for %s", email)

	// Send password reset confirmation via notification service
	if err := s.sendPasswordResetConfirm(email, account.Name); err != nil {
		log.Printf("Failed to send password reset confirmation: %v", err)
	}

	return nil
}

// ============================================================
// PRIVATE HELPERS
// ============================================================

// sendPasswordResetOTP sends password reset OTP via notification service
func (s *service) sendPasswordResetOTP(to, name, otp string) error {
	return s.notifSvc.SendPasswordResetOTP(context.Background(), authdomain.SendOTPRequest{
		To:      to,
		Name:    name,
		OTP:     otp,
		Expires: "5 minutes",
		Purpose: "password_reset",
		Meta:    nil,
	})
}

// sendPasswordResetConfirm sends password reset confirmation via notification service
func (s *service) sendPasswordResetConfirm(to, name string) error {
	return s.notifSvc.SendPasswordResetConfirm(context.Background(), authdomain.SendPasswordResetConfirmRequest{
		To:   to,
		Name: name,
	})
}