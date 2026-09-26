// internal/modules/events/service/ai_generate_draft_types.go

package service

import "time"

// ============================================================
// AI OUTPUT DTO
// ============================================================
//
// GeneratedEventDraft is the shape the AI returns. It mirrors the
// create-event form: identity, discovery, pricing, policy, and ONE
// canonical schedule.
//
// Event-level timing, venue, virtual/hybrid flags, and meeting links
// are NOT part of the AI's job. They are derived from schedules by
// deriveEventFromSchedules when the draft is turned into a real event.

type GeneratedEventDraft struct {
	Name               string               `json:"name"`
	Description        string               `json:"description"`
	ShortDescription   string               `json:"short_description"`
	Tags               []string             `json:"tags"`
	Language           string               `json:"language"`
	Schedules          []GeneratedSchedule  `json:"schedules"` // exactly one
	IsRecurring        bool                 `json:"is_recurring"`
	Recurrence         *GeneratedRecurrence `json:"recurrence"`
	IsFree             bool                 `json:"is_free"`
	Capacity           int                  `json:"capacity"`
	Tickets            []GeneratedTicket    `json:"tickets"`
	Visibility         string               `json:"visibility"`
	InviteOnly         bool                 `json:"invite_only"`
	IsFeatured         bool                 `json:"is_featured"`
	CertificateEnabled bool                 `json:"certificate_enabled"`
}

// GeneratedSchedule is a single canonical session. The AI always
// returns exactly one. If the event repeats, the recurrence block
// conveys the cadence and this schedule conveys the shape of one
// occurrence.
type GeneratedSchedule struct {
	StartDate     string `json:"start_date"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Timezone      string `json:"timezone"`
	SessionName   string `json:"session_name"`
	SessionNumber int    `json:"session_number"`
	Location      string `json:"location"`
	IsVirtual     bool   `json:"is_virtual"`
	MaxAttendees  *int   `json:"max_attendees"`
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




// ensure time import is used (GeneratedRecurrence keeps the *string
// for ends_on, so this file doesn't strictly need time — kept for
// future timestamp additions).
var _ = time.Time{}