// internal/modules/events/infrastructure/postgres/mappers.go

package postgres

import (
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
	"gorm.io/gorm"
)

// ============================================================
// DOMAIN ENTITY → DATABASE MODEL MAPPERS
// ============================================================

// ToModelEvent converts domain.Event to EventModel
func ToModelEvent(event *domain.Event) *EventModel {
	if event == nil {
		return nil
	}

	// ✅ Handle DeletedAt properly for restore operations
	var deletedAt gorm.DeletedAt
	if event.DeletedAt != nil && !event.DeletedAt.IsZero() {
		deletedAt = gorm.DeletedAt{Time: *event.DeletedAt, Valid: true}
	} else {
		// ✅ IMPORTANT: Set Valid: false to clear the deleted_at column
		deletedAt = gorm.DeletedAt{Valid: false}
	}

	// ✅ Convert empty string to nil for UUID fields
	var deletedBy *string
	if event.DeletedBy != "" {
		deletedBy = &event.DeletedBy
	}

	var restoredBy *string
	if event.RestoredBy != "" {
		restoredBy = &event.RestoredBy
	}

	// ✅ Handle RestoredAt
	var restoredAt *time.Time
	if event.RestoredAt != nil && !event.RestoredAt.IsZero() {
		restoredAt = event.RestoredAt
	}

	return &EventModel{
		ID:               event.ID,
		Slug:             event.Slug,
		Name:             event.Name,
		DisplayName:      event.DisplayName,
		Description:      event.Description,
		EventTypeID:      event.EventTypeID,
		EventStatusID:    event.EventStatusID,
		ImageURL:         event.ImageURL,
		ThumbnailURL:     event.ThumbnailURL,
		Date:             event.Date,
		Time:             event.Time,
		Duration:         event.Duration,
		Price:            event.Price,
		CertificatePrice: event.CertificatePrice,
		Location:         event.Location,
		IsVirtual:        event.IsVirtual,
		ZoomLink:         event.ZoomLink,
		MeetLink:         event.MeetLink,
		MaxAttendees:     event.MaxAttendees,
		CurrentAttendees: event.CurrentAttendees,
		AccountID:        event.AccountID,
		CreatedBy:        event.CreatedBy,
		IsActive:         event.IsActive,
		CreatedAt:        event.CreatedAt,
		UpdatedAt:        event.UpdatedAt,
		DeletedAt:        deletedAt,
		DeletedBy:        deletedBy,
		RestoredAt:       restoredAt,
		RestoredBy:       restoredBy,
		IsFeatured:       event.IsFeatured,
		IsPrivate:        event.IsPrivate,
	}
}

// ToDomainEvent converts EventModel to domain.Event
func ToDomainEvent(model *EventModel) *domain.Event {
	if model == nil {
		return nil
	}

	var deletedAt *time.Time
	if model.DeletedAt.Valid {
		deletedAt = &model.DeletedAt.Time
	}

	// ✅ Convert *string to string (empty string if nil)
	deletedBy := ""
	if model.DeletedBy != nil {
		deletedBy = *model.DeletedBy
	}

	restoredBy := ""
	if model.RestoredBy != nil {
		restoredBy = *model.RestoredBy
	}

	// ✅ Handle RestoredAt
	var restoredAt *time.Time
	if model.RestoredAt != nil && !model.RestoredAt.IsZero() {
		restoredAt = model.RestoredAt
	}

	return &domain.Event{
		ID:               model.ID,
		Slug:             model.Slug,
		Name:             model.Name,
		DisplayName:      model.DisplayName,
		Description:      model.Description,
		EventTypeID:      model.EventTypeID,
		EventStatusID:    model.EventStatusID,
		ImageURL:         model.ImageURL,
		ThumbnailURL:     model.ThumbnailURL,
		Date:             model.Date,
		Time:             model.Time,
		Duration:         model.Duration,
		Price:            model.Price,
		CertificatePrice: model.CertificatePrice,
		Location:         model.Location,
		IsVirtual:        model.IsVirtual,
		ZoomLink:         model.ZoomLink,
		MeetLink:         model.MeetLink,
		MaxAttendees:     model.MaxAttendees,
		CurrentAttendees: model.CurrentAttendees,
		AccountID:        model.AccountID,
		CreatedBy:        model.CreatedBy,
		IsActive:         model.IsActive,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
		DeletedAt:        deletedAt,
		DeletedBy:        deletedBy,
		RestoredAt:       restoredAt,
		RestoredBy:       restoredBy,
		IsFeatured:       model.IsFeatured,
		IsPrivate:        model.IsPrivate,
	}
}

// ============================================================
// EVENT TYPE - DOMAIN VALUE MAPPERS
// ============================================================

// ToModelEventTypeFromValue converts domain.EventTypeValue to EventTypeModel
func ToModelEventTypeFromValue(value domain.EventTypeValue) *EventTypeModel {
	// ✅ Use GetEventTypeInfo helper instead of value.Info()
	info, ok := domain.GetEventTypeInfo(value)
	if !ok {
		return nil
	}
	return ToModelEventTypeFromInfo(info)
}

// ToModelEventTypeFromInfo converts domain.EventTypeInfo to EventTypeModel
func ToModelEventTypeFromInfo(info domain.EventTypeInfo) *EventTypeModel {
	return &EventTypeModel{
		Slug:                info.Slug, // ✅ info.Slug is already a string
		Name:                info.Name,
		DisplayName:         info.DisplayName,
		Description:         info.Description,
		Icon:                info.Icon,
		Color:               info.Color,
		SortOrder:           info.SortOrder,
		SupportsCertificate: info.SupportsCertificate,
		MinDuration:         info.MinDuration,
		MaxDuration:         info.MaxDuration,
		IsActive:            info.IsActive,
	}
}

// ToDomainEventTypeInfo converts EventTypeModel to domain.EventTypeInfo
func ToDomainEventTypeInfo(model *EventTypeModel) (*domain.EventTypeInfo, error) {
	if model == nil {
		return nil, nil
	}

	// ✅ Use shared types to parse
	value, valid := types.ParseEventType(model.Name)
	if !valid {
		// Try parsing by slug
		value, valid = types.ParseEventTypeBySlug(model.Slug)
		if !valid {
			return nil, domain.ErrEventTypeNotFound
		}
	}

	// ✅ Use GetEventTypeInfo helper
	info, ok := domain.GetEventTypeInfo(value)
	if !ok {
		return nil, domain.ErrEventTypeNotFound
	}

	// Override with database values
	info.Name = model.Name
	info.DisplayName = model.DisplayName
	info.Description = model.Description
	info.Icon = model.Icon
	info.Color = model.Color
	info.SortOrder = model.SortOrder
	info.SupportsCertificate = model.SupportsCertificate
	info.MinDuration = model.MinDuration
	info.MaxDuration = model.MaxDuration
	info.IsActive = model.IsActive

	return &info, nil
}

// ============================================================
// EVENT STATUS - DOMAIN VALUE MAPPERS
// ============================================================

// ToModelEventStatusFromValue converts domain.EventStatusValue to EventStatusModel
func ToModelEventStatusFromValue(value domain.EventStatusValue) *EventStatusModel {
	// ✅ Use GetEventStatusInfo helper instead of value.Info()
	info, ok := domain.GetEventStatusInfo(value)
	if !ok {
		return nil
	}
	return ToModelEventStatusFromInfo(info)
}

// ToModelEventStatusFromInfo converts domain.EventStatusInfo to EventStatusModel
func ToModelEventStatusFromInfo(info domain.EventStatusInfo) *EventStatusModel {
	return &EventStatusModel{
		Slug:        info.Slug, // ✅ info.Slug is already a string
		Name:        info.Name,
		DisplayName: info.DisplayName,
		Description: info.Description,
		Color:       info.Color,
		Icon:        info.Icon,
		SortOrder:   info.SortOrder,
		IsFinal:     info.IsFinal,
		IsActive:    info.IsActive,
	}
}

// ToDomainEventStatusInfo converts EventStatusModel to domain.EventStatusInfo
func ToDomainEventStatusInfo(model *EventStatusModel) (*domain.EventStatusInfo, error) {
	if model == nil {
		return nil, nil
	}

	// ✅ Use shared types to parse
	value, valid := types.ParseEventStatus(model.Name)
	if !valid {
		// Try parsing by slug
		value, valid = types.ParseEventStatusBySlug(model.Slug)
		if !valid {
			return nil, domain.ErrEventStatusNotFound
		}
	}

	// ✅ Use GetEventStatusInfo helper
	info, ok := domain.GetEventStatusInfo(value)
	if !ok {
		return nil, domain.ErrEventStatusNotFound
	}

	// Override with database values
	info.Name = model.Name
	info.DisplayName = model.DisplayName
	info.Description = model.Description
	info.Color = model.Color
	info.Icon = model.Icon
	info.SortOrder = model.SortOrder
	info.IsFinal = model.IsFinal
	info.IsActive = model.IsActive

	return &info, nil
}

// ============================================================
// ENTITY MAPPERS (For database entities with ID)
// ============================================================

// ToDomainEventTypeEntity converts EventTypeModel to domain.EventType
func ToDomainEventTypeEntity(model *EventTypeModel) *domain.EventType {
	if model == nil {
		return nil
	}
	return &domain.EventType{
		ID:                  model.ID,
		Slug:                model.Slug,
		Name:                model.Name,
		DisplayName:         model.DisplayName,
		Description:         model.Description,
		Icon:                model.Icon,
		Color:               model.Color,
		SortOrder:           model.SortOrder,
		SupportsCertificate: model.SupportsCertificate,
		MinDuration:         model.MinDuration,
		MaxDuration:         model.MaxDuration,
		IsActive:            model.IsActive,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}

// ToModelEventTypeEntity converts domain.EventType to EventTypeModel
func ToModelEventTypeEntity(eventType *domain.EventType) *EventTypeModel {
	if eventType == nil {
		return nil
	}
	return &EventTypeModel{
		ID:                  eventType.ID,
		Slug:                eventType.Slug,
		Name:                eventType.Name,
		DisplayName:         eventType.DisplayName,
		Description:         eventType.Description,
		Icon:                eventType.Icon,
		Color:               eventType.Color,
		SortOrder:           eventType.SortOrder,
		SupportsCertificate: eventType.SupportsCertificate,
		MinDuration:         eventType.MinDuration,
		MaxDuration:         eventType.MaxDuration,
		IsActive:            eventType.IsActive,
		CreatedAt:           eventType.CreatedAt,
		UpdatedAt:           eventType.UpdatedAt,
	}
}

// ToDomainEventStatusEntity converts EventStatusModel to domain.EventStatus
func ToDomainEventStatusEntity(model *EventStatusModel) *domain.EventStatus {
	if model == nil {
		return nil
	}
	return &domain.EventStatus{
		ID:          model.ID,
		Slug:        model.Slug,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Description: model.Description,
		Color:       model.Color,
		Icon:        model.Icon,
		SortOrder:   model.SortOrder,
		IsFinal:     model.IsFinal,
		IsActive:    model.IsActive,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// ToModelEventStatusEntity converts domain.EventStatus to EventStatusModel
func ToModelEventStatusEntity(eventStatus *domain.EventStatus) *EventStatusModel {
	if eventStatus == nil {
		return nil
	}
	return &EventStatusModel{
		ID:          eventStatus.ID,
		Slug:        eventStatus.Slug,
		Name:        eventStatus.Name,
		DisplayName: eventStatus.DisplayName,
		Description: eventStatus.Description,
		Color:       eventStatus.Color,
		Icon:        eventStatus.Icon,
		SortOrder:   eventStatus.SortOrder,
		IsFinal:     eventStatus.IsFinal,
		IsActive:    eventStatus.IsActive,
		CreatedAt:   eventStatus.CreatedAt,
		UpdatedAt:   eventStatus.UpdatedAt,
	}
}

// ============================================================
// EVENT WITH CREATOR MAPPERS
// ============================================================

// toDomainEventWithCreator converts EventModel to domain.Event with creator info
func toDomainEventWithCreator(model *EventModel) *domain.Event {
	if model == nil {
		return nil
	}

	event := ToDomainEvent(model)

	// ✅ Populate creator info from join fields
	if model.CreatorName != "" || model.CreatorEmail != "" {
		creator := &domain.AccountInfo{
			ID:              model.CreatedBy,
			Name:            model.CreatorName,
			DisplayName:     model.CreatorDisplayName,
			Email:           model.CreatorEmail,
			Phone:           model.CreatorPhone,
			AccountType:     model.CreatorAccountType,
			InstitutionName: model.CreatorInstitutionName,
		}
		event.WithCreator(creator)
	}

	return event
}

// toDomainEventsWithCreator converts multiple EventModels to domain.Events with creator info
func toDomainEventsWithCreator(models []EventModel) []*domain.Event {
	events := make([]*domain.Event, len(models))
	for i, model := range models {
		events[i] = toDomainEventWithCreator(&model)
	}
	return events
}