// internal/modules/events/infrastructure/postgres/repository.go

package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
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

// isAllowedSortField whitelists columns the client can sort by. This
// prevents SQL injection via the sort_by query param.
func isAllowedSortField(field string) bool {
	switch field {
	case "created_at", "updated_at", "start_date", "end_date",
		"display_name", "name", "current_attendees", "capacity",
		"event_status_id", "event_type_id", "visibility":
		return true
	}
	return false
}

// resolveSortExpression returns a safe ORDER BY expression for the given
// sort field. Falls back to updated_at when the field isn't whitelisted.
func resolveSortExpression(field, order string) string {
	sortField := "updated_at"
	if field != "" && isAllowedSortField(field) {
		sortField = field
	}
	sortOrder := "DESC"
	if strings.EqualFold(order, "asc") {
		sortOrder = "ASC"
	}
	return sortField + " " + sortOrder
}

// ============================================================
// EVENT CRUD
// ============================================================

func (r *PostgresRepository) CreateEvent(ctx context.Context, event *domain.Event) error {
	model := toModelEvent(event)

	log.Printf("🔍 Creating event with model: %+v", model)
	log.Printf("🔍 Schedules: %+v", event.Schedules)
	log.Printf("🔍 Tickets: %+v", event.Tickets)
	log.Printf("🔍 Speakers: %+v", event.Speakers)
	log.Printf("🔍 Materials: %+v", event.Materials)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		log.Printf("❌ Failed to create event: %v", err)
		return err
	}

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

	if err := r.loadChildEntities(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (r *PostgresRepository) GetEventByIDIncludingDeleted(ctx context.Context, id string) (*domain.Event, error) {
	var model EventModel
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

	result := r.db.WithContext(ctx).Unscoped().
		Model(&EventModel{}).
		Where("id = ?", event.ID).
		Select("*").
		Updates(model)

	if result.Error != nil {
		return result.Error
	}

	if err := r.upsertSchedules(ctx, event.ID, event.Schedules); err != nil {
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

// ListEvents returns a paginated list of events with flexible filtering.
//
// Filters applied (all optional):
//   - Team / TeamID / Account       — scope
//   - UserID                        — creator
//   - EventTypeID, EventStatusID,
//     CategoryID, Visibility        — scalar filters
//   - SearchQuery                   — ILIKE against name, display_name,
//                                     description, short_description
//   - IncludeDeleted / OnlyDeleted  — soft-delete view
//   - SortBy / SortOrder            — whitelisted column + direction
func (r *PostgresRepository) ListEvents(ctx context.Context, filters domain.ListEventsFilters) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	log.Printf("🔍 REPOSITORY: IncludeDeleted=%v, OnlyDeleted=%v, SearchQuery=%q, SortBy=%q %q",
		filters.IncludeDeleted, filters.OnlyDeleted, filters.SearchQuery,
		filters.SortBy, filters.SortOrder)

	query := r.db.WithContext(ctx).Unscoped().Model(&EventModel{})

	// Deleted filter
	if filters.OnlyDeleted {
		query = query.Where("deleted_at IS NOT NULL")
	} else if !filters.IncludeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	// Team filter
	if filters.TeamID != "" {
		query = query.Where("team_id = ?", filters.TeamID)
	} else if filters.Team.ID != "" && filters.Team.Type != "" {
		query = query.Where("team_id = ?", filters.Team.ID)
	}

	// Account filter
	if filters.Account.ID != "" {
		query = query.Joins("JOIN teams ON teams.id = events.team_id").
			Where("teams.account_id = ?", filters.Account.ID)
	}

	// Creator
	if filters.UserID != "" {
		query = query.Where("created_by = ?", filters.UserID)
	}

	// Scalar filters
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

	// Full-text search
	if filters.SearchQuery != "" {
		q := "%" + filters.SearchQuery + "%"
		query = query.Where(
			"name ILIKE ? OR display_name ILIKE ? OR description ILIKE ? OR short_description ILIKE ?",
			q, q, q, q,
		)
	}

	// Count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Safe sort
	query = query.Order(resolveSortExpression(filters.SortBy, filters.SortOrder))

	if err := query.Find(&models).Error; err != nil {
		return nil, 0, err
	}

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

// ListEventsByAccount returns all events across all teams under an account.
func (r *PostgresRepository) ListEventsByAccount(ctx context.Context, account domain.AccountFilter, filters domain.ListEventsFilters) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	query := r.db.WithContext(ctx).Unscoped().Model(&EventModel{}).
		Joins("JOIN teams ON teams.id = events.team_id").
		Where("teams.account_id = ?", account.ID)

	if filters.OnlyDeleted {
		query = query.Where("events.deleted_at IS NOT NULL")
	} else if !filters.IncludeDeleted {
		query = query.Where("events.deleted_at IS NULL")
	}

	if filters.UserID != "" {
		query = query.Where("events.created_by = ?", filters.UserID)
	}
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
	if filters.SearchQuery != "" {
		q := "%" + filters.SearchQuery + "%"
		query = query.Where(
			"events.name ILIKE ? OR events.display_name ILIKE ? OR events.description ILIKE ? OR events.short_description ILIKE ?",
			q, q, q, q,
		)
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

	sortExpr := resolveSortExpression(filters.SortBy, filters.SortOrder)
	query = query.Order("events." + sortExpr)

	if err := query.Find(&models).Error; err != nil {
		return nil, 0, err
	}

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

func (r *PostgresRepository) ListEventsByTeam(ctx context.Context, teamID string, filters domain.ListEventsFilters) ([]*domain.Event, int64, error) {
	filters.TeamID = teamID
	return r.ListEvents(ctx, filters)
}

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

	if err := query.Order("start_date ASC").Find(&models).Error; err != nil {
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

	if err := query.Order("start_date DESC").Find(&models).Error; err != nil {
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

// SearchEvents performs a full-text search across event fields.
// Now honors EventStatusID, SortBy, SortOrder (matching ListEvents).
func (r *PostgresRepository) SearchEvents(ctx context.Context, query string, filters domain.SearchFilters) ([]*domain.Event, int64, error) {
	var models []EventModel
	var total int64

	dbQuery := r.db.WithContext(ctx).Unscoped().Model(&EventModel{})

	if filters.OnlyDeleted {
		dbQuery = dbQuery.Where("deleted_at IS NOT NULL")
	} else if !filters.IncludeDeleted {
		dbQuery = dbQuery.Where("deleted_at IS NULL")
	}

	if filters.TeamID != "" {
		dbQuery = dbQuery.Where("team_id = ?", filters.TeamID)
	} else if filters.Team.ID != "" && filters.Team.Type != "" {
		dbQuery = dbQuery.Where("team_id = ?", filters.Team.ID)
	}

	if filters.Account.ID != "" {
		dbQuery = dbQuery.Joins("JOIN teams ON teams.id = events.team_id").
			Where("teams.account_id = ?", filters.Account.ID)
	}

	if query != "" {
		searchTerm := "%" + query + "%"
		dbQuery = dbQuery.Where(
			"name ILIKE ? OR display_name ILIKE ? OR description ILIKE ? OR short_description ILIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	if filters.UserID != "" {
		dbQuery = dbQuery.Where("created_by = ?", filters.UserID)
	}
	if filters.EventTypeID != "" {
		dbQuery = dbQuery.Where("event_type_id = ?", filters.EventTypeID)
	}
	if filters.EventStatusID != "" {
		dbQuery = dbQuery.Where("event_status_id = ?", filters.EventStatusID)
	}
	if filters.CategoryID != "" {
		dbQuery = dbQuery.Where("category_id = ?", filters.CategoryID)
	}
	if filters.Visibility != "" {
		dbQuery = dbQuery.Where("visibility = ?", filters.Visibility)
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

	dbQuery = dbQuery.Order(resolveSortExpression(filters.SortBy, filters.SortOrder))

	if err := dbQuery.Find(&models).Error; err != nil {
		return nil, 0, err
	}

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
	var scheduleModels []EventScheduleModel
	if err := r.db.WithContext(ctx).
		Where("event_id = ? AND deleted_at IS NULL", event.ID).
		Order("session_number ASC").
		Find(&scheduleModels).Error; err != nil {
		return err
	}
	event.Schedules = toDomainSchedules(scheduleModels)

	var ticketModels []EventTicketModel
	if err := r.db.WithContext(ctx).
		Preload("TicketType").
		Where("event_id = ? AND deleted_at IS NULL", event.ID).
		Find(&ticketModels).Error; err != nil {
		return err
	}
	event.Tickets = toDomainTickets(ticketModels)

	var speakerModels []EventSpeakerModel
	if err := r.db.WithContext(ctx).Where("event_id = ? AND deleted_at IS NULL", event.ID).Find(&speakerModels).Error; err != nil {
		return err
	}
	event.Speakers = toDomainSpeakers(speakerModels)

	var materialModels []EventMaterialModel
	if err := r.db.WithContext(ctx).Where("event_id = ? AND deleted_at IS NULL", event.ID).Find(&materialModels).Error; err != nil {
		return err
	}
	event.Materials = toDomainMaterials(materialModels)

	if event.EventTypeID != "" {
		var eventTypeModel EventTypeModel
		if err := r.db.WithContext(ctx).Where("id = ?", event.EventTypeID).First(&eventTypeModel).Error; err == nil {
			event.EventType = toDomainEventType(&eventTypeModel)
		}
	}

	if event.EventStatusID != "" {
		var eventStatusModel EventStatusModel
		if err := r.db.WithContext(ctx).Where("id = ?", event.EventStatusID).First(&eventStatusModel).Error; err == nil {
			event.EventStatus = toDomainEventStatus(&eventStatusModel)
		}
	}

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
	log.Printf("[debug] saveSchedules: in=%d models=%d firstSessionNum=%d",
		len(schedules), len(models), models[0].SessionNumber)
	err := r.db.WithContext(ctx).Create(&models).Error
	log.Printf("[debug] saveSchedules err=%v", err)
	return err
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

func (r *PostgresRepository) AdjustAttendeeCount(ctx context.Context, eventID string, delta int) error {
	res := r.db.WithContext(ctx).
		Model(&EventModel{}).
		Where("id = ?", eventID).
		UpdateColumn("current_attendees", gorm.Expr("current_attendees + ?", delta))

	if res.Error != nil {
		return fmt.Errorf("adjust attendee count: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrEventNotFound
	}
	return nil
}

func (r *PostgresRepository) upsertSchedules(
	ctx context.Context,
	eventID string,
	schedules []domain.EventSchedule,
) error {
	var existingIDs []string
	if err := r.db.WithContext(ctx).
		Model(&EventScheduleModel{}).
		Where("event_id = ? AND deleted_at IS NULL", eventID).
		Pluck("id", &existingIDs).Error; err != nil {
		return err
	}

	incoming := make(map[string]struct{}, len(schedules))
	for _, s := range schedules {
		if s.ID != "" {
			incoming[s.ID] = struct{}{}
		}
	}

	toDelete := make([]string, 0)
	for _, id := range existingIDs {
		if _, ok := incoming[id]; !ok {
			toDelete = append(toDelete, id)
		}
	}
	if len(toDelete) > 0 {
		if err := r.db.WithContext(ctx).
			Where("id IN ?", toDelete).
			Delete(&EventScheduleModel{}).Error; err != nil {
			return err
		}
	}

	for i, s := range schedules {
		model := toScheduleModel(eventID, s, i+1)

		if s.ID == "" || !containsString(existingIDs, s.ID) {
			if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
				return err
			}
			continue
		}

		if err := r.db.WithContext(ctx).
			Model(&EventScheduleModel{}).
			Where("id = ?", s.ID).
			Updates(map[string]any{
				"session_name":              model.SessionName,
				"session_number":            model.SessionNumber,
				"start_date":                model.StartDate,
				"end_date":                  model.EndDate,
				"start_time":                model.StartTime,
				"end_time":                  model.EndTime,
				"timezone":                  model.Timezone,
				"location":                  model.Location,
				"is_virtual":                model.IsVirtual,
				"platform":                  model.Platform,
				"zoom_link":                 model.ZoomLink,
				"meet_link":                 model.MeetLink,
				"video_meeting_id":          model.VideoMeetingID,
				"video_meeting_external_id": model.VideoMeetingExternalID,
				"max_attendees":             model.MaxAttendees,
				"updated_at":                time.Now().UTC(),
			}).Error; err != nil {
			return err
		}
	}

	return nil
}

func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func (r *PostgresRepository) ReorderEventSchedules(
	ctx context.Context,
	eventID string,
	orderedIDs []string,
) error {
	if len(orderedIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&EventScheduleModel{}).
			Where("event_id = ? AND id IN ?", eventID, orderedIDs).
			Update("session_number", gorm.Expr("-session_number")).Error; err != nil {
			return err
		}
		for i, id := range orderedIDs {
			if err := tx.Model(&EventScheduleModel{}).
				Where("event_id = ? AND id = ?", eventID, id).
				Update("session_number", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PostgresRepository) ListEventIDsWithUnprovisionedSchedules(
	ctx context.Context,
	withinDays int,
) ([]string, error) {
	if withinDays <= 0 {
		withinDays = 30
	}

	var ids []string
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT e.id
			FROM events e
			JOIN event_schedules s
			  ON s.event_id = e.id
			 AND s.deleted_at IS NULL
			WHERE e.deleted_at IS NULL
			  AND s.is_virtual = true
			  AND s.platform != ''
			  AND s.video_meeting_id IS NULL
			  AND s.start_date <= NOW() + (INTERVAL '1 day' * ?)
		`, withinDays).
		Scan(&ids).Error
	if err != nil {
		return nil, fmt.Errorf("list unprovisioned events: %w", err)
	}
	return ids, nil
}