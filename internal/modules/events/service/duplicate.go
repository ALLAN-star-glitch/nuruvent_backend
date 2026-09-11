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
// DUPLICATE - Single Event
// ============================================================

// DuplicateEvent creates a copy of an existing event
func (s *eventService) DuplicateEvent(ctx context.Context, id string, cmd DuplicateEventCommand) (*domain.Event, error) {
	log.Printf("📋 Duplicating event: %s", id)

	if cmd.CreatedBy == "" {
		return nil, errors.New("user ID (CreatedBy) is required")
	}

	// 1. Get original event
	original, err := s.getOriginalEvent(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Check permissions using 2-tier resolution pattern
	if err := s.checkDuplicatePermission(ctx, original, cmd); err != nil {
		return nil, err
	}

	// 3. Prepare duplication data
	newName, newDate := s.prepareDuplicateData(original, cmd)

	// 4. Build create command from original
	createCmd := s.buildDuplicateDraftCommand(original, newName, newDate, cmd.CreatedBy)

	// 5. Create draft using existing CreateDraft method
	newEvent, err := s.CreateDraft(ctx, createCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to duplicate event: %w", err)
	}

	log.Printf("✅ Event duplicated successfully: %s -> %s", id, newEvent.ID)
	return newEvent, nil
}

// ============================================================
// DUPLICATE - Bulk Events
// ============================================================

// BulkDuplicateEvents creates copies of multiple existing events
func (s *eventService) BulkDuplicateEvents(ctx context.Context, ids []string, cmd BulkDuplicateCommand) (*BulkDuplicateResult, error) {
	if len(ids) == 0 {
		return nil, errors.New("at least one event ID is required")
	}
	if cmd.CreatedBy == "" {
		return nil, errors.New("user ID (CreatedBy) is required")
	}

	result := &BulkDuplicateResult{
		DuplicatedCount: 0,
		CreatedEvents:   []DuplicatedEvent{},
		FailedIDs:       []string{},
		Errors:          []string{},
	}

	for _, id := range ids {
		dupCmd := s.buildDuplicateCommandForBulk(ctx, id, cmd)
		if dupCmd == nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: failed to build duplicate command", id))
			continue
		}

		event, err := s.DuplicateEvent(ctx, id, *dupCmd)
		if err != nil {
			result.FailedIDs = append(result.FailedIDs, id)
			result.Errors = append(result.Errors, fmt.Sprintf("event %s: %v", id, err))
			continue
		}

		result.CreatedEvents = append(result.CreatedEvents, DuplicatedEvent{
			ID:   event.ID,
			Name: event.Name,
			Slug: event.Slug,
		})
		result.DuplicatedCount++
	}

	return result, nil
}

// ============================================================
// PRIVATE HELPER FUNCTIONS
// ============================================================

// getOriginalEvent retrieves the original event by ID
func (s *eventService) getOriginalEvent(ctx context.Context, id string) (*domain.Event, error) {
	original, err := s.repo.GetEventByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if original == nil {
		return nil, domain.ErrEventNotFound
	}
	return original, nil
}

// resolveDuplicateTeamDomain dynamically determines personal or institution team domain
func (s *eventService) resolveDuplicateTeamDomain(original *domain.Event, cmd DuplicateEventCommand) string {
	targetTeamID := original.TeamID
	if cmd.TeamID != "" {
		targetTeamID = cmd.TeamID
	}

	if cmd.TeamType == "personal" {
		return domain.PersonalTeamDomain(targetTeamID)
	}

	if cmd.TeamType == "institution" {
		return domain.InstitutionTeamDomain(targetTeamID)
	}

	// Heuristic fallback: if team ID matches the creator's user ID
	if cmd.CreatedBy != "" && cmd.CreatedBy == targetTeamID {
		return domain.PersonalTeamDomain(cmd.CreatedBy)
	}

	return domain.InstitutionTeamDomain(targetTeamID)
}

// checkDuplicatePermission performs 2-tier domain permission check for duplication
func (s *eventService) checkDuplicatePermission(ctx context.Context, original *domain.Event, cmd DuplicateEventCommand) error {
	userID := cmd.CreatedBy
	if userID == "" {
		return errors.New("user ID is required")
	}

	// Resolve Team domain
	teamDomain := s.resolveDuplicateTeamDomain(original, cmd)

	log.Printf("🔍 Tier 1: Checking manage permission for user=%s on team domain=%s", userID, teamDomain)

	// Tier 1: Check Team Domain
	allowed, err := s.permChecker.CanManageEvent(ctx, userID, teamDomain)
	if err != nil {
		return fmt.Errorf("permission check failed: %w", err)
	}

	// Tier 2: Account Domain Fallback
	accountID := cmd.AccountID
	if accountID == "" {
		accountID = original.AccountID
	}

	if !allowed && accountID != "" {
		accountDomain := domain.AccountDomain(accountID)
		log.Printf("🔍 Tier 2: Team check failed. Falling back to Account domain=%s", accountDomain)

		allowed, err = s.permChecker.CanManageEvent(ctx, userID, accountDomain)
		if err != nil {
			return fmt.Errorf("account permission check failed: %w", err)
		}
	}

	if !allowed {
		log.Printf("❌ Permission denied: user %s cannot duplicate event %s", userID, original.ID)
		return domain.ErrForbidden
	}

	return nil
}

// prepareDuplicateData prepares the new name and date for the duplicated event
func (s *eventService) prepareDuplicateData(original *domain.Event, cmd DuplicateEventCommand) (string, string) {
	newName := cmd.Name
	if newName == "" {
		newName = fmt.Sprintf("Copy of %s", original.Name)
	}

	var newDate string
	if cmd.Date != "" {
		newDate = cmd.Date
	} else if len(original.Schedules) > 0 {
		newDate = original.Schedules[0].StartDate.Format("2006-01-02")
	} else {
		newDate = time.Now().Format("2006-01-02")
	}

	return newName, newDate
}

// buildDuplicateDraftCommand builds a CreateDraftCommand from the original event
func (s *eventService) buildDuplicateDraftCommand(
	original *domain.Event,
	newName, newDate, createdBy string,
) CreateDraftCommand {
	return CreateDraftCommand{
		Name:             newName,
		Description:      original.Description,
		ShortDescription: original.ShortDescription,
		EventTypeID:      original.EventTypeID,
		CategoryID:       original.CategoryID,
		Tags:             original.Tags,
		Language:         original.Language,
		CreatedBy:        createdBy,
		TeamID:           original.TeamID,
		AccountID:        original.AccountID,

		// Schedule - use the new date
		IsMultiDay:  original.IsMultiDay,
		IsRecurring: original.IsRecurring,
		Schedules:   s.convertSchedulesToInputWithDate(original.Schedules, newDate),
		Recurrence:  s.convertRecurrenceToInput(original),

		// Venue
		IsVirtual:          original.IsVirtual,
		IsHybrid:           original.IsHybrid,
		InPersonLocation:   original.InPersonLocation,
		VirtualPlatform:    original.VirtualPlatform,
		VirtualPlatformURL: original.VirtualPlatformURL,
		ZoomLink:           original.ZoomLink,
		MeetLink:           original.MeetLink,
		VenueName:          original.VenueName,
		VenueAddress:       original.VenueAddress,
		VenueCity:          original.VenueCity,
		VenueCountry:       original.VenueCountry,

		// Tickets
		IsFree:   original.IsFreeEvent,
		Capacity: original.Capacity,
		Tickets:  s.convertTicketsToInput(original.Tickets),

		// Access & Privacy
		Visibility:    original.Visibility,
		Password:      original.Password,
		InviteOnly:    original.InviteOnly,
		InvitedEmails: original.InvitedEmails,

		// Monetization
		IsFeatured:            original.IsFeatured,
		CertificateEnabled:    original.CertificateEnabled,
		CertificatePrice:      original.CertificatePrice,
		CertificateTemplateID: original.CertificateTemplateID,

		// Speakers
		Speakers: s.convertSpeakersToInput(original.Speakers),

		// Materials
		Materials: s.convertMaterialsToInput(original.Materials),

		// SEO
		SEO: s.convertSEOToInput(original),
	}
}

// buildDuplicateCommandForBulk builds a DuplicateEventCommand for bulk operations
func (s *eventService) buildDuplicateCommandForBulk(ctx context.Context, id string, cmd BulkDuplicateCommand) *DuplicateEventCommand {
	dupCmd := &DuplicateEventCommand{
		Name:      cmd.NamePrefix,
		IsDraft:   cmd.IsDraft,
		CreatedBy: cmd.CreatedBy,
		AccountID: cmd.AccountID,
		Date:      "",
	}

	if cmd.DateOffsetDays != 0 {
		event, err := s.repo.GetEventByID(ctx, id)
		if err != nil {
			return nil
		}
		if event != nil && len(event.Schedules) > 0 {
			newDate := event.Schedules[0].StartDate.AddDate(0, 0, cmd.DateOffsetDays)
			dupCmd.Date = newDate.Format("2006-01-02")
		}
	}

	return dupCmd
}

// ============================================================
// CONVERTER HELPERS FOR DUPLICATE
// ============================================================

// convertSchedulesToInputWithDate converts schedules with a new date
func (s *eventService) convertSchedulesToInputWithDate(schedules []domain.EventSchedule, newDate string) []ScheduleInput {
	if len(schedules) == 0 {
		return nil
	}

	inputs := make([]ScheduleInput, len(schedules))
	for i, sch := range schedules {
		var endDate *string
		if sch.EndDate != nil {
			d := sch.EndDate.Format("2006-01-02")
			endDate = &d
		}

		inputs[i] = ScheduleInput{
			StartDate:     newDate,
			EndDate:       endDate,
			StartTime:     sch.StartTime,
			EndTime:       sch.EndTime,
			Timezone:      sch.Timezone,
			SessionName:   sch.SessionName,
			SessionNumber: sch.SessionNumber,
			Location:      sch.Location,
			IsVirtual:     sch.IsVirtual,
			ZoomLink:      sch.ZoomLink,
			MeetLink:      sch.MeetLink,
			MaxAttendees:  sch.MaxAttendees,
		}
	}
	return inputs
}

// convertSchedulesToInput converts schedules without changing dates
func (s *eventService) convertSchedulesToInput(schedules []domain.EventSchedule) []ScheduleInput {
	if len(schedules) == 0 {
		return nil
	}

	inputs := make([]ScheduleInput, len(schedules))
	for i, sch := range schedules {
		var endDate *string
		if sch.EndDate != nil {
			d := sch.EndDate.Format("2006-01-02")
			endDate = &d
		}

		inputs[i] = ScheduleInput{
			StartDate:     sch.StartDate.Format("2006-01-02"),
			EndDate:       endDate,
			StartTime:     sch.StartTime,
			EndTime:       sch.EndTime,
			Timezone:      sch.Timezone,
			SessionName:   sch.SessionName,
			SessionNumber: sch.SessionNumber,
			Location:      sch.Location,
			IsVirtual:     sch.IsVirtual,
			ZoomLink:      sch.ZoomLink,
			MeetLink:      sch.MeetLink,
			MaxAttendees:  sch.MaxAttendees,
		}
	}
	return inputs
}

// convertRecurrenceToInput converts recurrence from domain to input
func (s *eventService) convertRecurrenceToInput(event *domain.Event) *RecurrenceInput {
	if !event.IsRecurring {
		return nil
	}

	var endsOn *string
	if event.RecurrenceEndsOn != nil {
		d := event.RecurrenceEndsOn.Format("2006-01-02")
		endsOn = &d
	}

	var pattern string
	if event.RecurrencePatternID != nil {
		pattern = *event.RecurrencePatternID
	}

	return &RecurrenceInput{
		Pattern:     pattern,
		Interval:    event.RecurrenceInterval,
		DaysOfWeek:  event.RecurrenceDaysOfWeek,
		DayOfMonth:  event.RecurrenceDayOfMonth,
		WeekOfMonth: event.RecurrenceWeekOfMonth,
		EndsOn:      endsOn,
		Occurrences: event.RecurrenceOccurrences,
	}
}

// convertTicketsToInput converts tickets from domain to input
func (s *eventService) convertTicketsToInput(tickets []domain.EventTicket) []TicketInput {
	if len(tickets) == 0 {
		return nil
	}

	inputs := make([]TicketInput, len(tickets))
	for i, t := range tickets {
		var earlyBirdDeadline *string
		if t.EarlyBirdDeadline != nil {
			d := t.EarlyBirdDeadline.Format(time.RFC3339)
			earlyBirdDeadline = &d
		}

		inputs[i] = TicketInput{
			TicketTypeID:      t.TicketTypeID,
			Name:              t.Name,
			Description:       t.Description,
			Price:             t.Price,
			Quantity:          t.Quantity,
			MaxPerPerson:      t.MaxPerPerson,
			EarlyBirdDeadline: earlyBirdDeadline,
			GroupMinAttendees: t.GroupMinAttendees,
			GroupDiscount:     t.GroupDiscount,
		}
	}
	return inputs
}

// convertSpeakersToInput converts speakers from domain to input
func (s *eventService) convertSpeakersToInput(speakers []domain.EventSpeaker) []SpeakerInput {
	if len(speakers) == 0 {
		return nil
	}

	inputs := make([]SpeakerInput, len(speakers))
	for i, sp := range speakers {
		inputs[i] = SpeakerInput{
			Name:        sp.Name,
			Title:       sp.Title,
			Bio:         sp.Bio,
			PhotoURL:    sp.PhotoURL,
			SocialLinks: sp.SocialLinks,
			IsKeynote:   sp.IsKeynote,
			SortOrder:   sp.SortOrder,
		}
	}
	return inputs
}

// convertMaterialsToInput converts materials from domain to input
func (s *eventService) convertMaterialsToInput(materials []domain.EventMaterial) []MaterialInput {
	if len(materials) == 0 {
		return nil
	}

	inputs := make([]MaterialInput, len(materials))
	for i, m := range materials {
		inputs[i] = MaterialInput{
			Title:          m.Title,
			MaterialTypeID: m.MaterialTypeID,
			URL:            m.URL,
			Description:    m.Description,
			IsPreEvent:     m.IsPreEvent,
			SortOrder:      m.SortOrder,
		}
	}
	return inputs
}

// convertSEOToInput converts SEO from domain to input
func (s *eventService) convertSEOToInput(event *domain.Event) *SEOInput {
	return &SEOInput{
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