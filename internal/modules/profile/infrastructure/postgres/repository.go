// internal/modules/profile/infrastructure/postgres/repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
	"gorm.io/gorm"
)

type ProfileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) domain.Repository {
	return &ProfileRepository{db: db}
}

// ============================================================
// USER CRUD OPERATIONS
// ============================================================

func (r *ProfileRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return model.ToDomain(), nil
}

func (r *ProfileRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return model.ToDomain(), nil
}

func (r *ProfileRepository) GetUsersByIDs(ctx context.Context, ids []string) ([]*domain.User, error) {
	if len(ids) == 0 {
		return []*domain.User{}, nil
	}

	var models []UserModel
	err := r.db.WithContext(ctx).
		Where("id IN ? AND deleted_at IS NULL", ids).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get users by IDs: %w", err)
	}

	users := make([]*domain.User, len(models))
	for i, model := range models {
		users[i] = model.ToDomain()
	}
	return users, nil
}

func (r *ProfileRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	var model UserModel
	model.FromDomain(user)

	result := r.db.WithContext(ctx).
		Model(&UserModel{}).
		Where("id = ?", user.ID).
		Updates(&model)

	if result.Error != nil {
		return fmt.Errorf("failed to update user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// ============================================================
// ACCOUNT CRUD OPERATIONS
// ============================================================

func (r *ProfileRepository) GetAccountByID(ctx context.Context, id string) (*domain.Account, error) {
	var model AccountModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get account by ID: %w", err)
	}

	return model.ToDomain(), nil
}

func (r *ProfileRepository) GetAccountBySlug(ctx context.Context, slug string) (*domain.Account, error) {
	var model AccountModel
	err := r.db.WithContext(ctx).
		Where("slug = ? AND deleted_at IS NULL", slug).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get account by slug: %w", err)
	}

	return model.ToDomain(), nil
}

func (r *ProfileRepository) GetAccountsByIDs(ctx context.Context, ids []string) ([]*domain.Account, error) {
	if len(ids) == 0 {
		return []*domain.Account{}, nil
	}

	var models []AccountModel
	err := r.db.WithContext(ctx).
		Where("id IN ? AND deleted_at IS NULL", ids).
		Find(&models).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by IDs: %w", err)
	}

	accounts := make([]*domain.Account, len(models))
	for i, model := range models {
		accounts[i] = model.ToDomain()
	}
	return accounts, nil
}

func (r *ProfileRepository) UpdateAccount(ctx context.Context, account *domain.Account) error {
	var model AccountModel
	model.FromDomain(account)

	result := r.db.WithContext(ctx).
		Model(&AccountModel{}).
		Where("id = ?", account.ID).
		Updates(&model)

	if result.Error != nil {
		return fmt.Errorf("failed to update account: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

// ============================================================
// QUERY OPERATIONS
// ============================================================

func (r *ProfileRepository) ListUsers(ctx context.Context, filters domain.ListUsersFilters) ([]*domain.User, int64, error) {
	query := r.db.WithContext(ctx).Model(&UserModel{}).Where("deleted_at IS NULL")

	// Apply filters
	if filters.UserID != "" {
		query = query.Where("id = ?", filters.UserID)
	}

	if filters.Search != "" {
		searchTerm := "%" + filters.Search + "%"
		query = query.Where("name ILIKE ? OR display_name ILIKE ? OR email ILIKE ?",
			searchTerm, searchTerm, searchTerm)
	}

	if filters.IncludeDeleted {
		query = query.Unscoped()
	}

	if filters.OnlyDeleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	}

	// Team filtering - this is a simplified version
	// For personal teams: just the user themselves
	if filters.Team.Type == "personal" && filters.Team.ID != "" {
		query = query.Where("id = ?", filters.Team.ID)
	}

	// For institution teams: users who are members of the account
	if filters.Team.Type == "institution" && filters.Team.ID != "" {
		// This requires a join with account_members table
		query = query.Joins("JOIN account_members ON account_members.user_id = users.id").
			Where("account_members.account_id = ? AND account_members.deleted_at IS NULL", filters.Team.ID)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Apply sorting
	if filters.SortBy != "" {
		sortOrder := "ASC"
		if filters.SortOrder == "desc" {
			sortOrder = "DESC"
		}
		query = query.Order(filters.SortBy + " " + sortOrder)
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Execute query
	var models []UserModel
	if err := query.Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]*domain.User, len(models))
	for i, model := range models {
		users[i] = model.ToDomain()
	}

	return users, total, nil
}

func (r *ProfileRepository) ListAccounts(ctx context.Context, filters domain.ListAccountsFilters) ([]*domain.Account, int64, error) {
	query := r.db.WithContext(ctx).Model(&AccountModel{}).Where("deleted_at IS NULL")

	// Apply filters
	if filters.AccountID != "" {
		query = query.Where("id = ?", filters.AccountID)
	}

	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.KYCStatus != "" {
		query = query.Where("kyc_status = ?", filters.KYCStatus)
	}

	if filters.Search != "" {
		searchTerm := "%" + filters.Search + "%"
		query = query.Where("name ILIKE ? OR display_name ILIKE ? OR email ILIKE ?",
			searchTerm, searchTerm, searchTerm)
	}

	if filters.IncludeDeleted {
		query = query.Unscoped()
	}

	if filters.OnlyDeleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count accounts: %w", err)
	}

	// Apply sorting
	if filters.SortBy != "" {
		sortOrder := "ASC"
		if filters.SortOrder == "desc" {
			sortOrder = "DESC"
		}
		query = query.Order(filters.SortBy + " " + sortOrder)
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Execute query
	var models []AccountModel
	if err := query.Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list accounts: %w", err)
	}

	accounts := make([]*domain.Account, len(models))
	for i, model := range models {
		accounts[i] = model.ToDomain()
	}

	return accounts, total, nil
}