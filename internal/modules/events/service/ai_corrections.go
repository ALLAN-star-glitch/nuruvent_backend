// internal/modules/events/service/ai_corrections.go

package service

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ============================================================
// CORRECTION ENGINE — v11.2
// ============================================================
//
// The prompt is the source of truth. The corrector enforces the
// three-shape contract but does NOT drop a model-inferred recurrence
// just because the toggle was off. It only repairs genuine
// contradictions:
//
//   - is_recurring = true with multiple schedules -> series.
//   - is_recurring = false with a stray recurrence -> drop it.
//
// v11.1 normalizes platform-named virtual locations ("Virtual on
// Zoom") to the platform-agnostic label "Virtual". The platform is
// chosen downstream by the video integration.
//
// v11.2 makes the recurrence toggle a reinforcement, not an
// override. When the caller supplies a recurrence block:
//   - If the AI produced no recurrence, use the caller's block.
//   - If the AI produced one, keep the AI's values and only fill
//     fields it left empty (interval, days_of_week, ends_on, ...).
// This means toggling the UI on for a prompt that already describes
// the cadence produces the same draft as leaving it off.

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
	d.Description = strings.TrimSpace(d.Description)
	if len(d.Description) < 100 {
		warnings = append(warnings, fmt.Sprintf(
			"Description is %d chars (minimum 100) - will be retried.",
			len(d.Description)))
	}

	// --- short_description -----------------------------------
	if strings.TrimSpace(d.ShortDescription) == "" {
		d.ShortDescription = truncateAt(d.Description, 160)
		warnings = append(warnings, "Generated short description from full description.")
	} else if len(d.ShortDescription) > 160 {
		d.ShortDescription = truncateAt(d.ShortDescription, 160)
		warnings = append(warnings, "Truncated short description to 160 characters.")
	}

	d.Language = cctx.Language

	// --- tags ------------------------------------------------
	d.Tags = normalizeTags(d.Tags)
	if len(d.Tags) == 0 {
		d.Tags = []string{"event", "nuruvent"}
		warnings = append(warnings, "Added default tags.")
	}

	// --- SHAPE detection + repair ----------------------------
	shapeWarnings, err := repairShape(d, cctx)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, shapeWarnings...)

	// --- schedules (per-shape) -------------------------------
	schedWarnings, err := correctSchedulesForShape(d, cctx.Timezone)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, schedWarnings...)

	// --- duration from prompt (authoritative) ----------------
	warnings = append(warnings, applyPromptDuration(d, cctx.Request.Prompt)...)

	// --- weekday alignment (SHAPE B only) --------------------
	if d.Shape() == ShapeRecurring && d.Recurrence != nil &&
		d.Recurrence.Pattern == "weekly" &&
		len(d.Recurrence.DaysOfWeek) > 0 &&
		len(d.Schedules) > 0 {
		warnings = append(warnings,
			alignScheduleToWeekday(&d.Schedules[0], d.Recurrence.DaysOfWeek)...)
	}

	// --- capacity + tickets ----------------------------------
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

	ticketWarnings, err := correctTickets(d, cctx)
	if err != nil {
		return nil, err
	}
	warnings = append(warnings, ticketWarnings...)

	if d.Capacity > cctx.MaxCapacity {
		d.Capacity = cctx.MaxCapacity
		warnings = append(warnings, fmt.Sprintf(
			"Re-capped capacity at %d after ticket reconciliation.",
			cctx.MaxCapacity))
	}

	// --- recurrence ------------------------------------------
	warnings = append(warnings, correctRecurrence(d, cctx)...)

	// --- event-type duration bounds --------------------------
	warnings = append(warnings, checkEventTypeDurationBounds(d, cctx)...)

	// --- visibility ------------------------------------------
	if d.Visibility != "public" && d.Visibility != "private" && d.Visibility != "unlisted" {
		d.Visibility = "public"
		warnings = append(warnings, "Reset visibility to public.")
	}

	return warnings, nil
}

// ============================================================
// SHAPE REPAIR
// ============================================================

// repairShape enforces the three-shape contract without overriding
// the model's shape decision based on the toggle. It only repairs
// genuine contradictions.
func repairShape(d *GeneratedEventDraft, cctx correctionContext) ([]string, error) {
	var warnings []string

	if len(d.Schedules) == 0 {
		return nil, fmt.Errorf("at least one schedule is required")
	}

	// Caller supplied a recurrence block -> force SHAPE B.
	if cctx.Request.Recurrence != nil {
		if len(d.Schedules) > 1 {
			warnings = append(warnings, fmt.Sprintf(
				"Caller-supplied recurrence forces SHAPE B; collapsed %d schedules to the first.",
				len(d.Schedules)))
			d.Schedules = d.Schedules[:1]
		}
		d.IsRecurring = true
		return warnings, nil
	}

	// Contradiction: is_recurring = true with multiple schedules.
	// Prefer the schedules (series) and drop the recurrence.
	if d.IsRecurring && len(d.Schedules) > 1 {
		warnings = append(warnings, fmt.Sprintf(
			"Model returned is_recurring=true with %d schedules; treating as a series and dropping recurrence.",
			len(d.Schedules)))
		d.IsRecurring = false
		d.Recurrence = nil
		return warnings, nil
	}

	// Stray recurrence block on a non-recurring event.
	if !d.IsRecurring && d.Recurrence != nil {
		warnings = append(warnings, "Dropped recurrence block on non-recurring event.")
		d.Recurrence = nil
	}

	return warnings, nil
}

// ============================================================
// SCHEDULE CORRECTION
// ============================================================

func correctSchedulesForShape(d *GeneratedEventDraft, fallbackTZ string) ([]string, error) {
	var warnings []string

	if len(d.Schedules) == 0 {
		return nil, fmt.Errorf("at least one schedule is required")
	}

	now := time.Now().UTC()

	if d.Shape() == ShapeSeries {
		sort.SliceStable(d.Schedules, func(i, j int) bool {
			di, _ := time.Parse("2006-01-02", d.Schedules[i].StartDate)
			dj, _ := time.Parse("2006-01-02", d.Schedules[j].StartDate)
			return di.Before(dj)
		})
		for i := range d.Schedules {
			if d.Schedules[i].SessionNumber != i+1 {
				d.Schedules[i].SessionNumber = i + 1
			}
		}
	}

	for i := range d.Schedules {
		s := &d.Schedules[i]

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

		if s.SessionNumber < 1 {
			s.SessionNumber = 1
		}

		if strings.TrimSpace(s.SessionName) == "" {
			if d.Shape() == ShapeSeries {
				s.SessionName = fmt.Sprintf("Session %d", s.SessionNumber)
			} else {
				s.SessionName = d.Name
			}
			warnings = append(warnings, fmt.Sprintf(
				"Set schedule %d session_name to a default.", i+1))
		}

		if strings.TrimSpace(s.Location) == "" {
			if s.IsVirtual {
				s.Location = "Virtual"
			} else {
				s.Location = "To be announced"
			}
			warnings = append(warnings, fmt.Sprintf(
				"Set schedule %d location to a default.", i+1))
		} else if s.IsVirtual && namesPlatform(s.Location) {
			s.Location = "Virtual"
			warnings = append(warnings, fmt.Sprintf(
				"Normalized schedule %d location to \"Virtual\".", i+1))
		}
	}

	return warnings, nil
}

func alignScheduleToWeekday(s *GeneratedSchedule, daysOfWeek []string) []string {
	targetDays := make(map[time.Weekday]bool)
	for _, name := range daysOfWeek {
		if wd := parseWeekdayName(name); wd >= 0 {
			targetDays[wd] = true
		}
	}
	if len(targetDays) == 0 {
		return nil
	}

	start, err := time.Parse("2006-01-02", s.StartDate)
	if err != nil {
		return nil
	}
	if targetDays[start.Weekday()] {
		return nil
	}

	shifted := start
	for i := 0; i < 7 && !targetDays[shifted.Weekday()]; i++ {
		shifted = shifted.AddDate(0, 0, 1)
	}

	old := s.StartDate
	s.StartDate = shifted.Format("2006-01-02")
	return []string{fmt.Sprintf(
		"Shifted schedule from %s to %s to match declared weekday %s.",
		old, s.StartDate, weekdayName(shifted.Weekday()))}
}

// ============================================================
// DURATION CORRECTOR
// ============================================================

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
		"Set schedule 1 end_time to %s (was %s) to match the stated duration %s.",
		newEnd, old, dur)}
}

func parseDurationFromPrompt(prompt string) (time.Duration, bool) {
	p := strings.ToLower(prompt)

	if strings.Contains(p, "half a day") || strings.Contains(p, "half-day") {
		return 4 * time.Hour, true
	}
	if strings.Contains(p, "full day") || strings.Contains(p, "all day") ||
		strings.Contains(p, "full-day") {
		return 8 * time.Hour, true
	}

	combined := regexp.MustCompile(
		`(\d+(?:\.\d+)?)\s*(?:hours?|hrs?|h)\b(?:\s*(?:and\s*)?(\d+)\s*(?:minutes?|mins?|m)\b)?`,
	)
	if m := combined.FindStringSubmatch(p); len(m) >= 2 && m[1] != "" {
		hours, err := strconv.ParseFloat(m[1], 64)
		if err == nil && hours > 0 {
			total := time.Duration(hours * float64(time.Hour))
			if len(m) >= 3 && m[2] != "" {
				if mins, err := strconv.Atoi(m[2]); err == nil && mins > 0 {
					total += time.Duration(mins) * time.Minute
				}
			}
			return total, true
		}
	}

	minRe := regexp.MustCompile(`(\d+)\s*(?:minutes?|mins?|m)\b`)
	if m := minRe.FindStringSubmatch(p); len(m) == 2 {
		if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
			return time.Duration(n) * time.Minute, true
		}
	}

	return 0, false
}

// ============================================================
// RECURRENCE CORRECTION — v11.2
// ============================================================

func correctRecurrence(d *GeneratedEventDraft, cctx correctionContext) []string {
	var warnings []string

	if !d.IsRecurring {
		if d.Recurrence != nil {
			d.Recurrence = nil
			warnings = append(warnings, "Dropped recurrence block on non-recurring event.")
		}
		return warnings
	}

	// ---- Caller-supplied block (toggle ON) ----
	// The toggle reinforces; it does not override. If the AI read
	// the prompt and produced a coherent recurrence, keep it. Use
	// the caller's block only to (a) build a recurrence when the
	// AI produced none, or (b) fill fields the AI left empty.
	if cctx.Request.Recurrence != nil {
		src := cctx.Request.Recurrence

		if d.Recurrence == nil {
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
			d.Recurrence = rec
			warnings = append(warnings,
				"Applied caller-supplied recurrence (AI did not infer one).")
			return warnings
		}

		// AI already produced a recurrence. Fill only empty fields.
		r := d.Recurrence
		filled := []string{}

		if r.Interval < 1 && src.Interval > 0 {
			r.Interval = src.Interval
			filled = append(filled, "interval")
		}
		if len(r.DaysOfWeek) == 0 && len(src.DaysOfWeek) > 0 {
			r.DaysOfWeek = dedupeStrings(normalizeWeekdays(src.DaysOfWeek))
			filled = append(filled, "days_of_week")
		}
		if r.DayOfMonth == nil && src.DayOfMonth != nil {
			r.DayOfMonth = src.DayOfMonth
			filled = append(filled, "day_of_month")
		}
		if r.WeekOfMonth == nil && src.WeekOfMonth != nil && *src.WeekOfMonth != "" {
			wom := *src.WeekOfMonth
			r.WeekOfMonth = &wom
			filled = append(filled, "week_of_month")
		}
		if r.EndsOn == nil && src.EndsOn != nil && *src.EndsOn != "" {
			endsOn := *src.EndsOn
			r.EndsOn = &endsOn
			filled = append(filled, "ends_on")
		}
		if r.Occurrences == nil && src.Occurrences != nil {
			r.Occurrences = src.Occurrences
			filled = append(filled, "occurrences")
		}
		if len(filled) > 0 {
			warnings = append(warnings, fmt.Sprintf(
				"Filled missing recurrence fields from the toggle: %s.",
				strings.Join(filled, ", ")))
		}
		// Fall through to the normalisation below.
	}

	// ---- Model-inferred path ----
	if d.Recurrence == nil {
		return warnings
	}

	r := d.Recurrence
	r.Pattern = strings.ToLower(strings.TrimSpace(r.Pattern))
	if r.Interval < 1 {
		r.Interval = 1
	}

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

	switch r.Pattern {
	case "daily":
		r.DaysOfWeek = nil
		r.DayOfMonth = nil
		r.WeekOfMonth = nil
	case "weekly":
		r.DayOfMonth = nil
		r.WeekOfMonth = nil
	case "monthly":
		r.DaysOfWeek = nil
	case "custom":
	default:
	}

	if (r.Pattern == "weekly" || r.Pattern == "custom") &&
		len(r.DaysOfWeek) == 0 && len(d.Schedules) > 0 {
		if t, err := time.Parse("2006-01-02", d.Schedules[0].StartDate); err == nil {
			if day := weekdayName(t.Weekday()); day != "" {
				r.DaysOfWeek = []string{day}
				warnings = append(warnings, fmt.Sprintf(
					"Derived days_of_week from schedule date: [%s]", day))
			}
		}
	}

	if r.EndsOn == nil && r.Occurrences == nil {
		warnings = append(warnings,
			"Recurrence has no end condition (ends_on and occurrences both null); will be retried.")
	}

	return warnings
}

// ============================================================
// EVENT-TYPE DURATION BOUNDS
// ============================================================

func checkEventTypeDurationBounds(d *GeneratedEventDraft, cctx correctionContext) []string {
	if cctx.EventTypeMinDur <= 0 && cctx.EventTypeMaxDur <= 0 {
		return nil
	}

	var warnings []string
	for i, s := range d.Schedules {
		start, err1 := time.Parse("15:04:05", s.StartTime)
		if err1 != nil {
			start, err1 = time.Parse("15:04", s.StartTime)
			if err1 != nil {
				continue
			}
		}
		end, err2 := time.Parse("15:04:05", s.EndTime)
		if err2 != nil {
			end, err2 = time.Parse("15:04", s.EndTime)
			if err2 != nil {
				continue
			}
		}
		mins := int(end.Sub(start).Minutes())
		if mins <= 0 {
			continue
		}
		if cctx.EventTypeMinDur > 0 && mins < cctx.EventTypeMinDur {
			warnings = append(warnings, fmt.Sprintf(
				"Schedule %d duration %d min is below the event type minimum %d min.",
				i+1, mins, cctx.EventTypeMinDur))
		}
		if cctx.EventTypeMaxDur > 0 && mins > cctx.EventTypeMaxDur {
			warnings = append(warnings, fmt.Sprintf(
				"Schedule %d duration %d min exceeds the event type maximum %d min.",
				i+1, mins, cctx.EventTypeMaxDur))
		}
	}
	return warnings
}

// ============================================================
// TICKET CORRECTION
// ============================================================

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

		if strings.TrimSpace(t.Name) == "" {
			t.Name = "Admission"
			warnings = append(warnings, fmt.Sprintf(
				"Set ticket %d name to a default.", i+1))
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

// ============================================================
// HELPERS
// ============================================================

func normalizeTags(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, t := range in {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	if len(out) > 10 {
		out = out[:10]
	}
	return out
}

// namesPlatform reports whether a virtual location string names a
// specific platform. Such strings are replaced with "Virtual" so
// the downstream video integration can pick the platform.
func namesPlatform(loc string) bool {
	l := strings.ToLower(loc)
	for _, p := range []string{"zoom", "meet", "teams", "webex", "jitsi"} {
		if strings.Contains(l, p) {
			return true
		}
	}
	return false
}

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