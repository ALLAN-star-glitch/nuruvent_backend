// internal/modules/events/infrastructure/postgres/mappers.go

package postgres

import (
	"encoding/json"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
	"gorm.io/gorm"
)

// ============================================================
// JSONB HELPER FUNCTIONS
// ============================================================

// toJSONB converts any value to JSONB
func toJSONB(v interface{}) JSONB {
	if v == nil {
		return nil
	}

	// If it's already JSONB, return it
	if j, ok := v.(JSONB); ok {
		return j
	}

	// Convert map[string]string to JSONB
	if m, ok := v.(map[string]string); ok {
		if len(m) == 0 {
			return nil
		}
		jsonb := make(JSONB)
		for k, val := range m {
			jsonb[k] = val
		}
		return jsonb
	}

	// Convert map[string]float64 to JSONB
	if m, ok := v.(map[string]float64); ok {
		if len(m) == 0 {
			return nil
		}
		jsonb := make(JSONB)
		for k, val := range m {
			jsonb[k] = val
		}
		return jsonb
	}

	// Convert map[string]interface{} to JSONB
	if m, ok := v.(map[string]interface{}); ok {
		if len(m) == 0 {
			return nil
		}
		return JSONB(m)
	}

	// ✅ FIX: Convert []string to proper JSON array
	if s, ok := v.([]string); ok {
		if len(s) == 0 {
			return nil
		}
		// Marshal to JSON array and then unmarshal to JSONB
		data, err := json.Marshal(s)
		if err != nil {
			return nil
		}
		var result JSONB
		if err := json.Unmarshal(data, &result); err != nil {
			return nil
		}
		return result
	}

	// Try marshaling to JSON
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}

	var result JSONB
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}
	return result
}

// fromJSONBToStringArray converts JSONB to []string
func fromJSONBToStringArray(j JSONB) []string {
	if j == nil {
		return nil
	}

	result := make([]string, 0, len(j))
	for _, v := range j {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// fromJSONBToMapString converts JSONB to map[string]string
func fromJSONBToMapString(j JSONB) map[string]string {
	if j == nil {
		return nil
	}

	result := make(map[string]string)
	for k, v := range j {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result
}

func fromJSONBToMapFloat64(j JSONB) map[string]float64 {
	if j == nil {
		return nil
	}

	result := make(map[string]float64)
	for k, v := range j {
		switch val := v.(type) {
		case float64:
			result[k] = val
		case int:
			result[k] = float64(val)
		case int64:
			result[k] = float64(val)
		}
	}
	return result
}

func fromJSONBToMapInterface(j JSONB) map[string]interface{} {
	if j == nil {
		return nil
	}
	return map[string]interface{}(j)
}

// ============================================================
// DOMAIN ENTITY → DATABASE MODEL MAPPERS
// ============================================================

// toModelEvent converts domain.Event to EventModel
func toModelEvent(event *domain.Event) *EventModel {
	if event == nil {
		return nil
	}

	var deletedAt gorm.DeletedAt
	if event.DeletedAt != nil && !event.DeletedAt.IsZero() {
		deletedAt = gorm.DeletedAt{Time: *event.DeletedAt, Valid: true}
	}

	var deletedBy *string
	if event.DeletedBy != nil {
		deletedBy = event.DeletedBy
	}

	var restoredBy *string
	if event.RestoredBy != nil {
		restoredBy = event.RestoredBy
	}

	return &EventModel{
		// Core Identity
		ID:               event.ID,
		Slug:             event.Slug,
		Name:             event.Name,
		DisplayName:      event.DisplayName,
		Description:      event.Description,
		ShortDescription: event.ShortDescription,
		Tags:             toJSONB(event.Tags),
		Language:         event.Language,

		// Relations
		EventTypeID:          event.EventTypeID,
		EventStatusID:        event.EventStatusID,
		CategoryID:           event.CategoryID,
		EventFormatID:        event.EventFormatID,
		CertificateTemplateID: event.CertificateTemplateID,

		// Ownership
		InstitutionID: event.InstitutionID,
		CreatedBy:     event.CreatedBy,

		// Schedule & Venue
		StartDate:   event.StartDate,
		EndDate:     event.EndDate,
		IsMultiDay:  event.IsMultiDay,
		IsRecurring: event.IsRecurring,

		// Recurrence
		RecurrencePatternID:   event.RecurrencePatternID,
		RecurrenceInterval:    event.RecurrenceInterval,
		RecurrenceEndsOn:      event.RecurrenceEndsOn,
		RecurrenceOccurrences: event.RecurrenceOccurrences,
		RecurrenceDaysOfWeek:  toJSONB(event.RecurrenceDaysOfWeek),
		RecurrenceDayOfMonth:  event.RecurrenceDayOfMonth,
		RecurrenceWeekOfMonth: event.RecurrenceWeekOfMonth,

		// Venue
		VenueName:          event.VenueName,
		VenueAddress:       event.VenueAddress,
		VenueCity:          event.VenueCity,
		VenueCountry:       event.VenueCountry,
		VenueCoordinates:   toJSONB(event.VenueCoordinates),
		IsVirtual:          event.IsVirtual,
		IsHybrid:           event.IsHybrid,
		VirtualPlatform:    event.VirtualPlatform,
		VirtualPlatformURL: event.VirtualPlatformURL,
		InPersonLocation:   event.InPersonLocation,
		ZoomLink:           event.ZoomLink,
		MeetLink:           event.MeetLink,

		// Ticketing & Capacity
		IsFree:             event.IsFreeEvent,
		Capacity:           event.Capacity,
		CurrentAttendees:   event.CurrentAttendees,
		WaitlistEnabled:    event.WaitlistEnabled,
		WaitlistCapacity:   event.WaitlistCapacity,
		TicketSalesStart:   event.TicketSalesStart,
		TicketSalesEnd:     event.TicketSalesEnd,
		MinTicketsPerOrder: event.MinTicketsPerOrder,
		MaxTicketsPerOrder: event.MaxTicketsPerOrder,

		// Access & Privacy
		Visibility:          event.Visibility,
		Password:            event.Password,
		InviteOnly:          event.InviteOnly,
		InvitedEmails:       toJSONB(event.InvitedEmails),
		RequiresApproval:    event.RequiresApproval,
		ApprovalRequiredFor: toJSONB(event.ApprovalRequiredFor),

		// Monetization
		IsFeatured:          event.IsFeatured,
		FeaturedUntil:       event.FeaturedUntil,
		CertificateEnabled:  event.CertificateEnabled,
		CertificatePrice:    event.CertificatePrice,
		EarlyBirdDiscountPercentage: event.EarlyBirdDiscountPercentage,
		GroupDiscountPercentage: event.GroupDiscountPercentage,
		GroupMinAttendees:   event.GroupMinAttendees,

		// SEO
		SEOTitle:        event.SEO.Title,
		SEODescription:  event.SEO.Description,
		SEOKeywords:     toJSONB(event.SEO.Keywords),
		SEOCanonicalURL: event.SEO.CanonicalURL,
		SEORobots:       event.SEO.Robots,
		SEONoIndex:      event.SEO.NoIndex,

		// Open Graph
		OGTitle:       event.OpenGraph.Title,
		OGDescription: event.OpenGraph.Description,
		OGImageURL:    event.OpenGraph.ImageURL,
		OGType:        event.OpenGraph.Type,

		// Twitter
		TwitterCard:        event.Twitter.Card,
		TwitterTitle:       event.Twitter.Title,
		TwitterDescription: event.Twitter.Description,
		TwitterImageURL:    event.Twitter.ImageURL,

		// Schema
		SchemaOrg: toJSONB(event.SchemaOrg),

		// Media
		ImageURL:     event.ImageURL,
		ThumbnailURL: event.ThumbnailURL,

		// Social
		SocialLinks:        toJSONB(event.SocialLinks),
		HasLivestream:      event.HasLivestream,
		LivestreamURL:      event.LivestreamURL,
		RecordingAvailable: event.RecordingAvailable,
		RecordingURL:       event.RecordingURL,

		// Metadata
		Metadata:           toJSONB(event.Metadata),
		Version:            event.Version,
		PublishedAt:        event.PublishedAt,
		ScheduledPublishAt: event.ScheduledPublishAt,
		LastPublishedAt:    event.LastPublishedAt,

		// Audit
		IsActive:   event.IsActive,
		CreatedAt:  event.CreatedAt,
		UpdatedAt:  event.UpdatedAt,
		DeletedAt:  deletedAt,
		DeletedBy:  deletedBy,
		RestoredAt: event.RestoredAt,
		RestoredBy: restoredBy,
	}
}

// toDomainEvent converts EventModel to domain.Event
func toDomainEvent(model *EventModel) *domain.Event {
	if model == nil {
		return nil
	}

	var deletedAt *time.Time
	if model.DeletedAt.Valid {
		deletedAt = &model.DeletedAt.Time
	}

	var deletedBy *string
	if model.DeletedBy != nil {
		deletedBy = model.DeletedBy
	}

	var restoredBy *string
	if model.RestoredBy != nil {
		restoredBy = model.RestoredBy
	}

	event := &domain.Event{
		// Core Identity
		ID:               model.ID,
		Slug:             model.Slug,
		Name:             model.Name,
		DisplayName:      model.DisplayName,
		Description:      model.Description,
		ShortDescription: model.ShortDescription,
		Tags:             fromJSONBToStringArray(model.Tags),
		Language:         model.Language,

		// Relations
		EventTypeID:          model.EventTypeID,
		EventStatusID:        model.EventStatusID,
		CategoryID:           model.CategoryID,
		EventFormatID:        model.EventFormatID,
		CertificateTemplateID: model.CertificateTemplateID,

		// Ownership
		InstitutionID: model.InstitutionID,
		CreatedBy:     model.CreatedBy,

		// Schedule & Venue
		StartDate:   model.StartDate,
		EndDate:     model.EndDate,
		IsMultiDay:  model.IsMultiDay,
		IsRecurring: model.IsRecurring,

		// Recurrence
		RecurrencePatternID:   model.RecurrencePatternID,
		RecurrenceInterval:    model.RecurrenceInterval,
		RecurrenceEndsOn:      model.RecurrenceEndsOn,
		RecurrenceOccurrences: model.RecurrenceOccurrences,
		RecurrenceDaysOfWeek:  fromJSONBToStringArray(model.RecurrenceDaysOfWeek),
		RecurrenceDayOfMonth:  model.RecurrenceDayOfMonth,
		RecurrenceWeekOfMonth: model.RecurrenceWeekOfMonth,

		// Venue
		VenueName:          model.VenueName,
		VenueAddress:       model.VenueAddress,
		VenueCity:          model.VenueCity,
		VenueCountry:       model.VenueCountry,
		VenueCoordinates:   fromJSONBToMapFloat64(model.VenueCoordinates),
		IsVirtual:          model.IsVirtual,
		IsHybrid:           model.IsHybrid,
		VirtualPlatform:    model.VirtualPlatform,
		VirtualPlatformURL: model.VirtualPlatformURL,
		InPersonLocation:   model.InPersonLocation,
		ZoomLink:           model.ZoomLink,
		MeetLink:           model.MeetLink,

		// Ticketing & Capacity
		IsFreeEvent:        model.IsFree,
		Capacity:           model.Capacity,
		CurrentAttendees:   model.CurrentAttendees,
		WaitlistEnabled:    model.WaitlistEnabled,
		WaitlistCapacity:   model.WaitlistCapacity,
		TicketSalesStart:   model.TicketSalesStart,
		TicketSalesEnd:     model.TicketSalesEnd,
		MinTicketsPerOrder: model.MinTicketsPerOrder,
		MaxTicketsPerOrder: model.MaxTicketsPerOrder,

		// Access & Privacy
		Visibility:          model.Visibility,
		Password:            model.Password,
		InviteOnly:          model.InviteOnly,
		InvitedEmails:       fromJSONBToStringArray(model.InvitedEmails),
		RequiresApproval:    model.RequiresApproval,
		ApprovalRequiredFor: fromJSONBToStringArray(model.ApprovalRequiredFor),

		// Monetization
		IsFeatured:          model.IsFeatured,
		FeaturedUntil:       model.FeaturedUntil,
		CertificateEnabled:  model.CertificateEnabled,
		CertificatePrice:    model.CertificatePrice,
		EarlyBirdDiscountPercentage: model.EarlyBirdDiscountPercentage,
		GroupDiscountPercentage: model.GroupDiscountPercentage,
		GroupMinAttendees:   model.GroupMinAttendees,

		// Media
		ImageURL:     model.ImageURL,
		ThumbnailURL: model.ThumbnailURL,

		// Social
		SocialLinks:        fromJSONBToMapString(model.SocialLinks),
		HasLivestream:      model.HasLivestream,
		LivestreamURL:      model.LivestreamURL,
		RecordingAvailable: model.RecordingAvailable,
		RecordingURL:       model.RecordingURL,

		// Metadata
		Metadata:           fromJSONBToMapInterface(model.Metadata),
		Version:            model.Version,
		PublishedAt:        model.PublishedAt,
		ScheduledPublishAt: model.ScheduledPublishAt,
		LastPublishedAt:    model.LastPublishedAt,

		// Audit
		IsActive:   model.IsActive,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		DeletedAt:  deletedAt,
		DeletedBy:  deletedBy,
		RestoredAt: model.RestoredAt,
		RestoredBy: restoredBy,
	}

	// Set SEO
	event.SEO.Title = model.SEOTitle
	event.SEO.Description = model.SEODescription
	event.SEO.Keywords = fromJSONBToStringArray(model.SEOKeywords)
	event.SEO.CanonicalURL = model.SEOCanonicalURL
	event.SEO.Robots = model.SEORobots
	event.SEO.NoIndex = model.SEONoIndex

	// Set Open Graph
	event.OpenGraph.Title = model.OGTitle
	event.OpenGraph.Description = model.OGDescription
	event.OpenGraph.ImageURL = model.OGImageURL
	event.OpenGraph.Type = model.OGType

	// Set Twitter
	event.Twitter.Card = model.TwitterCard
	event.Twitter.Title = model.TwitterTitle
	event.Twitter.Description = model.TwitterDescription
	event.Twitter.ImageURL = model.TwitterImageURL

	// Set SchemaOrg
	event.SchemaOrg = fromJSONBToMapInterface(model.SchemaOrg)

	return event
}

// ============================================================
// USER INFO MAPPERS
// ============================================================

// toDomainUserInfo converts UserModel to domain.UserInfo
func toDomainUserInfo(model *UserModel) *domain.UserInfo {
	if model == nil {
		return nil
	}

	return &domain.UserInfo{
		ID:          model.ID,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Email:       model.Email,
		Phone:       model.Phone,
	}
}

// ============================================================
// EVENT WITH CREATOR MAPPERS
// ============================================================

// toDomainEventsWithCreator converts multiple EventModels to domain.Events with creator info
func toDomainEventsWithCreator(models []EventModel) []*domain.Event {
	if len(models) == 0 {
		return []*domain.Event{}
	}

	events := make([]*domain.Event, len(models))
	for i, model := range models {
		event := toDomainEvent(&model)

		// Populate creator info from join fields
		if model.CreatorName != "" || model.CreatorEmail != "" {
			creator := &domain.UserInfo{
				ID:          model.CreatedBy,
				Name:        model.CreatorName,
				DisplayName: model.CreatorDisplayName,
				Email:       model.CreatorEmail,
				Phone:       model.CreatorPhone,
			}
			event.Creator = creator
		}

		events[i] = event
	}
	return events
}

// ============================================================
// CHILD ENTITY MAPPERS
// ============================================================

// toDomainSchedules converts EventScheduleModel to domain.EventSchedule
func toDomainSchedules(models []EventScheduleModel) []domain.EventSchedule {
	if len(models) == 0 {
		return []domain.EventSchedule{}
	}

	schedules := make([]domain.EventSchedule, len(models))
	for i, m := range models {
		schedules[i] = domain.EventSchedule{
			ID:            m.ID,
			EventID:       m.EventID,
			SessionName:   m.SessionName,
			SessionNumber: m.SessionNumber,
			StartDate:     m.StartDate,
			EndDate:       m.EndDate,
			StartTime:     m.StartTime,
			EndTime:       m.EndTime,
			Timezone:      m.Timezone,
			Location:      m.Location,
			IsVirtual:     m.IsVirtual,
			ZoomLink:      m.ZoomLink,
			MeetLink:      m.MeetLink,
			MaxAttendees:  m.MaxAttendees,
		}
	}
	return schedules
}

// toModelSchedules converts domain.EventSchedule to EventScheduleModel
func toModelSchedules(eventID string, schedules []domain.EventSchedule) []EventScheduleModel {
	models := make([]EventScheduleModel, len(schedules))
	for i, s := range schedules {
		models[i] = EventScheduleModel{
			EventID:       eventID,
			SessionName:   s.SessionName,
			SessionNumber: s.SessionNumber,
			StartDate:     s.StartDate,
			EndDate:       s.EndDate,
			StartTime:     s.StartTime,
			EndTime:       s.EndTime,
			Timezone:      s.Timezone,
			Location:      s.Location,
			IsVirtual:     s.IsVirtual,
			ZoomLink:      s.ZoomLink,
			MeetLink:      s.MeetLink,
			MaxAttendees:  s.MaxAttendees,
		}
	}
	return models
}

// toDomainTickets converts EventTicketModel to domain.EventTicket
func toDomainTickets(models []EventTicketModel) []domain.EventTicket {
	if len(models) == 0 {
		return []domain.EventTicket{}
	}

	tickets := make([]domain.EventTicket, len(models))
	for i, m := range models {
		tickets[i] = domain.EventTicket{
			ID:                 m.ID,
			EventID:            m.EventID,
			TicketTypeID:       m.TicketTypeID,
			Name:               m.Name,
			Description:        m.Description,
			Price:              m.Price,
			Quantity:           m.Quantity,
			MaxPerPerson:       m.MaxPerPerson,
			EarlyBirdDeadline:  m.EarlyBirdDeadline,
			GroupMinAttendees:  m.GroupMinAttendees,
			GroupDiscount:      m.GroupDiscount,
			SortOrder:          m.SortOrder,
			IsActive:           m.IsActive,
		}
	}
	return tickets
}

// toModelTickets converts domain.EventTicket to EventTicketModel
func toModelTickets(eventID string, tickets []domain.EventTicket) []EventTicketModel {
	models := make([]EventTicketModel, len(tickets))
	for i, t := range tickets {
		models[i] = EventTicketModel{
			EventID:            eventID,
			TicketTypeID:       t.TicketTypeID,
			Name:               t.Name,
			Description:        t.Description,
			Price:              t.Price,
			Quantity:           t.Quantity,
			MaxPerPerson:       t.MaxPerPerson,
			EarlyBirdDeadline:  t.EarlyBirdDeadline,
			GroupMinAttendees:  t.GroupMinAttendees,
			GroupDiscount:      t.GroupDiscount,
			SortOrder:          t.SortOrder,
			IsActive:           t.IsActive,
		}
	}
	return models
}

// toDomainSpeakers converts EventSpeakerModel to domain.EventSpeaker
func toDomainSpeakers(models []EventSpeakerModel) []domain.EventSpeaker {
	if len(models) == 0 {
		return []domain.EventSpeaker{}
	}

	speakers := make([]domain.EventSpeaker, len(models))
	for i, m := range models {
		speakers[i] = domain.EventSpeaker{
			ID:          m.ID,
			EventID:     m.EventID,
			Name:        m.Name,
			Title:       m.Title,
			Bio:         m.Bio,
			PhotoURL:    m.PhotoURL,
			SocialLinks: fromJSONBToMapString(m.SocialLinks),
			SortOrder:   m.SortOrder,
			IsKeynote:   m.IsKeynote,
		}
	}
	return speakers
}

// toModelSpeakers converts domain.EventSpeaker to EventSpeakerModel
func toModelSpeakers(eventID string, speakers []domain.EventSpeaker) []EventSpeakerModel {
	models := make([]EventSpeakerModel, len(speakers))
	for i, s := range speakers {
		models[i] = EventSpeakerModel{
			EventID:     eventID,
			Name:        s.Name,
			Title:       s.Title,
			Bio:         s.Bio,
			PhotoURL:    s.PhotoURL,
			SocialLinks: toJSONB(s.SocialLinks),
			SortOrder:   s.SortOrder,
			IsKeynote:   s.IsKeynote,
		}
	}
	return models
}

// toDomainMaterials converts EventMaterialModel to domain.EventMaterial
func toDomainMaterials(models []EventMaterialModel) []domain.EventMaterial {
	if len(models) == 0 {
		return []domain.EventMaterial{}
	}

	materials := make([]domain.EventMaterial, len(models))
	for i, m := range models {
		materials[i] = domain.EventMaterial{
			ID:             m.ID,
			EventID:        m.EventID,
			MaterialTypeID: m.MaterialTypeID,
			Title:          m.Title,
			Description:    m.Description,
			URL:            m.URL,
			IsPreEvent:     m.IsPreEvent,
			SortOrder:      m.SortOrder,
			FileSize:       m.FileSize,
			MimeType:       m.MimeType,
		}
	}
	return materials
}

// toModelMaterials converts domain.EventMaterial to EventMaterialModel
func toModelMaterials(eventID string, materials []domain.EventMaterial) []EventMaterialModel {
	models := make([]EventMaterialModel, len(materials))
	for i, m := range materials {
		models[i] = EventMaterialModel{
			EventID:        eventID,
			MaterialTypeID: m.MaterialTypeID,
			Title:          m.Title,
			Description:    m.Description,
			URL:            m.URL,
			IsPreEvent:     m.IsPreEvent,
			SortOrder:      m.SortOrder,
			FileSize:       m.FileSize,
			MimeType:       m.MimeType,
		}
	}
	return models
}

// ============================================================
// EVENT TYPE - DOMAIN VALUE MAPPERS
// ============================================================

// toDomainEventTypeEntity converts EventTypeModel to domain.EventType
func toDomainEventTypeEntity(model *EventTypeModel) *domain.EventType {
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

// toDomainEventStatusEntity converts EventStatusModel to domain.EventStatus
func toDomainEventStatusEntity(model *EventStatusModel) *domain.EventStatus {
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