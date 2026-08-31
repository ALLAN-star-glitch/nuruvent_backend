// internal/modules/events/infrastructure/postgres/repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

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

// internal/modules/events/postgres/repository.go

func (r *PostgresRepository) CreateEvent(ctx context.Context, event *domain.Event) error {
    model := toModelEvent(event)
    
    // ✅ Add debug logging
    log.Printf("🔍 Creating event with model: %+v", model)
    log.Printf("🔍 Schedules: %+v", event.Schedules)
    log.Printf("🔍 Tickets: %+v", event.Tickets)
    log.Printf("🔍 Speakers: %+v", event.Speakers)
    log.Printf("🔍 Materials: %+v", event.Materials)
    
    // Create the main event
    if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
        log.Printf("❌ Failed to create event: %v", err)
        return err
    }

    // Save child entities
    if err := r.saveSchedules(ctx, event.ID, event.Schedules); err != nil {
        return err
    }
    if err := r.saveTickets(ctx, event.ID, event.Tickets); err != nil {
        return err
    }
    if err := r.saveSpeakers(ctx, event.ID, event.Speakers); err != nil {
        return err
    }
    if err := r.saveMaterials(ctx, event.ID, event.Materials); err != nil {
        return err
    }

    return nil
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

	event := toDomainEvent(&model)
	if event == nil {
		return nil, nil
	}

	// Load child entities
	if err := r.loadChildEntities(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
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

	event := toDomainEvent(&model)
	if event == nil {
		return nil, nil
	}

	// Load child entities
	if err := r.loadChildEntities(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (r *PostgresRepository) GetEventByName(ctx context.Context, name string) (*domain.Event, error) {
	var model EventModel
	err := r.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	event := toDomainEvent(&model)
	if event == nil {
		return nil, nil
	}

	if err := r.loadChildEntities(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
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

	event := toDomainEvent(&model)
	if event == nil {
		return nil, nil
	}

	if err := r.loadChildEntities(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (r *PostgresRepository) UpdateEvent(ctx context.Context, event *domain.Event) error {
	model := toModelEvent(event)

	// Update main event
	result := r.db.WithContext(ctx).Unscoped().
		Model(&EventModel{}).
		Where("id = ?", event.ID).
		Select("*").
		Updates(model)

	if result.Error != nil {
		return result.Error
	}

	// Update child entities - delete and recreate
	if err := r.deleteSchedules(ctx, event.ID); err != nil {
		return err
	}
	if err := r.saveSchedules(ctx, event.ID, event.Schedules); err != nil {
		return err
	}

	if err := r.deleteTickets(ctx, event.ID); err != nil {
		return err
	}
	if err := r.saveTickets(ctx, event.ID, event.Tickets); err != nil {
		return err
	}

	if err := r.deleteSpeakers(ctx, event.ID); err != nil {
		return err
	}
	if err := r.saveSpeakers(ctx, event.ID, event.Speakers); err != nil {
		return err
	}

	if err := r.deleteMaterials(ctx, event.ID); err != nil {
		return err
	}
	if err := r.saveMaterials(ctx, event.ID, event.Materials); err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) DeleteEvent(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&EventModel{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}

func (r *PostgresRepository) PermanentlyDeleteEvent(ctx context.Context, id string) error {
	// Delete child entities first
	if err := r.deleteSchedules(ctx, id); err != nil {
		return err
	}
	if err := r.deleteTickets(ctx, id); err != nil {
		return err
	}
	if err := r.deleteSpeakers(ctx, id); err != nil {
		return err
	}
	if err := r.deleteMaterials(ctx, id); err != nil {
		return err
	}

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
	if filters.CategoryID != "" {
		query = query.Where("category_id = ?", filters.CategoryID)
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

	// Use start_date for sorting
	sortField := "start_date"
	if filters.SortBy != "" {
		sortField = filters.SortBy
	}
	sortOrder := "DESC"
	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query = query.Order(sortField + " " + sortOrder)

	err := query.Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = toDomainEvent(&m)
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

	if err := query.Order("start_date DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	events := make([]*domain.Event, len(models))
	for i, model := range models {
		events[i] = toDomainEvent(&model)
	}
	return events, total, nil
}

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

	err := query.Order("start_date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = toDomainEvent(&m)
	}
	return events, total, nil
}

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

	err := query.Order("start_date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = toDomainEvent(&m)
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
		Where("start_date >= CURRENT_DATE AND deleted_at IS NULL").
		Where("event_status_id = ?", publishedStatusID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Order("start_date ASC").Find(&models).Error
	if err != nil {
		log.Printf("❌ Query error: %v", err)
		return nil, err
	}

	log.Printf("✅ Found %d upcoming published events", len(models))

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = toDomainEvent(&m)
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
		Where("start_date < CURRENT_DATE AND deleted_at IS NULL").
		Where("event_status_id = ?", publishedStatusID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Order("start_date DESC").Find(&models).Error
	if err != nil {
		log.Printf("❌ Query error: %v", err)
		return nil, err
	}

	log.Printf("✅ Found %d past published events", len(models))

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = toDomainEvent(&m)
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
		dbQuery = dbQuery.Where("name ILIKE ? OR description ILIKE ? OR display_name ILIKE ?", searchTerm, searchTerm, searchTerm)
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
	if filters.CategoryID != "" {
		dbQuery = dbQuery.Where("category_id = ?", filters.CategoryID)
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

	err := dbQuery.Order("start_date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = toDomainEvent(&m)
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
		Where("events.start_date >= CURRENT_DATE AND events.deleted_at IS NULL").
		Where("events.event_status_id = ?", publishedStatusID).
		Order("events.start_date ASC").
		Limit(limit)

	err = query.Find(&models).Error
	if err != nil {
		log.Printf("❌ Query error: %v", err)
		return nil, err
	}

	log.Printf("✅ Found %d upcoming published events with creator info", len(models))
	events := toDomainEventsWithCreator(models)

	// Load child entities for each event
	for _, event := range events {
		if err := r.loadChildEntities(ctx, event); err != nil {
			return nil, err
		}
	}

	return events, nil
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
	event := events[0]

	if err := r.loadChildEntities(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
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
	event := events[0]

	if err := r.loadChildEntities(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
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
	event := events[0]

	if err := r.loadChildEntities(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

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

	err := query.Order("events.start_date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := toDomainEventsWithCreator(models)

	// Load child entities for each event
	for _, event := range events {
		if err := r.loadChildEntities(ctx, event); err != nil {
			return nil, 0, err
		}
	}

	return events, total, nil
}

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

	err := query.Order("events.start_date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := toDomainEventsWithCreator(models)

	// Load child entities for each event
	for _, event := range events {
		if err := r.loadChildEntities(ctx, event); err != nil {
			return nil, 0, err
		}
	}

	return events, total, nil
}

// ============================================================
// EVENT QUERIES WITH CREATOR INFO AND FILTERS (NEW)
// ============================================================

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

	err = query.Order("events.start_date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	events := toDomainEventsWithCreator(models)

	// Load child entities for each event
	for _, event := range events {
		if err := r.loadChildEntities(ctx, event); err != nil {
			return nil, 0, err
		}
	}

	return events, total, nil
}

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
	return toDomainUserInfo(&model), nil
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
		userInfos[i] = toDomainUserInfo(&m)
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
	return toDomainEventTypeEntity(&model), nil
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
	return toDomainEventTypeEntity(&model), nil
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
	return toDomainEventTypeEntity(&model), nil
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
		types[i] = toDomainEventTypeEntity(&m)
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
	return toDomainEventStatusEntity(&model), nil
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
	return toDomainEventStatusEntity(&model), nil
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
	return toDomainEventStatusEntity(&model), nil
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
		statuses[i] = toDomainEventStatusEntity(&m)
	}
	return statuses, nil
}

// ============================================================
// CHILD ENTITY HELPERS
// ============================================================

func (r *PostgresRepository) loadChildEntities(ctx context.Context, event *domain.Event) error {
	// Load schedules
	var scheduleModels []EventScheduleModel
	if err := r.db.WithContext(ctx).Where("event_id = ? AND deleted_at IS NULL", event.ID).Find(&scheduleModels).Error; err != nil {
		return err
	}
	event.Schedules = toDomainSchedules(scheduleModels)

	// Load tickets
	var ticketModels []EventTicketModel
	if err := r.db.WithContext(ctx).Where("event_id = ? AND deleted_at IS NULL", event.ID).Find(&ticketModels).Error; err != nil {
		return err
	}
	event.Tickets = toDomainTickets(ticketModels)

	// Load speakers
	var speakerModels []EventSpeakerModel
	if err := r.db.WithContext(ctx).Where("event_id = ? AND deleted_at IS NULL", event.ID).Find(&speakerModels).Error; err != nil {
		return err
	}
	event.Speakers = toDomainSpeakers(speakerModels)

	// Load materials
	var materialModels []EventMaterialModel
	if err := r.db.WithContext(ctx).Where("event_id = ? AND deleted_at IS NULL", event.ID).Find(&materialModels).Error; err != nil {
		return err
	}
	event.Materials = toDomainMaterials(materialModels)

	return nil
}

func (r *PostgresRepository) saveSchedules(ctx context.Context, eventID string, schedules []domain.EventSchedule) error {
	if len(schedules) == 0 {
		return nil
	}
	models := toModelSchedules(eventID, schedules)
	return r.db.WithContext(ctx).Create(&models).Error
}

func (r *PostgresRepository) deleteSchedules(ctx context.Context, eventID string) error {
	return r.db.WithContext(ctx).Where("event_id = ?", eventID).Delete(&EventScheduleModel{}).Error
}

func (r *PostgresRepository) saveTickets(ctx context.Context, eventID string, tickets []domain.EventTicket) error {
	if len(tickets) == 0 {
		return nil
	}
	models := toModelTickets(eventID, tickets)
	return r.db.WithContext(ctx).Create(&models).Error
}

func (r *PostgresRepository) deleteTickets(ctx context.Context, eventID string) error {
	return r.db.WithContext(ctx).Where("event_id = ?", eventID).Delete(&EventTicketModel{}).Error
}

func (r *PostgresRepository) saveSpeakers(ctx context.Context, eventID string, speakers []domain.EventSpeaker) error {
	if len(speakers) == 0 {
		return nil
	}
	models := toModelSpeakers(eventID, speakers)
	return r.db.WithContext(ctx).Create(&models).Error
}

func (r *PostgresRepository) deleteSpeakers(ctx context.Context, eventID string) error {
	return r.db.WithContext(ctx).Where("event_id = ?", eventID).Delete(&EventSpeakerModel{}).Error
}

func (r *PostgresRepository) saveMaterials(ctx context.Context, eventID string, materials []domain.EventMaterial) error {
	if len(materials) == 0 {
		return nil
	}
	models := toModelMaterials(eventID, materials)
	return r.db.WithContext(ctx).Create(&models).Error
}

func (r *PostgresRepository) deleteMaterials(ctx context.Context, eventID string) error {
	return r.db.WithContext(ctx).Where("event_id = ?", eventID).Delete(&EventMaterialModel{}).Error
}