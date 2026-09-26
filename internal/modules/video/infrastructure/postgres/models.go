// internal/modules/video/infrastructure/postgres/models.go

package postgres

import (
	"time"

)

// ============================================================
// VIDEO CONNECTIONS
// ============================================================

// VideoConnectionModel maps to the `video_connections` table.
type VideoConnectionModel struct {
	ID       string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID   string `gorm:"type:uuid;not null;index:idx_video_connections_user"`
	Platform string `gorm:"type:varchar(30);not null"`

	ExternalUserID string `gorm:"type:varchar(255);not null"`
	ExternalEmail  string `gorm:"type:varchar(255);not null;default:''"`
	ExternalOrgID  string `gorm:"type:varchar(255);not null;default:'';index:idx_video_connections_platform_org,priority:2"`

	AccessTokenEncrypted  string    `gorm:"type:text;not null"`
	RefreshTokenEncrypted string    `gorm:"type:text;not null;default:''"`
	TokenExpiresAt        time.Time `gorm:"not null"`
	Scopes                string    `gorm:"type:text;not null;default:''"`

	ConnectedAt time.Time  `gorm:"not null;default:now()"`
	RevokedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (VideoConnectionModel) TableName() string { return "video_connections" }

// ============================================================
// VIDEO OAUTH STATES
// ============================================================

// VideoOAuthStateModel maps to the `video_oauth_states` table.
type VideoOAuthStateModel struct {
	State     string `gorm:"primaryKey;type:varchar(64)"`
	UserID    string `gorm:"type:uuid;not null;index:idx_video_oauth_states_user"`
	Platform  string `gorm:"type:varchar(30);not null"`
	ReturnURL string `gorm:"type:text;not null;default:''"`

	CreatedAt  time.Time  `gorm:"not null;default:now()"`
	ExpiresAt  time.Time  `gorm:"not null;index:idx_video_oauth_states_expiry"`
	ConsumedAt *time.Time
}

func (VideoOAuthStateModel) TableName() string { return "video_oauth_states" }

// ============================================================
// VIDEO MEETINGS
// ============================================================

// VideoMeetingModel maps to the `video_meetings` table.
type VideoMeetingModel struct {
	ID       string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID   string `gorm:"type:uuid;not null;index:idx_video_meetings_user"`
	Platform string `gorm:"type:varchar(30);not null;index:idx_video_meetings_external,priority:1"`

	ExternalID string `gorm:"type:varchar(255);not null;index:idx_video_meetings_external,priority:2"`
	JoinURL    string `gorm:"type:text;not null"`
	StartURL   string `gorm:"type:text;not null;default:''"`
	Password   string `gorm:"type:varchar(255);not null;default:''"`

	Topic       string    `gorm:"type:varchar(500);not null"`
	StartTime   time.Time `gorm:"not null"`
	DurationSec int       `gorm:"not null;default:0"`
	Timezone    string    `gorm:"type:varchar(50);not null;default:''"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (VideoMeetingModel) TableName() string { return "video_meetings" }