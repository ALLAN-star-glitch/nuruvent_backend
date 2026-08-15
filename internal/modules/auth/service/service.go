// internal/modules/auth/service/service.go

package service

import (
	"context"

	authdomain "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/config"
	sharedRedis "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
)

// ============================================================
// INBOUND PORT: Service Interface
// ============================================================

type Service interface {
	// Registration
	RegisterAccount(ctx context.Context, req RegisterRequest) error
	VerifyOTPAndCreateAccount(ctx context.Context, email, otp string) (*authdomain.Account, map[string]interface{}, error)

	// Login
	LoginAccount(ctx context.Context, email, password, ipAddress, userAgent string) (*authdomain.Account, string, error)
	VerifyTwoFactorAndLogin(ctx context.Context, email, otp, ipAddress, userAgent string) (*authdomain.Account, string, string, error)

	// Token management
	GenerateTokens(ctx context.Context, account *authdomain.Account) (string, string, error)
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

	// Email - now handled by notification service
	SendOTPEmail(to, name, otp string) error

	// Two-factor OTP
	StoreTwoFactorOTP(email, otp string) error
	GetTwoFactorOTP(email string) (string, error)
	DeleteTwoFactorOTP(email string) error

	// User data
	StoreUserData(email string, data map[string]interface{}) error
	GetUserData(email string) (map[string]string, error)
	DeleteUserData(email string) error

	// Password reset data
	StoreResetData(email, otp, newPassword string) error
	GetResetData(email string) (map[string]string, error)
	DeleteResetData(email string) error
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
	redisClient *sharedRedis.Client          // ✅ Injected Redis client
	queue       authdomain.QueueService
	permService authdomain.PermissionService
	tokenSvc    authdomain.TokenService
	notifSvc    authdomain.NotificationService
}

func NewService(
	repo authdomain.Repository,
	cfg *config.Config,
	redisClient *sharedRedis.Client,        // ✅ Inject Redis client
	queueClient authdomain.QueueService,
	permService authdomain.PermissionService,
	tokenSvc authdomain.TokenService,
	notifSvc authdomain.NotificationService,
) Service {
	return &service{
		repo:        repo,
		config:      cfg,
		redisClient: redisClient,
		queue:       queueClient,
		permService: permService,
		tokenSvc:    tokenSvc,
		notifSvc:    notifSvc,
	}
}