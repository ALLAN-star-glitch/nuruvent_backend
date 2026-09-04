// internal/modules/profile/infrastructure/postgres/repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/profile/domain"
)

type ProfileRepository struct {
    db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) domain.Repository {
    return &ProfileRepository{db: db}
}

// ============================================================
// USER METHODS
// ============================================================

func (r *ProfileRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
    if id == "" {
        return nil, domain.ErrInvalidUserID
    }

    var model UserModel
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    return model.ToDomain(), nil
}

func (r *ProfileRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
    if email == "" {
        return nil, errors.New("email is required")
    }

    var model UserModel
    if err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user by email: %w", err)
    }

    return model.ToDomain(), nil
}


func (r *ProfileRepository) UpdateUser(ctx context.Context, user *domain.User) error {
    if user == nil {
        return errors.New("user is nil")
    }
    if user.ID == "" {
        return domain.ErrInvalidUserID
    }

    // ✅ Update all profile fields now
    updates := map[string]interface{}{
        "name":          user.Name,
        "display_name":  user.DisplayName,
        "phone":         user.Phone,
        "avatar_url":    user.AvatarURL,
        "bio":           user.Bio,
        "location":      user.Location,
        "website":       user.Website,
        "social_links":  user.SocialLinks,
        "updated_at":    time.Now(),
    }

    result := r.db.WithContext(ctx).Model(&UserModel{}).
        Where("id = ?", user.ID).
        Updates(updates)

    if result.Error != nil {
        return fmt.Errorf("failed to update user: %w", result.Error)
    }

    if result.RowsAffected == 0 {
        return domain.ErrUserNotFound
    }

    return nil
}

func (r *ProfileRepository) GetUsersByIDs(ctx context.Context, ids []string) ([]*domain.User, error) {
    if len(ids) == 0 {
        return []*domain.User{}, nil
    }

    var models []UserModel
    if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&models).Error; err != nil {
        return nil, fmt.Errorf("failed to get users: %w", err)
    }

    users := make([]*domain.User, len(models))
    for i, model := range models {
        users[i] = model.ToDomain()
    }
    return users, nil
}

func (r *ProfileRepository) ListUsers(ctx context.Context, filters domain.ListUsersFilters) ([]*domain.User, int64, error) {
    query := r.db.WithContext(ctx).Model(&UserModel{})

    // Apply team filter
    switch filters.Team.Type {
    case "personal":
        query = query.Where("id = ?", filters.Team.ID)
    case "institution":
        // For institution team, we need to find users who are members of this institution
        // This would require a join with team_members table
        // For now, we'll just filter by institution_id if available
        query = query.Where("institution_id = ?", filters.Team.ID)
    }

    // Apply filters
    if filters.UserID != "" {
        query = query.Where("id = ?", filters.UserID)
    }
    if filters.Search != "" {
        searchTerm := "%" + filters.Search + "%"
        query = query.Where("name ILIKE ? OR email ILIKE ?", searchTerm, searchTerm)
    }

    // Soft delete filters
    if filters.OnlyDeleted {
        query = query.Unscoped().Where("deleted_at IS NOT NULL")
    } else if !filters.IncludeDeleted {
        query = query.Where("deleted_at IS NULL")
    }

    // Count total
    var total int64
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to count users: %w", err)
    }

    // Pagination
    if filters.Limit > 0 {
        query = query.Limit(filters.Limit)
        if filters.Offset > 0 {
            query = query.Offset(filters.Offset)
        }
    }

    // Sorting
    sortField := "created_at"
    if filters.SortBy != "" {
        sortField = filters.SortBy
    }
    sortOrder := "ASC"
    if filters.SortOrder == "desc" {
        sortOrder = "DESC"
    }
    query = query.Order(sortField + " " + sortOrder)

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

// ============================================================
// INSTITUTION METHODS
// ============================================================

func (r *ProfileRepository) GetInstitutionByID(ctx context.Context, id string) (*domain.Institution, error) {
    if id == "" {
        return nil, domain.ErrInvalidInstitutionID
    }

    var model InstitutionModel
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrInstitutionNotFound
        }
        return nil, fmt.Errorf("failed to get institution: %w", err)
    }

    return model.ToDomain(), nil
}

func (r *ProfileRepository) GetInstitutionBySlug(ctx context.Context, slug string) (*domain.Institution, error) {
    if slug == "" {
        return nil, errors.New("slug is required")
    }

    var model InstitutionModel
    if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrInstitutionNotFound
        }
        return nil, fmt.Errorf("failed to get institution by slug: %w", err)
    }

    return model.ToDomain(), nil
}

func (r *ProfileRepository) UpdateInstitution(ctx context.Context, institution *domain.Institution) error {
    if institution == nil {
        return errors.New("institution is nil")
    }
    if institution.ID == "" {
        return domain.ErrInvalidInstitutionID
    }

    var model InstitutionModel
    model.FromDomain(institution)

    if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
        return fmt.Errorf("failed to update institution: %w", err)
    }

    return nil
}

func (r *ProfileRepository) GetInstitutionsByIDs(ctx context.Context, ids []string) ([]*domain.Institution, error) {
    if len(ids) == 0 {
        return []*domain.Institution{}, nil
    }

    var models []InstitutionModel
    if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&models).Error; err != nil {
        return nil, fmt.Errorf("failed to get institutions: %w", err)
    }

    institutions := make([]*domain.Institution, len(models))
    for i, model := range models {
        institutions[i] = model.ToDomain()
    }
    return institutions, nil
}

func (r *ProfileRepository) ListInstitutions(ctx context.Context, filters domain.ListInstitutionsFilters) ([]*domain.Institution, int64, error) {
    query := r.db.WithContext(ctx).Model(&InstitutionModel{})

    // Apply team filter
    if filters.Team.Type == "institution" {
        query = query.Where("id = ?", filters.Team.ID)
    }

    // Apply filters
    if filters.InstitutionID != "" {
        query = query.Where("id = ?", filters.InstitutionID)
    }
    if filters.Search != "" {
        searchTerm := "%" + filters.Search + "%"
        query = query.Where("name ILIKE ? OR email ILIKE ?", searchTerm, searchTerm)
    }

    // Soft delete filters
    if filters.OnlyDeleted {
        query = query.Unscoped().Where("deleted_at IS NOT NULL")
    } else if !filters.IncludeDeleted {
        query = query.Where("deleted_at IS NULL")
    }

    // Count total
    var total int64
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to count institutions: %w", err)
    }

    // Pagination
    if filters.Limit > 0 {
        query = query.Limit(filters.Limit)
        if filters.Offset > 0 {
            query = query.Offset(filters.Offset)
        }
    }

    // Sorting
    sortField := "created_at"
    if filters.SortBy != "" {
        sortField = filters.SortBy
    }
    sortOrder := "ASC"
    if filters.SortOrder == "desc" {
        sortOrder = "DESC"
    }
    query = query.Order(sortField + " " + sortOrder)

    var models []InstitutionModel
    if err := query.Find(&models).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to list institutions: %w", err)
    }

    institutions := make([]*domain.Institution, len(models))
    for i, model := range models {
        institutions[i] = model.ToDomain()
    }

    return institutions, total, nil
}