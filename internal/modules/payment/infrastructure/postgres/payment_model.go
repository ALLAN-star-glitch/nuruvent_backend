// internal/modules/payment/infrastructure/postgres/payment_model.go

package postgres

import "time"

// PaymentModel maps to the `payments` table.
type PaymentModel struct {
	ID      string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	OrderID string `gorm:"type:uuid;not null;index"`

	Provider string `gorm:"type:varchar(30);not null"`
	Method   string `gorm:"type:varchar(20);not null"`

	Amount   int64  `gorm:"not null"`
	Currency string `gorm:"type:varchar(3);not null;default:'KES'"`

	Status         string  `gorm:"type:varchar(20);not null;default:'pending';index"`
	IdempotencyKey string  `gorm:"type:varchar(100);not null"`
	ProviderReference *string `gorm:"type:varchar(255)"`
	RedirectURL string `gorm:"column:redirect_url"`
	FailureReason  *string `gorm:"type:text"`

	InitiatedAt time.Time  `gorm:"not null"`
	ExpiresAt   time.Time  `gorm:"not null"`
	CompletedAt *time.Time
	FailedAt    *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (PaymentModel) TableName() string { return "payments" }