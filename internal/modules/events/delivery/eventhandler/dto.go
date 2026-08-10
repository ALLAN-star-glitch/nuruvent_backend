package eventhandler

// ============================================================
// REQUEST DTOS
// ============================================================

type CreateEventRequest struct {
	Name             string  `json:"name" binding:"required"`
	Description      string  `json:"description"`
	EventTypeID      string  `json:"event_type_id" binding:"required"`
	Date             string  `json:"date" binding:"required"` // YYYY-MM-DD
	Time             string  `json:"time" binding:"required"` // HH:MM
	Duration         int     `json:"duration" binding:"required"`
	Price            float64 `json:"price"`
	CertificatePrice float64 `json:"certificate_price"`
	Location         string  `json:"location"`
	IsVirtual        bool    `json:"is_virtual"`
	ZoomLink         string  `json:"zoom_link"`
	MeetLink         string  `json:"meet_link"`
	MaxAttendees     int     `json:"max_attendees"`
}

type UpdateEventRequest struct {
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	EventTypeID      string  `json:"event_type_id"`
	EventStatusID    string  `json:"event_status_id"`
	Date             string  `json:"date"` // YYYY-MM-DD
	Time             string  `json:"time"` // HH:MM
	Duration         int     `json:"duration"`
	Price            float64 `json:"price"`
	CertificatePrice float64 `json:"certificate_price"`
	Location         string  `json:"location"`
	IsVirtual        bool    `json:"is_virtual"`
	ZoomLink         string  `json:"zoom_link"`
	MeetLink         string  `json:"meet_link"`
	MaxAttendees     int     `json:"max_attendees"`
}

// ============================================================
// RESPONSE DTOS
// ============================================================

type EventResponse struct {
	ID               string  `json:"id"`
	Slug             string  `json:"slug"`
	Name             string  `json:"name"`
	DisplayName      string  `json:"display_name,omitempty"`
	Description      string  `json:"description,omitempty"`
	EventTypeID      string  `json:"event_type_id"`
	EventStatusID    string  `json:"event_status_id"`
	ImageURL         string  `json:"image_url,omitempty"`
	ThumbnailURL     string  `json:"thumbnail_url,omitempty"`
	Date             string  `json:"date"`
	Time             string  `json:"time"`
	Duration         int     `json:"duration"`
	Price            float64 `json:"price"`
	CertificatePrice float64 `json:"certificate_price"`
	Location         string  `json:"location,omitempty"`
	IsVirtual        bool    `json:"is_virtual"`
	ZoomLink         string  `json:"zoom_link,omitempty"`
	MeetLink         string  `json:"meet_link,omitempty"`
	MaxAttendees     int     `json:"max_attendees"`
	CurrentAttendees int     `json:"current_attendees"`
	AccountID        string  `json:"account_id"`
	CreatedBy        string  `json:"created_by"`
	IsActive         bool    `json:"is_active"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// CreateEventWithImageRequest represents the request for creating an event with image
type CreateEventWithImageRequest struct {
	Name             string  `form:"name" binding:"required"`
	Description      string  `form:"description"`
	EventTypeID      string  `form:"event_type_id" binding:"required"`
	Date             string  `form:"date" binding:"required"`
	Time             string  `form:"time" binding:"required"`
	Duration         int     `form:"duration" binding:"required"`
	Price            float64 `form:"price"`
	CertificatePrice float64 `form:"certificate_price"`
	Location         string  `form:"location"`
	IsVirtual        bool    `form:"is_virtual"`
	ZoomLink         string  `form:"zoom_link"`
	MeetLink         string  `form:"meet_link"`
	MaxAttendees     int     `form:"max_attendees"`
}