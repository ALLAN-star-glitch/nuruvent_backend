// internal/modules/auth/domain/repository.go

package domain

import (
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/google/uuid"
)

// Repository defines the data access interface
type Repository interface {
	// ================================================
	// ACCOUNT OPERATIONS
	// ================================================
	
	// Account existence checks
	AccountExistsByEmail(email string) (bool, error)
	AccountExistsByPhone(phone string) (bool, error)
	
	// Account retrieval
	GetAccountByEmail(email string) (*models.Account, error)
	GetAccountByPhone(phone string) (*models.Account, error)
	GetAccountByID(id uuid.UUID) (*models.Account, error)
	
	// Account CRUD
	CreateAccount(account *models.Account) error
	UpdateAccount(account *models.Account) error

	// Account type
	GetAccountTypeBySlug(slug string) (*models.AccountType, error)

	// ================================================
	// REFRESH TOKEN OPERATIONS
	// ================================================
	
	CreateRefreshToken(token *models.RefreshToken) error
	GetRefreshTokenByToken(token string) (*models.RefreshToken, error)
	RevokeRefreshToken(token string) error
	RevokeAllAccountRefreshTokens(accountID string) error

	// ================================================
	// INSTITUTION OPERATIONS
	// ================================================
	
	CreateInstitution(institution *models.Institution) error
	GetInstitutionTypeBySlug(slug string) (*models.InstitutionType, error)

	// ================================================
	// TEAM MEMBER OPERATIONS
	// ================================================
	
	CreateTeamMember(member *models.TeamMember) error
}