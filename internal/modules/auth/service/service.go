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
	
	RegisterUser(ctx context.Context, req RegisterRequest) error
	VerifyOTPAndCreateUser(ctx context.Context, email, otp string) (*authdomain.User, map[string]interface{}, error)

	// ============================================================
	// LOGIN
	// ============================================================
	
	LoginUser(ctx context.Context, email, password, ipAddress, userAgent string) (*authdomain.User, string, error)
	VerifyTwoFactorAndLogin(ctx context.Context, email, otp, ipAddress, userAgent string) (*authdomain.User, string, string, error)

	// ============================================================
	// TOKEN MANAGEMENT
	// ============================================================
	
	GenerateTokens(ctx context.Context, user *authdomain.User) (string, string, error)
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
	
	GenerateOTP() string
	StoreOTP(ctx context.Context, email, otp, purpose string) error
	GetOTP(ctx context.Context, email, purpose string) (string, error)
	DeleteOTP(ctx context.Context, email, purpose string) error
	VerifyOTP(ctx context.Context, email, otp, purpose string) error

	// ============================================================
	// CONVENIENCE OTP METHOD
	// ============================================================
	
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

	// ============================================================
	// PROFESSIONAL TYPE
	// ============================================================
	
	GetProfessionalTypeBySlug(ctx context.Context, slug string) (*authdomain.ProfessionalType, error)
	ListProfessionalTypes(ctx context.Context) ([]*authdomain.ProfessionalType, error)
	GetAccountTypeByID(ctx context.Context, id string) (*authdomain.AccountType, error)
	GetProfessionalTypeByID(ctx context.Context, id string) (*authdomain.ProfessionalType, error)
}

// ============================================================
// COMMANDS
// ============================================================

type RegisterRequest struct {
	// User fields
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8"`
	Name           string `json:"name" validate:"required"`
	Phone          string `json:"phone" validate:"required"`
	AccountType    string `json:"account_type" validate:"required,oneof=personal institution"`
	
	// Professional Type (for personal accounts)
	ProfessionalType string `json:"professional_type,omitempty"`

	// Institution fields (for institution accounts)
	InstitutionName  string `json:"institution_name,omitempty"`
	InstitutionEmail string `json:"institution_email,omitempty"`
	InstitutionPhone string `json:"institution_phone,omitempty"`
	InstitutionType  string `json:"institution_type,omitempty"`
}

// ============================================================
// SERVICE IMPLEMENTATION
// ============================================================

type service struct {
	repo          authdomain.Repository
	config        *config.Config
	redisClient   *sharedRedis.Client
	queue         authdomain.QueueService
	permChecker   authdomain.PermissionChecker
	roleManager   authdomain.RoleManager
	policyManager authdomain.PolicyManager
	tokenSvc      authdomain.TokenService
	notifSvc      authdomain.NotificationService
	enforcer      *authorization.Enforcer
}

func NewService(
	repo authdomain.Repository,
	cfg *config.Config,
	redisClient *sharedRedis.Client,
	queueClient authdomain.QueueService,
	permChecker authdomain.PermissionChecker,
	roleManager authdomain.RoleManager,
	policyManager authdomain.PolicyManager,
	tokenSvc authdomain.TokenService,
	notifSvc authdomain.NotificationService,
	enforcer *authorization.Enforcer,
) Service {
	return &service{
		repo:          repo,
		config:        cfg,
		redisClient:   redisClient,
		queue:         queueClient,
		permChecker:   permChecker,
		roleManager:   roleManager,
		policyManager: policyManager,
		tokenSvc:      tokenSvc,
		notifSvc:      notifSvc,
		enforcer:      enforcer,
	}
}

// ============================================================
// PROFESSIONAL TYPE METHODS
// ============================================================

// GetProfessionalTypeBySlug gets a professional type by slug
func (s *service) GetProfessionalTypeBySlug(ctx context.Context, slug string) (*authdomain.ProfessionalType, error) {
	// Repository doesn't take context, but we keep it for interface consistency
	return s.repo.GetProfessionalTypeBySlug(slug)
}

// ListProfessionalTypes lists all professional types
func (s *service) ListProfessionalTypes(ctx context.Context) ([]*authdomain.ProfessionalType, error) {
	return s.repo.ListProfessionalTypes(ctx)
}

// GetAccountTypeByID gets an account type by ID
func (s *service) GetAccountTypeByID(ctx context.Context, id string) (*authdomain.AccountType, error) {
	return s.repo.GetAccountTypeByID(id)
}

// GetProfessionalTypeByID gets a professional type by ID
func (s *service) GetProfessionalTypeByID(ctx context.Context, id string) (*authdomain.ProfessionalType, error) {
	return s.repo.GetProfessionalTypeByID(id)
}