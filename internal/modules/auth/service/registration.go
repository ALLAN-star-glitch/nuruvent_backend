package service

import (
	"context"
	"fmt"
	"log"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/email"
)

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

	// 5. Store user data (individual only)
	userData := map[string]any{
		"email":    req.Email,
		"password": req.Password,
		"name":     req.Name,
		"phone":    req.Phone,
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

	// 6.  Send OTP via Queue (async)
	if err := s.enqueueOTPEmail(req.Email, req.Name, otp, "registration"); err != nil {
		log.Printf("Failed to enqueue OTP email: %v", err)
		// Fallback: send synchronously if queue fails
		if err := s.emailSvc.SendVerificationOTP(req.Email, req.Name, otp, "5 minutes", "registration", nil); err != nil {
			log.Printf("Failed to send OTP: %v", err)
		}
	}

	return nil
}

// enqueueOTPEmail enqueues an OTP email task
func (s *service) enqueueOTPEmail(to, name, otp, purpose string) error {
	if s.queue == nil {
		return fmt.Errorf("queue client not available")
	}

	task := email.VerificationOTPTask{
		To:      to,
		Name:    name,
		OTP:     otp,
		Expires: "5 minutes",
		Purpose: email.VerificationPurpose(purpose),
		Meta:    nil,
	}

	payload, err := task.Payload()
	if err != nil {
		return err
	}

	return s.queue.Enqueue(email.TypeVerificationOTP, payload)
}