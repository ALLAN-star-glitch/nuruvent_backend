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
	IsFeatured       bool    `form:"is_featured"`
	IsPrivate        bool    `form:"is_private"`
	TeamType         string  `form:"team_type"` // ✅ NEW: "personal" or "institution"
}

// ✅ CreateEventRequest - All fields required for published events (application/json)
type CreateEventRequest struct {
	Name             string  `json:"name" binding:"required"`
	Description      string  `json:"description"`
	EventTypeID      string  `json:"event_type_id" binding:"required"`
	Date             string  `json:"date" binding:"required"`
	Time             string  `json:"time" binding:"required"`
	Duration         int     `json:"duration" binding:"required"`
	Price            float64 `json:"price"`
	CertificatePrice float64 `json:"certificate_price"`
	Location         string  `json:"location"`
	IsVirtual        bool    `json:"is_virtual"`
	ZoomLink         string  `json:"zoom_link"`
	MeetLink         string  `json:"meet_link"`
	MaxAttendees     int     `json:"max_attendees"`
	TeamType         string  `json:"team_type"` // ✅ NEW: "personal" or "institution"
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
	IsFeatured       bool    `form:"is_featured"`
	IsPrivate        bool    `form:"is_private"`
	TeamType         string  `form:"team_type"` // ✅ NEW: "personal" or "institution"
}

// ✅ UpdateEventRequest - All fields optional for updates (application/json)
type UpdateEventRequest struct {
	Name             *string  `json:"name,omitempty"`
	DisplayName      *string  `json:"display_name,omitempty"`
	Description      *string  `json:"description,omitempty"`
	EventTypeID      *string  `json:"event_type_id,omitempty"`
	EventStatusID    *string  `json:"event_status_id,omitempty"`
	Date             *string  `json:"date,omitempty"`
	Time             *string  `json:"time,omitempty"`
	Duration         *int     `json:"duration,omitempty"`
	Price            *float64 `json:"price,omitempty"`
	CertificatePrice *float64 `json:"certificate_price,omitempty"`
	Location         *string  `json:"location,omitempty"`
	IsVirtual        *bool    `json:"is_virtual,omitempty"`
	ZoomLink         *string  `json:"zoom_link,omitempty"`
	MeetLink         *string  `json:"meet_link,omitempty"`
	MaxAttendees     *int     `json:"max_attendees,omitempty"`
	IsFeatured       *bool    `json:"is_featured,omitempty"`
	IsPrivate        *bool    `json:"is_private,omitempty"`
}

// ✅ BulkIDsRequest - For bulk operations
type BulkIDsRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
}

// ✅ DuplicateEventRequest - For duplicating a single event
type DuplicateEventRequest struct {
	Name    string `json:"name"`
	Date    string `json:"date"`
	IsDraft bool   `json:"is_draft"`
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

// ✅ CreatorDTO - Creator information in API responses
type CreatorDTO struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	DisplayName     string `json:"display_name,omitempty"`
	Email           string `json:"email"`
	Phone           string `json:"phone,omitempty"`
	AccountType     string `json:"account_type"`
	InstitutionName string `json:"institution_name,omitempty"`
}

// ✅ EventResponse - API response for event data
type EventResponse struct {
	ID               string     `json:"id"`
	Slug             string     `json:"slug"`
	Name             string     `json:"name"`
	DisplayName      string     `json:"display_name,omitempty"`
	Description      string     `json:"description,omitempty"`
	EventTypeID      string     `json:"event_type_id"`
	EventStatusID    string     `json:"event_status_id"`
	InstitutionID    string     `json:"institution_id,omitempty"` // ✅ NULL for personal events
	TeamType         string     `json:"team_type"`               // ✅ NEW: "personal" or "institution"
	ImageURL         string     `json:"image_url,omitempty"`
	ThumbnailURL     string     `json:"thumbnail_url,omitempty"`
	Date             string     `json:"date"`
	Time             string     `json:"time"`
	Duration         int        `json:"duration"`
	Price            float64    `json:"price"`
	CertificatePrice float64    `json:"certificate_price"`
	Location         string     `json:"location,omitempty"`
	IsVirtual        bool       `json:"is_virtual"`
	ZoomLink         string     `json:"zoom_link,omitempty"`
	MeetLink         string     `json:"meet_link,omitempty"`
	MaxAttendees     int        `json:"max_attendees"`
	CurrentAttendees int        `json:"current_attendees"`
	IsActive         bool       `json:"is_active"`
	IsFeatured       bool       `json:"is_featured"`
	IsPrivate        bool       `json:"is_private"`
	DeletedAt        *string    `json:"deleted_at,omitempty"`
	DeletedBy        string     `json:"deleted_by,omitempty"`
	RestoredAt       *string    `json:"restored_at,omitempty"`
	RestoredBy       string     `json:"restored_by,omitempty"`
	CreatedAt        string     `json:"created_at"`
	UpdatedAt        string     `json:"updated_at"`
	Creator          CreatorDTO `json:"creator"`
}

// ✅ MediaInfoResponse - Response for media data
type MediaInfoResponse struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	MediaType  string `json:"media_type"`
	EntityID   string `json:"entity_id"`
	UploadedBy string `json:"uploaded_by"`
	CreatedAt  string `json:"created_at"`
}

// ✅ BulkDeleteResultResponse
type BulkDeleteResultResponse struct {
	DeletedCount int      `json:"deleted_count"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
	Errors       []string `json:"errors,omitempty"`
}

// ✅ BulkRestoreResultResponse
type BulkRestoreResultResponse struct {
	RestoredCount int      `json:"restored_count"`
	FailedIDs     []string `json:"failed_ids,omitempty"`
	Errors        []string `json:"errors,omitempty"`
}

// ✅ BulkStatusResultResponse
type BulkStatusResultResponse struct {
	ProcessedCount int      `json:"processed_count"`
	FailedIDs      []string `json:"failed_ids,omitempty"`
	Errors         []string `json:"errors,omitempty"`
}

// ✅ BulkDuplicateResultResponse
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