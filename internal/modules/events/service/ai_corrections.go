// internal/modules/events/service/ai_corrections.go

package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

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
	desc := strings.TrimSpace(d.Description)
	d.Description = desc
	if len(desc) < 100 {
		warnings = append(warnings, fmt.Sprintf(
			"Description is %d chars (minimum 100) - will be retried.", len(desc)))
	}

	// --- short_description -----------------------------------
	if strings.TrimSpace(d.ShortDescription) == "" {
		d.ShortDescription = truncateAt(d.Description, 160)
		warnings = append(warnings, "Generated short description from full description.")
	}

	// --- language forced to request value --------------------
	d.Language = cctx.Language

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
	// Exactly one schedule. If the AI returned more, keep the first.
	// Recurrence defines the rest.
	if len(d.Schedules) > 1 {
		originalCount := len(d.Schedules)
		d.Schedules = d.Schedules[:1]
		warnings = append(warnings, fmt.Sprintf(
			"Collapsed %d schedules to 1 - the recurrence block defines the cadence.",
			originalCount))
	}
	warnings = append(warnings, correctSchedules(d.Schedules, cctx.Timezone)...)

	// --- duration from prompt (authoritative) ----------------
	// Runs AFTER correctSchedules so the base times are valid, and
	// BEFORE the weekday shift so a duration change doesn't interact
	// with the date.
	warnings = append(warnings, applyPromptDuration(d, cctx.Request.Prompt)...)

	// --- align schedule date with declared weekday ---
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
				warnings = append(warnings, fmt.Sprintf(
					"Shifted schedule from %s to %s to match declared weekday %s.",
					oldDate, d.Schedules[0].StartDate, weekdayName(shifted.Weekday())))
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

	// --- schedule-level consistency --------------------------
	scheduleIssues := validateScheduleConsistency(d)
	warnings = append(warnings, scheduleIssues...)

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

// correctSchedules normalizes the single canonical schedule.
func correctSchedules(schedules []GeneratedSchedule, fallbackTZ string) []string {
	var warnings []string
	now := time.Now().UTC()

	for i := range schedules {
		s := &schedules[i]

		if s.Timezone == "" {
			s.Timezone = fallbackTZ
		}

		// Start date must be in the future.
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

		if s.SessionNumber < 1 {
			s.SessionNumber = 1
			warnings = append(warnings, fmt.Sprintf(
				"Reset schedule %d session_number to 1.", i+1))
		}

		if strings.TrimSpace(s.SessionName) == "" {
			s.SessionName = fmt.Sprintf("Session %d", s.SessionNumber)
			warnings = append(warnings, fmt.Sprintf(
				"Set schedule %d session_name to a default.", i+1))
		}
	}

	return warnings
}

// ============================================================
// DURATION CORRECTOR
// ============================================================

// applyPromptDuration scans the prompt for an explicit session
// duration and, if found, overrides the schedule's end_time.
//
// The AI is unreliable at arithmetic: it will produce end_time
// values that don't match a stated duration. This corrector makes
// the duration authoritative: whatever the prompt says wins.
//
// Recognizes:
//   "90 minutes", "90 min", "2 hours", "2 hrs", "2h", "1.5 hours"
//
// Called from applyCorrections after correctSchedules.
func applyPromptDuration(d *GeneratedEventDraft, prompt string) []string {
	if len(d.Schedules) == 0 {
		return nil
	}

	dur, ok := parseDurationFromPrompt(prompt)
	if !ok || dur <= 0 {
		return nil
	}

	s := &d.Schedules[0]

	start, err := time.Parse("15:04:05", s.StartTime)
	if err != nil {
		start, err = time.Parse("15:04", s.StartTime)
		if err != nil {
			return nil
		}
	}

	end := start.Add(dur)
	newEnd := end.Format("15:04:05")

	if s.EndTime == newEnd {
		return nil
	}

	old := s.EndTime
	s.EndTime = newEnd

	return []string{fmt.Sprintf(
		"Set end_time to %s (was %s) to match the stated duration %s.",
		newEnd, old, dur)}
}

// parseDurationFromPrompt extracts a session duration from free text.
// Returns (0, false) when no duration phrase is present.
func parseDurationFromPrompt(prompt string) (time.Duration, bool) {
	p := strings.ToLower(prompt)

	// hour-based: "1.5 hours", "2 hours", "2 hrs", "2h"
	hourRe := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:hours?|hrs?|h)\b`)
	if m := hourRe.FindStringSubmatch(p); len(m) == 2 {
		if f, err := strconv.ParseFloat(m[1], 64); err == nil && f > 0 {
			return time.Duration(f * float64(time.Hour)), true
		}
	}

	// minute-based: "90 minutes", "90 min", "90m"
	minRe := regexp.MustCompile(`(\d+)\s*(?:minutes?|mins?|m)\b`)
	if m := minRe.FindStringSubmatch(p); len(m) == 2 {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			return time.Duration(n) * time.Minute, true
		}
	}

	return 0, false
}

// correctRecurrence normalizes the AI's recurrence block in place.
//
// If the caller supplied a structured recurrence (cctx.Request.Recurrence),
// the AI's output is discarded and replaced with exactly what the caller
// asked for.
//
// Otherwise, the function normalizes what the AI produced and can
// INFER missing fields from the schedule's date.
func correctRecurrence(d *GeneratedEventDraft, cctx correctionContext) []string {
	var warnings []string

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

	// Normalize custom + interval > 1 + weekdays into weekly.
	if r.Pattern == "custom" && r.Interval > 1 && len(r.DaysOfWeek) > 0 {
		r.Pattern = "weekly"
		r.Interval = 1
		warnings = append(warnings,
			"Normalized custom+interval to weekly recurrence.")
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

	// --- Inference from the schedule's date ---
	if r.Pattern == "weekly" || r.Pattern == "custom" {
		if len(r.DaysOfWeek) == 0 && len(d.Schedules) > 0 {
			if t, err := time.Parse("2006-01-02", d.Schedules[0].StartDate); err == nil {
				if day := weekdayName(t.Weekday()); day != "" {
					r.DaysOfWeek = []string{day}
					warnings = append(warnings, fmt.Sprintf(
						"Derived days_of_week from schedule date: [%s]", day))
				}
			}
		}
	}

	// --- Ensure the recurrence has an end condition ---
	if r.EndsOn == nil && r.Occurrences == nil {
		defaults := map[string]int{
			"daily":   10,
			"weekly":  4,
			"monthly": 3,
			"custom":  4,
		}
		n := defaults[r.Pattern]
		if n == 0 {
			n = 4
		}
		r.Occurrences = &n
		warnings = append(warnings, fmt.Sprintf(
			"Applied default occurrences=%d for %s recurrence.",
			n, r.Pattern))
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

// validateScheduleConsistency checks the single schedule for the
// invariants that matter at the schedule level. Event-level venue and
// virtual/hybrid checks are gone — those are derived downstream.
func validateScheduleConsistency(d *GeneratedEventDraft) []string {
	var issues []string

	if len(d.Schedules) == 0 {
		return []string{"at least one schedule is required"}
	}

	s := d.Schedules[0]

	if !s.IsVirtual && strings.TrimSpace(s.Location) == "" {
		issues = append(issues,
			"in-person schedules require a location")
	}

	if s.IsVirtual && strings.TrimSpace(s.Location) == "" {
		issues = append(issues,
			"virtual schedules require a location label (e.g. \"Virtual on Zoom\")")
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

