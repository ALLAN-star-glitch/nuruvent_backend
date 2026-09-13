package service

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================
// AI OUTPUT DTO
// ============================================================
//
// GeneratedEventDraft mirrors the JSON the AI is asked to return.
// It is intentionally decoupled from domain.Event so the AI contract
// can evolve without touching the internal aggregate.
//
// NOTE: The AI produces only ONE identity field — `Name` — which is
// the human-readable title. The service's generateEventIdentifiers()
// derives display_name, internal snake_case name, and slug from it.

type GeneratedEventDraft struct {
	Name               string              `json:"name"`   // human-readable title, e.g. "Rust Async Programming Meetup"
	Description        string              `json:"description"`
	ShortDescription   string              `json:"short_description"`
	Tags               []string            `json:"tags"`
	Language           string              `json:"language"`
	Schedules          []GeneratedSchedule `json:"schedules"`
	IsMultiDay         bool                `json:"is_multi_day"`
	IsVirtual          bool                `json:"is_virtual"`
	IsHybrid           bool                `json:"is_hybrid"`
	InPersonLocation   string              `json:"in_person_location"`
	VenueName          string              `json:"venue_name"`
	VenueAddress       string              `json:"venue_address"`
	VenueCity          string              `json:"venue_city"`
	VenueCountry       string              `json:"venue_country"`
	VirtualPlatform    string              `json:"virtual_platform"`
	VirtualPlatformURL string              `json:"virtual_platform_url"`
	Timezone           string              `json:"timezone"`
	IsFree             bool                `json:"is_free"`
	Capacity           int                 `json:"capacity"`
	Tickets            []GeneratedTicket   `json:"tickets"`
	Visibility         string              `json:"visibility"`
	InviteOnly         bool                `json:"invite_only"`
	IsFeatured         bool                `json:"is_featured"`
	CertificateEnabled bool                `json:"certificate_enabled"`
}

type GeneratedSchedule struct {
	StartDate     string  `json:"start_date"`
	EndDate       *string `json:"end_date"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	Timezone      string  `json:"timezone"`
	SessionName   string  `json:"session_name"`
	SessionNumber int     `json:"session_number"`
	Location      string  `json:"location"`
	IsVirtual     bool    `json:"is_virtual"`
	ZoomLink      *string `json:"zoom_link"`
	MeetLink      *string `json:"meet_link"`
}

type GeneratedTicket struct {
	TicketTypeID string  `json:"ticket_type_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Quantity     int     `json:"quantity"`
	MaxPerPerson *int    `json:"max_per_person"`
}

// ============================================================
// CORRECTION CONTEXT
// ============================================================

type correctionContext struct {
	Request       GenerateEventDraftRequest
	EventTypeID   string
	CategoryID    *string
	TicketTypeIDs map[string]struct{}
	Timezone      string
	Language      string
	MinCapacity   int
	MaxCapacity   int
}

// ============================================================
// CORRECTION ENGINE
// ============================================================

// applyCorrections validates a parsed draft and corrects recoverable
// issues in place. Returns human-readable warnings.
//
// Hard-fail cases (return a non-nil error):
//   - empty name (no fallback exists)
//   - empty schedules (retry will be asked)
//   - empty tickets (retry will be asked)
//   - description that cannot be padded to the minimum
//
// Everything else is corrected in place and reported via warnings.
func applyCorrections(d *GeneratedEventDraft, cctx correctionContext) ([]string, error) {
	warnings := make([]string, 0)

	// --- name (the raw title the service will derive everything from) ---
	d.Name = strings.TrimSpace(d.Name)
	if d.Name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	if len(d.Name) < 3 {
		return nil, fmt.Errorf("name is too short (min 3 characters)")
	}
	if len(d.Name) > 120 {
		d.Name = d.Name[:120]
		warnings = append(warnings, "Truncated event name to 120 characters.")
	}

	// --- description -----------------------------------------
	desc := strings.TrimSpace(d.Description)
	if len(desc) < 100 {
		// Pad using short_description + venue context, up to the minimum.
		// The retry path will ask the AI to regenerate properly if this
		// still isn't good enough.
		pad := strings.TrimSpace(d.ShortDescription)
		if pad == "" {
			pad = d.Name
		}
		if d.VenueCity != "" {
			pad += fmt.Sprintf(" Taking place in %s", d.VenueCity)
			if d.VenueCountry != "" {
				pad += fmt.Sprintf(", %s", d.VenueCountry)
			}
			pad += "."
		}
		// Append until we hit the minimum (or give up after a few tries)
		for len(desc) < 100 && pad != "" && !strings.Contains(desc, pad) {
			desc = desc + " " + pad
		}

		if len(desc) >= 100 {
			d.Description = desc
			warnings = append(warnings,
				"Padded description to meet minimum length.")
		} else {
			return nil, fmt.Errorf("description is too short and could not be padded")
		}
	} else {
		d.Description = desc
	}

	// --- short_description -----------------------------------
	if strings.TrimSpace(d.ShortDescription) == "" {
		d.ShortDescription = truncateAt(d.Description, 160)
		warnings = append(warnings, "Generated short description from full description.")
	}

	// --- language / timezone are forced to request values ----
	d.Language = cctx.Language
	d.Timezone = cctx.Timezone

	// --- tags ------------------------------------------------
	if len(d.Tags) == 0 {
		d.Tags = []string{"event", "nuruvent"}
		warnings = append(warnings, "Added default tags.")
	}
	if len(d.Tags) > 10 {
		d.Tags = d.Tags[:10]
	}

	// --- schedules -------------------------------------------
	if len(d.Schedules) == 0 {
		return nil, fmt.Errorf("at least one schedule is required")
	}
	warnings = append(warnings, correctSchedules(d.Schedules, cctx.Timezone)...)

	// Propagate event-level virtuality to schedules.
	if d.IsVirtual || d.IsHybrid {
		for i := range d.Schedules {
			if !d.Schedules[i].IsVirtual {
				d.Schedules[i].IsVirtual = true
				warnings = append(warnings, fmt.Sprintf(
					"Marked schedule %d as virtual to match event format.", i+1))
			}
		}
	}

	// --- capacity --------------------------------------------
	if d.Capacity < cctx.MinCapacity {
		d.Capacity = cctx.MinCapacity
		warnings = append(warnings, fmt.Sprintf(
			"Capacity increased to minimum %d.", cctx.MinCapacity))
	}
	if d.Capacity > cctx.MaxCapacity {
		d.Capacity = cctx.MaxCapacity
		warnings = append(warnings, fmt.Sprintf(
			"Capacity capped at maximum %d.", cctx.MaxCapacity))
	}

	// --- tickets ---------------------------------------------
	ticketWarnings, err := correctTickets(d, cctx)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, ticketWarnings...)

	// --- venue consistency -----------------------------------
	// Soft warnings — checkPublishReadiness will re-check and trigger
	// the retry if they remain unresolved.
	venueIssues := validateVenueConsistency(d)
	warnings = append(warnings, venueIssues...)

	// --- visibility ------------------------------------------
	if d.Visibility != "public" && d.Visibility != "private" && d.Visibility != "unlisted" {
		d.Visibility = "public"
		warnings = append(warnings, "Reset visibility to public.")
	}

	return warnings, nil
}

// ============================================================
// FIELD-LEVEL CORRECTORS
// ============================================================

func correctSchedules(schedules []GeneratedSchedule, fallbackTZ string) []string {
	var warnings []string
	now := time.Now().UTC()

	for i := range schedules {
		s := &schedules[i]

		if s.Timezone == "" {
			s.Timezone = fallbackTZ
		}

		// start_date must be future
		if start, err := time.Parse("2006-01-02", s.StartDate); err == nil {
			if start.Before(now) {
				// Shift forward in yearly steps until the date is in the future.
				// A single +1y shift is not enough when the AI returns a date
				// more than a year behind (e.g. 2025 when now is 2026).
				shifted := start
				for shifted.Before(now) {
					shifted = shifted.AddDate(1, 0, 0)
				}
				s.StartDate = shifted.Format("2006-01-02")
				warnings = append(warnings, fmt.Sprintf(
					"Shifted schedule %d start date to %s (was in the past).",
					i+1, s.StartDate))
			}
		} else {
			s.StartDate = now.AddDate(0, 0, 90).Format("2006-01-02")
			warnings = append(warnings, fmt.Sprintf(
				"Reset schedule %d start date to a default future date.", i+1))
		}

		// start_time
		if !isValidTime(s.StartTime) {
			s.StartTime = "09:00:00"
			warnings = append(warnings, fmt.Sprintf(
				"Reset schedule %d start_time to 09:00:00.", i+1))
		}

		// end_time must be > start_time
		if !isValidTime(s.EndTime) || s.EndTime <= s.StartTime {
			s.EndTime = addHours(s.StartTime, 8)
			warnings = append(warnings, fmt.Sprintf(
				"Adjusted schedule %d end_time.", i+1))
		}

		// end_date >= start_date
		if s.EndDate != nil && *s.EndDate != "" && *s.EndDate < s.StartDate {
			*s.EndDate = s.StartDate
			warnings = append(warnings, fmt.Sprintf(
				"Adjusted schedule %d end_date to match start_date.", i+1))
		}

		// session_number floor — AI frequently sends 0 (off-by-one)
		if s.SessionNumber < 1 {
			s.SessionNumber = i + 1
			warnings = append(warnings, fmt.Sprintf(
				"Reset schedule %d session_number to %d.", i+1, s.SessionNumber))
		}

		// session_name fallback
		if strings.TrimSpace(s.SessionName) == "" {
			s.SessionName = fmt.Sprintf("Session %d", s.SessionNumber)
			warnings = append(warnings, fmt.Sprintf(
				"Set schedule %d session_name to a default.", i+1))
		}
	}

	return warnings
}

func correctTickets(d *GeneratedEventDraft, cctx correctionContext) ([]string, error) {
	var warnings []string

	if len(d.Tickets) == 0 {
		return nil, fmt.Errorf("at least one ticket is required")
	}

	// Pick a fallback ticket type from the allowed set
	var fallbackTicketType string
	for id := range cctx.TicketTypeIDs {
		fallbackTicketType = id
		break
	}

	for i := range d.Tickets {
		t := &d.Tickets[i]

		// ticket_type_id must be in the allowed set
		if _, ok := cctx.TicketTypeIDs[t.TicketTypeID]; !ok {
			t.TicketTypeID = fallbackTicketType
			warnings = append(warnings, fmt.Sprintf(
				"Replaced unknown ticket type on ticket %d.", i+1))
		}

		// price clamp
		if t.Price < 0 {
			t.Price = 0
			warnings = append(warnings, fmt.Sprintf(
				"Clamped negative price on ticket %d to 0.", i+1))
		}

		// quantity floor
		if t.Quantity < 1 {
			t.Quantity = 1
			warnings = append(warnings, fmt.Sprintf(
				"Reset ticket %d quantity to 1.", i+1))
		}
	}

	// capacity >= sum(quantities)
	var totalQty int
	for _, t := range d.Tickets {
		totalQty += t.Quantity
	}
	if d.Capacity < totalQty {
		warnings = append(warnings, fmt.Sprintf(
			"Increased capacity from %d to %d to cover ticket quantities.",
			d.Capacity, totalQty))
		d.Capacity = totalQty
	}

	// is_free must match ticket prices
	hasPaid := false
	for _, t := range d.Tickets {
		if t.Price > 0 {
			hasPaid = true
			break
		}
	}
	if d.IsFree && hasPaid {
		d.IsFree = false
		warnings = append(warnings, "Corrected is_free to match ticket prices.")
	}
	if !d.IsFree && !hasPaid {
		d.IsFree = true
		warnings = append(warnings, "Corrected is_free to match ticket prices.")
	}

	return warnings, nil
}

// validateVenueConsistency checks venue fields against is_virtual / is_hybrid.
// Returns a list of human-readable issues. An empty slice means the draft
// is consistent. These are NOT fatal — the retry path can fix them.
func validateVenueConsistency(d *GeneratedEventDraft) []string {
	var issues []string

	// Physical-only events need location + venue name
	if !d.IsVirtual && !d.IsHybrid {
		if d.InPersonLocation == "" {
			issues = append(issues, "in-person events require in_person_location")
		}
		if d.VenueName == "" {
			issues = append(issues, "in-person events require venue_name")
		}
	}

	// Virtual / hybrid events need a link (event-level or schedule-level)
	if d.IsVirtual || d.IsHybrid {
		hasLink := d.VirtualPlatformURL != ""
		for _, s := range d.Schedules {
			if (s.ZoomLink != nil && *s.ZoomLink != "") ||
				(s.MeetLink != nil && *s.MeetLink != "") {
				hasLink = true
				break
			}
		}
		if !hasLink {
			issues = append(issues,
				"virtual or hybrid events require virtual_platform_url or a zoom_link/meet_link on at least one schedule")
		}
	}

	// Hybrid events: in-person schedules should carry a location
	if d.IsHybrid {
		for i, s := range d.Schedules {
			if !s.IsVirtual && s.Location == "" {
				issues = append(issues, fmt.Sprintf(
					"hybrid event schedule %d is marked in-person but has no location", i+1))
			}
		}
	}

	return issues
}

// ============================================================
// UTILITIES
// ============================================================

// isValidTime accepts both "HH:MM:SS" and "HH:MM" formats so the
// correction layer doesn't rewrite times the AI produced correctly.
func isValidTime(s string) bool {
	if _, err := time.Parse("15:04:05", s); err == nil {
		return true
	}
	if _, err := time.Parse("15:04", s); err == nil {
		return true
	}
	return false
}

func addHours(timeStr string, hours int) string {
	// Try full format first, then short.
	if t, err := time.Parse("15:04:05", timeStr); err == nil {
		return t.Add(time.Duration(hours) * time.Hour).Format("15:04:05")
	}
	if t, err := time.Parse("15:04", timeStr); err == nil {
		return t.Add(time.Duration(hours) * time.Hour).Format("15:04:05")
	}
	return "17:00:00"
}

func truncateAt(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := s[:n]
	if idx := strings.LastIndex(cut, " "); idx > 0 {
		return cut[:idx]
	}
	return cut
}