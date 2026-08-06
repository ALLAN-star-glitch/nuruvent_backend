// internal/modules/events/eventrepo/repository.go

package eventrepo

import (
	"errors"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

// ================================================
// CRUD OPERATIONS
// ================================================

// Create creates a new event
func (r *EventRepository) Create(event *models.Event) error {
	return r.db.Create(event).Error
}

// GetByID gets an event by ID with relations
func (r *EventRepository) GetByID(id uuid.UUID) (*models.Event, error) {
	var event models.Event
	err := r.db.
		Preload("EventType").
		Preload("EventStatus").
		Preload("Business").
		Where("id = ? AND is_active = ?", id, true).
		First(&event).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &event, nil
}

// GetBySlug gets an event by slug
func (r *EventRepository) GetBySlug(slug string) (*models.Event, error) {
	var event models.Event
	err := r.db.
		Preload("EventType").
		Preload("EventStatus").
		Preload("Business").
		Where("slug = ? AND is_active = ?", slug, true).
		First(&event).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &event, nil
}

// Update updates an event
func (r *EventRepository) Update(event *models.Event) error {
	return r.db.Save(event).Error
}

// Delete soft deletes an event
func (r *EventRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Event{}, id).Error
}

// ================================================
// QUERY OPERATIONS
// ================================================

// GetEventsByBusiness gets all events for a business with pagination and filters
func (r *EventRepository) GetEventsByBusiness(businessID uuid.UUID, eventTypeID *uuid.UUID, eventStatusID *uuid.UUID, limit, offset int) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64

	query := r.db.Model(&models.Event{}).
		Where("business_id = ? AND is_active = ?", businessID, true)

	if eventTypeID != nil {
		query = query.Where("event_type_id = ?", *eventTypeID)
	}
	if eventStatusID != nil {
		query = query.Where("event_status_id = ?", *eventStatusID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.
		Preload("EventType").
		Preload("EventStatus").
		Preload("Business").
		Order("date DESC, time DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error

	return events, total, err
}

// GetEventsByType gets all events of a specific type
func (r *EventRepository) GetEventsByType(eventTypeID uuid.UUID, limit, offset int) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64

	query := r.db.Model(&models.Event{}).
		Where("event_type_id = ? AND is_active = ?", eventTypeID, true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("EventType").
		Preload("EventStatus").
		Preload("Business").
		Order("date DESC, time DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error

	return events, total, err
}

// GetUpcomingEvents gets upcoming published events
func (r *EventRepository) GetUpcomingEvents(limit int) ([]models.Event, error) {
	var events []models.Event
	now := time.Now()
	
	// Get published status ID
	var publishedStatus models.EventStatus
	err := r.db.Where("name = ?", "published").First(&publishedStatus).Error
	if err != nil {
		return nil, err
	}

	err = r.db.
		Where("is_active = ? AND event_status_id = ? AND date >= ?", true, publishedStatus.ID, now).
		Preload("EventType").
		Preload("EventStatus").
		Preload("Business").
		Order("date ASC, time ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

// GetUpcomingEventsByBusiness gets upcoming events for a business
func (r *EventRepository) GetUpcomingEventsByBusiness(businessID uuid.UUID, limit int) ([]models.Event, error) {
	var events []models.Event
	now := time.Now()
	
	// Get published status ID
	var publishedStatus models.EventStatus
	err := r.db.Where("name = ?", "published").First(&publishedStatus).Error
	if err != nil {
		return nil, err
	}

	err = r.db.
		Where("business_id = ? AND is_active = ? AND event_status_id = ? AND date >= ?", businessID, true, publishedStatus.ID, now).
		Preload("EventType").
		Preload("EventStatus").
		Preload("Business").
		Order("date ASC, time ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

// GetPastEvents gets past published events
func (r *EventRepository) GetPastEvents(limit int) ([]models.Event, error) {
	var events []models.Event
	now := time.Now()
	
	// Get published and completed status IDs
	var statuses []models.EventStatus
	err := r.db.Where("name IN (?)", []string{"published", "completed"}).Find(&statuses).Error
	if err != nil {
		return nil, err
	}
	
	statusIDs := make([]uuid.UUID, len(statuses))
	for i, s := range statuses {
		statusIDs[i] = s.ID
	}

	err = r.db.
		Where("is_active = ? AND event_status_id IN (?) AND date < ?", true, statusIDs, now).
		Preload("EventType").
		Preload("EventStatus").
		Preload("Business").
		Order("date DESC, time DESC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

// SearchEvents searches events by title or description
func (r *EventRepository) SearchEvents(query string, businessID *uuid.UUID, eventTypeID *uuid.UUID, limit, offset int) ([]models.Event, int64, error) {
	var events []models.Event
	var total int64

	dbQuery := r.db.Model(&models.Event{}).
		Where("is_active = ?", true)

	if businessID != nil {
		dbQuery = dbQuery.Where("business_id = ?", *businessID)
	}
	if eventTypeID != nil {
		dbQuery = dbQuery.Where("event_type_id = ?", *eventTypeID)
	}
	if query != "" {
		dbQuery = dbQuery.Where("title ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := dbQuery.
		Preload("EventType").
		Preload("EventStatus").
		Preload("Business").
		Order("date ASC, time ASC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error

	return events, total, err
}

// ================================================
// COUNT OPERATIONS
// ================================================

// CountEventsByBusiness counts events for a business
func (r *EventRepository) CountEventsByBusiness(businessID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Event{}).
		Where("business_id = ? AND is_active = ?", businessID, true).
		Count(&count).Error
	return count, err
}

// CountUpcomingEventsByBusiness counts upcoming events for a business
func (r *EventRepository) CountUpcomingEventsByBusiness(businessID uuid.UUID) (int64, error) {
	var count int64
	now := time.Now()
	
	var publishedStatus models.EventStatus
	err := r.db.Where("name = ?", "published").First(&publishedStatus).Error
	if err != nil {
		return 0, err
	}

	err = r.db.Model(&models.Event{}).
		Where("business_id = ? AND is_active = ? AND event_status_id = ? AND date >= ?", businessID, true, publishedStatus.ID, now).
		Count(&count).Error
	return count, err
}

// ================================================
// EVENT STATUS OPERATIONS
// ================================================

// GetEventStatusByID gets an event status by ID
func (r *EventRepository) GetEventStatusByID(id uuid.UUID) (*models.EventStatus, error) {
	var eventStatus models.EventStatus
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&eventStatus).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &eventStatus, nil
}

// GetEventStatusByName gets an event status by name
func (r *EventRepository) GetEventStatusByName(name string) (*models.EventStatus, error) {
	var eventStatus models.EventStatus
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&eventStatus).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &eventStatus, nil
}

// GetAllEventStatuses gets all event statuses
func (r *EventRepository) GetAllEventStatuses() ([]models.EventStatus, error) {
	var eventStatuses []models.EventStatus
	err := r.db.
		Where("is_active = ?", true).
		Order("sort_order ASC, display_name ASC").
		Find(&eventStatuses).Error
	return eventStatuses, err
}

// ================================================
// EVENT TYPE OPERATIONS
// ================================================

// GetEventTypeByID gets an event type by ID
func (r *EventRepository) GetEventTypeByID(id uuid.UUID) (*models.EventType, error) {
	var eventType models.EventType
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&eventType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &eventType, nil
}

// GetEventTypeByName gets an event type by name
func (r *EventRepository) GetEventTypeByName(name string) (*models.EventType, error) {
	var eventType models.EventType
	err := r.db.Where("name = ? AND is_active = ?", name, true).First(&eventType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &eventType, nil
}

// GetAllEventTypes gets all event types
func (r *EventRepository) GetAllEventTypes() ([]models.EventType, error) {
	var eventTypes []models.EventType
	err := r.db.
		Where("is_active = ?", true).
		Order("sort_order ASC, display_name ASC").
		Find(&eventTypes).Error
	return eventTypes, err
}