// internal/modules/auth/service/service.go

package service

import (
	"context"
	"fmt"

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

	// ============================================================
	// USER QUERIES (For Team Module)
	// ============================================================

	GetUserByID(ctx context.Context, userID string) (*authdomain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*authdomain.User, error)
	UserExists(ctx context.Context, email string) (bool, error)
	GetTokenContext(ctx context.Context, user *authdomain.User) (*authdomain.TokenContext, error)

	// ✅ GetUserByIDWithAccount retrieves a user by ID with their account ID
	// Returns: user, accountID, error
	GetUserByIDWithAccount(ctx context.Context, userID string) (*authdomain.User, string, error)

	// ✅ GetUserByEmailWithAccount retrieves a user by email with their account ID
	// Returns: user, accountID, error
	GetUserByEmailWithAccount(ctx context.Context, email string) (*authdomain.User, string, error)

	// ✅ GetAccountIDByUserID gets the account ID for a user
	GetAccountIDByUserID(ctx context.Context, userID string) (string, error)

	AddAccountMember(ctx context.Context, accountID, userID, role string) error
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

	InviteToken string

	// Professional Type (for personal accounts)
	ProfessionalType string

	// Institution fields (for institution accounts)
	InstitutionName  string
	InstitutionEmail string
	InstitutionPhone string
	InstitutionType  string
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
	teamSvc       TeamService
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
	teamSvc TeamService,
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
		teamSvc:       teamSvc,
	}
}

// ============================================================
// PROFESSIONAL TYPE METHODS
// ============================================================

func (s *service) GetProfessionalTypeBySlug(ctx context.Context, slug string) (*authdomain.ProfessionalType, error) {
	return s.repo.GetProfessionalTypeBySlug(ctx, slug)
}

func (s *service) ListProfessionalTypes(ctx context.Context) ([]*authdomain.ProfessionalType, error) {
	return s.repo.ListProfessionalTypes(ctx)
}

func (s *service) GetAccountTypeByID(ctx context.Context, id string) (*authdomain.AccountType, error) {
	return s.repo.GetAccountTypeByID(ctx, id)
}

func (s *service) GetProfessionalTypeByID(ctx context.Context, id string) (*authdomain.ProfessionalType, error) {
	return s.repo.GetProfessionalTypeByID(ctx, id)
}

func (s *service) AddAccountMember(ctx context.Context, accountID, userID, role string) error {
	if accountID == "" || userID == "" || role == "" {
		return fmt.Errorf("accountID, userID, and role are required")
	}
	if !authdomain.IsAccountRole(role) {
		return fmt.Errorf("invalid role: %q", role)
	}
	member, err := authdomain.NewAccountMember(accountID, userID, role, userID)
	if err != nil {
		return err
	}
	return s.repo.CreateAccountMember(ctx, member)
}