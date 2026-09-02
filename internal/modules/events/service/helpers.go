// internal/modules/events/service/helpers.go

package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// CONVERTER HELPERS
// ============================================================

// convertSchedules converts schedule inputs to domain schedules
func (s *eventService) convertSchedules(inputs []ScheduleInput) ([]domain.EventSchedule, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	schedules := make([]domain.EventSchedule, len(inputs))
	for i, input := range inputs {
		startDate, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format: %w", err)
		}

		var endDate *time.Time
		if input.EndDate != nil && *input.EndDate != "" {
			parsed, err := time.Parse("2006-01-02", *input.EndDate)
			if err != nil {
				return nil, fmt.Errorf("invalid end_date format: %w", err)
			}
			endDate = &parsed
		}

		schedules[i] = domain.EventSchedule{
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

// convertTickets converts ticket inputs to domain tickets
func (s *eventService) convertTickets(inputs []TicketInput) ([]domain.EventTicket, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	tickets := make([]domain.EventTicket, len(inputs))
	for i, input := range inputs {
		var earlyBirdDeadline *time.Time
		if input.EarlyBirdDeadline != nil && *input.EarlyBirdDeadline != "" {
			parsed, err := time.Parse(time.RFC3339, *input.EarlyBirdDeadline)
			if err != nil {
				return nil, fmt.Errorf("invalid early_bird_deadline format: %w", err)
			}
			earlyBirdDeadline = &parsed
		}

		tickets[i] = domain.EventTicket{
			TicketTypeID:      input.TicketTypeID,
			Name:              input.Name,
			Description:       input.Description,
			Price:             input.Price,
			Quantity:          input.Quantity,
			MaxPerPerson:      input.MaxPerPerson,
			EarlyBirdDeadline: earlyBirdDeadline,
			GroupMinAttendees: input.GroupMinAttendees,
			GroupDiscount:     input.GroupDiscount,
			IsActive:          true,
		}
	}
	return tickets, nil
}

// convertSpeakers converts speaker inputs to domain speakers
func (s *eventService) convertSpeakers(inputs []SpeakerInput) ([]domain.EventSpeaker, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	speakers := make([]domain.EventSpeaker, len(inputs))
	for i, input := range inputs {
		speakers[i] = domain.EventSpeaker{
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

// convertMaterials converts material inputs to domain materials
func (s *eventService) convertMaterials(inputs []MaterialInput) ([]domain.EventMaterial, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	materials := make([]domain.EventMaterial, len(inputs))
	for i, input := range inputs {
		materials[i] = domain.EventMaterial{
			Title:          input.Title,
			MaterialTypeID: input.MaterialTypeID,
			URL:            input.URL,
			Description:    input.Description,
			IsPreEvent:     input.IsPreEvent,
			SortOrder:      input.SortOrder,
		}
	}
	return materials, nil
}

// applyRecurrence applies recurrence to an event
func (s *eventService) applyRecurrence(event *domain.Event, input *RecurrenceInput) {
	if input == nil {
		return
	}

	event.IsRecurring = true
	event.RecurrencePatternID = &input.Pattern
	event.RecurrenceInterval = input.Interval
	event.RecurrenceDaysOfWeek = input.DaysOfWeek
	event.RecurrenceDayOfMonth = input.DayOfMonth
	event.RecurrenceWeekOfMonth = input.WeekOfMonth

	if input.EndsOn != nil && *input.EndsOn != "" {
		if endsOn, err := time.Parse("2006-01-02", *input.EndsOn); err == nil {
			event.RecurrenceEndsOn = &endsOn
		}
	}
	event.RecurrenceOccurrences = input.Occurrences
}

// applySEO applies SEO to an event
func (s *eventService) applySEO(event *domain.Event, input *SEOInput) {
	if input == nil {
		return
	}

	event.SEO.Title = input.MetaTitle
	event.SEO.Description = input.MetaDescription
	event.SEO.Keywords = input.MetaKeywords
	event.SEO.CanonicalURL = input.CanonicalURL
	event.SEO.Robots = input.Robots
	event.SEO.NoIndex = input.NoIndex

	event.OpenGraph.Title = input.OGTitle
	event.OpenGraph.Description = input.OGDescription
	event.OpenGraph.ImageURL = input.OGImageURL
	event.OpenGraph.Type = input.OGType

	event.Twitter.Card = input.TwitterCard
	event.Twitter.Title = input.TwitterTitle
	event.Twitter.Description = input.TwitterDescription
	event.Twitter.ImageURL = input.TwitterImageURL
}

// ============================================================
// PERMISSION HELPERS (Shared across services)
// ============================================================

// getEventAndCheckUpdatePermission gets event and checks update permission using Scope
func (s *eventService) getEventAndCheckUpdatePermission(ctx context.Context, eventID, userID string) (*domain.Event, error) {
	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	// Create scope from event
	scope := s.getScopeFromEvent(event)

	// Check if user can update events in this scope
	allowed, err := s.permChecker.CanUpdateEvent(ctx, userID, scope)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return nil, errors.New("insufficient permissions to update this event")
	}

	return event, nil
}

// getEventAndCheckDeletePermission gets event and checks delete permission using Scope
func (s *eventService) getEventAndCheckDeletePermission(ctx context.Context, eventID, userID string) (*domain.Event, error) {
	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	// Create scope from event
	scope := s.getScopeFromEvent(event)

	// Check if user can delete events in this scope
	allowed, err := s.permChecker.CanDeleteEvent(ctx, userID, scope)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return nil, errors.New("insufficient permissions to delete this event")
	}

	return event, nil
}

// ============================================================
// LOGGING HELPERS
// ============================================================

// logBulkStatusResult logs the result of a bulk status operation
func (s *eventService) logBulkStatusResult(operation string, result *BulkStatusResult, total int) {
	if result.ProcessedCount == total {
		log.Printf("✅ Bulk %s complete: %d events processed", operation, result.ProcessedCount)
	} else if result.ProcessedCount > 0 {
		log.Printf("⚠️ Bulk %s partial: %d succeeded, %d failed",
			operation, result.ProcessedCount, len(result.FailedIDs))
	} else {
		log.Printf("❌ Bulk %s failed: all %d events failed", operation, total)
	}
}