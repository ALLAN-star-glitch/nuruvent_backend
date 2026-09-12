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

// convertSchedules converts schedule inputs to domain schedules.
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

// convertTickets converts ticket inputs to domain tickets.
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

// convertSpeakers converts speaker inputs to domain speakers.
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

// convertMaterials converts material inputs to domain materials.
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

// applyRecurrence applies recurrence to an event.
func (s *eventService) applyRecurrence(ctx context.Context, event *domain.Event, input *RecurrenceInput) error {
	if input == nil {
		return nil
	}
	if input.Pattern == "" {
		return errors.New("recurrence pattern is required")
	}

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
	event.RecurrencePatternID = &pattern.ID
	event.RecurrencePatternSlug = pattern.Slug
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

// applySEO applies SEO to an event.
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
// PERMISSION HELPERS (shared across services)
// ============================================================
//
// POST-REVAMP: every permission check runs against the event's parent
// ACCOUNT domain. Teams are not authorization domains. Team membership
// is enforced by the service as data (team_members) when it matters.
//
// Every event must have an AccountID. If it's missing on a loaded event,
// the caller is responsible for backfilling it (usually by resolving
// through the team).

// validateUpdatePermission checks whether the user may update this event.
func (s *eventService) validateUpdatePermission(ctx context.Context, event *domain.Event, userID string) error {
	if event == nil {
		return errors.New("event is nil")
	}

	accountID := event.AccountID
	if accountID == "" {
		return errors.New("event has no account ID; cannot authorize update")
	}

	accountDomain := domain.AccountDomain(accountID)
	log.Printf("🔍 AUTHZ: event:update check user=%s domain=%s event=%s",
		userID, accountDomain, event.ID)

	allowed, err := s.permChecker.CanUpdateEvent(ctx, userID, accountDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		log.Printf("❌ Permission denied: user %s cannot update event %s", userID, event.ID)
		return domain.ErrForbidden
	}

	return nil
}

// validateViewCreatorPermission checks whether the user may see the
// event's creator details.
//
// Returns false on denial. Permission-check errors are logged and treated
// as denial (fail closed).
func (s *eventService) validateViewCreatorPermission(ctx context.Context, event *domain.Event, userID string) bool {
	if event == nil {
		return false
	}

	accountID := event.AccountID
	if accountID == "" {
		log.Printf("⚠️ event %s has no AccountID; denying view_creator", event.ID)
		return false
	}

	accountDomain := domain.AccountDomain(accountID)
	allowed, err := s.permChecker.CanViewCreator(ctx, userID, accountDomain)
	if err != nil {
		log.Printf("⚠️ Failed to check view_creator permission: %v", err)
		return false
	}
	return allowed
}

// getEventAndCheckUpdatePermission loads the event and verifies the user
// may update it.
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

// canViewCreatorInfo checks whether a user may see creator info for an event.
func (s *eventService) canViewCreatorInfo(ctx context.Context, userID string, event *domain.Event) bool {
	if userID == "" {
		return false
	}
	if event.CreatedBy == userID {
		return true
	}
	return s.validateViewCreatorPermission(ctx, event, userID)
}

// getCreatorInfo fetches creator information using the UserInfoProvider.
func (s *eventService) getCreatorInfo(ctx context.Context, userID string) *domain.UserInfo {
	if userID == "" {
		return nil
	}

	user, err := s.userInfo.GetUserByIDWithDetails(ctx, userID)
	if err != nil {
		log.Printf("⚠️ Failed to get user info for %s: %v", userID, err)
		return nil
	}
	return user
}

// getOrganizerInfo returns organizer information for an event.
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

// checkEventCreatePermission is the shared authorization check used by
// CreateDraft, CreateEvent, and any other event-creation path.
//
// POST-REVAMP: single check against the account domain. The teamID and
// teamType arguments are accepted for signature compatibility but are no
// longer used for authorization.
func (s *eventService) checkEventCreatePermission(
	ctx context.Context,
	userID, teamID, teamType, accountID string,
) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	if accountID == "" {
		return errors.New("account ID is required for permission check")
	}

	accountDomain := domain.AccountDomain(accountID)
	log.Printf("🔍 AUTHZ: event:create check user=%s domain=%s", userID, accountDomain)

	allowed, err := s.permChecker.CanCreateEvent(ctx, userID, accountDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return nil
}

// ============================================================
// LOGGING HELPERS
// ============================================================

// logBulkStatusResult logs the result of a bulk status operation.
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

// isEmptyTeam checks if a TeamFilter is empty.
func (s *eventService) isEmptyTeam(team domain.TeamFilter) bool {
	return team.ID == "" || team.Type == ""
}

// ============================================================
// TEAM INFO (kept for handlers that still need it)
// ============================================================

// TeamInfo represents basic team information for handlers.
//
// Note: this is a read-only DTO for display and query purposes. It is not
// used for authorization. If you need the team's parent account ID, call
// the repository's AccountIDForTeam method directly.
type TeamInfo struct {
	ID        string
	Type      string // "personal" or "institution"
	AccountID string
}