// internal/modules/account/infrastructure/postgres/account_repository.go

package postgres

import (
    "context"
    "errors"
    "fmt"
    "time"

    "gorm.io/gorm"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/accountdomain"
)

// AccountRepository implements accountdomain.Repository using PostgreSQL
type AccountRepository struct {
    db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) accountdomain.Repository {
    return &AccountRepository{db: db}
}


func (r *AccountRepository) GetAccountIDByTeamID(ctx context.Context, teamID string) (string, error) {
    if teamID == "" {
        return "", fmt.Errorf("team ID is required")
    }

    var result struct {
        AccountID *string `gorm:"column:account_id"`
    }

    err := r.db.WithContext(ctx).
        Table("teams").
        Select("account_id").
        Where("id = ? AND deleted_at IS NULL", teamID).
        Scan(&result).Error

    if err != nil {
        return "", fmt.Errorf("failed to resolve team %s: %w", teamID, err)
    }
    if result.AccountID == nil {
        return "", nil  // team not found, or has no account
    }
    return *result.AccountID, nil
}
// ============================================================
// ACCOUNT TYPE OPERATIONS
// ============================================================

func (r *AccountRepository) GetAccountTypeByID(ctx context.Context, id string) (*accountdomain.AccountType, error) {
    if id == "" {
        return nil, accountdomain.ErrAccountTypeNotFound
    }

    var model AccountTypeModel
    if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, accountdomain.ErrAccountTypeNotFound
        }
        return nil, fmt.Errorf("failed to get account type: %w", err)
    }

    return model.ToDomain(), nil
}

func (r *AccountRepository) GetAccountTypeByName(ctx context.Context, name string) (*accountdomain.AccountType, error) {
    if name == "" {
        return nil, fmt.Errorf("name is required")
    }

    var model AccountTypeModel
    if err := r.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, accountdomain.ErrAccountTypeNotFound
        }
        return nil, fmt.Errorf("failed to get account type by name: %w", err)
    }

    return model.ToDomain(), nil
}

func (r *AccountRepository) GetAccountTypeBySlug(ctx context.Context, slug string) (*accountdomain.AccountType, error) {
    if slug == "" {
        return nil, fmt.Errorf("slug is required")
    }

    var model AccountTypeModel
    if err := r.db.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, accountdomain.ErrAccountTypeNotFound
        }
        return nil, fmt.Errorf("failed to get account type by slug: %w", err)
    }

    return model.ToDomain(), nil
}

func (r *AccountRepository) GetAccountTypes(ctx context.Context) ([]*accountdomain.AccountType, error) {
    var models []AccountTypeModel
    if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Order("sort_order ASC").Find(&models).Error; err != nil {
        return nil, fmt.Errorf("failed to get account types: %w", err)
    }

    accountTypes := make([]*accountdomain.AccountType, len(models))
    for i, model := range models {
        accountTypes[i] = model.ToDomain()
    }
    return accountTypes, nil
}

// ============================================================
// ACCOUNT OPERATIONS
// ============================================================

func (r *AccountRepository) CreateAccount(ctx context.Context, account *accountdomain.Account) error {
    if account == nil {
        return fmt.Errorf("account is nil")
    }

    model := AccountModel{}
    model.FromDomain(account)

    if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
        return fmt.Errorf("failed to create account: %w", err)
    }

    return nil
}

func (r *AccountRepository) GetAccountByID(ctx context.Context, id string) (*accountdomain.Account, error) {
	if id == "" {
		return nil, accountdomain.ErrAccountNotFound
	}

	var result struct {
		AccountModel
		AccountTypeSlug string `gorm:"column:account_type_slug"`
	}

	err := r.db.WithContext(ctx).
		Table("accounts AS a").
		Select("a.*, at.slug AS account_type_slug").
		Joins("LEFT JOIN account_types at ON at.id = a.account_type_id").
		Where("a.id = ? AND a.deleted_at IS NULL", id).
		Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	if result.ID == "" {
		return nil, accountdomain.ErrAccountNotFound
	}

	account := result.AccountModel.ToDomain()
	account.Type = result.AccountTypeSlug   // populate from the join
	return account, nil
}

func (r *AccountRepository) GetAccountByEmail(ctx context.Context, email string) (*accountdomain.Account, error) {
    if email == "" {
        return nil, fmt.Errorf("email is required")
    }

    var model AccountModel
    if err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, accountdomain.ErrAccountNotFound
        }
        return nil, fmt.Errorf("failed to get account by email: %w", err)
    }

    return model.ToDomain(), nil
}

func (r *AccountRepository) GetAccountBySlug(ctx context.Context, slug string) (*accountdomain.Account, error) {
    if slug == "" {
        return nil, fmt.Errorf("slug is required")
    }

    var model AccountModel
    if err := r.db.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, accountdomain.ErrAccountNotFound
        }
        return nil, fmt.Errorf("failed to get account by slug: %w", err)
    }

    return model.ToDomain(), nil
}

func (r *AccountRepository) GetAccountsByUserID(ctx context.Context, userID string) ([]*accountdomain.Account, error) {
    if userID == "" {
        return nil, fmt.Errorf("user ID is required")
    }

    var models []AccountModel
    if err := r.db.WithContext(ctx).
        Joins("JOIN account_members ON account_members.account_id = accounts.id").
        Where("account_members.user_id = ? AND account_members.deleted_at IS NULL AND accounts.deleted_at IS NULL", userID).
        Find(&models).Error; err != nil {
        return nil, fmt.Errorf("failed to get accounts by user: %w", err)
    }

    accounts := make([]*accountdomain.Account, len(models))
    for i, model := range models {
        accounts[i] = model.ToDomain()
    }
    return accounts, nil
}

func (r *AccountRepository) UpdateAccount(ctx context.Context, account *accountdomain.Account) error {
    if account == nil {
        return fmt.Errorf("account is nil")
    }
    if account.ID == "" {
        return accountdomain.ErrAccountNotFound
    }

    model := AccountModel{}
    model.FromDomain(account)

    if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
        return fmt.Errorf("failed to update account: %w", err)
    }

    return nil
}

func (r *AccountRepository) DeleteAccount(ctx context.Context, id string) error {
    if id == "" {
        return accountdomain.ErrAccountNotFound
    }

    // Soft delete
    result := r.db.WithContext(ctx).Model(&AccountModel{}).Where("id = ?", id).Update("deleted_at", time.Now())
    if result.Error != nil {
        return fmt.Errorf("failed to delete account: %w", result.Error)
    }

    if result.RowsAffected == 0 {
        return accountdomain.ErrAccountNotFound
    }

    return nil
}

// ============================================================
// ACCOUNT MEMBER OPERATIONS
// ============================================================

func (r *AccountRepository) CreateAccountMember(ctx context.Context, member *accountdomain.AccountMember) error {
    if member == nil {
        return fmt.Errorf("member is nil")
    }

    model := AccountMemberModel{}
    model.FromDomain(member)

    if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
        return fmt.Errorf("failed to create account member: %w", err)
    }

    return nil
}

func (r *AccountRepository) GetAccountMemberByID(ctx context.Context, id string) (*accountdomain.AccountMember, error) {
    if id == "" {
        return nil, fmt.Errorf("member ID is required")
    }

    var model AccountMemberModel
    if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, accountdomain.ErrAccountMemberNotFound
        }
        return nil, fmt.Errorf("failed to get account member: %w", err)
    }

    return model.ToDomain(), nil
}

func (r *AccountRepository) GetAccountMemberByAccountAndUser(ctx context.Context, accountID, userID string) (*accountdomain.AccountMember, error) {
    if accountID == "" || userID == "" {
        return nil, fmt.Errorf("account ID and user ID are required")
    }

    var model AccountMemberModel
    if err := r.db.WithContext(ctx).
        Where("account_id = ? AND user_id = ? AND deleted_at IS NULL", accountID, userID).
        First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, accountdomain.ErrAccountMemberNotFound
        }
        return nil, fmt.Errorf("failed to get account member: %w", err)
    }

    return model.ToDomain(), nil
}

func (r *AccountRepository) GetAccountMembersByAccount(ctx context.Context, accountID string) ([]*accountdomain.AccountMember, error) {
    if accountID == "" {
        return nil, fmt.Errorf("account ID is required")
    }

    var models []AccountMemberModel
    if err := r.db.WithContext(ctx).
        Where("account_id = ? AND deleted_at IS NULL", accountID).
        Find(&models).Error; err != nil {
        return nil, fmt.Errorf("failed to get account members: %w", err)
    }

    members := make([]*accountdomain.AccountMember, len(models))
    for i, model := range models {
        members[i] = model.ToDomain()
    }
    return members, nil
}

func (r *AccountRepository) GetAccountMembersByUser(ctx context.Context, userID string) ([]*accountdomain.AccountMember, error) {
    if userID == "" {
        return nil, fmt.Errorf("user ID is required")
    }

    var models []AccountMemberModel
    if err := r.db.WithContext(ctx).
        Where("user_id = ? AND deleted_at IS NULL", userID).
        Find(&models).Error; err != nil {
        return nil, fmt.Errorf("failed to get account memberships: %w", err)
    }

    members := make([]*accountdomain.AccountMember, len(models))
    for i, model := range models {
        members[i] = model.ToDomain()
    }
    return members, nil
}

func (r *AccountRepository) UpdateAccountMember(ctx context.Context, member *accountdomain.AccountMember) error {
    if member == nil {
        return fmt.Errorf("member is nil")
    }
    if member.ID == "" {
        return accountdomain.ErrAccountMemberNotFound
    }

    model := AccountMemberModel{}
    model.FromDomain(member)

    if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
        return fmt.Errorf("failed to update account member: %w", err)
    }

    return nil
}

func (r *AccountRepository) DeleteAccountMember(ctx context.Context, accountID, userID string) error {
    if accountID == "" || userID == "" {
        return fmt.Errorf("account ID and user ID are required")
    }

    // Soft delete
    result := r.db.WithContext(ctx).Model(&AccountMemberModel{}).
        Where("account_id = ? AND user_id = ?", accountID, userID).
        Update("deleted_at", time.Now())

    if result.Error != nil {
        return fmt.Errorf("failed to delete account member: %w", result.Error)
    }

    if result.RowsAffected == 0 {
        return accountdomain.ErrAccountMemberNotFound
    }

    return nil
}

func (r *AccountRepository) CountAccountMembers(ctx context.Context, accountID string) (int64, error) {
    if accountID == "" {
        return 0, fmt.Errorf("account ID is required")
    }

    var count int64
    if err := r.db.WithContext(ctx).Model(&AccountMemberModel{}).
        Where("account_id = ? AND deleted_at IS NULL", accountID).
        Count(&count).Error; err != nil {
        return 0, fmt.Errorf("failed to count account members: %w", err)
    }

    return count, nil
}