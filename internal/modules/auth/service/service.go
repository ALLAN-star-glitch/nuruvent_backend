// internal/modules/auth/service/service.go

package service

import (
	"context"

	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authorization"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	sharedRedis "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
)

// ============================================================
// INBOUND PORT: Service Interface
// ============================================================

type Service interface {
	// ============================================================
	// REGISTRATION
	// ============================================================
	
	RegisterAccount(ctx context.Context, req RegisterRequest) error
	VerifyOTPAndCreateAccount(ctx context.Context, email, otp string) (*authdomain.Account, map[string]interface{}, error)

	// ============================================================
	// LOGIN
	// ============================================================
	
	LoginAccount(ctx context.Context, email, password, ipAddress, userAgent string) (*authdomain.Account, string, error)
	VerifyTwoFactorAndLogin(ctx context.Context, email, otp, ipAddress, userAgent string) (*authdomain.Account, string, string, error)

	// ============================================================
	// TOKEN MANAGEMENT
	// ============================================================
	
	GenerateTokens(ctx context.Context, account *authdomain.Account) (string, string, error)
	RefreshTokens(ctx context.Context, refreshToken, userAgent, ip string) (string, string, error)
	RevokeToken(ctx context.Context, refreshToken string) error

	// ============================================================
	// PASSWORD RESET
	// ============================================================
	
	InitiatePasswordReset(ctx context.Context, email, newPassword string) error
	VerifyResetOTPAndResetPassword(ctx context.Context, email, otp string) error

	// ============================================================
	// UNIFIED OTP METHODS
	// ============================================================
	
	// GenerateOTP generates a 6-digit OTP
	GenerateOTP() string
	
	// StoreOTP stores an OTP with purpose
	StoreOTP(ctx context.Context, email, otp, purpose string) error
	
	// GetOTP retrieves an OTP by email and purpose
	GetOTP(ctx context.Context, email, purpose string) (string, error)
	
	// DeleteOTP deletes an OTP by email and purpose
	DeleteOTP(ctx context.Context, email, purpose string) error
	
	// VerifyOTP verifies an OTP for a specific purpose
	VerifyOTP(ctx context.Context, email, otp, purpose string) error

	// ============================================================
	// CONVENIENCE OTP METHOD
	// ============================================================
	
	// SendOTPEmail is a convenience method that generates, stores, and sends an OTP
	// Purpose can be: "registration", "two_factor", "password_reset", "email_change", "phone_change"
	SendOTPEmail(ctx context.Context, to, name, purpose string, meta map[string]string) error


	ResendOTP(ctx context.Context, email, name, purpose string) error

	// ============================================================
	// USER DATA (Registration flow)
	// ============================================================
	
	StoreUserData(ctx context.Context, email string, data map[string]interface{}) error
	GetUserData(ctx context.Context, email string) (map[string]string, error)
	DeleteUserData(ctx context.Context, email string) error

	// ============================================================
	// PASSWORD RESET DATA
	// ============================================================
	
	StoreResetData(ctx context.Context, email, otp, newPassword string) error
	GetResetData(ctx context.Context, email string) (map[string]string, error)
	DeleteResetData(ctx context.Context, email string) error
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
	repo        authdomain.Repository
	config      *config.Config
	redisClient *sharedRedis.Client
	queue       authdomain.QueueService
	permService authdomain.PermissionService
	tokenSvc    authdomain.TokenService
	notifSvc    authdomain.NotificationService
	enforcer    *authorization.Enforcer
}

func NewService(
	repo authdomain.Repository,
	cfg *config.Config,
	redisClient *sharedRedis.Client,
	queueClient authdomain.QueueService,
	permService authdomain.PermissionService,
	tokenSvc authdomain.TokenService,
	notifSvc authdomain.NotificationService,
	enforcer *authorization.Enforcer,
) Service {
	return &service{
		repo:        repo,
		config:      cfg,
		redisClient: redisClient,
		queue:       queueClient,
		permService: permService,
		tokenSvc:    tokenSvc,
		notifSvc:    notifSvc,
		enforcer:    enforcer,
	}
}