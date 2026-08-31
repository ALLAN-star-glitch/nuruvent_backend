// internal/modules/events/service/helpers.go

package service

import (
	"fmt"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// CONVERTER FUNCTIONS
// ============================================================

// convertSchedules converts ScheduleInput to domain.EventSchedule
func (s *eventService) convertSchedules(inputs []ScheduleInput) ([]domain.EventSchedule, error) {
	schedules := make([]domain.EventSchedule, len(inputs))
	for i, input := range inputs {
		startDate, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format: %w", err)
		}

		var endDate *time.Time
		if input.EndDate != nil && *input.EndDate != "" {
			d, err := time.Parse("2006-01-02", *input.EndDate)
			if err == nil {
				endDate = &d
			}
		}

		// Handle pointer to string for ID
		var id string
		if input.ID != nil {
			id = *input.ID
		}

		schedules[i] = domain.EventSchedule{
			ID:            id,
			SessionName:   input.SessionName,
			SessionNumber: input.SessionNumber,
			StartDate:     startDate,
			EndDate:       endDate,
			StartTime:     input.StartTime,
			EndTime:       input.EndTime,
			Timezone:      input.Timezone,
			Location:      input.Location,
			IsVirtual:     input.IsVirtual,
			ZoomLink:      input.ZoomLink,
			MeetLink:      input.MeetLink,
			MaxAttendees:  input.MaxAttendees,
		}
	}
	return schedules, nil
}

// convertTickets converts TicketInput to domain.EventTicket
func (s *eventService) convertTickets(inputs []TicketInput) ([]domain.EventTicket, error) {
	tickets := make([]domain.EventTicket, len(inputs))
	for i, input := range inputs {
		var earlyBirdDeadline *time.Time
		if input.EarlyBirdDeadline != nil && *input.EarlyBirdDeadline != "" {
			d, err := time.Parse(time.RFC3339, *input.EarlyBirdDeadline)
			if err == nil {
				earlyBirdDeadline = &d
			}
		}

		// Handle pointer to string for ID
		var id string
		if input.ID != nil {
			id = *input.ID
		}

		tickets[i] = domain.EventTicket{
			ID:                 id,
			TicketTypeID:       input.TicketTypeID,
			Name:               input.Name,
			Description:        input.Description,
			Price:              input.Price,
			Quantity:           input.Quantity,
			MaxPerPerson:       input.MaxPerPerson,
			EarlyBirdDeadline:  earlyBirdDeadline,
			GroupMinAttendees:  input.GroupMinAttendees,
			GroupDiscount:      input.GroupDiscount,
		}
	}
	return tickets, nil
}

// convertSpeakers converts SpeakerInput to domain.EventSpeaker
func (s *eventService) convertSpeakers(inputs []SpeakerInput) ([]domain.EventSpeaker, error) {
	speakers := make([]domain.EventSpeaker, len(inputs))
	for i, input := range inputs {
		// Handle pointer to string for ID
		var id string
		if input.ID != nil {
			id = *input.ID
		}

		speakers[i] = domain.EventSpeaker{
			ID:          id,
			Name:        input.Name,
			Title:       input.Title,
			Bio:         input.Bio,
			PhotoURL:    input.PhotoURL,
			SocialLinks: input.SocialLinks,
			IsKeynote:   input.IsKeynote,
			SortOrder:   input.SortOrder,
		}
	}
	return speakers, nil
}

// convertMaterials converts MaterialInput to domain.EventMaterial
func (s *eventService) convertMaterials(materials []MaterialInput) ([]domain.EventMaterial, error) {
    result := make([]domain.EventMaterial, len(materials))
    for i, m := range materials {
        result[i] = domain.EventMaterial{
            Title:          m.Title,
            MaterialTypeID: m.MaterialTypeID,  // ✅ Make sure this is being set
            URL:            m.URL,
            Description:    m.Description,
            IsPreEvent:     m.IsPreEvent,
            SortOrder:      m.SortOrder,
        }
    }
    return result, nil
}

// ============================================================
// APPLICATION FUNCTIONS
// ============================================================

// applyRecurrence applies recurrence settings to an event
func (s *eventService) applyRecurrence(event *domain.Event, input *RecurrenceInput) error {
	if input == nil {
		return nil
	}

	event.IsRecurring = true
	event.RecurrencePatternID = &input.Pattern
	event.RecurrenceInterval = input.Interval
	if input.Interval == 0 {
		event.RecurrenceInterval = 1
	}
	event.RecurrenceDaysOfWeek = input.DaysOfWeek
	event.RecurrenceDayOfMonth = input.DayOfMonth
	event.RecurrenceWeekOfMonth = input.WeekOfMonth

	if input.EndsOn != nil && *input.EndsOn != "" {
		endDate, err := time.Parse("2006-01-02", *input.EndsOn)
		if err == nil {
			event.RecurrenceEndsOn = &endDate
		}
	}
	event.RecurrenceOccurrences = input.Occurrences

	return nil
}

// applySEO applies SEO settings to an event
func (s *eventService) applySEO(event *domain.Event, input *SEOInput) {
	if input == nil {
		return
	}

	// Meta
	event.SEO.Title = input.MetaTitle
	event.SEO.Description = input.MetaDescription
	event.SEO.Keywords = input.MetaKeywords
	event.SEO.CanonicalURL = input.CanonicalURL
	event.SEO.Robots = input.Robots
	event.SEO.NoIndex = input.NoIndex

	// Open Graph
	event.OpenGraph.Title = input.OGTitle
	event.OpenGraph.Description = input.OGDescription
	event.OpenGraph.ImageURL = input.OGImageURL
	if input.OGType != "" {
		event.OpenGraph.Type = input.OGType
	}

	// Twitter
	event.Twitter.Card = input.TwitterCard
	event.Twitter.Title = input.TwitterTitle
	event.Twitter.Description = input.TwitterDescription
	event.Twitter.ImageURL = input.TwitterImageURL
}