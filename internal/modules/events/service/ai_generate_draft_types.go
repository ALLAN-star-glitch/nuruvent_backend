// internal/modules/events/service/ai_generate_draft_types.go

package service

// ============================================================
// AI OUTPUT DTO
// ============================================================
//
// GeneratedEventDraft mirrors the create-event form for the fields
// the AI is responsible for: identity, discovery, pricing, policy,
// and the schedule(s).
//
// Shape contract:
//   ONE-OFF   -> len(Schedules) == 1, IsRecurring == false
//   RECURRING -> len(Schedules) == 1, IsRecurring == true
//   SERIES    -> len(Schedules) >= 2, IsRecurring == false
//
// Event-level timing, venue, virtual/hybrid flags, and meeting links
// are derived downstream from the full schedule list. They are NOT
// the AI's job.

type GeneratedEventDraft struct {
	Name               string               `json:"name"`
	Description        string               `json:"description"`
	ShortDescription   string               `json:"short_description"`
	Tags               []string             `json:"tags"`
	Language           string               `json:"language"`
	Schedules          []GeneratedSchedule  `json:"schedules"`
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

type DraftShape string

const (
	ShapeInvalid   DraftShape = "invalid"
	ShapeOneOff    DraftShape = "one_off"
	ShapeRecurring DraftShape = "recurring"
	ShapeSeries    DraftShape = "series"
)

func (d *GeneratedEventDraft) Shape() DraftShape {
	switch {
	case d.IsRecurring && len(d.Schedules) == 1:
		return ShapeRecurring
	case !d.IsRecurring && len(d.Schedules) == 1:
		return ShapeOneOff
	case !d.IsRecurring && len(d.Schedules) >= 2:
		return ShapeSeries
	default:
		return ShapeInvalid
	}
}

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

type correctionContext struct {
	Request         GenerateEventDraftRequest
	EventTypeID     string
	CategoryID      *string
	TicketTypeIDs   map[string]struct{}
	Timezone        string
	Language        string
	Currency        string
	MinCapacity     int
	MaxCapacity     int
	EventTypeMinDur int
	EventTypeMaxDur int
}