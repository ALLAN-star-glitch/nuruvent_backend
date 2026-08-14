package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"
	"golang.org/x/crypto/bcrypt"
)

func (s *service) InitiatePasswordReset(ctx context.Context, email, newPassword string) error {
	account, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return err
	}
	if account == nil {
		return nil
	}

	otp := s.otpSvc.GenerateOTP()
	if err := s.otpSvc.StoreResetData(email, otp, newPassword); err != nil {
		return fmt.Errorf("failed to store reset data: %w", err)
	}

	// Send password reset OTP via notification service
	if err := s.sendPasswordResetOTP(email, account.Name, otp); err != nil {
		log.Printf("Failed to send password reset OTP: %v", err)
	}

	return nil
}

func (s *service) VerifyResetOTPAndResetPassword(ctx context.Context, email, otp string) error {
	data, err := s.otpSvc.GetResetData(email)
	if err != nil {
		return domain.ErrInvalidOTP
	}

	storedOTP, ok := data["otp"]
	if !ok || otp != storedOTP {
		return domain.ErrInvalidOTP
	}

	newPassword, ok := data["new_password"]
	if !ok {
		return errors.New("invalid reset data")
	}

	account, err := s.repo.GetAccountByEmail(email)
	if err != nil {
		return err
	}
	if account == nil {
		return domain.ErrAccountNotFound
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	account.UpdatePassword(string(hashedPassword))

	if err := s.repo.UpdateAccount(account); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.otpSvc.DeleteResetData(email)

	// Send password reset confirmation via notification service
	if err := s.sendPasswordResetConfirm(email, account.Name); err != nil {
		log.Printf("Failed to send password reset confirmation: %v", err)
	}

	return nil
}

// sendPasswordResetOTP sends password reset OTP via notification service
func (s *service) sendPasswordResetOTP(to, name, otp string) error {
	return s.notifSvc.SendPasswordResetOTP(context.Background(), domain.SendOTPRequest{
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
	return s.notifSvc.SendPasswordResetConfirm(context.Background(), domain.SendPasswordResetConfirmRequest{
		To:   to,
		Name: name,
	})
}