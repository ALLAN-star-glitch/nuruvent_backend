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

func (r *PostgresRepository) CreateEvent(ctx context.Context, event *domain.Event) error {
	model := toModelEvent(event)

	// Debug logging
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

	// Hydrate recurrence pattern slug if recurring
	if event.RecurrencePatternID != nil && *event.RecurrencePatternID != "" {
		var pattern RecurrencePatternModel
		if err := r.db.WithContext(ctx).
			Where("id = ?", *event.RecurrencePatternID).
			First(&pattern).Error; err == nil {
			event.RecurrencePatternSlug = pattern.Slug
		}
	}

	return event, nil
}

func (r *PostgresRepository) GetEventByIDIncludingDeleted(ctx context.Context, id string) (*domain.Event, error) {
	var model EventModel
	// Use Unscoped() to include soft-deleted records
	err := r.db.WithContext(ctx).Unscoped().
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

func (r *PostgresRepository) RestoreEvent(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&EventModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

// ============================================================
// QUERY OPERATIONS - TEAM AWARE
// ============================================================

// ListEvents returns a paginated list of events with flexible filtering
func (r *PostgresRepository) ListEvents(ctx context.Context, filters domain.ListEventsFilters) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	log.Printf("🔍 REPOSITORY: IncludeDeleted=%v, OnlyDeleted=%v", filters.IncludeDeleted, filters.OnlyDeleted)

	query := r.db.WithContext(ctx).Unscoped().Model(&EventModel{})

	// Handle deleted filter logic
	if filters.OnlyDeleted {
		query = query.Where("deleted_at IS NOT NULL")
		log.Printf("🔍 REPOSITORY: OnlyDeleted=true - adding deleted_at IS NOT NULL")
	} else if !filters.IncludeDeleted {
		query = query.Where("deleted_at IS NULL")
		log.Printf("🔍 REPOSITORY: IncludeDeleted=false - adding deleted_at IS NULL")
	}

	// Team filtering - use TeamID directly (events belong to teams)
	if filters.TeamID != "" {
		query = query.Where("team_id = ?", filters.TeamID)
	} else if filters.Team.ID != "" && filters.Team.Type != "" {
		// For backward compatibility with TeamFilter
		query = query.Where("team_id = ?", filters.Team.ID)
	}

	// Account filtering - all teams under an account
	if filters.Account.ID != "" {
		// Join with teams table to filter by account
		query = query.Joins("JOIN teams ON teams.id = events.team_id").
			Where("teams.account_id = ?", filters.Account.ID)
	}

	// User filter (creator)
	if filters.UserID != "" {
		query = query.Where("created_by = ?", filters.UserID)
	}

	// Other filters
	if filters.EventTypeID != "" {
		query = query.Where("event_type_id = ?", filters.EventTypeID)
	}
	if filters.EventStatusID != "" {
		query = query.Where("event_status_id = ?", filters.EventStatusID)
	}
	if filters.CategoryID != "" {
		query = query.Where("category_id = ?", filters.CategoryID)
	}
	if filters.Visibility != "" {
		query = query.Where("visibility = ?", filters.Visibility)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Apply sorting
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

	// Convert to domain events
	events := make([]*domain.Event, len(models))
	for i, m := range models {
		event := toDomainEvent(&m)
		if event != nil {
			if err := r.loadChildEntities(ctx, event); err != nil {
				log.Printf("⚠️ Failed to load child entities for event %s: %v", event.ID, err)
			}
		}
		events[i] = event
	}

	return events, total, nil
}

// ListEventsByAccount returns all events across all teams under an account
func (r *PostgresRepository) ListEventsByAccount(ctx context.Context, account domain.AccountFilter, filters domain.ListEventsFilters) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	query := r.db.WithContext(ctx).Unscoped().Model(&EventModel{}).
		Joins("JOIN teams ON teams.id = events.team_id").
		Where("teams.account_id = ?", account.ID)

	// Handle deleted filter logic
	if filters.OnlyDeleted {
		query = query.Where("events.deleted_at IS NOT NULL")
	} else if !filters.IncludeDeleted {
		query = query.Where("events.deleted_at IS NULL")
	}

	// User filter (creator)
	if filters.UserID != "" {
		query = query.Where("events.created_by = ?", filters.UserID)
	}

	// Other filters
	if filters.EventTypeID != "" {
		query = query.Where("events.event_type_id = ?", filters.EventTypeID)
	}
	if filters.EventStatusID != "" {
		query = query.Where("events.event_status_id = ?", filters.EventStatusID)
	}
	if filters.CategoryID != "" {
		query = query.Where("events.category_id = ?", filters.CategoryID)
	}
	if filters.Visibility != "" {
		query = query.Where("events.visibility = ?", filters.Visibility)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Apply sorting
	sortField := "events.start_date"
	if filters.SortBy != "" {
		sortField = "events." + filters.SortBy
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

	// Convert to domain events
	events := make([]*domain.Event, len(models))
	for i, m := range models {
		event := toDomainEvent(&m)
		if event != nil {
			if err := r.loadChildEntities(ctx, event); err != nil {
				log.Printf("⚠️ Failed to load child entities for event %s: %v", event.ID, err)
			}
		}
		events[i] = event
	}

	return events, total, nil
}

// ListEventsByTeam returns all events for a specific team
func (r *PostgresRepository) ListEventsByTeam(ctx context.Context, teamID string, filters domain.ListEventsFilters) ([]*domain.Event, int64, error) {
	filters.TeamID = teamID
	return r.ListEvents(ctx, filters)
}

// GetUpcomingEvents returns upcoming events for a team
func (r *PostgresRepository) GetUpcomingEvents(ctx context.Context, teamID string, limit int) ([]*domain.Event, error) {
	var models []EventModel

	publishedStatusID, err := r.getStatusIDByName(ctx, domain.EventStatusPublished.GetName())
	if err != nil {
		log.Printf("❌ Failed to get published status: %v", err)
		return nil, fmt.Errorf("failed to get published status: %w", err)
	}

	query := r.db.WithContext(ctx).Model(&EventModel{}).
		Where("start_date >= CURRENT_DATE AND deleted_at IS NULL").
		Where("event_status_id = ?", publishedStatusID)

	if teamID != "" {
		query = query.Where("team_id = ?", teamID)
	}

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
		event := toDomainEvent(&m)
		if event != nil {
			if err := r.loadChildEntities(ctx, event); err != nil {
				log.Printf("⚠️ Failed to load child entities: %v", err)
			}
		}
		events[i] = event
	}
	return events, nil
}

// GetPastEvents returns past events for a team
func (r *PostgresRepository) GetPastEvents(ctx context.Context, teamID string, limit int) ([]*domain.Event, error) {
	var models []EventModel

	publishedStatusID, err := r.getStatusIDByName(ctx, domain.EventStatusPublished.GetName())
	if err != nil {
		log.Printf("❌ Failed to get published status: %v", err)
		return nil, fmt.Errorf("failed to get published status: %w", err)
	}

	query := r.db.WithContext(ctx).Model(&EventModel{}).
		Where("start_date < CURRENT_DATE AND deleted_at IS NULL").
		Where("event_status_id = ?", publishedStatusID)

	if teamID != "" {
		query = query.Where("team_id = ?", teamID)
	}

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
		event := toDomainEvent(&m)
		if event != nil {
			if err := r.loadChildEntities(ctx, event); err != nil {
				log.Printf("⚠️ Failed to load child entities: %v", err)
			}
		}
		events[i] = event
	}
	return events, nil
}

// SearchEvents performs a full-text search across event fields with team filtering
func (r *PostgresRepository) SearchEvents(ctx context.Context, query string, filters domain.SearchFilters) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	dbQuery := r.db.WithContext(ctx).Unscoped().Model(&EventModel{})

	// Handle deleted filter logic
	if filters.OnlyDeleted {
		dbQuery = dbQuery.Where("deleted_at IS NOT NULL")
	} else if !filters.IncludeDeleted {
		dbQuery = dbQuery.Where("deleted_at IS NULL")
	}

	// Team filtering
	if filters.TeamID != "" {
		dbQuery = dbQuery.Where("team_id = ?", filters.TeamID)
	} else if filters.Team.ID != "" && filters.Team.Type != "" {
		dbQuery = dbQuery.Where("team_id = ?", filters.Team.ID)
	}

	// Account filtering
	if filters.Account.ID != "" {
		dbQuery = dbQuery.Joins("JOIN teams ON teams.id = events.team_id").
			Where("teams.account_id = ?", filters.Account.ID)
	}

	// Apply search query
	if query != "" {
		searchTerm := "%" + query + "%"
		dbQuery = dbQuery.Where("name ILIKE ? OR display_name ILIKE ? OR description ILIKE ? OR short_description ILIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// User filter
	if filters.UserID != "" {
		dbQuery = dbQuery.Where("created_by = ?", filters.UserID)
	}

	// Other filters
	if filters.EventTypeID != "" {
		dbQuery = dbQuery.Where("event_type_id = ?", filters.EventTypeID)
	}
	if filters.CategoryID != "" {
		dbQuery = dbQuery.Where("category_id = ?", filters.CategoryID)
	}
	if filters.Visibility != "" {
		dbQuery = dbQuery.Where("visibility = ?", filters.Visibility)
	}

	// Count total
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filters.Limit > 0 {
		dbQuery = dbQuery.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		dbQuery = dbQuery.Offset(filters.Offset)
	}

	// Apply sorting
	err := dbQuery.Order("start_date DESC").Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	// Convert to domain events
	events := make([]*domain.Event, len(models))
	for i, m := range models {
		event := toDomainEvent(&m)
		if event != nil {
			if err := r.loadChildEntities(ctx, event); err != nil {
				log.Printf("⚠️ Failed to load child entities: %v", err)
			}
		}
		events[i] = event
	}
	return events, total, nil
}

// ============================================================
// VALUE OBJECT QUERIES
// ============================================================

// --- Event Types ---

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

// --- Event Statuses ---

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

	// Load EventType
	if event.EventTypeID != "" {
		var eventTypeModel EventTypeModel
		if err := r.db.WithContext(ctx).Where("id = ?", event.EventTypeID).First(&eventTypeModel).Error; err == nil {
			event.EventType = toDomainEventType(&eventTypeModel)
		}
	}

	// Load EventStatus
	if event.EventStatusID != "" {
		var eventStatusModel EventStatusModel
		if err := r.db.WithContext(ctx).Where("id = ?", event.EventStatusID).First(&eventStatusModel).Error; err == nil {
			event.EventStatus = toDomainEventStatus(&eventStatusModel)
		}
	}

	// Load Category
	if event.CategoryID != nil && *event.CategoryID != "" {
		var categoryModel CategoryModel
		if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", *event.CategoryID).First(&categoryModel).Error; err == nil {
			event.Category = toDomainCategory(&categoryModel)
		}
	}

	return nil
}

func (r *PostgresRepository) GetAllCategories(ctx context.Context) ([]*domain.Category, error) {
	var models []CategoryModel
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("sort_order ASC, display_name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	categories := make([]*domain.Category, len(models))
	for i, m := range models {
		categories[i] = toDomainCategory(&m)
	}
	return categories, nil
}

// ============================================================
// CHILD ENTITY SAVE/UPDATE HELPERS
// ============================================================

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


func (r *PostgresRepository) GetAllTicketTypes(ctx context.Context) ([]*domain.TicketTypeRow, error) {
	var models []TicketTypeModel

	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	out := make([]*domain.TicketTypeRow, len(models))
	for i, m := range models {
		out[i] = &domain.TicketTypeRow{
			ID:          m.ID,
			Slug:        m.Slug,
			Name:        m.Name,
			DisplayName: m.DisplayName,
			Description: m.Description,
			SortOrder:   m.SortOrder,
			IsActive:    m.IsActive,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
		}
	}
	return out, nil
}

func (r *PostgresRepository) GetRecurrencePatternBySlug(ctx context.Context, slug string) (*domain.RecurrencePattern, error) {
	if slug == "" {
		return nil, errors.New("slug is required")
	}

	var model RecurrencePatternModel
	err := r.db.WithContext(ctx).
		Where("slug = ? AND deleted_at IS NULL", slug).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get recurrence pattern: %w", err)
	}

	return &domain.RecurrencePattern{
		ID:          model.ID,
		Slug:        model.Slug,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Description: model.Description,
		IsActive:    model.IsActive,
	}, nil
}

func (r *PostgresRepository) AccountIDForTeam(ctx context.Context, teamID string) (string, error) {
    var accountID string
    err := r.db.WithContext(ctx).
        Table("teams").
        Select("account_id").
        Where("id = ? AND deleted_at IS NULL", teamID).
        Scan(&accountID).Error
    if err != nil {
        return "", err
    }
    return accountID, nil
}