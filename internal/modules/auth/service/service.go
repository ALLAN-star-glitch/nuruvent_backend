package service

import (
	"context"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/email"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/queue"
)

// ============================================================
// INBOUND PORT: Service Interface
// ============================================================

type Service interface {
    // Registration
    RegisterAccount(ctx context.Context, req RegisterRequest) error
    VerifyOTPAndCreateAccount(ctx context.Context, email, otp string) (*domain.Account, map[string]interface{}, error)

    // Login
    LoginAccount(ctx context.Context, email, password, ipAddress, userAgent string) (*domain.Account, string, error)
    VerifyTwoFactorAndLogin(ctx context.Context, email, otp, ipAddress, userAgent string) (*domain.Account, string, string, error)

    // Token management
    GenerateTokens(ctx context.Context, account *domain.Account) (string, string, error)
    RefreshTokens(ctx context.Context, refreshToken, userAgent, ip string) (string, string, error)
    RevokeToken(ctx context.Context, refreshToken string) error

    // Password reset
    InitiatePasswordReset(ctx context.Context, email, newPassword string) error
    VerifyResetOTPAndResetPassword(ctx context.Context, email, otp string) error

    // OTP
    GenerateOTP() string
    StoreOTP(email, otp string) error
    GetOTP(email string) (string, error)
    DeleteOTP(email string) error

    // Email
    SendOTPEmail(to, name, otp string) error
}

// ============================================================
// COMMANDS
// ============================================================

type RegisterRequest struct {
    Email       string
    Password    string
    Name        string
    Phone       string
    AccountType string

    InstitutionName  string
    InstitutionEmail string
    InstitutionPhone string
    InstitutionType  string
}

// ============================================================
// SERVICE IMPLEMENTATION
// ============================================================

type service struct {
    repo        domain.Repository
    config      *config.Config
    queue       *queue.Client
    permService domain.PermissionService // ✅ Fixed: Interface type (no pointer)
    tokenSvc    domain.TokenService
    otpSvc      domain.OTPService
    emailSvc    *email.EmailService
}

func NewService(
    repo domain.Repository,
    cfg *config.Config,
    queueClient *queue.Client,
    permService domain.PermissionService, // ✅ Fixed: Interface type (no pointer)
    tokenSvc domain.TokenService,
    otpSvc domain.OTPService,
    emailSvc *email.EmailService,
) Service {
    return &service{ // This must satisfy the interface 
        repo:        repo,
        config:      cfg,
        queue:       queueClient,
        permService: permService,
        tokenSvc:    tokenSvc,
        otpSvc:      otpSvc,
        emailSvc:    emailSvc,
    }
}
