package postgres

import (
	"time"

	"gorm.io/gorm"
)

// ============================================================
// ATTENDEES
// ============================================================

// AttendeeModel maps to the `attendees` table.
type AttendeeModel struct {
	ID           string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ExternalType string `gorm:"type:varchar(50);not null;index:idx_attendees_external,priority:1"`
	ExternalID   string `gorm:"type:uuid;not null;index:idx_attendees_external,priority:2"`

	DisplayName string `gorm:"type:varchar(255);not null"`
	Email       string `gorm:"type:varchar(255);not null;index:idx_attendees_email"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (AttendeeModel) TableName() string { return "attendees" }

// ============================================================
// SESSIONS
// ============================================================

// SessionModel maps to the `sessions` table.
type SessionModel struct {
	ID           string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ExternalType string `gorm:"type:varchar(50);not null;index:idx_sessions_external,priority:1"`
	ExternalID   string `gorm:"type:uuid;not null;index:idx_sessions_external,priority:2"`

	ProviderSessionID string `gorm:"type:varchar(255);not null;default:'';index:idx_sessions_provider_session"`

	Title           string    `gorm:"type:varchar(255);not null"`
	ScheduledStart  time.Time `gorm:"not null;index:idx_sessions_scheduled"`
	ScheduledEnd    time.Time `gorm:"not null"`
	DurationMinutes int       `gorm:"not null"`

	Provider          string `gorm:"type:varchar(20);not null;index:idx_sessions_provider"`
	ProviderMeetingID string `gorm:"type:varchar(255)"`
	ProviderURL       string `gorm:"type:text"`

	Status string `gorm:"type:varchar(20);not null;default:'scheduled';index:idx_sessions_status"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (SessionModel) TableName() string { return "sessions" }



// ============================================================
// JOIN TOKENS
// ============================================================

// JoinTokenModel maps to the `join_tokens` table.
type JoinTokenModel struct {
	ID         string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	AttendeeID string `gorm:"type:uuid;not null;index:idx_join_tokens_attendee"`
	SessionID  string `gorm:"type:uuid;not null;index:idx_join_tokens_session"`

	TokenHash string `gorm:"type:char(64);not null;uniqueIndex:idx_join_tokens_hash"`

	IssuedAt  time.Time  `gorm:"not null"`
	ExpiresAt time.Time  `gorm:"not null;index:idx_join_tokens_expires"`
	RevokedAt *time.Time
}

func (JoinTokenModel) TableName() string { return "join_tokens" }

// ============================================================
// ATTENDANCE RECORDS
// ============================================================

// AttendanceRecordModel maps to the `attendance_records` table.
type AttendanceRecordModel struct {
	ID         string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	AttendeeID string `gorm:"type:uuid;not null;index:idx_attendance_records_attendee_session,priority:1"`
	SessionID  string `gorm:"type:uuid;not null;index:idx_attendance_records_attendee_session,priority:2"`

	JoinTime  time.Time  `gorm:"not null;index:idx_attendance_records_join_time"`
	LeaveTime *time.Time `gorm:"index:idx_attendance_records_open"`

	DurationSeconds int64  `gorm:"not null;default:0"`
	Source          string `gorm:"type:varchar(30);not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (AttendanceRecordModel) TableName() string { return "attendance_records" }

// ============================================================
// ATTENDEE SESSION STATUS
// ============================================================

// AttendeeSessionStatusModel maps to the `attendee_session_statuses`
// table. Composite primary key on (attendee_id, session_id).
type AttendeeSessionStatusModel struct {
	AttendeeID string `gorm:"primaryKey;type:uuid"`
	SessionID  string `gorm:"primaryKey;type:uuid"`

	DerivedStatus        string `gorm:"type:varchar(20);not null;index:idx_ass_status"`
	TotalDurationSeconds int64  `gorm:"not null;default:0"`

	HostConfirmed   bool    `gorm:"not null;default:false;index:idx_ass_confirmed"`
	ConfirmedStatus *string `gorm:"type:varchar(20)"` // nullable; empty means "no override"
	ConfirmedBy     *string `gorm:"type:uuid"`
	ConfirmedAt     *time.Time
	ConfirmReason   string `gorm:"type:text;not null;default:''"`

	LastDerivedAt time.Time `gorm:"not null"`
}

func (AttendeeSessionStatusModel) TableName() string { return "attendee_session_statuses" }

// ============================================================
// ATTENDEE ROLLUP STATUS
// ============================================================

// AttendeeRollupStatusModel maps to the `attendee_rollup_statuses`
// table. Composite primary key on (attendee_id, external_type, external_id).
type AttendeeRollupStatusModel struct {
	AttendeeID   string `gorm:"primaryKey;type:uuid"`
	ExternalType string `gorm:"primaryKey;type:varchar(50)"`
	ExternalID   string `gorm:"primaryKey;type:uuid"`

	DerivedStatus        string `gorm:"type:varchar(20);not null;index:idx_ars_status"`
	SessionsTotal        int    `gorm:"not null;default:0"`
	SessionsAttended     int    `gorm:"not null;default:0"`
	SessionsConfirmed    int    `gorm:"not null;default:0"`
	TotalDurationSeconds int64  `gorm:"not null;default:0"`

	LastDerivedAt time.Time `gorm:"not null"`
}

func (AttendeeRollupStatusModel) TableName() string { return "attendee_rollup_statuses" }

// ============================================================
// ATTENDANCE OVERRIDES (audit)
// ============================================================

// AttendanceOverrideModel maps to the `attendance_overrides` table.
// Append-only.
type AttendanceOverrideModel struct {
	ID         string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	AttendeeID string `gorm:"type:uuid;not null;index:idx_overrides_attendee_session,priority:1"`
	SessionID  string `gorm:"type:uuid;not null;index:idx_overrides_attendee_session,priority:2"`

	ActorID     string `gorm:"type:uuid;not null;index:idx_overrides_actor"`
	PriorStatus string `gorm:"type:varchar(20);not null"`
	NewStatus   string `gorm:"type:varchar(20);not null"`
	Reason      string `gorm:"type:text;not null;default:''"`

	CreatedAt time.Time `gorm:"not null;index:idx_overrides_created"`
}

func (AttendanceOverrideModel) TableName() string { return "attendance_overrides" }