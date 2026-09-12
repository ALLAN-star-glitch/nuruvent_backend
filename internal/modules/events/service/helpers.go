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
func (s *eventService) applyRecurrence(ctx context.Context, event *domain.Event, input *RecurrenceInput) error {
	if input == nil {
		return nil
	}
	if input.Pattern == "" {
		return errors.New("recurrence pattern is required")
	}

	// Resolve slug → UUID via lookup table
	pattern, err := s.repo.GetRecurrencePatternBySlug(ctx, input.Pattern)
	if err != nil {
		return fmt.Errorf("failed to resolve recurrence pattern %q: %w", input.Pattern, err)
	}
	if pattern == nil {
		return fmt.Errorf("unknown recurrence pattern: %q", input.Pattern)
	}
	if !pattern.IsActive {
		return fmt.Errorf("recurrence pattern %q is not active", input.Pattern)
	}

	event.IsRecurring = true
	event.RecurrencePatternID = &pattern.ID      // ✅ UUID → persisted
	event.RecurrencePatternSlug = pattern.Slug   // ✅ slug → runtime
	event.RecurrenceInterval = input.Interval
	event.RecurrenceDaysOfWeek = input.DaysOfWeek
	event.RecurrenceDayOfMonth = input.DayOfMonth
	event.RecurrenceWeekOfMonth = input.WeekOfMonth
	event.RecurrenceOccurrences = input.Occurrences

	if input.EndsOn != nil && *input.EndsOn != "" {
		endsOn, err := time.Parse("2006-01-02", *input.EndsOn)
		if err != nil {
			return fmt.Errorf("invalid recurrence ends_on: %w", err)
		}
		event.RecurrenceEndsOn = &endsOn
	}

	return nil
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

// resolveEventTeamDomain computes personal vs institution team domain string
func (s *eventService) resolveEventTeamDomain(event *domain.Event) string {
	if event == nil || event.TeamID == "" {
		return ""
	}

	// Dynamic evaluation check
	if event.CreatedBy != "" && event.CreatedBy == event.TeamID {
		return domain.PersonalTeamDomain(event.CreatedBy)
	}

	return domain.InstitutionTeamDomain(event.TeamID)
}

// validateUpdatePermission checks update permissions using 2-tier domain resolution
func (s *eventService) validateUpdatePermission(ctx context.Context, event *domain.Event, userID string) error {
	teamDomain := s.resolveEventTeamDomain(event)

	// Tier 1: Check Team Domain
	allowed, err := s.permChecker.CanUpdateEvent(ctx, userID, teamDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}

	// Tier 2: Check Account Domain fallback (if present)
	if !allowed && event.AccountID != "" {
		accountDomain := domain.AccountDomain(event.AccountID)
		allowed, err = s.permChecker.CanUpdateEvent(ctx, userID, accountDomain)
		if err != nil {
			return fmt.Errorf("account permission check failed: %w", err)
		}
	}

	if !allowed {
		log.Printf("❌ Permission denied: user %s cannot update event %s", userID, event.ID)
		return domain.ErrForbidden
	}

	return nil
}

// validateViewCreatorPermission checks view creator permissions using 2-tier domain resolution
func (s *eventService) validateViewCreatorPermission(ctx context.Context, event *domain.Event, userID string) bool {
	teamDomain := s.resolveEventTeamDomain(event)

	// Tier 1: Check Team Domain
	allowed, err := s.permChecker.CanViewCreator(ctx, userID, teamDomain)
	if err != nil {
		log.Printf("⚠️ Failed to check view_creator permission: %v", err)
		allowed = false
	}

	// Tier 2: Check Account Domain fallback (if present)
	if !allowed && event.AccountID != "" {
		accountDomain := domain.AccountDomain(event.AccountID)
		allowedAccount, err := s.permChecker.CanViewCreator(ctx, userID, accountDomain)
		if err != nil {
			log.Printf("⚠️ Failed to check account view_creator permission: %v", err)
		} else {
			allowed = allowedAccount
		}
	}

	return allowed
}

// getEventAndCheckUpdatePermission gets event and checks update permission using 2-tier domain fallback
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

	if err := s.validateUpdatePermission(ctx, event, userID); err != nil {
		return nil, err
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

	return s.validateViewCreatorPermission(ctx, event, userID)
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

func (s *eventService) getOrganizerInfo(ctx context.Context, event *domain.Event) (*domain.OrganizerInfo, error) {
    if event == nil {
        return nil, errors.New("event is nil")
    }
    if event.Organizer != nil {
        return event.Organizer, nil
    }
    if event.TeamID == "" {
        return nil, errors.New("event has no team ID")
    }
    if s.organizer == nil {
        return nil, errors.New("organizer provider is not configured")
    }
    return s.organizer.GetOrganizer(ctx, event.TeamID)
}

// getTeamByID retrieves a team by ID using the repository
func (s *eventService) getTeamByID(ctx context.Context, teamID string) (*TeamInfo, error) {
	if teamID == "" {
		return nil, errors.New("team ID is required")
	}

	return &TeamInfo{
		ID:        teamID,
		Type:      "institution",
		AccountID: "",
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



// checkEventCreatePermission is the shared authorization check used by
// CreateDraft, CreateEvent, and GenerateEventDraft.
//
// It runs a two-tier Casbin check:
//   Tier 1: event:create against the resolved team domain
//   Tier 2: event:create against the account domain (fallback)
func (s *eventService) checkEventCreatePermission(
	ctx context.Context,
	userID, teamID, teamType, accountID string,
) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if teamID == "" {
		return fmt.Errorf("team ID is required")
	}

	teamDomain := s.resolveTeamDomainFromParts(teamID, teamType, userID)
	log.Printf("🔍 Tier 1: event:create check user=%s domain=%s", userID, teamDomain)

	allowed, err := s.permChecker.CanCreateEvent(ctx, userID, teamDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}

	if !allowed && accountID != "" {
		accountDomain := domain.AccountDomain(accountID)
		log.Printf("🔍 Tier 2: event:create check user=%s domain=%s", userID, accountDomain)

		allowed, err = s.permChecker.CanCreateEvent(ctx, userID, accountDomain)
		if err != nil {
			return fmt.Errorf("account permission check failed: %w", err)
		}
	}

	if !allowed {
		return domain.ErrForbidden
	}
	return nil
}

// resolveTeamDomainFromParts is the primitive form of resolveTeamDomain.
// Handy when callers don't have a CreateDraftCommand.
func (s *eventService) resolveTeamDomainFromParts(teamID, teamType, userID string) string {
	if teamType == "personal" {
		return domain.PersonalTeamDomain(teamID)
	}
	if teamType == "institution" {
		return domain.InstitutionTeamDomain(teamID)
	}
	if userID != "" && userID == teamID {
		return domain.PersonalTeamDomain(userID)
	}
	return domain.InstitutionTeamDomain(teamID)
}