package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

func (s *service) RegisterAccount(ctx context.Context, req RegisterRequest) error {
	// 1. Check email uniqueness
	exists, err := s.repo.AccountExistsByEmail(req.Email)
	if err != nil {
		return err
	}
	if exists {
		return authdomain.ErrAccountExists
	}

	// 2. Check phone uniqueness
	exists, err = s.repo.AccountExistsByPhone(req.Phone)
	if err != nil {
		return err
	}
	if exists {
		return authdomain.ErrInvalidPhone
	}

	// 3. Generate OTP
	otp := s.GenerateOTP()

	// 4. Store OTP
	if err := s.StoreOTP(req.Email, otp); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// 5. Store user data
	userData := map[string]any{
		"email":        req.Email,
		"password":     req.Password,
		"name":         req.Name,
		"phone":        req.Phone,
		"account_type": req.AccountType,
	}

	if req.AccountType == "institution" {
		userData["institution_name"] = req.InstitutionName
		userData["institution_email"] = req.InstitutionEmail
		userData["institution_phone"] = req.InstitutionPhone
		userData["institution_type"] = req.InstitutionType
	}

	if err := s.StoreUserData(req.Email, userData); err != nil {
		return fmt.Errorf("failed to store user data: %w", err)
	}

	// 6. Send OTP via Notification Service (using authdomain port)
	if err := s.sendOTPNotification(req.Email, req.Name, otp, "registration"); err != nil {
		log.Printf("Failed to send OTP notification: %v", err)
		// Don't fail the registration - OTP can be resent later
	}

	return nil
}

// sendOTPNotification sends OTP via notification service
func (s *service) sendOTPNotification(to, name, otp, purpose string) error {
	// Use the notification service port
	if err := s.notifSvc.SendVerificationOTP(context.Background(), authdomain.SendOTPRequest{
		To:      to,
		Name:    name,
		OTP:     otp,
		Expires: "5 minutes",
		Purpose: purpose,
		Meta:    nil,
	}); err != nil {
		return fmt.Errorf("failed to send OTP: %w", err)
	}
	return nil
}