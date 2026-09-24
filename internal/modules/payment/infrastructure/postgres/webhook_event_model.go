// internal/modules/payment/infrastructure/postgres/webhook_event_model.go

package postgres

import "time"

// WebhookEventModel maps to the `webhook_events` table.
type WebhookEventModel struct {
	ID              string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Provider        string `gorm:"type:varchar(30);not null"`
	ProviderEventID string `gorm:"type:varchar(255);not null"`
	SignatureValid  bool   `gorm:"not null;default:false"`

	// Payload is the raw webhook body, stored as BYTEA.
	Payload []byte `gorm:"type:bytea;not null"`

	ReceivedAt      time.Time  `gorm:"not null"`
	ProcessedAt     *time.Time
	ProcessingError *string `gorm:"type:text"`
}

func (WebhookEventModel) TableName() string { return "webhook_events" }