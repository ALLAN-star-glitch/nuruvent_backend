package postgres

import (
	"context"
	"errors"

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

func (r *PostgresRepository) GetEventByID(ctx context.Context, id string) (*domain.Event, error) {
	var model EventModel
	err := r.db.WithContext(ctx).
		Preload("EventType"). // ✅ Preload the event type
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
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *PostgresRepository) DeleteEvent(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&EventModel{}, "id = ?", id).Error
}

// ============================================================
// EVENT QUERIES
// ============================================================

func (r *PostgresRepository) ListEvents(ctx context.Context, filters domain.ListEventsFilters) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	query := r.db.WithContext(ctx).Model(&EventModel{}).Preload("EventType")

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
		Where("event_type_id = ?", eventTypeID).
		Preload("EventType")

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
		Where("account_id = ?", accountID)

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
		Where("date >= CURRENT_DATE")

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
		Where("date < CURRENT_DATE")

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

	dbQuery := r.db.WithContext(ctx).Model(&EventModel{})

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