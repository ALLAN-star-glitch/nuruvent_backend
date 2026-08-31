// internal/modules/events/delivery/eventhandler/response.go

package eventhandler

import (
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// NewEventResponseFromEvent converts a domain Event to EventResponse
// This version does NOT include creator info - use NewEventResponseFromEventWithCreator for that
func NewEventResponseFromEvent(event *domain.Event) EventResponse {
	resp := EventResponse{
		// Core Identity
		ID:               event.ID,
		Slug:             event.Slug,
		Name:             event.Name,
		DisplayName:      event.DisplayName,
		Description:      event.Description,
		ShortDescription: event.ShortDescription,
		Tags:             event.Tags,
		Language:         event.Language,

		// Ownership
		InstitutionID: "",
		OwnerType:     "personal", // Default, will be determined from InstitutionID

		// Schedule
		IsMultiDay:  event.IsMultiDay,
		IsRecurring: event.IsRecurring,

		// Venue
		IsVirtual:          event.IsVirtual,
		IsHybrid:           event.IsHybrid,
		InPersonLocation:   event.InPersonLocation,
		VirtualPlatform:    event.VirtualPlatform,
		VirtualPlatformURL: event.VirtualPlatformURL,
		ZoomLink:           event.ZoomLink,
		MeetLink:           event.MeetLink,

		// Tickets
		IsFree:            event.IsFreeEvent,
		Capacity:          0,
		CurrentAttendees:  event.CurrentAttendees,
		WaitlistEnabled:   event.WaitlistEnabled,
		MinTicketsPerOrder: event.MinTicketsPerOrder,

		// Access & Privacy
		Visibility:    event.Visibility,
		IsPrivate:     event.Visibility == "private",
		InviteOnly:    event.InviteOnly,
		InvitedEmails: event.InvitedEmails,

		// Monetization
		IsFeatured:          event.IsFeatured,
		CertificateEnabled:  event.CertificateEnabled,
		CertificatePrice:    event.CertificatePrice,

		// Media
		ImageURL:     event.ImageURL,
		ThumbnailURL: event.ThumbnailURL,

		// Social
		SocialLinks:        event.SocialLinks,
		HasLivestream:      event.HasLivestream,
		LivestreamURL:      event.LivestreamURL,
		RecordingAvailable: event.RecordingAvailable,
		RecordingURL:       event.RecordingURL,

		// Metadata
		Version: event.Version,

		// Audit
		IsActive:  event.IsActive,
		CreatedAt: event.CreatedAt.Format(time.RFC3339),
		UpdatedAt: event.UpdatedAt.Format(time.RFC3339),
	}

	// Set InstitutionID and OwnerType
	if event.InstitutionID != nil && *event.InstitutionID != "" {
		resp.InstitutionID = *event.InstitutionID
		resp.OwnerType = "institution"
	} else {
		resp.OwnerType = "personal"
	}

	// Set Capacity
	if event.Capacity != nil {
		resp.Capacity = *event.Capacity
	}

	// Set WaitlistCapacity
	if event.WaitlistCapacity != nil {
		resp.WaitlistCapacity = *event.WaitlistCapacity
	}

	// Set MaxTicketsPerOrder
	if event.MaxTicketsPerOrder != nil {
		resp.MaxTicketsPerOrder = *event.MaxTicketsPerOrder
	}

	// Set StartDate and EndDate
	if !event.StartDate.IsZero() {
		resp.StartDate = event.StartDate.Format(time.RFC3339)
	}
	if event.EndDate != nil {
		resp.EndDate = event.EndDate.Format(time.RFC3339)
	}

	// Set Venue
	if event.VenueName != "" || event.VenueAddress != "" || event.VenueCity != "" || event.VenueCountry != "" {
		resp.Venue = &VenueDTO{
			Name:        event.VenueName,
			Address:     event.VenueAddress,
			City:        event.VenueCity,
			Country:     event.VenueCountry,
			Coordinates: event.VenueCoordinates,
		}
	}

	// Set Recurrence
	if event.IsRecurring {
		recurrence := &RecurrenceDTO{
			Interval:    event.RecurrenceInterval,
			DaysOfWeek:  event.RecurrenceDaysOfWeek,
			Occurrences: 0,
		}
		if event.RecurrencePatternID != nil {
			recurrence.Pattern = *event.RecurrencePatternID
		}
		if event.RecurrenceDayOfMonth != nil {
			recurrence.DayOfMonth = *event.RecurrenceDayOfMonth
		}
		if event.RecurrenceWeekOfMonth != nil {
			recurrence.WeekOfMonth = *event.RecurrenceWeekOfMonth
		}
		if event.RecurrenceEndsOn != nil {
			recurrence.EndsOn = event.RecurrenceEndsOn.Format("2006-01-02")
		}
		if event.RecurrenceOccurrences != nil {
			recurrence.Occurrences = *event.RecurrenceOccurrences
		}
		resp.Recurrence = recurrence
	}

	// Set Schedules
	if len(event.Schedules) > 0 {
		schedules := make([]ScheduleDTO, len(event.Schedules))
		for i, s := range event.Schedules {
			schedule := ScheduleDTO{
				ID:            s.ID,
				SessionName:   s.SessionName,
				SessionNumber: s.SessionNumber,
				StartDate:     s.StartDate.Format("2006-01-02"),
				StartTime:     s.StartTime,
				EndTime:       s.EndTime,
				Timezone:      s.Timezone,
				Location:      s.Location,
				IsVirtual:     s.IsVirtual,
				ZoomLink:      s.ZoomLink,
				MeetLink:      s.MeetLink,
			}
			if s.EndDate != nil {
				schedule.EndDate = s.EndDate.Format("2006-01-02")
			}
			if s.MaxAttendees != nil {
				schedule.MaxAttendees = *s.MaxAttendees
			}
			schedules[i] = schedule
		}
		resp.Schedules = schedules
	}

	// Set Tickets
	if len(event.Tickets) > 0 {
		tickets := make([]TicketDTO, len(event.Tickets))
		for i, t := range event.Tickets {
			ticket := TicketDTO{
				ID:          t.ID,
				Name:        t.Name,
				Description: t.Description,
				Price:       t.Price,
				Quantity:    t.Quantity,
				SortOrder:   t.SortOrder,
				IsActive:    t.IsActive,
			}
			if t.MaxPerPerson != nil {
				ticket.MaxPerPerson = *t.MaxPerPerson
			}
			if t.EarlyBirdDeadline != nil {
				ticket.EarlyBirdDeadline = t.EarlyBirdDeadline.Format(time.RFC3339)
			}
			if t.GroupMinAttendees != nil {
				ticket.GroupMinAttendees = *t.GroupMinAttendees
			}
			if t.GroupDiscount != nil {
				ticket.GroupDiscount = *t.GroupDiscount
			}
			tickets[i] = ticket
		}
		resp.Tickets = tickets
	}

	// Set Speakers
	if len(event.Speakers) > 0 {
		speakers := make([]SpeakerDTO, len(event.Speakers))
		for i, s := range event.Speakers {
			speakers[i] = SpeakerDTO{
				ID:          s.ID,
				Name:        s.Name,
				Title:       s.Title,
				Bio:         s.Bio,
				PhotoURL:    s.PhotoURL,
				SocialLinks: s.SocialLinks,
				IsKeynote:   s.IsKeynote,
				SortOrder:   s.SortOrder,
			}
		}
		resp.Speakers = speakers
	}

	// Set Materials
	if len(event.Materials) > 0 {
		materials := make([]MaterialDTO, len(event.Materials))
		for i, m := range event.Materials {
			materials[i] = MaterialDTO{
				ID:          m.ID,
				Title:       m.Title,
				Type:        m.MaterialTypeID,
				URL:         m.URL,
				Description: m.Description,
				IsPreEvent:  m.IsPreEvent,
				SortOrder:   m.SortOrder,
			}
		}
		resp.Materials = materials
	}

	// Set SEO
	if event.SEO.Title != "" || event.SEO.Description != "" || len(event.SEO.Keywords) > 0 {
		resp.SEO = &SEODTO{
			MetaTitle:          event.SEO.Title,
			MetaDescription:    event.SEO.Description,
			MetaKeywords:       event.SEO.Keywords,
			CanonicalURL:       event.SEO.CanonicalURL,
			Robots:             event.SEO.Robots,
			NoIndex:            event.SEO.NoIndex,
			OGTitle:            event.OpenGraph.Title,
			OGDescription:      event.OpenGraph.Description,
			OGImageURL:         event.OpenGraph.ImageURL,
			OGType:             event.OpenGraph.Type,
			TwitterCard:        event.Twitter.Card,
			TwitterTitle:       event.Twitter.Title,
			TwitterDescription: event.Twitter.Description,
			TwitterImageURL:    event.Twitter.ImageURL,
		}
	}

	// Set SchemaOrg
	if len(event.SchemaOrg) > 0 {
		resp.SchemaOrg = event.SchemaOrg
	}

	// Set Published timestamps
	if event.PublishedAt != nil {
		publishedAt := event.PublishedAt.Format(time.RFC3339)
		resp.PublishedAt = &publishedAt
	}
	if event.ScheduledPublishAt != nil {
		scheduledPublishAt := event.ScheduledPublishAt.Format(time.RFC3339)
		resp.ScheduledPublishAt = &scheduledPublishAt
	}
	if event.LastPublishedAt != nil {
		lastPublishedAt := event.LastPublishedAt.Format(time.RFC3339)
		resp.LastPublishedAt = &lastPublishedAt
	}

	// Set DeletedAt
	if event.DeletedAt != nil {
		deletedAt := event.DeletedAt.Format(time.RFC3339)
		resp.DeletedAt = &deletedAt
	}

	return resp
}

// NewEventResponseFromEventWithCreator converts a domain Event to EventResponse with creator info
// This should only be used when the caller has permission to view creator info
func NewEventResponseFromEventWithCreator(event *domain.Event) EventResponse {
	resp := NewEventResponseFromEvent(event)

	// Set Creator if available
	if event.Creator != nil {
		resp.Creator = &CreatorDTO{
			ID:          event.Creator.ID,
			Name:        event.Creator.Name,
			DisplayName: event.Creator.DisplayName,
			Email:       event.Creator.Email,
			Phone:       event.Creator.Phone,
		}
	}

	return resp
}

// NewEventResponseFromEvents converts multiple domain Events to EventResponse slice
// This version does NOT include creator info
func NewEventResponseFromEvents(events []*domain.Event) []EventResponse {
	responses := make([]EventResponse, len(events))
	for i, event := range events {
		responses[i] = NewEventResponseFromEvent(event)
	}
	return responses
}

// NewEventResponseFromEventsWithCreator converts multiple domain Events to EventResponse slice with creator info
// This should only be used when the caller has permission to view creator info
func NewEventResponseFromEventsWithCreator(events []*domain.Event) []EventResponse {
	responses := make([]EventResponse, len(events))
	for i, event := range events {
		responses[i] = NewEventResponseFromEventWithCreator(event)
	}
	return responses
}