package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"

	"golang.org/x/crypto/bcrypt"
)

func (s *service) RegisterAccount(ctx context.Context, req RegisterRequest) error {
    exists, err := s.repo.AccountExistsByEmail(req.Email)
    if err != nil {
        return err
    }
    if exists {
        return domain.ErrAccountExists
    }

    exists, err = s.repo.AccountExistsByPhone(req.Phone)
    if err != nil {
        return err
    }
    if exists {
        return domain.ErrInvalidPhone
    }

    otp := s.otpSvc.GenerateOTP()

    if err := s.otpSvc.StoreOTP(req.Email, otp); err != nil {
        return fmt.Errorf("failed to store OTP: %w", err)
    }

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

    if err := s.otpSvc.SendOTPEmail(req.Email, req.Name, otp); err != nil {
        log.Printf("Failed to send OTP: %v", err)
    }

    return nil
}

func (s *service) VerifyOTPAndCreateAccount(ctx context.Context, email, otp string) (*domain.Account, map[string]interface{}, error) {
    storedOTP, err := s.otpSvc.GetOTP(email)
    if err != nil {
        return nil, nil, domain.ErrInvalidOTP
    }
    if otp != storedOTP {
        return nil, nil, domain.ErrInvalidOTP
    }

    userData, err := s.otpSvc.GetUserData(email)
    if err != nil {
        return nil, nil, errors.New("registration data not found")
    }

    s.otpSvc.DeleteOTP(email)
    s.otpSvc.DeleteUserData(email)

    accountType, err := s.repo.GetAccountTypeBySlug(userData["account_type"])
    if err != nil {
        return nil, nil, fmt.Errorf("failed to get account type: %w", err)
    }
    if accountType == nil {
        return nil, nil, domain.ErrAccountTypeNotFound
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData["password"]), bcrypt.DefaultCost)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to hash password: %w", err)
    }

    account, err := domain.NewAccount(
        userData["email"],
        string(hashedPassword),
        userData["name"],
        userData["phone"],
        accountType.ID,
    )
    if err != nil {
        return nil, nil, err
    }

    var institution *domain.Institution
    if userData["account_type"] == "institution" {
        institutionType, err := s.repo.GetInstitutionTypeBySlug(userData["institution_type"])
        if err != nil {
            return nil, nil, fmt.Errorf("failed to get institution type: %w", err)
        }
        if institutionType == nil {
            return nil, nil, domain.ErrInstitutionNotFound
        }

        institution, err = domain.NewInstitution(
            userData["institution_name"],
            userData["institution_email"],
            userData["institution_phone"],
            institutionType.ID,
        )
        if err != nil {
            return nil, nil, err
        }

        if err := s.repo.CreateInstitution(institution); err != nil {
            return nil, nil, fmt.Errorf("failed to create institution: %w", err)
        }

        account.InstitutionID = &institution.ID

        teamMember, err := domain.NewTeamMember(account.ID, account.ID, "admin")
        if err != nil {
            return nil, nil, err
        }

        if err := s.repo.CreateTeamMember(teamMember); err != nil {
            return nil, nil, fmt.Errorf("failed to create team member: %w", err)
        }
    }

    if err := s.repo.CreateAccount(account); err != nil {
        return nil, nil, fmt.Errorf("failed to create account: %w", err)
    }

    accessToken, refreshToken, err := s.GenerateTokens(ctx, account)
    if err != nil {
        return nil, nil, err
    }

    if err := s.emailSvc.SendWelcome(account.Email, account.Name); err != nil {
        log.Printf("Failed to send welcome email: %v", err)
    }

    result := map[string]interface{}{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "expires_in":    int64(s.config.JWT.AccessExpiration.Seconds()),
    }

    if institution != nil {
        result["institution"] = institution
    }

    return account, result, nil
}