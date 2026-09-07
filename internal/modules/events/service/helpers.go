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

// getEventAndCheckUpdatePermission gets event and checks update permission using team domain
func (s *eventService) getEventAndCheckUpdatePermission(ctx context.Context, eventID, userID string) (*domain.Event, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	event, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	// Create team domain from event
	teamDomain := domain.TeamDomain(event.TeamID)

	// Check if user can update events in this team
	allowed, err := s.permChecker.CanUpdateEvent(ctx, userID, teamDomain)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return nil, errors.New("insufficient permissions to update this event")
	}

	return event, nil
}

// canViewCreatorInfo checks if a user can view creator info for an event
func (s *eventService) canViewCreatorInfo(ctx context.Context, userID string, event *domain.Event) bool {
	// If no user, cannot view creator info
	if userID == "" {
		return false
	}

	// Event creator can always see their own info
	if event.CreatedBy == userID {
		return true
	}

	// Check if user has explicit view_creator permission
	teamDomain := domain.TeamDomain(event.TeamID)
	allowed, err := s.permChecker.CanViewCreator(ctx, userID, teamDomain)
	if err != nil {
		log.Printf("⚠️ Failed to check view_creator permission: %v", err)
		return false
	}

	return allowed
}

// getCreatorInfo fetches creator information using UserInfoProvider
func (s *eventService) getCreatorInfo(ctx context.Context, userID string) *domain.UserInfo {
	if userID == "" {
		return nil
	}

	// Get full user details with email, phone, etc.
	user, err := s.userInfo.GetUserByIDWithDetails(ctx, userID)
	if err != nil {
		log.Printf("⚠️ Failed to get user info for %s: %v", userID, err)
		return nil
	}
	return user
}

// getOrganizerInfo returns the public-facing organizer info for an event
func (s *eventService) getOrganizerInfo(ctx context.Context, event *domain.Event) (*domain.OrganizerInfo, error) {
	// First check if event has an organizer already set
	if event.Organizer != nil {
		return event.Organizer, nil
	}

	// Get the team to determine if it's personal or institution
	team, err := s.getTeamByID(ctx, event.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team info: %w", err)
	}
	if team == nil {
		return nil, fmt.Errorf("team not found: %s", event.TeamID)
	}

	if team.Type == "institution" {
		// Institution team - show institution name
		account, err := s.userInfo.GetAccountByID(ctx, team.AccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get account info: %w", err)
		}
		if account == nil {
			return nil, fmt.Errorf("account not found: %s", team.AccountID)
		}
		return &domain.OrganizerInfo{
			ID:          account.ID,
			Name:        account.Name,
			DisplayName: account.DisplayName,
			Type:        "institution",
			AvatarURL:   account.LogoURL,
			Slug:        account.Slug,
		}, nil
	}

	// Personal team - show user's name
	user, err := s.userInfo.GetUserByID(ctx, event.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found: %s", event.CreatedBy)
	}
	return &domain.OrganizerInfo{
		ID:          user.ID,
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Type:        "personal",
		AvatarURL:   user.AvatarURL,
		Slug:        user.Slug,
	}, nil
}

// getTeamByID retrieves a team by ID using the repository
// Note: This requires a Team repository method to be added
func (s *eventService) getTeamByID(ctx context.Context, teamID string) (*TeamInfo, error) {
	// This is a temporary implementation - you'll need to add a Team repository
	// or use the existing repository to query teams
	// For now, we'll return a basic team info
	if teamID == "" {
		return nil, errors.New("team ID is required")
	}
	
	// TODO: Implement team retrieval from database
	// This should be added to the repository interface
	
	return &TeamInfo{
		ID:        teamID,
		Type:      "institution", // This should be fetched from DB
		AccountID: "",            // This should be fetched from DB
	}, nil
}

// TeamInfo represents basic team information
type TeamInfo struct {
	ID        string
	Type      string // "personal" or "institution"
	AccountID string
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

// isEmptyTeam checks if a TeamFilter is empty
func (s *eventService) isEmptyTeam(team domain.TeamFilter) bool {
	return team.ID == "" || team.Type == ""
}