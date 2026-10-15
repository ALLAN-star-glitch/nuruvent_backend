// internal/modules/payment/infrastructure/postgres/order_model.go

package postgres

import (
	"time"

	"gorm.io/gorm"
)

type OrderModel struct {
	ID             string  `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Kind           string  `gorm:"type:varchar(20);not null;default:'registration';index"`
	RegistrationID *string `gorm:"type:uuid;index"` // nullable now
	UserID         *string `gorm:"type:uuid;index"`
	GuestEmail     *string `gorm:"type:varchar(255)"`

	Currency      string `gorm:"type:varchar(3);not null;default:'KES'"`
	Subtotal      int64  `gorm:"not null;default:0"`
	DiscountTotal int64  `gorm:"not null;default:0"`
	TotalAmount   int64  `gorm:"not null;default:0"`

	Status      string     `gorm:"type:varchar(20);not null;default:'pending';index"`
	ExpiresAt   time.Time  `gorm:"not null"`
	PaidAt      *time.Time
	CancelledAt *time.Time

	BilledAccountID string     `gorm:"column:billed_account_id;type:uuid;not null;index"`
	PlatformFeeRate float64    `gorm:"column:platform_fee_rate;type:numeric(5,4);not null;default:0.0450"`
	SettledAt       *time.Time `gorm:"column:settled_at"`
	PayoutRef       *string    `gorm:"column:payout_ref;type:varchar(255)"`

	CertificateContext JSONB `gorm:"type:jsonb"` // ← NEW


		// Batch order context (organizer-pays certificates).
	BatchContext       JSONB `gorm:"type:jsonb"`
	BatchAttendeeCount int   `gorm:"not null;default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Items []OrderItemModel `gorm:"foreignKey:OrderID"`
}

func (OrderModel) TableName() string { return "orders" }