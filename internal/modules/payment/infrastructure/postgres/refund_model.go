// internal/modules/payment/infrastructure/postgres/refund_model.go

package postgres

import "time"

// RefundModel maps to the `refunds` table.
type RefundModel struct {
	ID        string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	PaymentID string `gorm:"type:uuid;not null;index"`

	Amount   int64  `gorm:"not null"`
	Currency string `gorm:"type:varchar(3);not null;default:'KES'"`
	Reason   *string `gorm:"type:text"`
	ActorID  string `gorm:"type:uuid;not null;index"`

	ProviderReference *string `gorm:"type:varchar(255)"`
	Status            string  `gorm:"type:varchar(20);not null;default:'pending';index"`
	FailureReason     *string `gorm:"type:text"`

	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (RefundModel) TableName() string { return "refunds" }