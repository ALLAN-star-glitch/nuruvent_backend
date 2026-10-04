
// internal/modules/registration/delivery/http/dto.go

package http

import "time"

// ============================================================
// REQUESTS
// ============================================================

// RegisterRequest is the body for POST /events/:id/register.
type RegisterRequest struct {
	Selections []TicketSelectionRequest `json:"selections"`
	Guest      *GuestRequest            `json:"guest,omitempty"`
}

// TicketSelectionRequest carries a ticket type + quantity.
// Prices are NEVER accepted from the client.
type TicketSelectionRequest struct {
	TicketTypeID string `json:"ticket_type_id"`
	Quantity     int    `json:"quantity"`
}



// ============================================================
// CROSS-EVENT REGISTRATION LIST
// ============================================================

// CrossEventRegistrationResponse is one row in the cross-event
// registration list. Flat shape — each field maps to a table column
// on the frontend. Richer nested data (selections, pricing) is
// available via GET /registrations/:id.
type CrossEventRegistrationResponse struct {
	ID                 string    `json:"id"`
	RegistrationNumber string    `json:"registration_number"`
	Status             string    `json:"status"`
	StatusLabel        string    `json:"status_label"`
	AttendeeName       string    `json:"attendee_name"`
	Email              string    `json:"email"`
	Phone              string    `json:"phone,omitempty"`
	IsGuest            bool      `json:"is_guest"`
	UserID             string    `json:"user_id,omitempty"`
	EventID            string    `json:"event_id"`
	EventName          string    `json:"event_name"`
	EventStartDate     string    `json:"event_start_date,omitempty"`
	EventImageURL string `json:"event_image_url,omitempty"` 

	// Event format / location
	IsVirtual        bool   `json:"is_virtual"`
	IsHybrid         bool   `json:"is_hybrid"`
	VenueName        string `json:"venue_name,omitempty"`
	VenueAddress     string `json:"venue_address,omitempty"`
	VenueCity        string `json:"venue_city,omitempty"`
	VenueCountry     string `json:"venue_country,omitempty"`
	InPersonLocation string `json:"in_person_location,omitempty"`

	TicketName string    `json:"ticket_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// CrossEventRegistrationListResponse is the paginated envelope.
type CrossEventRegistrationListResponse struct {
	Registrations []CrossEventRegistrationResponse `json:"registrations"`
	Total         int                              `json:"total"`
	Page          int                              `json:"page"`
	PageSize      int                              `json:"page_size"`
}








// GuestRequest carries guest identity when the caller is unauthenticated.
type GuestRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone,omitempty"`
}

// JoinWaitlistRequest is the body for POST /events/:id/waitlist.
type JoinWaitlistRequest struct {
	TicketTypeID string        `json:"ticket_type_id,omitempty"`
	Guest *GuestRequest `json:"guest,omitempty"`
}

// CancelRequest is the body for DELETE /registrations/:id.
type CancelRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ============================================================
// RESPONSES
// ============================================================

// RegistrationResponse is the canonical output for a registration.
type RegistrationResponse struct {
	ID                 string                    `json:"id"`
	RegistrationNumber string                    `json:"registration_number"`
	Status             StatusResponse            `json:"status"`
	UserID             string                    `json:"user_id,omitempty"`
	Guest              *GuestResponse            `json:"guest,omitempty"`
	EventID            string                    `json:"event_id"`
	Selections         []TicketSelectionResponse `json:"selections"`
	Pricing            PricingResponse           `json:"pricing"`
	CreatedAt          time.Time                 `json:"created_at"`
	ConfirmedAt        *time.Time                `json:"confirmed_at,omitempty"`
	CancelledAt        *time.Time                `json:"cancelled_at,omitempty"`
	CancellationReason string                    `json:"cancellation_reason,omitempty"`
}

// StatusResponse is the status block embedded in a registration.
type StatusResponse struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// GuestResponse echoes guest identity.
type GuestResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone,omitempty"`
}

// TicketSelectionResponse echoes a ticket line.
type TicketSelectionResponse struct {
	TicketTypeID string `json:"ticket_type_id"`
	Quantity     int    `json:"quantity"`
	UnitPrice    int64  `json:"unit_price"` // minor units
	Discount     int64  `json:"discount"`
	LineTotal    int64  `json:"line_total"`
}

// PricingResponse echoes the pricing snapshot.
type PricingResponse struct {
	Currency      string `json:"currency"`
	Subtotal      int64  `json:"subtotal"`
	DiscountTotal int64  `json:"discount_total"`
	Total         int64  `json:"total"`
}

// WaitlistResponse is the output for a waitlist entry.
type WaitlistResponse struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

// ListResponse wraps a paginated list.
type ListResponse struct {
	Data     []RegistrationResponse `json:"data"`
	Total    int                    `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// SessionLinkResponse is the wire format for one session's
// personalized join URL.
type SessionLinkResponse struct {
	SessionID      string    `json:"session_id"`
	SessionTitle   string    `json:"session_title"`
	ScheduledStart time.Time `json:"scheduled_start"`
	ScheduledEnd   time.Time `json:"scheduled_end"`
	Platform       string    `json:"platform"`
	JoinURL        string    `json:"join_url"`
	ExpiresAt      time.Time `json:"expires_at"`
}


// SessionLinkGroupResponse groups session links under the event the
// user is registered for.
type SessionLinkGroupResponse struct {
	RegistrationID string                `json:"registration_id"`
	EventID        string                `json:"event_id"`
	EventName      string                `json:"event_name"`
	EventDate      time.Time             `json:"event_date"`
	Links          []SessionLinkResponse `json:"links"`
}

// MySessionLinksResponse is what GET /me/session-links returns.
type MySessionLinksResponse struct {
	Groups []SessionLinkGroupResponse `json:"groups"`
}