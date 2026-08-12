package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"

)

// ============================================================
// REGISTER ACCOUNT
// ============================================================

func (s *service) RegisterAccount(ctx context.Context, req RegisterRequest) error {
	// 1. Check email uniqueness
	exists, err := s.repo.AccountExistsByEmail(req.Email)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrAccountExists
	}

	// 2. Check phone uniqueness
	exists, err = s.repo.AccountExistsByPhone(req.Phone)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrInvalidPhone
	}

	// 3. Generate OTP
	otp := s.otpSvc.GenerateOTP()

	// 4. Store OTP
	if err := s.otpSvc.StoreOTP(req.Email, otp); err != nil {
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

	if err := s.otpSvc.StoreUserData(req.Email, userData); err != nil {
		return fmt.Errorf("failed to store user data: %w", err)
	}

	// 6. Send OTP email
	if err := s.otpSvc.SendOTPEmail(req.Email, req.Name, otp); err != nil {
		log.Printf("Failed to send OTP: %v", err)
	}

	return nil
}