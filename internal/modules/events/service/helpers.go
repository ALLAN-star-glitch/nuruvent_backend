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

	accountID, err := s.resolveEventAccountID(ctx, event)
	if err != nil {
		return fmt.Errorf("cannot authorize update: %w", err)
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


// resolveEventAccountID returns the account ID that owns the event.
//
// The events table has no account_id column today, so the account is
// resolved through the event's team. If the schema later gains a direct
// account_id on events (and the repository populates it on load), that
// value is used directly and no team lookup happens.
func (s *eventService) resolveEventAccountID(ctx context.Context, event *domain.Event) (string, error) {
	if event == nil {
		return "", errors.New("event is nil")
	}

	if event.AccountID != "" {
		return event.AccountID, nil
	}

	if event.TeamID == "" {
		return "", errors.New("event has no team; cannot resolve account")
	}

	accountID, err := s.repo.AccountIDForTeam(ctx, event.TeamID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve account for team %s: %w", event.TeamID, err)
	}
	if accountID == "" {
		return "", fmt.Errorf("team %s has no account", event.TeamID)
	}

	return accountID, nil
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
// ============================================================
// EVENT DERIVATION FROM SCHEDULES
// ============================================================
//
// Event-level timing, venue, and virtual/hybrid flags are derived from
// the schedules. Schedules are the single source of truth. The client
// may still send these fields on create/update — they are ignored.
//
// Called after schedules are populated, and on update when
// cmd.Schedules is non-nil.

// deriveEventFromSchedules computes event-level fields from schedules.
// No-op when there are no schedules.
func deriveEventFromSchedules(event *domain.Event) {
	if event == nil || len(event.Schedules) == 0 {
		return
	}

	// 1. Earliest start and latest end across all schedules.
	earliestStart := firstScheduleStart(event.Schedules)
	latestEnd := lastScheduleEnd(event.Schedules)

	event.StartDate = earliestStart
	event.Date = time.Date(
		earliestStart.Year(), earliestStart.Month(), earliestStart.Day(),
		0, 0, 0, 0, earliestStart.Location(),
	)
	event.Time = scheduleStartClock(event.Schedules[0])

	if !latestEnd.IsZero() {
		event.EndDate = &latestEnd
	}

	// 2. Multi-day detection.
	event.IsMultiDay = computeIsMultiDay(event.Schedules)

	// 3. Duration = sum of (end - start) across schedules, in minutes.
	event.Duration = computeTotalDurationMinutes(event.Schedules)

	// 4. Virtual / hybrid / platform.
	event.IsVirtual, event.IsHybrid = computeVirtualFlags(event.Schedules)
	event.VirtualPlatform = computeVirtualPlatform(event.Schedules)
	event.VirtualPlatformURL = virtualPlatformBaseURL(event.VirtualPlatform)

	// 5. Mirror the first matching schedule's links for display convenience.
	event.ZoomLink = firstNonEmptyLink(event.Schedules, "zoom")
	event.MeetLink = firstNonEmptyLink(event.Schedules, "meet")

	// 6. Venue from the first in-person schedule.
	applyVenueFromSchedules(event, event.Schedules)
}

// ------------------------------------------------------------
// Internals
// ------------------------------------------------------------

// firstScheduleStart returns the earliest start datetime across schedules.
// Combines StartDate + StartTime in the schedule's timezone.
func firstScheduleStart(schedules []domain.EventSchedule) time.Time {
	var earliest time.Time
	for _, s := range schedules {
		t := scheduleStartDateTime(s)
		if t.IsZero() {
			continue
		}
		if earliest.IsZero() || t.Before(earliest) {
			earliest = t
		}
	}
	return earliest
}

// lastScheduleEnd returns the latest end datetime across schedules.
func lastScheduleEnd(schedules []domain.EventSchedule) time.Time {
	var latest time.Time
	for _, s := range schedules {
		t := scheduleEndDateTime(s)
		if t.IsZero() {
			continue
		}
		if latest.IsZero() || t.After(latest) {
			latest = t
		}
	}
	return latest
}

// scheduleStartDateTime combines StartDate + StartTime in the schedule's tz.
func scheduleStartDateTime(s domain.EventSchedule) time.Time {
	if s.StartDate.IsZero() {
		return time.Time{}
	}
	return combineDateAndClock(s.StartDate, s.StartTime, s.Timezone)
}

// scheduleEndDateTime combines EndDate (or StartDate) + EndTime.
func scheduleEndDateTime(s domain.EventSchedule) time.Time {
	date := s.StartDate
	if s.EndDate != nil && !s.EndDate.IsZero() {
		date = *s.EndDate
	}
	if date.IsZero() {
		return time.Time{}
	}
	return combineDateAndClock(date, s.EndTime, s.Timezone)
}

// combineDateAndClock merges a date with an "HH:MM:SS" string in the
// given IANA timezone. Falls back to UTC if the timezone can't be loaded.
func combineDateAndClock(date time.Time, clock, tz string) time.Time {
	loc := time.UTC
	if tz != "" {
		if loaded, err := time.LoadLocation(tz); err == nil {
			loc = loaded
		}
	}

	// Parse the clock string. Accept "15:04:05" and "15:04".
	parsed, err := time.Parse("15:04:05", clock)
	if err != nil {
		parsed, err = time.Parse("15:04", clock)
	}
	if err != nil {
		return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
	}
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		parsed.Hour(), parsed.Minute(), parsed.Second(), 0, loc,
	)
}

// scheduleStartClock extracts the "HH:MM:SS" start of a schedule. Falls
// back to parsing StartTime if needed.
func scheduleStartClock(s domain.EventSchedule) string {
	if s.StartTime != "" {
		return s.StartTime
	}
	return ""
}

// computeIsMultiDay reports whether any schedule spans more than one day,
// or whether the schedules as a whole span more than one calendar day.
func computeIsMultiDay(schedules []domain.EventSchedule) bool {
	var minDay, maxDay string
	for _, s := range schedules {
		if s.StartDate.IsZero() {
			continue
		}
		d := s.StartDate.Format("2006-01-02")
		if minDay == "" || d < minDay {
			minDay = d
		}
		if maxDay == "" || d > maxDay {
			maxDay = d
		}

		// Schedule itself spans days?
		if s.EndDate != nil && !s.EndDate.IsZero() && !s.EndDate.Equal(s.StartDate) {
			return true
		}
	}
	return minDay != "" && maxDay != "" && minDay != maxDay
}

// computeTotalDurationMinutes sums (end - start) across schedules, in
// minutes. Schedules with unparseable times contribute 0.
func computeTotalDurationMinutes(schedules []domain.EventSchedule) int {
	total := 0
	for _, s := range schedules {
		start := scheduleStartDateTime(s)
		end := scheduleEndDateTime(s)
		if start.IsZero() || end.IsZero() || end.Before(start) {
			continue
		}
		total += int(end.Sub(start).Minutes())
	}
	return total
}

// computeVirtualFlags returns (allVirtual, isHybrid).
//
//   - All virtual   → IsVirtual = true, IsHybrid = false
//   - All in-person → IsVirtual = false, IsHybrid = false
//   - Mixed         → IsVirtual = false, IsHybrid = true
func computeVirtualFlags(schedules []domain.EventSchedule) (bool, bool) {
	virtual, inPerson := 0, 0
	for _, s := range schedules {
		if s.IsVirtual {
			virtual++
		} else {
			inPerson++
		}
	}
	switch {
	case virtual > 0 && inPerson == 0:
		return true, false
	case virtual == 0 && inPerson > 0:
		return false, false
	case virtual > 0 && inPerson > 0:
		return false, true
	}
	return false, false
}

// computeVirtualPlatform returns "zoom", "google_meet", or "" based on
// the first link found in schedules.
func computeVirtualPlatform(schedules []domain.EventSchedule) string {
	for _, s := range schedules {
		if s.ZoomLink != "" {
			return "zoom"
		}
		if s.MeetLink != "" {
			return "google_meet"
		}
	}
	return ""
}

// virtualPlatformBaseURL maps a platform slug to its public base URL.
func virtualPlatformBaseURL(platform string) string {
	switch platform {
	case "zoom":
		return "https://zoom.us/"
	case "google_meet":
		return "https://meet.google.com/"
	default:
		return ""
	}
}

// firstNonEmptyLink returns the first schedule's link of the given kind.
func firstNonEmptyLink(schedules []domain.EventSchedule, kind string) string {
	for _, s := range schedules {
		switch kind {
		case "zoom":
			if s.ZoomLink != "" {
				return s.ZoomLink
			}
		case "meet":
			if s.MeetLink != "" {
				return s.MeetLink
			}
		}
	}
	return ""
}

// applyVenueFromSchedules copies the first in-person schedule's location
// into the event-level venue fields.
func applyVenueFromSchedules(event *domain.Event, schedules []domain.EventSchedule) {
	for _, s := range schedules {
		if s.IsVirtual {
			continue
		}
		if s.Location == "" {
			continue
		}
		event.InPersonLocation = s.Location
		event.VenueName = s.Location
		// We only have a free-text location string, so don't attempt
		// to split into street/city/country. That's a future feature.
		return
	}
}