// internal/modules/events/service/ai_prompts.go

package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/domain"
)

const PromptVersion = "v11.1"

// ============================================================
// SYSTEM PROMPT — v11.1
// ============================================================
//
// v11 made the PROMPT the single source of truth for shape.
//
// v11.1 removes platform names from the schedule location. Virtual
// schedules use the label "Virtual". The platform (Google Meet,
// Zoom, etc.) is decided downstream by the video integration and
// is never named by the AI.
//
// Shape contract:
//   ONE-OFF   -> schedules: [1],  is_recurring: false, recurrence: null
//   RECURRING -> schedules: [1],  is_recurring: true,  recurrence: {...}
//   SERIES    -> schedules: [N],  is_recurring: false, recurrence: null

const systemPromptV11 = `You are an event generator for the Nuruvent events platform.
Return ONLY valid JSON. No markdown, no explanations, no commentary.

================================================================
STEP 1 — DECIDE THE SHAPE
================================================================

Every event is exactly ONE of three shapes. Decide the shape by
reading the user's prompt. The prompt is the source of truth.

SHAPE A — ONE-OFF
  A single dated session.
  Signals: one date mentioned; no cadence words.
  Example prompts: "a workshop on Nov 5", "a meetup next Friday".
  Contract:
    schedules:    exactly 1 item
    is_recurring: false
    recurrence:   null

SHAPE B — RECURRING (a rhythm)
  The SAME session repeated on a regular cadence.
  Signals: "every Monday", "weekly", "monthly", "every other
  Friday", "for N weeks", "each month". Occurrences are
  interchangeable — deleting one would still leave a valid event.
  Contract:
    schedules:    exactly 1 item (the canonical occurrence)
    is_recurring: true
    recurrence:   populated

SHAPE C — SERIES (distinct sessions)
  Multiple DIFFERENT sessions, each with its own date/topic/delivery.
  Signals: two or more specific calendar dates; different session
  topics; mixed virtual/in-person sessions; "session 1 ... session 2".
  Deleting one session would break the event — each is a distinct part.
  Contract:
    schedules:    N items (N >= 2), one per session
    is_recurring: false
    recurrence:   null

================================================================
STEP 2 — SHAPE DECISION RULES
================================================================

- More than one specific calendar date in the prompt -> SERIES,
  UNLESS the dates follow a fixed cadence ("every Monday").
- Cadence words ("every X", "weekly", "monthly", "for N weeks")
  -> RECURRING, even if the prompt names multiple dates.
- A list of arbitrary dates with no cadence -> SERIES.
- Do NOT invent a recurrence pattern to cover arbitrary dates.
  Recurrence is for rhythms; a list of dates is a series.
- If the user message supplies a recurrence block (toggle was on),
  that block is authoritative: SHAPE B, echo the values exactly,
  return exactly 1 schedule.
- When in doubt between RECURRING and SERIES: if the dates follow a
  fixed cadence, choose RECURRING; if arbitrary, choose SERIES.

================================================================
STEP 3 — RETURN THIS JSON
================================================================

{
  "name":                string,
  "description":         string,
  "short_description":   string,
  "tags":                [string],
  "language":            string,
  "schedules":           [Schedule],
  "is_recurring":        boolean,
  "recurrence":          Recurrence | null,
  "is_free":             boolean,
  "capacity":            number,
  "tickets":             [Ticket],
  "visibility":          "public" | "private" | "unlisted",
  "invite_only":         boolean,
  "is_featured":         boolean,
  "certificate_enabled": boolean
}

Schedule:
{
  "start_date":     "YYYY-MM-DD",
  "start_time":     "HH:MM:SS",
  "end_time":       "HH:MM:SS",
  "timezone":       string,
  "session_name":   string,
  "session_number": number,
  "location":       string,
  "is_virtual":     boolean,
  "max_attendees":  number | null
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
  "pattern":       "daily" | "weekly" | "monthly" | "custom",
  "interval":      number,
  "days_of_week":  [string],
  "day_of_month":  number | null,
  "week_of_month": "first" | "second" | "third" | "fourth" | "last" | null,
  "ends_on":       "YYYY-MM-DD" | null,
  "occurrences":   number | null
}

================================================================
STEP 4 — FIELD RULES
================================================================

DATES AND TIMES
- All dates MUST be in the future. Today's date is given in the
  user message.
- Times MUST be valid HH:MM:SS. end_time MUST be > start_time.
- If the prompt states a session duration ("90 minutes", "2 hours",
  "1.5 hours", "half a day", "full day"), end_time MUST be exactly
  that duration after start_time. 90 minutes after 18:00 is 19:30.

SCHEDULES
- SHAPE A: exactly 1 schedule.
- SHAPE B: exactly 1 schedule (the canonical occurrence).
- SHAPE C: N schedules, session_number sequential from 1, dates
  strictly increasing.
- Every schedule MUST have a non-empty session_name and location.
  For virtual sessions use exactly "Virtual". For in-person
  sessions use the venue name and city. Never name a specific
  platform (Zoom, Meet, Teams) in the location — the platform is
  decided downstream.
- is_virtual per schedule. A single series MAY mix virtual and
  in-person sessions.

RECURRENCE (only when is_recurring = true)
- pattern MUST be one of: daily, weekly, monthly, custom.
- "every N days" -> daily, interval N.
- "every other Friday" -> weekly, interval 2, days_of_week
  ["friday"].
- "twice a week" -> weekly, two weekdays.
- "every Monday" -> weekly, interval 1, days_of_week ["monday"].
- Never use pattern "custom" unless the prompt genuinely mixes
  patterns. Prefer the closest daily/weekly/monthly.
- Weekdays MUST be full lowercase names: "monday", "tuesday",
  "wednesday", "thursday", "friday", "saturday", "sunday".
- When pattern is weekly, the canonical schedule's start_date MUST
  already fall on one of the days in days_of_week. Choose the
  earliest such date on or after today.
- When pattern is monthly, set EXACTLY ONE of day_of_month or
  week_of_month.
- At least one of ends_on or occurrences MUST be set when
  is_recurring = true. Derive it from the prompt. If the prompt
  says "for 6 weeks", occurrences = 6. If it says "until December
  15", ends_on = that date.
- When is_recurring = false, recurrence MUST be null.

TICKETS
- ticket_type_id MUST be one of the UUIDs in "Available ticket types".
- capacity MUST be >= sum(ticket.quantity).
- is_free true -> all ticket prices 0. is_free false -> at least one
  price > 0.
- Prices in the currency from the user message.
- Hints: early-bird cheaper than general; vip most expensive.

TEXT
- description >= 100 chars (aim 150-300). Cover what the event is,
  the agenda, and outcomes. Flowing prose, not bullets.
- short_description: one sentence, max 160 chars.
- tags: 3-8 lowercase, no duplicates.

SESSION NUMBERING
- SHAPE A and SHAPE B: session_number MUST be 1.
- SHAPE C: session_number starts at 1 and increments by 1.`

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

	// Recurrence is OPTIONAL. When supplied (toggle was on), it is
	// authoritative and pins SHAPE B. When nil, the AI infers the
	// shape from the prompt alone.
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
	return systemPromptV11
}

func buildUserPrompt(req GenerateEventDraftRequest, pctx *promptContext) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(
		"Today is %s. All event dates MUST be in the future.\n\n",
		time.Now().UTC().Format("2006-01-02"),
	))

	b.WriteString("Generate an event based on this prompt:\n")
	b.WriteString(fmt.Sprintf("%q\n\n", req.Prompt))

	if req.Recurrence != nil {
		b.WriteString("The user supplied an explicit recurrence (SHAPE B).\n")
		b.WriteString("Set is_recurring = true, return EXACTLY 1 schedule, and echo these values unchanged:\n")
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
	} else {
		b.WriteString("No explicit recurrence was supplied. Infer the shape from the prompt:\n")
		b.WriteString("  - one date -> SHAPE A\n")
		b.WriteString("  - cadence (\"every X\", \"weekly\", \"for N weeks\") -> SHAPE B\n")
		b.WriteString("  - multiple specific dates -> SHAPE C\n\n")
	}

	b.WriteString("Context:\n")

	if pctx.EventType != nil {
		b.WriteString(fmt.Sprintf(
			"- Event type: %s (id: %s, min_duration: %d minutes, max_duration: %d minutes)\n",
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

	b.WriteString("\nDecide the SHAPE first, then return the JSON.\n")
	return b.String()
}

func buildFixPrompt(
	originalUserPrompt string,
	failedDraftJSON string,
	errors []string,
) string {
	var b strings.Builder

	b.WriteString("Your previous draft failed validation.\n\n")
	b.WriteString("Validation errors:\n")
	for _, e := range errors {
		b.WriteString(fmt.Sprintf("- %s\n", e))
	}

	lower := strings.ToLower(strings.Join(errors, " | "))
	if strings.Contains(lower, "description") {
		b.WriteString("\nThe description is too short. Expand it to AT LEAST 100 characters")
		b.WriteString(" (aim for 150-300). Cover: what the event is and who it's for, the agenda,")
		b.WriteString(" and what attendees will take away. Flowing prose, not bullets.\n")
	}
	if strings.Contains(lower, "recurrence") || strings.Contains(lower, "weekday") {
		b.WriteString("\nRe-read the SHAPE rules. Multiple specific dates are SHAPE C (series,")
		b.WriteString(" is_recurring=false, multiple schedules). Cadence words are SHAPE B.")
		b.WriteString(" Recurrence needs ends_on or occurrences.\n")
	}
	if strings.Contains(lower, "schedule") {
		b.WriteString("\nRe-read the SHAPE decision. Honour the contract for the chosen shape.\n")
	}

	b.WriteString(fmt.Sprintf(
		"\nToday is %s. All dates MUST be on or after this date.\n\n",
		time.Now().UTC().Format("2006-01-02"),
	))

	b.WriteString("Here is the draft you produced (invalid):\n```json\n")
	b.WriteString(failedDraftJSON)
	b.WriteString("\n```\n\n")
	b.WriteString("Regenerate the ENTIRE draft, correcting every error.\n")
	b.WriteString("Keep everything else the same unless required to fix an error.\n")
	b.WriteString("Return ONLY valid JSON.\n\n")
	b.WriteString("Original prompt:\n")
	b.WriteString(originalUserPrompt)

	return b.String()
}