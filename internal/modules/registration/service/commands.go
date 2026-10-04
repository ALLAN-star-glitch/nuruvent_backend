package service

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"

// RegisterCommand is the input for RegisterForEvent.
type RegisterCommand struct {
    EventID   string
    UserID    string        // empty for guest
    Guest     *GuestIdentity // nil for authenticated user
    Selections []TicketSelectionInput
}

// GuestIdentity carries guest registration details.
type GuestIdentity struct {
    Name  string
    Email string
    Phone string
}

// TicketSelectionInput is a requested ticket type + quantity.
// Prices are NOT supplied by the caller — the server computes them.
type TicketSelectionInput struct {
    TicketTypeID string
    Quantity     int
}

// JoinWaitlistCommand is the input for JoinWaitlist.
type JoinWaitlistCommand struct {
    EventID string
    UserID  string
    Guest   *GuestIdentity
	TicketTypeID string
}

// CancelCommand is the input for CancelRegistration.
type CancelCommand struct {
    RegistrationID string
    ActorID        string
    Reason         string
}

// ListFilterInput carries pagination and status filters from HTTP.
type ListFilterInput struct {
    Statuses []string
    Page     int
    PageSize int
}


// ListAllRegistrationsCommand is the input for the cross-event
// registration list.
type ListAllRegistrationsCommand struct {
	UserID    string
	EventID   string
	Search    string
	Statuses  []string
	SortBy    string
	SortOrder string
	Page      int
	PageSize  int
}

// ListAllRegistrationsResult is the output. Registrations is always
// non-nil so JSON marshals as [] rather than null.
type ListAllRegistrationsResult struct {
    Registrations []*registrationdomain.CrossEventRegistrationRow
    Total         int
    Page          int
    PageSize      int
}

// ListMineRegistrationsCommand is the input for the current user's
// own registration list.
type ListMineRegistrationsCommand struct {
	UserID    string
	EventID   string
	Search    string
	Statuses  []string
	SortBy    string
	SortOrder string
	Page      int
	PageSize  int
}

// ListMineRegistrationsResult is the output. Registrations is always
// non-nil so JSON marshals as [].
type ListMineRegistrationsResult struct {
	Registrations []*registrationdomain.CrossEventRegistrationRow
	Total         int
	Page          int
	PageSize      int
}