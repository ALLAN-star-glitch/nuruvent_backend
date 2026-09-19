package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

// ============================================================
// PROMPT VERSION
// ============================================================

const PromptVersion = "v5"

// ============================================================
// SYSTEM PROMPT — v5
// ============================================================

const systemPromptV1 = `You are an event generator for the Nuruvent events platform.
Return ONLY valid JSON. No markdown, no explanations, no commentary.

Schema:
{
  "name":              string,
  "description":       string,
  "short_description": string,
  "tags":              [string],
  "language":          string,
  "schedules":         [Schedule],
  "is_multi_day":      boolean,
  "is_recurring":      boolean,
  "recurrence":        Recurrence | null,
  "is_virtual":        boolean,
  "is_hybrid":         boolean,
  "in_person_location": string,
  "venue_name":        string,
  "venue_address":     string,
  "venue_city":        string,
  "venue_country":     string,
  "virtual_platform":  string,
  "virtual_platform_url": string,
  "timezone":          string,
  "is_free":           boolean,
  "capacity":          number,
  "tickets":           [Ticket],
  "visibility":        "public" | "private" | "unlisted",
  "invite_only":       boolean,
  "is_featured":       boolean,
  "certificate_enabled": boolean
}

Schedule:
{
  "start_date":    "YYYY-MM-DD",
  "end_date":      "YYYY-MM-DD" | null,
  "start_time":    "HH:MM:SS",
  "end_time":      "HH:MM:SS",
  "timezone":      string,
  "session_name":  string,
  "session_number": number,
  "location":      string,
  "is_virtual":    boolean,
  "zoom_link":     string | null,
  "meet_link":     string | null
}

Ticket:
{
  "ticket_type_id": "UUID",
  "name":           string,
  "description":    string,
  "price":          number,
  "quantity":       number,
  "max_per_person": number | null
}

Recurrence:
{
  "pattern":      "daily" | "weekly" | "monthly" | "custom",
  "interval":     number,
  "days_of_week": [string],
  "day_of_month": number | null,
  "week_of_month": "first" | "second" | "third" | "fourth" | "last" | null,
  "ends_on":      "YYYY-MM-DD" | null,
  "occurrences":  number | null
}

Rules:
- All dates MUST be in the future.
- All times MUST be valid HH:MM:SS.
- Each schedule's end_date >= start_date.
- Each schedule's end_time > start_time (same-day schedules).
- capacity MUST be >= sum(ticket.quantity).
- If is_free is true, all ticket prices MUST be 0.
- If is_free is false, at least one ticket price MUST be > 0.
- If is_virtual is false AND is_hybrid is false, in_person_location and venue_name MUST be set.
- If is_virtual is true OR is_hybrid is true, virtual_platform_url OR a zoom_link/meet_link MUST be set.
- ticket_type_id MUST be one of the UUIDs in the "Available ticket types" list.
- If the prompt implies repetition (e.g. "every Monday", "weekly series", "runs for 6 weeks"), set is_recurring = true and populate recurrence.
- If is_recurring is true and pattern is "weekly" or "custom", days_of_week MUST contain at least one weekday.
- Weekday values MUST be full lowercase names: "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday". Never "mon", "tue", "wed", etc.
- If is_recurring is true and pattern is "monthly", set exactly one of day_of_month or week_of_month.
- If is_recurring is true, at least one of ends_on or occurrences MUST be set.
- If the event is a one-off, set is_recurring = false and recurrence = null.
- description MUST be at least 100 characters (roughly 2-3 sentences). Describe the agenda, who should attend, and what attendees will take away.
- short_description MUST be one sentence, max 160 characters.
- When is_recurring is true, use exactly ONE schedule that describes the typical session (time of day, timezone, location, session duration). Do NOT expand the series into one schedule per occurrence. The recurrence block conveys the cadence; the schedule conveys the session shape.
- The schedule's start_date MUST fall on one of the weekdays listed in recurrence.days_of_week when pattern is "weekly".`

// ============================================================
// REQUEST DTO
// ============================================================

type GenerateEventDraftRequest struct {
	Prompt        string   `json:"prompt"`
	EventTypeID   string   `json:"event_type_id"`
	CategoryID    *string  `json:"category_id,omitempty"`
	TicketTypeIDs []string `json:"ticket_type_ids"`
	Language      string   `json:"language,omitempty"`
	Timezone      string   `json:"timezone,omitempty"`
	Currency      string   `json:"currency,omitempty"`
	MinCapacity   int      `json:"min_capacity,omitempty"`
	MaxCapacity   int      `json:"max_capacity,omitempty"`

	// Recurrence — when provided, the AI MUST use exactly this pattern
	// and not invent its own. Nil means "the event doesn't repeat" (or
	// the caller didn't specify; the AI may still infer from prompt).
	Recurrence *RecurrenceInput `json:"recurrence,omitempty"`

	CreatedBy string `json:"-"`
	TeamID    string `json:"-"`
	TeamType  string `json:"-"`
	AccountID string `json:"-"`
}

// ============================================================
// PROMPT CONTEXT
// ============================================================

type promptContext struct {
	EventType   *domain.EventType
	Category    *domain.Category
	TicketTypes []*domain.TicketTypeRow
	Language    string
	Timezone    string
	Currency    string
	MinCapacity int
	MaxCapacity int
}

// ============================================================
// BUILDERS
// ============================================================

func buildSystemPrompt() string {
	return systemPromptV1
}

func buildUserPrompt(req GenerateEventDraftRequest, pctx *promptContext) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(
		"Today is %s. All event dates MUST be in the future.\n\n",
		time.Now().UTC().Format("2006-01-02"),
	))

	b.WriteString("Generate an event based on this prompt:\n")
	b.WriteString(fmt.Sprintf("%q\n\n", req.Prompt))

	// If the caller supplied a structured recurrence, pin it as a
	// hard constraint. The AI must echo these exact values — the
	// corrector will force them onto the draft regardless, but this
	// instruction keeps the model aligned so the retry path doesn't
	// have to fire on a mismatch.
	if req.Recurrence != nil {
		b.WriteString("HARD CONSTRAINT — the recurrence below is FIXED.\n")
		b.WriteString("Set is_recurring = true and echo these exact values in the `recurrence` object.\n")
		b.WriteString("Do NOT change the pattern, interval, days_of_week, day_of_month, week_of_month, ends_on, or occurrences.\n")
		b.WriteString("    pattern:       " + req.Recurrence.Pattern + "\n")
		if req.Recurrence.Interval > 0 {
			b.WriteString(fmt.Sprintf("    interval:      %d\n", req.Recurrence.Interval))
		}
		if len(req.Recurrence.DaysOfWeek) > 0 {
			b.WriteString(fmt.Sprintf(
				"    days_of_week:  [%s]\n",
				strings.Join(req.Recurrence.DaysOfWeek, ", "),
			))
		}
		if req.Recurrence.DayOfMonth != nil {
			b.WriteString(fmt.Sprintf(
				"    day_of_month:  %d\n", *req.Recurrence.DayOfMonth))
		}
		if req.Recurrence.WeekOfMonth != nil && *req.Recurrence.WeekOfMonth != "" {
			b.WriteString(fmt.Sprintf(
				"    week_of_month: %s\n", *req.Recurrence.WeekOfMonth))
		}
		if req.Recurrence.EndsOn != nil && *req.Recurrence.EndsOn != "" {
			b.WriteString(fmt.Sprintf(
				"    ends_on:       %s\n", *req.Recurrence.EndsOn))
		}
		if req.Recurrence.Occurrences != nil {
			b.WriteString(fmt.Sprintf(
				"    occurrences:   %d\n", *req.Recurrence.Occurrences))
		}
		b.WriteString("\n")
	}

	b.WriteString("Context:\n")

	if pctx.EventType != nil {
		b.WriteString(fmt.Sprintf(
			"- Event type: %s (id: %s, min_duration: %d, max_duration: %d)\n",
			pctx.EventType.DisplayName,
			pctx.EventType.ID,
			pctx.EventType.MinDuration,
			pctx.EventType.MaxDuration,
		))
	}

	if pctx.Category != nil {
		b.WriteString(fmt.Sprintf(
			"- Category: %s (id: %s)\n",
			pctx.Category.DisplayName,
			pctx.Category.ID,
		))
	}

	b.WriteString("- Available ticket types:\n")
	for _, tt := range pctx.TicketTypes {
		b.WriteString(fmt.Sprintf(
			"    - %s: %s (id: %s)\n",
			tt.Slug, tt.DisplayName, tt.ID,
		))
	}

	b.WriteString(fmt.Sprintf("- Language: %s\n", pctx.Language))
	b.WriteString(fmt.Sprintf("- Timezone: %s\n", pctx.Timezone))
	b.WriteString(fmt.Sprintf("- Currency: %s\n", pctx.Currency))
	b.WriteString(fmt.Sprintf("- Capacity range: %d-%d\n", pctx.MinCapacity, pctx.MaxCapacity))

	b.WriteString("\nReturn the JSON now.\n")
	return b.String()
}

func buildFixPrompt(originalPrompt string, errors []string) string {
	var b strings.Builder

	b.WriteString("The previous draft failed validation with these errors:\n")
	for _, e := range errors {
		b.WriteString(fmt.Sprintf("- %s\n", e))
	}

	// If any error mentions the description, spell out the fix.
	for _, e := range errors {
		if strings.Contains(strings.ToLower(e), "description") {
			b.WriteString("\nThe description is too short. Expand it to AT LEAST 100 characters")
			b.WriteString(" (aim for 150–300). Cover these points in order:\n")
			b.WriteString("  1. What the event is and who it's for.\n")
			b.WriteString("  2. The agenda — what will be covered or taught.\n")
			b.WriteString("  3. What attendees will walk away with.\n")
			b.WriteString("Write it as flowing prose, not a bulleted list.\n")
			break
		}
	}

	b.WriteString(fmt.Sprintf(
		"\nToday is %s. All dates MUST be on or after this date.\n\n",
		time.Now().UTC().Format("2006-01-02"),
	))

	b.WriteString("Regenerate the entire draft, correcting every error.\n")
	b.WriteString("Keep everything else the same unless required to fix an error.\n")
	b.WriteString("Return ONLY valid JSON.\n\n")
	b.WriteString("Original prompt:\n")
	b.WriteString(originalPrompt)

	return b.String()
}