package service

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