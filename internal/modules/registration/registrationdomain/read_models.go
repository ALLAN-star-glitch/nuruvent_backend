package registrationdomain

import "time"

type CrossEventRegistrationRow struct {
	// Base registration
	ID                 string
	RegistrationNumber string
	Status             Status
	UserID             string
	CreatedAt          time.Time

	// Attendee identity
	AttendeeName string
	Email        string
	Phone        string
	IsGuest      bool

	// Event
	EventID        string
	EventName      string
	EventStartDate time.Time
	EventImageURL  string 

	// Event format / location
	IsVirtual        bool
	IsHybrid         bool
	VenueName        string
	VenueAddress     string
	VenueCity        string
	VenueCountry     string
	InPersonLocation string

	// Ticket
	TicketName string
}