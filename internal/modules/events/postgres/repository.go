// internal/modules/events/infrastructure/postgres/repository.go

package postgres

import (
	"context"
	"errors"
	"time"

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
// EVENT CRUD
// ============================================================

func (r *PostgresRepository) CreateEvent(ctx context.Context, event *domain.Event) error {
	model := ToModelEvent(event)
	return r.db.WithContext(ctx).Create(model).Error
}


// ✅ For public/regular access - excludes deleted events
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

// ✅ For restore operations - includes deleted events
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

func (r *PostgresRepository) GetEventBySlug(ctx context.Context, slug string) (*domain.Event, error) {
	var model EventModel
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&model).Error
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
	return r.db.WithContext(ctx).Updates(model).Error
}

func (r *PostgresRepository) DeleteEvent(ctx context.Context, id string) error {
	// ✅ Soft delete - sets deleted_at timestamp
	return r.db.WithContext(ctx).Model(&EventModel{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}

// ✅ NEW: Hard delete - permanently removes from database
func (r *PostgresRepository) PermanentlyDeleteEvent(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&EventModel{}, "id = ?", id).Error
}

// ============================================================
// EVENT QUERIES
// ============================================================


func (r *PostgresRepository) ListEvents(ctx context.Context, filters domain.ListEventsFilters) ([]*domain.Event, int64, error) {
    var models []EventModel
    var total int64

    // ✅ Start with Unscoped() to include soft-deleted records
    // This prevents GORM from automatically adding "deleted_at IS NULL"
    query := r.db.WithContext(ctx).Unscoped().Model(&EventModel{})

    // ✅ Handle deleted filter logic
    if filters.OnlyDeleted {
        // Show ONLY deleted events
        query = query.Where("deleted_at IS NOT NULL")
    } else if !filters.IncludeDeleted {
        // Show ONLY non-deleted events (default)
        query = query.Where("deleted_at IS NULL")
    }
    // If IncludeDeleted=true and OnlyDeleted=false, show ALL events (no deleted_at filter)

    if filters.AccountID != "" {
        query = query.Where("account_id = ?", filters.AccountID)
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

func (r *PostgresRepository) GetEventsByAccount(ctx context.Context, accountID string, limit, offset int) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	query := r.db.WithContext(ctx).Model(&EventModel{}).
		Where("account_id = ? AND deleted_at IS NULL", accountID)

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

	query := r.db.WithContext(ctx).Model(&EventModel{}).
		Where("date >= CURRENT_DATE AND deleted_at IS NULL")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("date ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = ToDomainEvent(&m)
	}
	return events, nil
}

func (r *PostgresRepository) GetPastEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	var models []EventModel

	query := r.db.WithContext(ctx).Model(&EventModel{}).
		Where("date < CURRENT_DATE AND deleted_at IS NULL")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("date DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}

	events := make([]*domain.Event, len(models))
	for i, m := range models {
		events[i] = ToDomainEvent(&m)
	}
	return events, nil
}



func (r *PostgresRepository) SearchEvents(ctx context.Context, query string, filters domain.SearchFilters) ([]*domain.Event, int64, error) {
    var models []EventModel
    var total int64

    // ✅ Start with Unscoped() to include soft-deleted records
    dbQuery := r.db.WithContext(ctx).Unscoped().Model(&EventModel{})

    // ✅ Handle deleted filter logic
    if filters.OnlyDeleted {
        // Show ONLY deleted events
        dbQuery = dbQuery.Where("deleted_at IS NOT NULL")
    } else if !filters.IncludeDeleted {
        // Show ONLY non-deleted events (default)
        dbQuery = dbQuery.Where("deleted_at IS NULL")
    }
    // If IncludeDeleted=true and OnlyDeleted=false, show ALL events

    if query != "" {
        searchTerm := "%" + query + "%"
        dbQuery = dbQuery.Where("name ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
    }
    if filters.AccountID != "" {
        dbQuery = dbQuery.Where("account_id = ?", filters.AccountID)
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
// EVENT TYPE OPERATIONS (Return entities with ID)
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
// EVENT STATUS OPERATIONS (Return entities with ID)
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

// GetUpcomingEventsWithCreator returns upcoming events with creator info populated


func (r *PostgresRepository) GetUpcomingEventsWithCreator(ctx context.Context, limit int) ([]*domain.Event, error) {
    var models []EventModel

    query := r.db.WithContext(ctx).Table("events").
        Select(`events.*,
                accounts.name as creator_name,
                accounts.display_name as creator_display_name,
                accounts.email as creator_email,
                accounts.phone as creator_phone,
                COALESCE(institutions.name, '') as creator_institution_name,
                COALESCE(account_types.slug, 'personal') as creator_account_type`).
        Joins("LEFT JOIN accounts ON events.created_by = accounts.id").
        Joins("LEFT JOIN institutions ON accounts.institution_id = institutions.id").
        Joins("LEFT JOIN account_types ON accounts.account_type_id = account_types.id").
        Where("events.date >= CURRENT_DATE AND events.deleted_at IS NULL").
        Where("events.event_status_id IN (SELECT id FROM event_statuses WHERE slug IN ('published', 'upcoming', 'live'))").
        Order("events.date ASC, events.time ASC").
        Limit(limit)

    // ✅ Debug: Print the SQL
    sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
        return tx.Find(&models)
    })
    log.Printf("🔍 SQL Query: %s", sql)

    err := query.Find(&models).Error
    if err != nil {
        log.Printf("❌ Query error: %v", err)
        return nil, err
    }

    // ✅ Debug: Print the raw model data
    for i, m := range models {
        log.Printf("📊 Raw Model %d: ID=%s, CreatedBy=%s, CreatorName='%s', CreatorEmail='%s', CreatorAccountType='%s'", 
            i, m.ID, m.CreatedBy, m.CreatorName, m.CreatorEmail, m.CreatorAccountType)
    }

    return toDomainEventsWithCreator(models), nil
}

// GetEventBySlugWithCreator returns an event by slug with creator info populated


func (r *PostgresRepository) GetEventBySlugWithCreator(ctx context.Context, slug string) (*domain.Event, error) {
    var model EventModel

    err := r.db.WithContext(ctx).Table("events").
        Select(`events.*,
                accounts.name as creator_name,
                accounts.display_name as creator_display_name,
                accounts.email as creator_email,
                accounts.phone as creator_phone,
                COALESCE(institutions.name, '') as creator_institution_name,
                COALESCE(account_types.slug, 'personal') as creator_account_type`).
        Joins("LEFT JOIN accounts ON events.created_by = accounts.id").
        Joins("LEFT JOIN institutions ON accounts.institution_id = institutions.id").
        Joins("LEFT JOIN account_types ON accounts.account_type_id = account_types.id").
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


// GetEventByIDWithCreator returns an event by ID with creator info populated
func (r *PostgresRepository) GetEventByIDWithCreator(ctx context.Context, id string) (*domain.Event, error) {
    var model EventModel

    err := r.db.WithContext(ctx).Table("events").
        Select(`events.*,
                accounts.name as creator_name,
                accounts.display_name as creator_display_name,
                accounts.email as creator_email,
                accounts.phone as creator_phone,
                COALESCE(institutions.name, '') as creator_institution_name,
                COALESCE(account_types.slug, 'personal') as creator_account_type`).
        Joins("LEFT JOIN accounts ON events.created_by = accounts.id").
        Joins("LEFT JOIN institutions ON accounts.institution_id = institutions.id").
        Joins("LEFT JOIN account_types ON accounts.account_type_id = account_types.id").
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

// GetEventsByAccountWithCreator returns events for an account with creator info populated
func (r *PostgresRepository) GetEventsByAccountWithCreator(ctx context.Context, accountID string, limit, offset int) ([]*domain.Event, int64, error) {
    var models []EventModel
    var total int64

    query := r.db.WithContext(ctx).Table("events").
        Select(`events.*,
                accounts.name as creator_name,
                accounts.display_name as creator_display_name,
                accounts.email as creator_email,
                accounts.phone as creator_phone,
                COALESCE(institutions.name, '') as creator_institution_name,
                COALESCE(account_types.slug, 'personal') as creator_account_type`).
        Joins("LEFT JOIN accounts ON events.created_by = accounts.id").
        Joins("LEFT JOIN institutions ON accounts.institution_id = institutions.id").
        Joins("LEFT JOIN account_types ON accounts.account_type_id = account_types.id").
        Where("events.account_id = ? AND events.deleted_at IS NULL", accountID)

    // Get total count
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Apply pagination
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

