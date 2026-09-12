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
//
// Bump this whenever the system prompt or JSON schema changes.
// It is logged on every AI call so production issues can be
// correlated with prompt revisions.
const PromptVersion = "v1"

// ============================================================
// SYSTEM PROMPT — v1
// ============================================================

const systemPromptV1 = `You are an event generator for the Nuruvent events platform.
Return ONLY valid JSON. No markdown, no explanations, no commentary.

Schema:
{
  "name":              string,   // the event's user-facing title (e.g. "Hands-on Kubernetes Workshop")
  "description":       string,   // 200-500 words
  "short_description": string,   // one sentence
  "tags":              [string], // 3-6 lowercase tags
  "language":          string,   // ISO 639-1
  "schedules":         [Schedule],
  "is_multi_day":      boolean,
  "is_virtual":        boolean,
  "is_hybrid":         boolean,
  "in_person_location": string,  // if in-person or hybrid
  "venue_name":        string,
  "venue_address":     string,
  "venue_city":        string,
  "venue_country":     string,
  "virtual_platform":  string,   // if virtual or hybrid
  "virtual_platform_url": string,
  "timezone":          string,   // IANA
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
  "ticket_type_id": "UUID",   // MUST be one of the provided ticket type IDs
  "name":           string,
  "description":    string,
  "price":          number,   // >= 0
  "quantity":       number,   // >= 1
  "max_per_person": number | null
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
- ticket_type_id MUST be one of the UUIDs in the "Available ticket types" list.`

// ============================================================
// REQUEST DTO
// ============================================================

// GenerateEventDraftRequest is the caller-supplied input for the
// AI event-draft endpoint.
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

	// Never bound from the request body — set by the handler from
	// the authenticated context. Ignored by json.Unmarshal.
	CreatedBy string `json:"-"`
	TeamID    string `json:"-"`
	TeamType  string `json:"-"`
	AccountID string `json:"-"`
}

// ============================================================
// PROMPT CONTEXT
// ============================================================
//
// promptContext carries the resolved DB rows used to build the
// user prompt. Populated by loadPromptContext in ai_generate_draft.go.

type promptContext struct {
	EventType   *domain.EventType
	Category    *domain.Category // nil if not provided
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

// buildSystemPrompt returns the versioned system prompt.
func buildSystemPrompt() string {
	return systemPromptV1
}

// buildUserPrompt assembles the dynamic user prompt from the request
// and the resolved DB context.
func buildUserPrompt(req GenerateEventDraftRequest, pctx *promptContext) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf(
		"Today is %s. All event dates MUST be in the future.\n\n",
		time.Now().UTC().Format("2006-01-02"),
	))

	b.WriteString("Generate an event based on this prompt:\n")
	b.WriteString(fmt.Sprintf("%q\n\n", req.Prompt))

	b.WriteString("Context:\n")

	// Event type
	if pctx.EventType != nil {
		b.WriteString(fmt.Sprintf(
			"- Event type: %s (id: %s, min_duration: %d, max_duration: %d)\n",
			pctx.EventType.DisplayName,
			pctx.EventType.ID,
			pctx.EventType.MinDuration,
			pctx.EventType.MaxDuration,
		))
	}

	// Category
	if pctx.Category != nil {
		b.WriteString(fmt.Sprintf(
			"- Category: %s (id: %s)\n",
			pctx.Category.DisplayName,
			pctx.Category.ID,
		))
	}

	// Ticket types — use Slug (stable, human-readable) + DisplayName
	b.WriteString("- Available ticket types:\n")
	for _, tt := range pctx.TicketTypes {
		b.WriteString(fmt.Sprintf(
			"    - %s: %s (id: %s)\n",
			tt.Slug, tt.DisplayName, tt.ID,
		))
	}

	// Environment defaults
	b.WriteString(fmt.Sprintf("- Language: %s\n", pctx.Language))
	b.WriteString(fmt.Sprintf("- Timezone: %s\n", pctx.Timezone))
	b.WriteString(fmt.Sprintf("- Currency: %s\n", pctx.Currency))
	b.WriteString(fmt.Sprintf("- Capacity range: %d-%d\n", pctx.MinCapacity, pctx.MaxCapacity))

	b.WriteString("\nReturn the JSON now.\n")
	return b.String()
}

// buildFixPrompt wraps the original user prompt with a list of
// validation errors for the single retry attempt.
func buildFixPrompt(originalPrompt string, errors []string) string {
	var b strings.Builder

	b.WriteString("The previous draft failed validation with these errors:\n")
	for _, e := range errors {
		b.WriteString(fmt.Sprintf("- %s\n", e))
	}

	// Anchor the model to the current date so it doesn't keep guessing
	// from its training cutoff.
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