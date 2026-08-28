// internal/modules/events/infrastructure/postgres/repository.go

package postgres

import (
	"context"
	"errors"
	"time"
	"fmt"

	"log"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"gorm.io/gorm"
)

// ============================================================
// POSTGRES REPOSITORY - Implements domain.Repository
// ============================================================

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) domain.Repository {
	return &PostgresRepository{db: db}
}

// ============================================================
// HELPER - Get status ID by NAME (internal lookup)
// ============================================================

func (r *PostgresRepository) getStatusIDByName(ctx context.Context, name string) (string, error) {
	var model EventStatusModel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("status with name '%s' not found", name)
		}
		return "", err
	}
	return model.ID, nil
}

// ============================================================
// EVENT CRUD
// ============================================================

func (r *PostgresRepository) CreateEvent(ctx context.Context, event *domain.Event) error {
	model := ToModelEvent(event)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *PostgresRepository) GetEventByID(ctx context.Context, id string) (*domain.Event, error) {
	var model EventModel
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainEvent(&model), nil
}

func (r *PostgresRepository) GetEventByIDIncludingDeleted(ctx context.Context, id string) (*domain.Event, error) {
	var model EventModel
	err := r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainEvent(&model), nil
}

func (r *PostgresRepository) GetEventByName(ctx context.Context, name string) (*domain.Event, error) {
	var model EventModel
	err := r.db.WithContext(ctx).Where("slug = ?", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainEvent(&model), nil
}

func (r *PostgresRepository) GetEventBySlug(ctx context.Context, slug string) (*domain.Event, error) {
	var model EventModel
	err := r.db.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainEvent(&model), nil
}

func (r *PostgresRepository) UpdateEvent(ctx context.Context, event *domain.Event) error {
	model := ToModelEvent(event)
	
	log.Printf("🔍 Updating event %s: DeletedAt.Valid=%v, DeletedAt.Time=%v", 
		event.ID, model.DeletedAt.Valid, model.DeletedAt.Time)
	
	result := r.db.WithContext(ctx).Unscoped().
		Model(&EventModel{}).
		Where("id = ?", event.ID).
		Select("*").
		Updates(model)
	
	if result.Error != nil {
		return result.Error
	}
	
	log.Printf("✅ Event updated: %s, Rows affected: %d", event.ID, result.RowsAffected)
	return nil
}

func (r *PostgresRepository) DeleteEvent(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&EventModel{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}

func (r *PostgresRepository) PermanentlyDeleteEvent(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&EventModel{}, "id = ?", id).Error
}

// ============================================================
// EVENT QUERIES
// ============================================================

func (r *PostgresRepository) ListEvents(ctx context.Context, filters domain.ListEventsFilters) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	query := r.db.WithContext(ctx).Unscoped().Model(&EventModel{})

	// Handle deleted filter logic
	if filters.OnlyDeleted {
		query = query.Where("deleted_at IS NOT NULL")
	} else if !filters.IncludeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	if filters.InstitutionID != "" {
		query = query.Where("institution_id = ?", filters.InstitutionID)
	}
	if filters.UserID != "" {
		query = query.Where("created_by = ?", filters.UserID)
	}
	if filters.EventTypeID != "" {
		query = query.Where("event_type_id = ?", filters.EventTypeID)
	}
	if filters.EventStatusID != "" {
		query = query.Where("event_status_id = ?", filters.EventStatusID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	err := query.Order("date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = ToDomainEvent(&m)
	}
	return events, total, nil
}

func (r *PostgresRepository) GetEventsByType(ctx context.Context, eventTypeID string, limit, offset int) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	query := r.db.WithContext(ctx).Model(&EventModel{}).
		Where("event_type_id = ? AND deleted_at IS NULL", eventTypeID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Order("date DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	events := make([]*domain.Event, len(models))
	for i, model := range models {
		events[i] = ToDomainEvent(&model)
	}
	return events, total, nil
}

// ✅ GetEventsByInstitution - all events for an organization
func (r *PostgresRepository) GetEventsByInstitution(ctx context.Context, institutionID string, limit, offset int) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	query := r.db.WithContext(ctx).Model(&EventModel{}).
		Where("institution_id = ? AND deleted_at IS NULL", institutionID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = ToDomainEvent(&m)
	}
	return events, total, nil
}

// ✅ GetEventsByUser - all events created by a specific user
func (r *PostgresRepository) GetEventsByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	query := r.db.WithContext(ctx).Model(&EventModel{}).
		Where("created_by = ? AND deleted_at IS NULL", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = ToDomainEvent(&m)
	}
	return events, total, nil
}

func (r *PostgresRepository) GetUpcomingEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	var models []EventModel

	publishedStatusID, err := r.getStatusIDByName(ctx, domain.EventStatusPublished.GetName())
	if err != nil {
		log.Printf("❌ Failed to get published status: %v", err)
		return nil, fmt.Errorf("failed to get published status: %w", err)
	}

	query := r.db.WithContext(ctx).Model(&EventModel{}).
		Where("date >= CURRENT_DATE AND deleted_at IS NULL").
		Where("event_status_id = ?", publishedStatusID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Order("date ASC").Find(&models).Error
	if err != nil {
		log.Printf("❌ Query error: %v", err)
		return nil, err
	}

	log.Printf("✅ Found %d upcoming published events", len(models))

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = ToDomainEvent(&m)
	}
	return events, nil
}

func (r *PostgresRepository) GetPastEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	var models []EventModel

	publishedStatusID, err := r.getStatusIDByName(ctx, domain.EventStatusPublished.GetName())
	if err != nil {
		log.Printf("❌ Failed to get published status: %v", err)
		return nil, fmt.Errorf("failed to get published status: %w", err)
	}

	query := r.db.WithContext(ctx).Model(&EventModel{}).
		Where("date < CURRENT_DATE AND deleted_at IS NULL").
		Where("event_status_id = ?", publishedStatusID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Order("date DESC").Find(&models).Error
	if err != nil {
		log.Printf("❌ Query error: %v", err)
		return nil, err
	}

	log.Printf("✅ Found %d past published events", len(models))

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = ToDomainEvent(&m)
	}
	return events, nil
}

func (r *PostgresRepository) SearchEvents(ctx context.Context, query string, filters domain.SearchFilters) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	dbQuery := r.db.WithContext(ctx).Unscoped().Model(&EventModel{})

	if filters.OnlyDeleted {
		dbQuery = dbQuery.Where("deleted_at IS NOT NULL")
	} else if !filters.IncludeDeleted {
		dbQuery = dbQuery.Where("deleted_at IS NULL")
	}

	if query != "" {
		searchTerm := "%" + query + "%"
		dbQuery = dbQuery.Where("name ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}
	if filters.InstitutionID != "" {
		dbQuery = dbQuery.Where("institution_id = ?", filters.InstitutionID)
	}
	if filters.UserID != "" {
		dbQuery = dbQuery.Where("created_by = ?", filters.UserID)
	}
	if filters.EventTypeID != "" {
		dbQuery = dbQuery.Where("event_type_id = ?", filters.EventTypeID)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filters.Limit > 0 {
		dbQuery = dbQuery.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		dbQuery = dbQuery.Offset(filters.Offset)
	}

	err := dbQuery.Order("date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = ToDomainEvent(&m)
	}
	return events, total, nil
}

// ============================================================
// EVENT QUERIES WITH CREATOR INFO
// ============================================================

func (r *PostgresRepository) GetUpcomingEventsWithCreator(ctx context.Context, limit int) ([]*domain.Event, error) {
	var models []EventModel

	publishedStatusID, err := r.getStatusIDByName(ctx, domain.EventStatusPublished.GetName())
	if err != nil {
		log.Printf("❌ Failed to get published status: %v", err)
		return nil, fmt.Errorf("failed to get published status: %w", err)
	}

	query := r.db.WithContext(ctx).Table("events").
		Select(`events.*,
				users.name as creator_name,
				users.display_name as creator_display_name,
				users.email as creator_email,
				users.phone as creator_phone,
				COALESCE(institutions.name, '') as creator_institution_name,
				COALESCE(account_types.slug, 'personal') as creator_account_type`).
		Joins("LEFT JOIN users ON events.created_by = users.id").
		Joins("LEFT JOIN institutions ON events.institution_id = institutions.id").
		Joins("LEFT JOIN account_types ON users.account_type_id = account_types.id").
		Where("events.date >= CURRENT_DATE AND events.deleted_at IS NULL").
		Where("events.event_status_id = ?", publishedStatusID).
		Order("events.date ASC, events.time ASC").
		Limit(limit)

	err = query.Find(&models).Error
	if err != nil {
		log.Printf("❌ Query error: %v", err)
		return nil, err
	}

	log.Printf("✅ Found %d upcoming published events with creator info", len(models))
	return toDomainEventsWithCreator(models), nil
}

func (r *PostgresRepository) GetEventBySlugWithCreator(ctx context.Context, slug string) (*domain.Event, error) {
	var model EventModel

	err := r.db.WithContext(ctx).Table("events").
		Select(`events.*,
				users.name as creator_name,
				users.display_name as creator_display_name,
				users.email as creator_email,
				users.phone as creator_phone,
				COALESCE(institutions.name, '') as creator_institution_name,
				COALESCE(account_types.slug, 'personal') as creator_account_type`).
		Joins("LEFT JOIN users ON events.created_by = users.id").
		Joins("LEFT JOIN institutions ON events.institution_id = institutions.id").
		Joins("LEFT JOIN account_types ON users.account_type_id = account_types.id").
		Where("events.slug = ? AND events.deleted_at IS NULL", slug).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	events := toDomainEventsWithCreator([]EventModel{model})
	if len(events) == 0 {
		return nil, nil
	}
	return events[0], nil
}

func (r *PostgresRepository) GetEventByNameWithCreator(ctx context.Context, name string) (*domain.Event, error) {
	var model EventModel

	err := r.db.WithContext(ctx).Table("events").
		Select(`events.*,
				users.name as creator_name,
				users.display_name as creator_display_name,
				users.email as creator_email,
				users.phone as creator_phone,
				COALESCE(institutions.name, '') as creator_institution_name,
				COALESCE(account_types.slug, 'personal') as creator_account_type`).
		Joins("LEFT JOIN users ON events.created_by = users.id").
		Joins("LEFT JOIN institutions ON events.institution_id = institutions.id").
		Joins("LEFT JOIN account_types ON users.account_type_id = account_types.id").
		Where("events.name = ? AND events.deleted_at IS NULL", name).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	events := toDomainEventsWithCreator([]EventModel{model})
	if len(events) == 0 {
		return nil, nil
	}
	return events[0], nil
}

func (r *PostgresRepository) GetEventByIDWithCreator(ctx context.Context, id string) (*domain.Event, error) {
	var model EventModel

	err := r.db.WithContext(ctx).Table("events").
		Select(`events.*,
				users.name as creator_name,
				users.display_name as creator_display_name,
				users.email as creator_email,
				users.phone as creator_phone,
				COALESCE(institutions.name, '') as creator_institution_name,
				COALESCE(account_types.slug, 'personal') as creator_account_type`).
		Joins("LEFT JOIN users ON events.created_by = users.id").
		Joins("LEFT JOIN institutions ON events.institution_id = institutions.id").
		Joins("LEFT JOIN account_types ON users.account_type_id = account_types.id").
		Where("events.id = ? AND events.deleted_at IS NULL", id).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	events := toDomainEventsWithCreator([]EventModel{model})
	if len(events) == 0 {
		return nil, nil
	}
	return events[0], nil
}

// ✅ GetEventsByInstitutionWithCreator - events for an institution with creator info
func (r *PostgresRepository) GetEventsByInstitutionWithCreator(ctx context.Context, institutionID string, limit, offset int) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	query := r.db.WithContext(ctx).Table("events").
		Select(`events.*,
				users.name as creator_name,
				users.display_name as creator_display_name,
				users.email as creator_email,
				users.phone as creator_phone,
				COALESCE(institutions.name, '') as creator_institution_name,
				COALESCE(account_types.slug, 'personal') as creator_account_type`).
		Joins("LEFT JOIN users ON events.created_by = users.id").
		Joins("LEFT JOIN institutions ON events.institution_id = institutions.id").
		Joins("LEFT JOIN account_types ON users.account_type_id = account_types.id").
		Where("events.institution_id = ? AND events.deleted_at IS NULL", institutionID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("events.date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	return toDomainEventsWithCreator(models), total, nil
}

// ✅ GetEventsByUserWithCreator - events created by a user with creator info
func (r *PostgresRepository) GetEventsByUserWithCreator(ctx context.Context, userID string, limit, offset int) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	query := r.db.WithContext(ctx).Table("events").
		Select(`events.*,
				users.name as creator_name,
				users.display_name as creator_display_name,
				users.email as creator_email,
				users.phone as creator_phone,
				COALESCE(institutions.name, '') as creator_institution_name,
				COALESCE(account_types.slug, 'personal') as creator_account_type`).
		Joins("LEFT JOIN users ON events.created_by = users.id").
		Joins("LEFT JOIN institutions ON events.institution_id = institutions.id").
		Joins("LEFT JOIN account_types ON users.account_type_id = account_types.id").
		Where("events.created_by = ? AND events.deleted_at IS NULL", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("events.date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	return toDomainEventsWithCreator(models), total, nil
}

// ============================================================
// EVENT QUERIES WITH CREATOR INFO AND FILTERS (NEW)
// ============================================================

// GetEventsByUserWithCreatorFiltered returns events with creator info, with privacy/status filters
// includePrivate: true to include private events (for team members/owners)
func (r *PostgresRepository) GetEventsByUserWithCreatorFiltered(ctx context.Context, userID string, includePrivate bool, limit, offset int) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	// Get published and completed status IDs
	publishedStatusID, err := r.getStatusIDByName(ctx, domain.EventStatusPublished.GetName())
	if err != nil {
		log.Printf("❌ Failed to get published status: %v", err)
		return nil, 0, fmt.Errorf("failed to get published status: %w", err)
	}

	completedStatusID, err := r.getStatusIDByName(ctx, domain.EventStatusCompleted.GetName())
	if err != nil {
		log.Printf("❌ Failed to get completed status: %v", err)
		return nil, 0, fmt.Errorf("failed to get completed status: %w", err)
	}

	query := r.db.WithContext(ctx).Table("events").
		Select(`events.*,
				users.name as creator_name,
				users.display_name as creator_display_name,
				users.email as creator_email,
				users.phone as creator_phone,
				COALESCE(institutions.name, '') as creator_institution_name,
				COALESCE(account_types.slug, 'personal') as creator_account_type`).
		Joins("LEFT JOIN users ON events.created_by = users.id").
		Joins("LEFT JOIN institutions ON events.institution_id = institutions.id").
		Joins("LEFT JOIN account_types ON users.account_type_id = account_types.id").
		Where("events.created_by = ? AND events.deleted_at IS NULL", userID).
		Where("events.event_status_id IN (?)", []string{publishedStatusID, completedStatusID})

	// Only include private events if viewer has permission
	if !includePrivate {
		query = query.Where("events.is_private = ?", false)
	}

	// Count total
	countQuery := r.db.WithContext(ctx).Table("events").
		Where("created_by = ? AND deleted_at IS NULL", userID).
		Where("event_status_id IN (?)", []string{publishedStatusID, completedStatusID})

	if !includePrivate {
		countQuery = countQuery.Where("is_private = ?", false)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err = query.Order("events.date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	return toDomainEventsWithCreator(models), total, nil
}

// GetEventsByUserWithCreatorPublic returns only public events (published, non-private)
func (r *PostgresRepository) GetEventsByUserWithCreatorPublic(ctx context.Context, userID string, limit, offset int) ([]*domain.Event, int64, error) {
	return r.GetEventsByUserWithCreatorFiltered(ctx, userID, false, limit, offset)
}

// ============================================================
// USER INFO
// ============================================================

func (r *PostgresRepository) GetUserInfoByID(ctx context.Context, userID string) (*domain.UserInfo, error) {
	var model UserModel
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.*, institutions.name as institution_name").
		Joins("LEFT JOIN institutions ON users.institution_id = institutions.id").
		Where("users.id = ? AND users.is_active = ?", userID, true).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainUserInfo(&model), nil
}

func (r *PostgresRepository) GetUserInfoByIDs(ctx context.Context, userIDs []string) ([]*domain.UserInfo, error) {
	if len(userIDs) == 0 {
		return []*domain.UserInfo{}, nil
	}

	var models []UserModel
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.*, institutions.name as institution_name").
		Joins("LEFT JOIN institutions ON users.institution_id = institutions.id").
		Where("users.id IN ? AND users.is_active = ?", userIDs, true).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	userInfos := make([]*domain.UserInfo, len(models))
	for i, m := range models {
		userInfos[i] = ToDomainUserInfo(&m)
	}
	return userInfos, nil
}

// ============================================================
// EVENT TYPE OPERATIONS
// ============================================================

func (r *PostgresRepository) GetEventTypeByID(ctx context.Context, id string) (*domain.EventType, error) {
	var model EventTypeModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainEventTypeEntity(&model), nil
}

func (r *PostgresRepository) GetEventTypeByName(ctx context.Context, name string) (*domain.EventType, error) {
	var model EventTypeModel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainEventTypeEntity(&model), nil
}

func (r *PostgresRepository) GetEventTypeBySlug(ctx context.Context, slug string) (*domain.EventType, error) {
	var model EventTypeModel
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainEventTypeEntity(&model), nil
}

func (r *PostgresRepository) GetAllEventTypes(ctx context.Context) ([]*domain.EventType, error) {
	var models []EventTypeModel
	err := r.db.WithContext(ctx).
		Order("sort_order ASC, display_name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	types := make([]*domain.EventType, len(models))
	for i, m := range models {
		types[i] = ToDomainEventTypeEntity(&m)
	}
	return types, nil
}

// ============================================================
// EVENT STATUS OPERATIONS
// ============================================================

func (r *PostgresRepository) GetEventStatusByID(ctx context.Context, id string) (*domain.EventStatus, error) {
	var model EventStatusModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainEventStatusEntity(&model), nil
}

func (r *PostgresRepository) GetEventStatusByName(ctx context.Context, name string) (*domain.EventStatus, error) {
	var model EventStatusModel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainEventStatusEntity(&model), nil
}

func (r *PostgresRepository) GetEventStatusBySlug(ctx context.Context, slug string) (*domain.EventStatus, error) {
	var model EventStatusModel
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return ToDomainEventStatusEntity(&model), nil
}

func (r *PostgresRepository) GetAllEventStatuses(ctx context.Context) ([]*domain.EventStatus, error) {
	var models []EventStatusModel
	err := r.db.WithContext(ctx).
		Order("sort_order ASC, display_name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	statuses := make([]*domain.EventStatus, len(models))
	for i, m := range models {
		statuses[i] = ToDomainEventStatusEntity(&m)
	}
	return statuses, nil
}