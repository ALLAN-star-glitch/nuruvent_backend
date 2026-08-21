// internal/modules/events/handler/dtos.go

package eventhandler

// ============================================================
// REQUEST DTOS
// ============================================================

// ✅ CreateDraftRequest - All fields optional for drafts (multipart/form-data)
type CreateDraftRequest struct {
	Name             string  `form:"name"`
	Description      string  `form:"description"`
	EventTypeID      string  `form:"event_type_id"`
	Date             string  `form:"date"`
	Time             string  `form:"time"`
	Duration         int     `form:"duration"`
	Price            float64 `form:"price"`
	CertificatePrice float64 `form:"certificate_price"`
	Location         string  `form:"location"`
	IsVirtual        bool    `form:"is_virtual"`
	ZoomLink         string  `form:"zoom_link"`
	MeetLink         string  `form:"meet_link"`
	MaxAttendees     int     `form:"max_attendees"`
	// ✅ NEW: Feature flags
	IsFeatured bool `form:"is_featured"`
	IsPrivate  bool `form:"is_private"`
}

// ✅ CreateEventRequest - All fields required for published events (application/json)
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

// ✅ CreateEventWithImageRequest - All fields required for published events (multipart/form-data)
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
	// ✅ NEW: Feature flags
	IsFeatured bool `form:"is_featured"`
	IsPrivate  bool `form:"is_private"`
}

// ✅ UpdateEventRequest - All fields optional for updates (application/json)
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
	// ✅ NEW: Feature flags
	IsFeatured bool `form:"is_featured"`
	IsPrivate  bool `form:"is_private"`
}

// ✅ BulkIDsRequest - For bulk operations (delete, publish, cancel, complete, restore)
type BulkIDsRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
}

// ✅ DuplicateEventRequest - For duplicating a single event
type DuplicateEventRequest struct {
	Name    string `json:"name"`
	Date    string `json:"date"`     // YYYY-MM-DD
	IsDraft bool   `json:"is_draft"` // Default: true
}

// ✅ BulkDuplicateRequest - For duplicating multiple events
type BulkDuplicateRequest struct {
	IDs            []string `json:"ids" binding:"required,min=1"`
	NamePrefix     string   `json:"name_prefix"`
	DateOffsetDays int      `json:"date_offset_days"`
	IsDraft        bool     `json:"is_draft"`
}

// ============================================================
// RESPONSE DTOS
// ============================================================

// ✅ EventResponse - API response for event data
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
	DeletedAt        *string `json:"deleted_at,omitempty"`
	DeletedBy        string  `json:"deleted_by,omitempty"`
	RestoredAt       *string `json:"restored_at,omitempty"`
	RestoredBy       string  `json:"restored_by,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// ✅ MediaInfoResponse - API response for media data
type MediaInfoResponse struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	MediaType  string `json:"media_type"`
	EntityID   string `json:"entity_id"`
	UploadedBy string `json:"uploaded_by"`
	CreatedAt  string `json:"created_at"`
}

// ✅ BulkDeleteResult - Response for bulk delete operations
type BulkDeleteResultResponse struct {
	DeletedCount int      `json:"deleted_count"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
	Errors       []string `json:"errors,omitempty"`
}

// ✅ BulkRestoreResult - Response for bulk restore operations
type BulkRestoreResultResponse struct {
	RestoredCount int      `json:"restored_count"`
	FailedIDs     []string `json:"failed_ids,omitempty"`
	Errors        []string `json:"errors,omitempty"`
}

// ✅ BulkStatusResult - Response for bulk status operations
type BulkStatusResultResponse struct {
	ProcessedCount int      `json:"processed_count"`
	FailedIDs      []string `json:"failed_ids,omitempty"`
	Errors         []string `json:"errors,omitempty"`
}

// ✅ BulkDuplicateResult - Response for bulk duplicate operations
type BulkDuplicateResultResponse struct {
	DuplicatedCount int `json:"duplicated_count"`
	CreatedEvents   []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"created_events"`
	FailedIDs []string `json:"failed_ids,omitempty"`
	Errors    []string `json:"errors,omitempty"`
}