package service

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// ============================================================
// AI OUTPUT DTO
// ============================================================

type GeneratedEventDraft struct {
	Name               string               `json:"name"`
	Description        string               `json:"description"`
	ShortDescription   string               `json:"short_description"`
	Tags               []string             `json:"tags"`
	Language           string               `json:"language"`
	Schedules          []GeneratedSchedule  `json:"schedules"`
	IsMultiDay         bool                 `json:"is_multi_day"`
	IsRecurring        bool                 `json:"is_recurring"`
	Recurrence         *GeneratedRecurrence `json:"recurrence"`
	IsVirtual          bool                 `json:"is_virtual"`
	IsHybrid           bool                 `json:"is_hybrid"`
	InPersonLocation   string               `json:"in_person_location"`
	VenueName          string               `json:"venue_name"`
	VenueAddress       string               `json:"venue_address"`
	VenueCity          string               `json:"venue_city"`
	VenueCountry       string               `json:"venue_country"`
	VirtualPlatform    string               `json:"virtual_platform"`
	VirtualPlatformURL string               `json:"virtual_platform_url"`
	Timezone           string               `json:"timezone"`
	IsFree             bool                 `json:"is_free"`
	Capacity           int                  `json:"capacity"`
	Tickets            []GeneratedTicket    `json:"tickets"`
	Visibility         string               `json:"visibility"`
	InviteOnly         bool                 `json:"invite_only"`
	IsFeatured         bool                 `json:"is_featured"`
	CertificateEnabled bool                 `json:"certificate_enabled"`
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

type GeneratedRecurrence struct {
	Pattern     string   `json:"pattern"`
	Interval    int      `json:"interval"`
	DaysOfWeek  []string `json:"days_of_week"`
	DayOfMonth  *int     `json:"day_of_month"`
	WeekOfMonth *string  `json:"week_of_month"`
	EndsOn      *string  `json:"ends_on"`
	Occurrences *int     `json:"occurrences"`
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

func applyCorrections(d *GeneratedEventDraft, cctx correctionContext) ([]string, error) {
	warnings := make([]string, 0)

	// --- name ------------------------------------------------
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
	// Do NOT pad short descriptions with repeated text — it produces
	// garbage that reaches the user. Leave the text as-is and let
	// checkPublishReadiness flag it, which triggers the retry path
	// with a specific prompt telling the model how to expand it.
	desc := strings.TrimSpace(d.Description)
	d.Description = desc
	if len(desc) < 100 {
		warnings = append(warnings, fmt.Sprintf(
			"Description is %d chars (minimum 100) — will be retried.", len(desc)))
	}

	// --- short_description -----------------------------------
	if strings.TrimSpace(d.ShortDescription) == "" {
		d.ShortDescription = truncateAt(d.Description, 160)
		warnings = append(warnings, "Generated short description from full description.")
	}

	// --- language / timezone forced to request values --------
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

	// --- collapse over-expanded recurring schedules ---
	//
	// The AI frequently returns one schedule per occurrence (e.g. 6
	// weekly webinars get 6 schedules dated 7 days apart) rather than
	// one schedule that describes the typical session. The recurrence
	// block already conveys the cadence, so multiple schedules are
	// redundant and often inconsistent (dates don't fall on the
	// declared weekdays).
	//
	// We collapse to a single schedule only when all schedules share
	// the same start_time, end_time, and timezone — that's the pattern
	// we can safely reduce. Non-uniform schedules are left alone.
	if d.IsRecurring && len(d.Schedules) > 1 {
		uniform := true
		for i := 1; i < len(d.Schedules); i++ {
			if d.Schedules[i].StartTime != d.Schedules[0].StartTime ||
				d.Schedules[i].EndTime != d.Schedules[0].EndTime ||
				d.Schedules[i].Timezone != d.Schedules[0].Timezone {
				uniform = false
				break
			}
		}
		if uniform {
			originalCount := len(d.Schedules)
			d.Schedules = d.Schedules[:1]
			warnings = append(warnings, fmt.Sprintf(
				"Collapsed %d schedules to 1 — the recurrence block defines the cadence.",
				originalCount))
		}
	}

	// --- align schedule date with declared weekday ---
	//
	// If the event is weekly and the schedule's start_date doesn't
	// fall on one of the declared weekdays, shift it forward to the
	// next matching day. Prevents "runs on Monday" + "schedule is
	// Saturday" contradictions.
	//
	// Because the shift can push the start_date past the existing
	// end_date (e.g. start=Thu end=Thu, shifted to Mon), we clear the
	// end_date afterward. A weekly session doesn't need an end_date —
	// the session duration is start_time → end_time.
	if d.IsRecurring && d.Recurrence != nil &&
		d.Recurrence.Pattern == "weekly" &&
		len(d.Recurrence.DaysOfWeek) > 0 &&
		len(d.Schedules) > 0 {
		targetDays := make(map[time.Weekday]bool)
		for _, name := range d.Recurrence.DaysOfWeek {
			if wd := parseWeekdayName(name); wd >= 0 {
				targetDays[wd] = true
			}
		}
		if start, err := time.Parse("2006-01-02", d.Schedules[0].StartDate); err == nil {
			if !targetDays[start.Weekday()] {
				shifted := start
				for i := 0; i < 7 && !targetDays[shifted.Weekday()]; i++ {
					shifted = shifted.AddDate(0, 0, 1)
				}
				oldDate := d.Schedules[0].StartDate
				d.Schedules[0].StartDate = shifted.Format("2006-01-02")

				// Clear a now-stale end_date. If it was equal to the old
				// start (single-day session) or falls before the new
				// start, drop it. The recurrence block defines the cadence
				// anyway — no need for a per-session end_date.
				if d.Schedules[0].EndDate != nil && *d.Schedules[0].EndDate != "" &&
					*d.Schedules[0].EndDate <= d.Schedules[0].StartDate {
					d.Schedules[0].EndDate = nil
					warnings = append(warnings,
						"Cleared stale end_date after shifting start_date.")
				}

				warnings = append(warnings, fmt.Sprintf(
					"Shifted schedule from %s to %s to match declared weekday %s.",
					oldDate, d.Schedules[0].StartDate, weekdayName(shifted.Weekday())))
			}
		}
	}

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

	// --- recurrence ------------------------------------------
	recurrenceWarnings := correctRecurrence(d, cctx)
	warnings = append(warnings, recurrenceWarnings...)

	// --- venue consistency -----------------------------------
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

		if start, err := time.Parse("2006-01-02", s.StartDate); err == nil {
			if start.Before(now) {
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

		if !isValidTime(s.StartTime) {
			s.StartTime = "09:00:00"
			warnings = append(warnings, fmt.Sprintf(
				"Reset schedule %d start_time to 09:00:00.", i+1))
		}

		if !isValidTime(s.EndTime) || s.EndTime <= s.StartTime {
			s.EndTime = addHours(s.StartTime, 8)
			warnings = append(warnings, fmt.Sprintf(
				"Adjusted schedule %d end_time.", i+1))
		}

		if s.EndDate != nil && *s.EndDate != "" && *s.EndDate < s.StartDate {
			*s.EndDate = s.StartDate
			warnings = append(warnings, fmt.Sprintf(
				"Adjusted schedule %d end_date to match start_date.", i+1))
		}

		if s.SessionNumber < 1 {
			s.SessionNumber = i + 1
			warnings = append(warnings, fmt.Sprintf(
				"Reset schedule %d session_number to %d.", i+1, s.SessionNumber))
		}

		if strings.TrimSpace(s.SessionName) == "" {
			s.SessionName = fmt.Sprintf("Session %d", s.SessionNumber)
			warnings = append(warnings, fmt.Sprintf(
				"Set schedule %d session_name to a default.", i+1))
		}
	}

	return warnings
}

// correctRecurrence normalizes the AI's recurrence block in place.
//
// If the caller supplied a structured recurrence (cctx.Request.Recurrence),
// the AI's output is discarded and replaced with exactly what the caller
// asked for. This is what makes the modal's "This event repeats" toggle
// authoritative — the AI cannot invent a different pattern.
//
// Otherwise, the function normalizes what the AI produced and can
// INFER missing fields from the schedules array (weekdays from schedule
// dates, ends_on from the last schedule date).
func correctRecurrence(d *GeneratedEventDraft, cctx correctionContext) []string {
	var warnings []string

	log.Printf("[AI] correctRecurrence: req.Recurrence=%+v", cctx.Request.Recurrence)
	// ---- Structured override path ----
	if cctx.Request.Recurrence != nil {
		src := cctx.Request.Recurrence

		rec := &GeneratedRecurrence{
			Pattern:  strings.ToLower(strings.TrimSpace(src.Pattern)),
			Interval: src.Interval,
		}
		if rec.Interval < 1 {
			rec.Interval = 1
		}

		if len(src.DaysOfWeek) > 0 {
			rec.DaysOfWeek = dedupeStrings(normalizeWeekdays(src.DaysOfWeek))
		}

		if src.DayOfMonth != nil {
			rec.DayOfMonth = src.DayOfMonth
		}
		if src.WeekOfMonth != nil && *src.WeekOfMonth != "" {
			wom := *src.WeekOfMonth
			rec.WeekOfMonth = &wom
		}
		if src.EndsOn != nil && *src.EndsOn != "" {
			endsOn := *src.EndsOn
			rec.EndsOn = &endsOn
		}
		if src.Occurrences != nil {
			rec.Occurrences = src.Occurrences
		}

		d.IsRecurring = true
		d.Recurrence = rec
		warnings = append(warnings, "Applied caller-supplied recurrence.")
		return warnings
	}

	// ---- Fallback path: normalize whatever the AI produced ----

	if !d.IsRecurring {
		if d.Recurrence != nil {
			d.Recurrence = nil
			warnings = append(warnings, "Dropped recurrence block on non-recurring event.")
		}
		return warnings
	}

	if d.Recurrence == nil {
		return warnings
	}

	r := d.Recurrence

	r.Pattern = strings.ToLower(strings.TrimSpace(r.Pattern))

	if r.Interval < 1 {
		r.Interval = 1
	}

	// Normalize weekday abbreviations to full names.
	normalized := make([]string, 0, len(r.DaysOfWeek))
	for _, day := range r.DaysOfWeek {
		full := normalizeWeekday(day)
		if full == "" {
			warnings = append(warnings, fmt.Sprintf(
				"Dropped unrecognized weekday %q.", day))
			continue
		}
		if full != day {
			warnings = append(warnings, fmt.Sprintf(
				"Normalized weekday %q to %q.", day, full))
		}
		normalized = append(normalized, full)
	}
	r.DaysOfWeek = dedupeStrings(normalized)

	// Clear fields that don't apply to the pattern.
	switch r.Pattern {
	case "daily":
		if len(r.DaysOfWeek) > 0 {
			warnings = append(warnings, "Dropped days_of_week on daily recurrence.")
		}
		r.DaysOfWeek = nil
		r.DayOfMonth = nil
		r.WeekOfMonth = nil
	case "weekly":
		r.DayOfMonth = nil
		r.WeekOfMonth = nil
	case "monthly":
		r.DaysOfWeek = nil
	case "custom":
		// keep everything the AI sent
	default:
		// Unknown pattern — checkPublishReadiness will reject.
	}

	// --- Inference from schedules ---
	if r.Pattern == "weekly" || r.Pattern == "custom" {
		if len(r.DaysOfWeek) == 0 && len(d.Schedules) > 0 {
			seen := make(map[string]struct{})
			derived := make([]string, 0, len(d.Schedules))
			for _, s := range d.Schedules {
				t, err := time.Parse("2006-01-02", s.StartDate)
				if err != nil {
					continue
				}
				day := weekdayName(t.Weekday())
				if day == "" {
					continue
				}
				if _, ok := seen[day]; ok {
					continue
				}
				seen[day] = struct{}{}
				derived = append(derived, day)
			}
			if len(derived) > 0 {
				r.DaysOfWeek = derived
				warnings = append(warnings, fmt.Sprintf(
					"Derived days_of_week from schedule dates: %v", derived))
			}
		}
	}

	// Derive ends_on from the last schedule's date — but ONLY when there
	// is more than one schedule.
	if r.EndsOn == nil && r.Occurrences == nil && len(d.Schedules) > 1 {
		last := d.Schedules[len(d.Schedules)-1]
		if last.StartDate != "" {
			endsOn := last.StartDate
			r.EndsOn = &endsOn
			warnings = append(warnings, fmt.Sprintf(
				"Derived recurrence ends_on from last schedule date: %s", last.StartDate))
		}
	}

	return warnings
}

// weekdayName converts a time.Weekday to the full lowercase name
// used by the recurrence contract.
func weekdayName(w time.Weekday) string {
	switch w {
	case time.Monday:
		return "monday"
	case time.Tuesday:
		return "tuesday"
	case time.Wednesday:
		return "wednesday"
	case time.Thursday:
		return "thursday"
	case time.Friday:
		return "friday"
	case time.Saturday:
		return "saturday"
	case time.Sunday:
		return "sunday"
	}
	return ""
}

// parseWeekdayName is the inverse of weekdayName. Returns -1 for
// unrecognized input so the caller can detect invalid entries.
func parseWeekdayName(s string) time.Weekday {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "monday":
		return time.Monday
	case "tuesday":
		return time.Tuesday
	case "wednesday":
		return time.Wednesday
	case "thursday":
		return time.Thursday
	case "friday":
		return time.Friday
	case "saturday":
		return time.Saturday
	case "sunday":
		return time.Sunday
	}
	return time.Weekday(-1)
}

func normalizeWeekday(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "monday", "mon":
		return "monday"
	case "tuesday", "tue", "tues":
		return "tuesday"
	case "wednesday", "wed":
		return "wednesday"
	case "thursday", "thu", "thur", "thurs":
		return "thursday"
	case "friday", "fri":
		return "friday"
	case "saturday", "sat":
		return "saturday"
	case "sunday", "sun":
		return "sunday"
	}
	return ""
}

// normalizeWeekdays applies normalizeWeekday to a slice and drops
// any entries that couldn't be recognized.
func normalizeWeekdays(in []string) []string {
	out := make([]string, 0, len(in))
	for _, d := range in {
		if full := normalizeWeekday(d); full != "" {
			out = append(out, full)
		}
	}
	return out
}

func isValidFullWeekday(day string) bool {
	switch day {
	case "monday", "tuesday", "wednesday", "thursday",
		"friday", "saturday", "sunday":
		return true
	}
	return false
}

func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func correctTickets(d *GeneratedEventDraft, cctx correctionContext) ([]string, error) {
	var warnings []string

	if len(d.Tickets) == 0 {
		return nil, fmt.Errorf("at least one ticket is required")
	}

	var fallbackTicketType string
	for id := range cctx.TicketTypeIDs {
		fallbackTicketType = id
		break
	}

	for i := range d.Tickets {
		t := &d.Tickets[i]

		if _, ok := cctx.TicketTypeIDs[t.TicketTypeID]; !ok {
			t.TicketTypeID = fallbackTicketType
			warnings = append(warnings, fmt.Sprintf(
				"Replaced unknown ticket type on ticket %d.", i+1))
		}

		if t.Price < 0 {
			t.Price = 0
			warnings = append(warnings, fmt.Sprintf(
				"Clamped negative price on ticket %d to 0.", i+1))
		}

		if t.Quantity < 1 {
			t.Quantity = 1
			warnings = append(warnings, fmt.Sprintf(
				"Reset ticket %d quantity to 1.", i+1))
		}
	}

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

func validateVenueConsistency(d *GeneratedEventDraft) []string {
	var issues []string

	if !d.IsVirtual && !d.IsHybrid {
		if d.InPersonLocation == "" {
			issues = append(issues, "in-person events require in_person_location")
		}
		if d.VenueName == "" {
			issues = append(issues, "in-person events require venue_name")
		}
	}

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